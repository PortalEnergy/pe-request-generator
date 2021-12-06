package module

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/portalenergy/pe-request-generator/actions"
	"github.com/portalenergy/pe-request-generator/db"
	"github.com/portalenergy/pe-request-generator/fields"
	"github.com/portalenergy/pe-request-generator/icontext"
	"github.com/portalenergy/pe-request-generator/response"
	"github.com/portalenergy/pe-request-generator/utils"
)

const (
	GeneratorErrorAdd    string = "Cannot create record"
	GeneratorErrorUpdate string = "Cannot update record"
	GeneratorErrorDelete string = "Cannot delete record"
)

type Generator struct {
	db      db.DBExecutor
	group   gin.RouterGroup
	Modules []BaseModule
}

func NewGenerator(
	db db.DBExecutor,
	group gin.RouterGroup,
	modules []BaseModule,
) *Generator {
	return &Generator{
		db:      db,
		group:   group,
		Modules: modules,
	}
}

func (generator *Generator) Run() {
	for _, module := range generator.Modules {
		for _, action := range module.Actions {
			switch action.Action() {
			case actions.ModuleActionNameList:
				listAction, _ := action.(actions.ListModuleAction)

				route := generator.group.GET(module.FullPath(), generator.actionList(module, listAction))
				if listAction.Auth {
					if module.AuthMiddleware == nil {
						panic(fmt.Sprintf("auth middleware not implemented in module: %s", module.Name))
					}
					route.Use(module.AuthMiddleware(listAction))
				}
				if len(listAction.Permission) > 0 {
					if module.PermissionMiddleware == nil {
						panic(fmt.Sprintf("permission middleware not implemented in module: %s", module.Name))
					}
					route.Use(module.PermissionMiddleware(listAction.Permission))
				}
			case actions.ModuleActionNameAdd:
				addAction, _ := action.(actions.AddModuleAction)
				route := generator.group.PUT(module.FullPath(), generator.actionAdd(module, addAction))
				generator.group.GET(fmt.Sprintf("%s/defrec", module.FullPath()), generator.actionDefrec(module))
				if addAction.Auth {
					if module.AuthMiddleware == nil {
						panic(fmt.Sprintf("auth middleware not implemented in module: %s", module.Name))
					}
					route.Use(module.AuthMiddleware(addAction))
				}
				if len(addAction.Permission) > 0 {
					if module.PermissionMiddleware == nil {
						panic(fmt.Sprintf("permission middleware not implemented in module: %s", module.Name))
					}
					route.Use(module.PermissionMiddleware(addAction.Permission))
				}
			case actions.ModuleActionNameView:
				viewAction, _ := action.(actions.ViewModuleAction)
				route := generator.group.GET(fmt.Sprintf("%s/view/:bykey/:value", module.FullPath()), generator.actionView(module, viewAction))
				if viewAction.Auth {
					if module.AuthMiddleware == nil {
						panic(fmt.Sprintf("auth middleware not implemented in module: %s", module.Name))
					}
					route.Use(module.AuthMiddleware(viewAction))
				}
				if len(viewAction.Permission) > 0 {
					if module.PermissionMiddleware == nil {
						panic(fmt.Sprintf("permission middleware not implemented in module: %s", module.Name))
					}
					route.Use(module.PermissionMiddleware(viewAction.Permission))
				}
			case actions.ModuleActionNameUpdate:
				updateAction, _ := action.(actions.UpdateModuleAction)
				route := generator.group.POST(fmt.Sprintf("%s/:bykey/:value", module.FullPath()), generator.actionUpdate(module, updateAction))
				if updateAction.Auth {
					if module.AuthMiddleware == nil {
						panic(fmt.Sprintf("auth middleware not implemented in module: %s", module.Name))
					}
					route.Use(module.AuthMiddleware(updateAction))
				}
				if len(updateAction.Permission) > 0 {
					if module.PermissionMiddleware == nil {
						panic(fmt.Sprintf("permission middleware not implemented in module: %s", module.Name))
					}
					route.Use(module.PermissionMiddleware(updateAction.Permission))
				}
			case actions.ModuleActionNameDelete:
				deleteAction, _ := action.(actions.DeleteModuleAction)
				route := generator.group.DELETE(fmt.Sprintf("%s/delete/:bykey/:value", module.FullPath()), generator.actionDelete(module, deleteAction))
				if deleteAction.Auth {
					if module.AuthMiddleware == nil {
						panic(fmt.Sprintf("auth middleware not implemented in module: %s", module.Name))
					}
					route.Use(module.AuthMiddleware(deleteAction))
				}
				if len(deleteAction.Permission) > 0 {
					if module.PermissionMiddleware == nil {
						panic(fmt.Sprintf("permission middleware not implemented in module: %s", module.Name))
					}
					route.Use(module.PermissionMiddleware(deleteAction.Permission))
				}
			}
		}
	}
}

