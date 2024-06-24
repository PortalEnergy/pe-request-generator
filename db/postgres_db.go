package db

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	_ "github.com/lib/pq"
	"github.com/portalenergy/pe-request-generator/actions"
	"github.com/portalenergy/pe-request-generator/fields"
	log "github.com/sirupsen/logrus"
)

type DB struct {
	DBExecutor
	sql *sql.DB
}
type Tx struct {
	sql *sql.Tx
}
/* type AddOutptut struct {
	Id int64 `json:"id"`
} */



func NewDB(sql *sql.DB) *DB {
	return &DB{
		sql: sql,
	}
}

func (db *DB) Begin() (*Tx, error) {
	tx, err := db.sql.Begin()
	if err != nil {
		return nil, err
	}
	return &Tx{tx}, nil
}

func (db *DB) RowExists(query string, args ...interface{}) bool {
	var exists bool
	query = fmt.Sprintf("SELECT exists (%s)", query)
	_ = db.sql.QueryRow(query, args...).Scan(&exists)

	return exists
}

func (db *DB) List(
	log *log.Entry,
	tableName string,
	primaryKey string,
	fields []fields.ModuleField,
	page int64,
	size int64,
	searchFields []string,
	searchText string,
	filter map[string]string,
	where *actions.ModuleActionWhere,
	joins []actions.ModuleActionJoin,
) (result []interface{}, rowsCount int64, err error) {
	fieldsString := make([]string, 0, 10)
	fieldsFunction := make(map[string]string)
	for _, field := range fields {
		fieldsString = append(fieldsString, field.Name)
		if field.SelectFunction != nil {
			fieldsFunction[field.Name] = *field.SelectFunction
		}
	}

	pq := PostgresQuery{
		TableName:      tableName,
		PrimaryKey:     primaryKey,
		Fields:         fieldsString,
		FieldsFunction: fieldsFunction,
		SearchFields:   searchFields,
		SearchText:     searchText,
		Filter:         filter,
		Joins:          joins,
		Where:          where,
		Page:           page,
		Size:           size,
	}
	query, values := pq.GetQuery(false)
	countQuery, _ := pq.GetQuery(true)

	log.Infoln("LIST QUERY: ", query)
	log.Infoln("LIST COUNT QUERY: ", countQuery)
	fmt.Println("LIST QUERY: ", query, values)
	fmt.Println("LIST COUNT QUERY: ", countQuery)

	var rows *sql.Rows
	var countResult *sql.Rows

	if len(values) > 0 {
		rows, err = db.sql.Query(query, values...)
		countResult, err = db.sql.Query(countQuery, values...)
	} else {
		rows, err = db.sql.Query(query)
		countResult, err = db.sql.Query(countQuery)
	}

	if err != nil {
		fmt.Println("LIST ERR: ", err)
		log.Errorln("LIST ERR: ", err)
		return nil, 0, err
	}
	defer rows.Close()

	results := make([]interface{}, 0, 10)
	for rows.Next() {
		columnValues := make([]interface{}, 0, 10)
		var primaryValue interface{}
		columnValues = append(columnValues, &primaryValue)

		for i := 0; i < len(fields); i++ {
			value := fields[i].ScanObject
			columnValues = append(columnValues, value)
		}
		for _, join := range joins {
			if len(join.Fields) == 0 {
				continue
			}
			var columnValue json.RawMessage
			columnValues = append(columnValues, &columnValue)
		}

		err = rows.Scan(columnValues...)
		if err != nil {
			fmt.Println("SCAN ERR: ", err)
			continue
		}

		currentResult := make(map[string]interface{})
		offset := 1
		for index, field := range fields {
			value, ok := columnValues[index+offset].(driver.Valuer)

			if ok {
				if field.ResultValueConverter != nil {
					currentResult[field.Name] = field.ResultValueConverter(value)
				} else {
					currentResult[field.Name], _ = value.Value()
				}
			} else {
				if field.ResultValueConverter != nil {
					currentResult[field.Name] = field.ResultValueConverter(value)
				} else {
					currentResult[field.Name] = value
				}
			}
		}

		if len(fields) > 0 {
			offset = offset + len(fields)
		}

		for index, join := range joins {
			joinValue := columnValues[index+offset]
			converted, ok := joinValue.(*json.RawMessage)
			if !ok {
				continue
			}

			var joinValues [][]interface{}
			err := json.Unmarshal(*converted, &joinValues)
			if err != nil {
				log.Errorln("VIEW JOIN ERR: ", err)
				continue
			}

			checkString := ""
			for _, val := range joinValues {
				if val == nil {
					continue
				}

				for _, v := range val {
					if v == nil {
						continue
					}
					checkString = fmt.Sprintf("%v%v", checkString, v)
				}
			}

			joinResults := make([]map[string]interface{}, 0, 10)

			if len(checkString) > 0 {
				for _, joinValue := range joinValues {
					resultMap := make(map[string]interface{})
					for index, field := range join.Fields {
						resultMap[field] = joinValue[index]
					}
					joinResults = append(joinResults, resultMap)
				}
				log.Infoln("VIEW JOIN RESULTS: ", joinResults)
			}

			joinStringsArray := make([]string, 0, 10)
			for _, res := range joinResults {
				jsonRes, err := json.Marshal(res)
				if err != nil {
					continue
				}

				joinStringsArray = append(joinStringsArray, string(jsonRes))
			}
			resultUnique := removeDuplicate(joinStringsArray)

			joinResultUnique := make([]map[string]interface{}, 0, 10)
			for _, res := range resultUnique {
				var mapResult map[string]interface{}
				err := json.Unmarshal([]byte(res), &mapResult)
				if err != nil {
					continue
				}

				joinResultUnique = append(joinResultUnique, mapResult)
			}

			currentResult[join.ResultArrayName] = joinResultUnique
			//offset += 1
		}

		results = append(results, currentResult)
	}

	result = append(result, results...)

	var count int64
	if len(joins) > 0 {
		for countResult.Next() {
			count++
		}
	} else {
		for countResult.Next() {
			var currentCount int64
			err = countResult.Scan(&currentCount)
			if err == nil {
				count += currentCount
			}
		}
	}

	fmt.Println("COUNT OF RESULT: ", count)

	return result, count, nil
}

