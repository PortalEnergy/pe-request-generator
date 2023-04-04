package actions

import "github.com/gin-gonic/gin"

type DeleteModuleAction struct {
	ModuleAction
	BeforeAction func(c *gin.Context) error
	AfterAction  func(c *gin.Context)
	Label        string        `json:"label"`
	Permission   []string      `json:"permission"`
	Auth         bool          `json:"auth"`
	By           []interface{} `json:"by"`
}

func (action DeleteModuleAction) Action() ModuleActionName {
	return ModuleActionNameDelete
}

func (action DeleteModuleAction) BeforeRequest(c *gin.Context) error {
	if action.BeforeAction == nil {
		return nil
	}

	return action.BeforeAction(c)
}
func (action DeleteModuleAction) AfterRequest(c *gin.Context) {
	if action.AfterAction == nil {
		return
	}

	action.AfterAction(c)
}
