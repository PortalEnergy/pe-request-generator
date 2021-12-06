package module

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/portalenergy/pe-request-generator/actions"
	"github.com/portalenergy/pe-request-generator/fields"
)

type BaseModule struct {
	Name                 string                                            `json:"name"`
	TableName            string                                            `json:"table_name"`
	PrimaryKey           string                                            `json:"primary_key"`
	Path                 string                                            `json:"path"`
	AuthMiddleware       func(module actions.ModuleAction) gin.HandlerFunc `json:"auth_middleware"`
	PermissionMiddleware func(permissions []string) gin.HandlerFunc        `json:"permission_middleware"`
	Fields               []fields.ModuleField                              `json:"fields"`
	Defrec               actions.DefrecModuleAction                        `json:"defrec"`
	Actions              []actions.ModuleAction                            `json:"actions"`
}

func (module *BaseModule) FullPath() string {
	return fmt.Sprintf("%s/%s", module.Path, module.Name)
}

func (module BaseModule) GetField(fieldName string) *fields.ModuleField {
	for _, field := range module.Fields {
		if field.Name == fieldName {
			return &field
		}
	}
	return nil
}

func (module BaseModule) GetRules(field fields.ModuleField, scenario fields.Scenario) []fields.CheckRules {
	checkRules := make([]fields.CheckRules, 0, 10)
	for _, rule := range field.Check {
		for _, checkScenario := range rule.GetScenarios() {
			if checkScenario == scenario {
				checkRules = append(checkRules, rule)
			}
		}
	}
	return checkRules
}
