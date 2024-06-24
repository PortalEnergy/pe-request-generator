package module

import (
	"github.com/gin-gonic/gin"
	"github.com/portalenergy/pe-request-generator/actions"
	"github.com/portalenergy/pe-request-generator/fields"
)

type BaseModule struct {
	Name       string                     `json:"name"`
	Label      string                     `json:"label"`
	TableName  string                     `json:"table_name"`
	PrimaryKey string                     `json:"primary_key"`
	Path       string                     `json:"path"`
	Fields     []fields.ModuleField       `json:"fields"`
	Defrec     actions.DefrecModuleAction `json:"defrec"`
	Actions    []actions.ModuleAction     `json:"actions"`
}

func (module BaseModule) GetField(fieldName string) *fields.ModuleField {
	for _, field := range module.Fields {
		if field.Name == fieldName {
			return &field
		}
	}
	return nil
}

func (module BaseModule) GetRules(context *gin.Context, field fields.ModuleField, scenario fields.Scenario) []fields.CheckRules {
	checkRules := make([]fields.CheckRules, 0, 10)
	if field.Check != nil {
		for _, rule := range field.Check {
			for _, checkScenario := range rule.GetScenarios() {
				if checkScenario == scenario {
					checkRules = append(checkRules, rule)
				}
			}
		}
	}
	if field.CheckFunc != nil {
		for _, rule := range field.CheckFunc(context) {
			for _, checkScenario := range rule.GetScenarios() {
				if checkScenario == scenario {
					checkRules = append(checkRules, rule)
				}
			}
		}
	}
	return checkRules
}