func (db *DB) View(
	log *log.Entry,
	tableName string,
	primaryKey string,
	fields []fields.ModuleField,
	keys []interface{},
	values []interface{},
	where *actions.ModuleActionWhere,
	joins []actions.ModuleActionJoin,
) (interface{}, error) {
	fieldsString := make([]string, 0, 10)
	fieldsFunction := make(map[string]string)
	for _, field := range fields {
		fieldsString = append(fieldsString, field.Name)
		if field.SelectFunction != nil {
			fieldsFunction[field.Name] = *field.SelectFunction
		}
	}

	//fmt.Println("SSSSSSSSSSSSS")
	//fmt.Printf("\n\n\nWhere 1 TEST: %+v\n\n\n", where)
	//fmt.Printf("\n\n\nWhere keys TEST: %+v\n\n\n", keys)

	for index, key := range keys {
		if where == nil || len(where.Fields) == 0 {
			where = &actions.ModuleActionWhere{
				Fields: make([]actions.ModuleActionWhereField, 0, 10),
				Values: make([]interface{}, 0, 10),
			}
			where.Fields = append(where.Fields, actions.ModuleActionWhereField{
				Name:          key.(string),
				ConditionType: actions.ModuleActionWhereConditionTypeAnd,
			})
			where.Values = append(where.Values, values[index])
		}

	}

	//fmt.Printf("\n\n\nWhere 2 TEST: %+v\n\n\n", where)

	pq := PostgresQuery{
		TableName:      tableName,
		PrimaryKey:     primaryKey,
		Fields:         fieldsString,
		FieldsFunction: fieldsFunction,
		SearchFields:   nil,
		SearchText:     "",
		Filter:         nil,
		Joins:          joins,
		Where:          where,
		Page:           0,
		Size:           1,
	}
	where = nil
	query, values := pq.GetQuery(false)
	log.Infoln("VIEW QUERY: ", query)
	fmt.Println("VIEW QUERY: ", query)

	var rows *sql.Rows
	var err error
	if len(values) > 0 {
		rows, err = db.sql.Query(query, values...)
	} else {
		rows, err = db.sql.Query(query)
	}

	if err != nil {
		log.Errorln("VIEW ERR: ", err)
		return nil, err
	}
	defer rows.Close()

	results := make([]interface{}, 0, 10)
	for rows.Next() {
		columnValues := make([]interface{}, 0, 10)
		var primaryValue interface{}
		columnValues = append(columnValues, &primaryValue)

		for i := 0; i < len(fields); i++ {
			value := fields[i].ScanObject
			columnValues = append(columnValues, value)
		}
		for _, join := range joins {
			if len(join.Fields) == 0 {
				continue
			}
			var columnValue json.RawMessage
			columnValues = append(columnValues, &columnValue)
		}

		err = rows.Scan(columnValues...)
		if err != nil {
			fmt.Println("ERROR: ", err)
			continue
		}

		currentResult := make(map[string]interface{})
		offset := 1
		for index, field := range fields {
			value, ok := columnValues[index+offset].(driver.Valuer)
			if ok {
				currentResult[field.Name], _ = value.Value()
			} else {
				currentResult[field.Name] = value
			}
		}

		if len(fields) > 0 {
			offset = offset + len(fields)
		}

		for index, join := range joins {
			joinValue := columnValues[index+offset]
			converted, ok := joinValue.(*json.RawMessage)
			if !ok {
				continue
			}

			var joinValues [][]interface{}
			err := json.Unmarshal(*converted, &joinValues)
			if err != nil {
				log.Errorln("VIEW JOIN ERR: ", err)
				continue
			}

			log.Infoln("VIEW JOIN VALUES: ", joinValues, join)

			checkString := ""
			for _, val := range joinValues {
				if val == nil {
					continue
				}

				for _, v := range val {
					if v == nil {
						continue
					}
					checkString = fmt.Sprintf("%v%v", checkString, v)
				}
			}

			joinResults := make([]map[string]interface{}, 0, 10)

			if len(checkString) > 0 {
				for _, joinValue := range joinValues {
					resultMap := make(map[string]interface{})
					for index, field := range join.Fields {
						resultMap[field] = joinValue[index]
					}
					joinResults = append(joinResults, resultMap)
				}
				log.Infoln("VIEW JOIN RESULTS: ", joinResults)
			}

			currentResult[join.ResultArrayName] = joinResults
			offset += 1
		}

		results = append(results, currentResult)
	}

	fmt.Println("RESULTS:  ", results)

	if len(results) > 0 {
		return results[0], nil
	}

	return nil, errors.New("Record not found")
}

