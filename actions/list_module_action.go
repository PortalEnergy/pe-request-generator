package actions

import (
	"github.com/gin-gonic/gin"
)

type ListModuleAction struct {
	ModuleAction
	BeforeAction func(c *gin.Context) error
	AfterAction  func(c *gin.Context)
	Label        string                                  `json:"label"`
	Fields       []string                                `json:"fields"`
	Size         int64                                   `json:"size,omitempty"`
	Maxsize      int64                                   `json:"maxsize"`
	Permission   []string                                `json:"permission"`
	Auth         bool                                    `json:"auth"`
	Join         []ModuleActionJoin                      `json:"join"`
	Where        func(c *gin.Context) *ModuleActionWhere `json:"where"`
	Extra        interface{}                             `json:"extra"`
	Search       []string                                `json:"search"`
	Filter       []string                                `json:"filter"`
}

func (action ListModuleAction) Action() ModuleActionName {
	return ModuleActionNameList
}

func (action ListModuleAction) BeforeRequest(c *gin.Context) error {
	if action.BeforeAction == nil {
		return nil
	}

	return action.BeforeAction(c)
}
func (action ListModuleAction) AfterRequest(c *gin.Context) {
	if action.AfterAction == nil {
		return
	}

	action.AfterAction(c)
}
