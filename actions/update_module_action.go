package actions

import (
	"github.com/gin-gonic/gin"
)

type UpdateModuleAction struct {
	ModuleAction
	BeforeAction func(c *gin.Context) error
	AfterAction  func(c *gin.Context)
	Label        string        `json:"label"`
	Fields       []string      `json:"fields"`
	Permission   []string      `json:"permission"`
	Auth         bool          `json:"auth"`
	By           []interface{} `json:"by"`
}

func (action UpdateModuleAction) Action() ModuleActionName {
	return ModuleActionNameUpdate
}

func (action UpdateModuleAction) BeforeRequest(c *gin.Context) error {
	if action.BeforeAction == nil {
		return nil
	}

	return action.BeforeAction(c)
}
func (action UpdateModuleAction) AfterRequest(c *gin.Context) {
	if action.AfterAction == nil {
		return
	}

	action.AfterAction(c)
}

func (action UpdateModuleAction) GetFields() []string {
	return action.Fields
}
