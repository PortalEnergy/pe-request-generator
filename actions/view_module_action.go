package actions

import "github.com/gin-gonic/gin"

type ViewModuleAction struct {
	ModuleAction
	BeforeAction func(c *gin.Context) error
	AfterAction  func(c *gin.Context)
	Label        string `json:"label"`

	Fields     []string           `json:"fields"`
	Permission []string           `json:"permission"`
	Auth       bool               `json:"auth"`
	Join       []ModuleActionJoin `json:"join"`
	Where      ModuleActionWhere  `json:"where"`
	By         []interface{}      `json:"by"`
	Extra      interface{}        `json:"extra"`
}

func (action ViewModuleAction) Action() ModuleActionName {
	return ModuleActionNameView
}

func (action ViewModuleAction) BeforeRequest(c *gin.Context) error {
	if action.BeforeAction == nil {
		return nil
	}

	return action.BeforeAction(c)
}
func (action ViewModuleAction) AfterRequest(c *gin.Context) {
	if action.AfterAction == nil {
		return
	}

	action.AfterAction(c)
}
