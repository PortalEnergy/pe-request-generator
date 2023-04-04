package actions

import (
	"github.com/gin-gonic/gin"
)

type AddModuleAction struct {
	ModuleAction
	BeforeAction func(c *gin.Context) error
	AfterAction  func(c *gin.Context)
	Label        string   `json:"label"`
	Fields       []string `json:"fields"`
	Permission   []string `json:"permission"`
	Auth         bool     `json:"auth"`
}

func (action AddModuleAction) Action() ModuleActionName {
	return ModuleActionNameAdd
}

func (action AddModuleAction) BeforeRequest(c *gin.Context) error {
	if action.BeforeAction == nil {
		return nil
	}

	return action.BeforeAction(c)
}
func (action AddModuleAction) AfterRequest(c *gin.Context) {
	if action.AfterAction == nil {
		return
	}

	action.AfterAction(c)
}

func (action AddModuleAction) GetFields() []string {
	return action.Fields
}
