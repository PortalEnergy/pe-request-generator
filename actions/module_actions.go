package actions

import (
	"github.com/gin-gonic/gin"
)

type ModuleActionName string

const (
	ModuleActionNameList   ModuleActionName = "list"
	ModuleActionNameAdd    ModuleActionName = "add"
	ModuleActionNameDefrec ModuleActionName = "defrec"
	ModuleActionNameView   ModuleActionName = "view"
	ModuleActionNameUpdate ModuleActionName = "update"
	ModuleActionNameDelete ModuleActionName = "delete"
)

type ModuleAction interface {
	GetModuleName() string
	Action() ModuleActionName
	BeforeRequest(c *gin.Context) error
	AfterRequest(c *gin.Context)
	GetFields() []string
}

func NewWhere(
	fields []ModuleActionWhereField,
	values []interface{},
) *ModuleActionWhere {
	return &ModuleActionWhere{
		Fields: fields,
		Values: values,
	}
}

func NewJoin(tableName string, joinType JoinType, onParentKey string, onKey string, fields []string, resultArrayName string) ModuleActionJoin {
	return ModuleActionJoin{
		TableName:       tableName,
		Type:            joinType,
		OnParentKey:     onParentKey,
		OnKey:           onKey,
		Fields:          fields,
		ResultArrayName: resultArrayName,
	}
}

type ModuleActionWhereConditionType string

const (
	ModuleActionWhereConditionTypeAnd ModuleActionWhereConditionType = "AND"
	ModuleActionWhereConditionTypeOR  ModuleActionWhereConditionType = "OR"
)

type ModuleActionWhere struct {
	Fields []ModuleActionWhereField `json:"fields"`
	Values []interface{}            `json:"values"`
}

type ModuleActionWhereField struct {
	Name          string
	ConditionType ModuleActionWhereConditionType
}

type JoinType string

const (
	JoinTypeLeft       JoinType = "LEFT"
	JoinTypeLeftOuter  JoinType = "LEFT OUTER"
	JoinTypeRight      JoinType = "RIGHT"
	JoinTypeRightOuter JoinType = "RIGHT OUTER"
	JoinTypeInner      JoinType = "INNER"
)

type ModuleActionJoin struct {
	TableName       string   `json:"table_name"`
	Type            JoinType `json:"type"`
	OnParentKey     string   `json:"on"`
	OnKey           string   `json:"on_key"`
	Fields          []string `json:"fields"`
	ResultArrayName string   `json:"result_array_name"`
}