func (generator *Generator) actionList(module BaseModule, action actions.ListModuleAction) func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		l, _ := icontext.GetLogger(ctx)

		err := action.BeforeRequest(c)
		if err != nil {
			response.ErrorResponse(l, c, http.StatusBadRequest, err.Error(), nil)
			return
		}

		page := int64QueryParam(c, "page", 0)
		size := int64QueryParam(c, "size", 10)
		filters := generator.normalizeFilters(c.QueryMap("filter"), module, action)
		searchText := c.Query("search")
		addFilters := c.Query("addFilters")
		addHeads := c.Query("addHeads")

		realFields := make([]fields.ModuleField, 0, 10)
		for _, realField := range module.Fields {
			if containsStrings(action.Fields, realField.Name) {
				realFields = append(realFields, realField)
			}
		}

		results, count, err := generator.db.List(
			l,
			module.TableName,
			module.PrimaryKey,
			realFields,
			page,
			size,
			action.Search,
			searchText,
			filters,
			&action.Where,
			action.Join,
		)
		if err != nil {
			response.ErrorResponse(l, c, http.StatusBadRequest, err.Error(), nil)
			return
		}

		if len(results) == 0 {
			response.ErrorResponse(l, c, http.StatusNotFound, "Not found", nil)
			return
		}

		var filter map[string]fields.ModuleFilterField
		if addFilters == "true" {
			filter = make(map[string]fields.ModuleFilterField)
			for _, realField := range module.Fields {
				if containsStrings(action.Filter, realField.Name) {
					filterField := fields.ModuleFilterField{
						ScanObject: realField.ScanObject,
						Name:       realField.Name,
						Title:      realField.Title,
						Type:       realField.Type,
						FormType:   realField.FormType,
						Example:    realField.Example,
						Options:    realField.Options,
						Check:      realField.Check,
						Convert:    realField.Convert,
					}
					filter[realField.Name] = filterField
				}
			}
		}

		var heads map[string]string
		if addHeads == "true" {
			heads = make(map[string]string)

			for _, realField := range module.Fields {
				if containsStrings(action.Fields, realField.Name) {
					heads[realField.Name] = realField.Title
				}
			}
		}

		output := struct {
			Count   int64                               `json:"count"`
			Size    int64                               `json:"size"`
			Page    int64                               `json:"page"`
			Extra   interface{}                         `json:"extra"`
			Rows    []interface{}                       `json:"rows"`
			Heads   map[string]string                   `json:"heads"`
			Filters map[string]fields.ModuleFilterField `json:"filters,omitempty"`
		}{
			Count:   count,
			Size:    size,
			Page:    page,
			Extra:   action.Extra,
			Rows:    results,
			Heads:   heads,
			Filters: filter,
		}

		response.Response(l, c, output)

		action.AfterRequest(c)
	}
}

func (generator *Generator) actionAdd(module BaseModule, action actions.AddModuleAction) func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		l, _ := icontext.GetLogger(ctx)

		err := action.BeforeRequest(c)
		if err != nil {
			response.ErrorResponse(l, c, http.StatusBadRequest, GeneratorErrorAdd, []string{
				err.Error(),
			})
			return
		}

		var input map[string]interface{}
		err = utils.ParseJson(c.Request, &input)
		if err != nil {
			response.ErrorResponse(l, c, http.StatusBadRequest, GeneratorErrorAdd, []string{
				"Parse Input Error",
			})
			return
		}

		errs := generator.checkRequest(input, module, action, fields.ScenarioAdd)
		if len(errs) > 0 {
			response.ErrorResponse(l, c, http.StatusBadRequest, GeneratorErrorAdd, errs)
			return
		}

		realFields := make([]fields.ModuleField, 0, 10)
		for _, realField := range module.Fields {
			if containsStrings(action.Fields, realField.Name) {
				realFields = append(realFields, realField)
			}
		}

		mapInput := generator.mapRequestInput(input, module)
		fmt.Println(mapInput)
		output, err := generator.db.Add(l, module.TableName, module.PrimaryKey, realFields, mapInput)
		if err != nil {
			response.ErrorResponse(l, c, http.StatusBadRequest, GeneratorErrorAdd, []string{
				err.Error(),
			})
			return
		}

		response.Response(l, c, output)

		action.AfterRequest(c)
	}
}

func (generator *Generator) actionDefrec(module BaseModule) func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		l, _ := icontext.GetLogger(ctx)

		err := module.Defrec.BeforeRequest(c)
		if err != nil {
			response.ErrorResponse(l, c, http.StatusBadRequest, err.Error(), nil)
			return
		}

		response.Response(l, c, response.NewDefrecResponse(nil, module.Fields))

		module.Defrec.AfterRequest(c)
	}
}