func (db *DB) Add(log *log.Entry, tableName string, primaryKey string, fields []fields.ModuleField, input map[string]interface{}) (interface{}, error) {
	query := fmt.Sprintf(`INSERT INTO public."%s"`, tableName)
	output := struct {
		Value int64 `json:"value"`
		PrimaryKey string `json:"primary_key"`
	}{}

	keys := make([]string, 0, 10)
	values := make([]interface{}, 0, 10)
	fieldsString := make([]string, 0, 10)

	sortedInput := make([]string, 0, len(input))
	for k := range input {
		sortedInput = append(sortedInput, k)
	}
	sort.Strings(sortedInput)

	for _, key := range sortedInput {
		value, _ := input[key]
		fieldsString = append(fieldsString, key)
		keys = append(keys, fmt.Sprintf(`"%s"`, key))
		values = append(values, value)
	}
	keys = append(keys, `"created_ts"`, `"updated_ts"`)
	values = append(values, time.Now().Unix(), time.Now().Unix())

	names := strings.Join(keys, ",")
	valueNumbers := make([]string, 0, 10)

	for index, _ := range values {
		valueNumbers = append(valueNumbers, fmt.Sprintf(`$%d`, index+1))
	}

	valueNumberString := strings.Join(valueNumbers, ",")

	query = fmt.Sprintf(`%s (%s) VALUES (%s) RETURNING "%s"`, query, names, valueNumberString, primaryKey)
	log.Infoln("ADD QUERY: ", query)

	fmt.Println(query)
	fmt.Println(values)


	err := db.sql.QueryRow(query, values...).Scan(&output.Value)
	if err != nil {
		fmt.Println("ERR: ", err)
		log.Errorln("ADD ERR: ", err)
		return nil, err
	}

	output.PrimaryKey = primaryKey;

	//fmt.Println("PK: ", primaryKey, output.Value)



	return output, nil;

	//return db.View(log, tableName, primaryKey, fields, []interface{}{primaryKey}, []interface{}{value}, nil, nil)
}

func (db *DB) Update(log *log.Entry, tableName string, primaryKey string, fields []fields.ModuleField, input map[string]interface{}, key interface{}, value interface{}) (interface{}, error) {
	query := fmt.Sprintf(`UPDATE public."%s" SET`, tableName)
	values := make([]interface{}, 0, 10)
	index := 1
	fieldsString := make([]string, 0, 10)
	for key, value := range input {
		fieldsString = append(fieldsString, key)
		query = fmt.Sprintf(`%s "%s" = $%d, `, query, key, index)
		values = append(values, value)
		index++
	}
	values = append(values, value)

	query = strings.TrimSpace(query)
	query = strings.TrimSuffix(query, ",")

	query = fmt.Sprintf(`%s WHERE "%s"=$%d`, query, key, index)

	log.Infoln(`UPDATE QUERY: `, query)
	log.Infoln(`UPDATE VALUES: `, values)

	result, err := db.sql.Exec(query, values...)
	if err != nil {
		return nil, err
	}

	updatedCount, err := result.RowsAffected()
	if err != nil {
		return nil, err
	}

	if updatedCount == 0 {
		return nil, errors.New("record not found")
	}

	return db.View(log, tableName, primaryKey, fields, []interface{}{key}, []interface{}{value}, nil, nil)
}

func (db *DB) Delete(log *log.Entry, tableName string, key interface{}, value interface{}) error {
	query := fmt.Sprintf(`DELETE FROM "%s" WHERE "%s"=$1`, tableName, key)
	log.Infoln("DELETE QUERY: ", query)
	result, err := db.sql.Exec(query, value)
	if err != nil {
		return err
	}

	countOfDeleted, err := result.RowsAffected()
	if err != nil {
		return err
	}

	log.Infoln("DELETE COUNT OF DELETED: ", countOfDeleted)
	if countOfDeleted == 0 {
		return errors.New("record not found")
	}

	return nil
}

func (db *DB) RawRequest(log *log.Entry, query string, params ...interface{}) (*sql.Rows, error) {
	return db.sql.Query(query, params...)
}

func removeDuplicate(sliceList []string) []string {
	allKeys := make(map[string]bool)
	var list []string
	for _, item := range sliceList {
		if _, value := allKeys[item]; !value {
			allKeys[item] = true
			list = append(list, item)
		}
	}
	return list
}
