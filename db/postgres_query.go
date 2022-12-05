package db

import (
	"fmt"
	"strings"

	"github.com/portalenergy/pe-request-generator/actions"
)

type PostgresQuery struct {
	TableName      string
	PrimaryKey     string
	Fields         []string
	FieldsFunction map[string]string
	SearchFields   []string
	SearchText     string
	Filter         map[string]string
	Joins          []actions.ModuleActionJoin
	Where          *actions.ModuleActionWhere
	Page           int64
	Size           int64
}

func (pq *PostgresQuery) GetQuery(isCount bool) (string, []interface{}) {
	values := make([]interface{}, 0, 10)

	fields := make([]string, 0, 10)
	fields = append(fields, fmt.Sprintf(`parent."%s"`, pq.PrimaryKey))

	fmt.Println("FIELDS: ", pq.Fields)
	fmt.Println("FIELD FUN: ", pq.FieldsFunction)
	for _, field := range pq.Fields {
		selectFunction, ok := pq.FieldsFunction[field]

		if !ok {
			fields = append(fields, fmt.Sprintf(`parent."%s"`, field))
		} else {
			fields = append(fields, fmt.Sprintf(`%s(parent."%s")`, selectFunction, field))
		}
	}
	fmt.Println("FIELDS: ", fields)

	for _, join := range pq.Joins {
		if len(join.Fields) == 0 {
			continue
		}

		joinQueries := make([]string, 0, 10)
		for _, field := range join.Fields {
			joinQueries = append(joinQueries, fmt.Sprintf(`%s."%s"`, join.ResultArrayName, field))
		}

		joinQueryField := fmt.Sprintf(`json_agg(json_build_array(%s))`, strings.Join(joinQueries, ", "))
		fields = append(fields, joinQueryField)
	}

	queryFields := strings.Join(fields, ", ")
	if isCount {
		queryFields = `COUNT(parent.*)`
	}

	query := fmt.Sprintf(`SELECT %s FROM public."%s" AS parent`, queryFields, pq.TableName)
	for _, join := range pq.Joins {
		if len(join.TableName) > 0 {
			query = fmt.Sprintf(
				`%s %s JOIN public."%s" AS %s ON parent."%s"=%s."%s"`,
				query,
				join.Type,
				join.TableName,
				join.ResultArrayName,
				join.OnParentKey,
				join.ResultArrayName,
				join.OnKey,
			)
		}
	}

	conditionIndex := 0
	if pq.Where != nil {
		if len(pq.Where.Fields) > 0 && len(pq.Where.Values) > 0 && len(pq.Where.Fields) == len(pq.Where.Values) {
			values = pq.Where.Values
			lastIndex := len(pq.Where.Fields) - 1

			whereQueries := make([]string, 0, 10)
			for index, whereKey := range pq.Where.Fields {
				conditionIndex += 1
				if lastIndex == index {
					whereQueries = append(whereQueries, fmt.Sprintf(`parent.%s=$%d`, whereKey.Name, conditionIndex))
				} else {
					whereQueries = append(whereQueries, fmt.Sprintf(`parent.%s=$%d %s`, whereKey.Name, conditionIndex, whereKey.ConditionType))
				}
			}

			query = fmt.Sprintf(`%s WHERE (%s)`, query, strings.Join(whereQueries, " "))
		}
	}

	if len(pq.SearchText) > 0 && len(pq.SearchFields) > 0 {
		values = append(values, strings.ToLower(pq.SearchText))
		searchQueries := make([]string, 0, 10)
		conditionIndex += 1

		for _, field := range pq.SearchFields {
			searchQueries = append(searchQueries, fmt.Sprintf(`LOWER(parent."%s") LIKE '%%' || $%d || '%%'`, field, conditionIndex))
		}

		if pq.Where != nil && len(pq.Where.Fields) > 0 && len(pq.Where.Values) > 0 && len(pq.Where.Fields) == len(pq.Where.Values) {
			query = fmt.Sprintf(`%s AND (%s)`, query, strings.Join(searchQueries, " OR "))
		} else {
			query = fmt.Sprintf("%s WHERE (%s)", query, strings.Join(searchQueries, " OR "))
		}
	}

	if len(pq.Filter) > 0 {
		filterQueries := make([]string, 0, 10)
		for key, value := range pq.Filter {
			conditionIndex += 1
			values = append(values, value)
			result := strings.Split(key, ".")
			if len(result) > 1 {
				placeholder := result[0]
				placeholderKey := result[1]
				filterQueries = append(filterQueries, fmt.Sprintf(`%s."%s"=$%d`, placeholder, placeholderKey, conditionIndex))
			} else {
				filterQueries = append(filterQueries, fmt.Sprintf(`parent."%s"=$%d`, key, conditionIndex))
			}
		}

		if (pq.Where != nil && len(pq.Where.Fields) > 0 && len(pq.Where.Values) > 0 && len(pq.Where.Fields) == len(pq.Where.Values)) || (len(pq.SearchText) > 0 && len(pq.SearchFields) > 0) {
			query = fmt.Sprintf(`%s AND (%s)`, query, strings.Join(filterQueries, " AND "))
		} else {
			query = fmt.Sprintf(`%s WHERE (%s)`, query, strings.Join(filterQueries, " AND "))
		}
	}

	if isCount {
		return fmt.Sprintf(`%s GROUP BY parent."%s"`, query, pq.PrimaryKey), values
	}

	return fmt.Sprintf(`%s GROUP BY parent."%s" LIMIT %d OFFSET %d`, query, pq.PrimaryKey, pq.Size, pq.Size*pq.Page), values
}
