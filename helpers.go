package module

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/portalenergy/pe-request-generator/actions"
	"github.com/portalenergy/pe-request-generator/fields"
)

func (generator *Generator) getPagination(page int64, size int64) (int64, int64, int64) {
	var limit int64
	if size <= 0 {
		limit = 10
	} else {
		limit = size
	}
	if page <= 0 {
		page = 0
	}
	offset := page * limit

	return limit, offset, page
}

func (generator *Generator) normalizeFilters(data map[string]string, module *BaseModule, listAction actions.ListModuleAction) map[string]string {
	resultFilterMap := make(map[string]string)

	filters := make(map[string]fields.ModuleField)
	for _, realField := range module.Fields {
		if containsStrings(listAction.Filter, realField.Name) {
			filters[realField.Name] = realField
		}
	}

parentLoop:
	for _, filter := range filters {
		filterValue, ok := data[filter.Name]
		if !ok || len(filterValue) == 0 {
			continue
		}

		for _, rule := range filter.Check {
			if err := rule.Validate(filterValue); err != nil {
				continue parentLoop
			}
		}

		resultFilterMap[filter.Name] = filterValue
	}

	for key, value := range data {
		result := strings.Split(key, ".")
		if len(result) > 1 {
			resultFilterMap[key] = value
		}
	}

	return resultFilterMap
}

func (generator *Generator) checkRequest(
	context *gin.Context,
	data map[string]interface{},
	module *BaseModule,
	action actions.ModuleAction,
	scenario fields.Scenario,
) map[string]string {
	errs := make(map[string]string)
	actionFields := action.GetFields()
	fmt.Println("fields: ", actionFields)

	for _, fieldName := range actionFields {
		value := data[fieldName]
		field := module.GetField(fieldName)
		if field == nil {
			continue
		}

		rules := module.GetRules(context, *field, scenario)

		for _, rule := range rules {
			err := rule.Validate(value)
			if err != nil {
				errs[fieldName] = err.Error()
			}
		}

		if field.Convert != nil && value != nil {
			_, err := field.Convert(value)
			if err != nil {
				errs[fieldName] = err.Error()
			}
		}
	}

	return errs
}

func (generator *Generator) mapRequestInput(
	data map[string]interface{},
	module *BaseModule,
	actionFields []string,
) map[string]interface{} {
	output := make(map[string]interface{})

	for _, field := range module.Fields {
		value, ok := data[field.Name]
		if ok && containsStrings(actionFields, field.Name) {
			if field.Convert != nil {
				convertedValue, err := field.Convert(value)
				if err != nil {
					continue
				}
				output[field.Name] = convertedValue
			} else {
				output[field.Name] = value
			}
		}
	}

	return output
}

func queryParam(c *gin.Context, param string) (interface{}, error) {
	result := c.Request.URL.Query().Get(param)
	if len(result) == 0 {
		return nil, fmt.Errorf("param %s incorrect", param)
	}
	return result, nil
}

func int64QueryParam(c *gin.Context, param string, defaultValue int64) int64 {
	resultInterface, err := queryParam(c, param)
	if err != nil {
		return defaultValue
	}

	resultString, ok := resultInterface.(string)
	if !ok {
		return defaultValue
	}

	result, err := strconv.ParseInt(resultString, 0, 10)
	if err != nil {
		fmt.Println("PARSE INT ERR: ", err)
		return defaultValue
	}

	return result
}

func containsStrings(coll []string, item string) bool {
	for _, a := range coll {
		if a == item {
			return true
		}
	}
	return false
}