func (generator *Generator) actionView(module BaseModule, action actions.ViewModuleAction) func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		l, _ := icontext.GetLogger(ctx)

		err := action.BeforeRequest(c)
		if err != nil {
			response.ErrorResponse(l, c, http.StatusBadRequest, err.Error(), nil)
			return
		}

		whereKey := c.Param("bykey")
		err = validation.In(action.By...).Error(fmt.Sprintf(`allowed keys %v`, action.By)).Validate(whereKey)
		if err != nil {
			response.ErrorResponse(l, c, http.StatusBadRequest, GeneratorErrorDelete, []string{
				err.Error(),
			})
			return
		}

		whereValue := c.Param("value")
		if len(whereValue) == 0 {
			response.ErrorResponse(l, c, http.StatusBadRequest, GeneratorErrorDelete, []string{
				"value param not found",
			})
			return
		}

		realFields := make([]fields.ModuleField, 0, 10)
		for _, realField := range module.Fields {
			if containsStrings(action.Fields, realField.Name) {
				realFields = append(realFields, realField)
			}
		}

		result, err := generator.db.View(l, module.TableName, module.PrimaryKey, realFields, []interface{}{whereKey}, []interface{}{whereValue}, &action.Where, action.Join)
		if err != nil {
			response.ErrorResponse(l, c, http.StatusBadRequest, err.Error(), nil)
			return
		}

		response.Response(l, c, result)

		action.AfterRequest(c)
	}
}

func (generator *Generator) actionUpdate(module BaseModule, action actions.UpdateModuleAction) func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		l, _ := icontext.GetLogger(ctx)

		err := action.BeforeRequest(c)
		if err != nil {
			response.ErrorResponse(l, c, http.StatusBadRequest, GeneratorErrorUpdate, nil)
			return
		}

		whereKey := c.Param("bykey")
		err = validation.In(action.By...).Error(fmt.Sprintf(`allowed keys %v`, action.By)).Validate(whereKey)
		if err != nil {
			response.ErrorResponse(l, c, http.StatusBadRequest, GeneratorErrorDelete, []string{
				err.Error(),
			})
			return
		}

		whereValue := c.Param("value")
		if len(whereValue) == 0 {
			response.ErrorResponse(l, c, http.StatusBadRequest, GeneratorErrorDelete, []string{
				"value param not found",
			})
			return
		}

		var input map[string]interface{}
		err = utils.ParseJson(c.Request, &input)
		if err != nil {
			response.ErrorResponse(l, c, http.StatusBadRequest, GeneratorErrorUpdate, nil)
			return
		}

		errs := generator.checkRequest(input, module, action, fields.ScenarioUpdate)
		if len(errs) > 0 {
			response.ErrorResponse(l, c, http.StatusBadRequest, GeneratorErrorUpdate, errs)
			return
		}

		realFields := make([]fields.ModuleField, 0, 10)
		for _, realField := range module.Fields {
			if containsStrings(action.Fields, realField.Name) {
				realFields = append(realFields, realField)
			}
		}

		mapInput := generator.mapRequestInput(input, module)
		output, err := generator.db.Update(l, module.TableName, module.PrimaryKey, realFields, mapInput, whereKey, whereValue)
		if err != nil {
			response.ErrorResponse(l, c, http.StatusBadRequest, GeneratorErrorUpdate, nil)
			return
		}

		response.Response(l, c, output)

		action.AfterRequest(c)
	}
}

func (generator *Generator) actionDelete(module BaseModule, action actions.DeleteModuleAction) func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		l, _ := icontext.GetLogger(ctx)

		err := action.BeforeRequest(c)
		if err != nil {
			response.ErrorResponse(l, c, http.StatusBadRequest, GeneratorErrorDelete, nil)
			return
		}

		whereKey := c.Param("bykey")
		err = validation.In(action.By...).Error(fmt.Sprintf(`allowed keys %v`, action.By)).Validate(whereKey)
		if err != nil {
			response.ErrorResponse(l, c, http.StatusBadRequest, GeneratorErrorDelete, []string{
				err.Error(),
			})
			return
		}

		whereValue := c.Param("value")
		if len(whereValue) == 0 {
			response.ErrorResponse(l, c, http.StatusBadRequest, GeneratorErrorDelete, nil)
			return
		}

		err = generator.db.Delete(l, module.TableName, whereKey, whereValue)
		if err != nil {
			response.ErrorResponse(l, c, http.StatusBadRequest, GeneratorErrorDelete, []string{
				err.Error(),
			})
			return
		}

		output := struct {
			Delete bool `json:"delete"`
		}{
			Delete: true,
		}
		response.Response(l, c, output)

		action.AfterRequest(c)
	}
}
