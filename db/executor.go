package db

import (
	"database/sql"

	"github.com/portalenergy/pe-request-generator/actions"
	"github.com/portalenergy/pe-request-generator/fields"
	log "github.com/sirupsen/logrus"
)

type DBExecutor interface {
	List(
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
	) (result []interface{}, rowsCount int64, err error)
	View(
		log *log.Entry,
		tableName string,
		primaryKey string,
		fields []fields.ModuleField,
		keys []interface{},
		values []interface{},
		where *actions.ModuleActionWhere,
		joins []actions.ModuleActionJoin,
	) (interface{}, error)
	Add(log *log.Entry, tableName string, primaryKey string, fields []fields.ModuleField, input map[string]interface{}) (interface{}, error)
	Update(log *log.Entry, tableName string, primaryKey string, fields []fields.ModuleField, input map[string]interface{}, key interface{}, value interface{}) (interface{}, error)
	Delete(log *log.Entry, tableName string, key interface{}, value interface{}) error
	RawRequest(log *log.Entry, query string, params ...interface{}) (*sql.Rows, error)
}
