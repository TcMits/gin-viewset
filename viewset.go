package viewset

import (
	"net/http"

	"github.com/TcMits/viewset/manager"
	"github.com/gin-gonic/gin"
)

const (
	DEFAULT_LIST_ACTION     = "list"
	DEFAULT_RETRIEVE_ACTION = "retrieve"
	DEFAULT_CREATE_ACTION   = "create"
	DEFAULT_UPDATE_ACTION   = "update"
	DEFAULT_DELETE_ACTION   = "delete"
)

type HandlerWithViewSetFunc[EntityType any] func(
	string,
	*ViewSet[EntityType],
	*gin.Context,
)

type Route[EntityType any] struct {
	Action  string
	SubPath string
	Method  string
	Handler HandlerWithViewSetFunc[EntityType]
}

type ViewSet[EntityType any] struct {
	BasePath string
	Actions  []Route[EntityType]

	ExceptionHandler
	PermissionChecker
	Manager       manager.Manager[EntityType, any]
	Serializer    Serializer[EntityType]
	FormValidator FormValidator[EntityType, any]
}

func NewViewSet[EntityType any, ValidatedType any, URIType any](
	basePath string,
	detailParams string,
	excludeDefaultActions []string,
	additionalActions []Route[EntityType],

	manager manager.Manager[EntityType, URIType],
	exceptionHandler ExceptionHandler,
	permissionChecker PermissionChecker,
	serializer Serializer[EntityType],
	formValidator FormValidator[EntityType, ValidatedType],
) *ViewSet[EntityType] {
	if manager == nil {
		panic("manager is required")
	}
	if exceptionHandler == nil {
		exceptionHandler = &DefaultExceptionHandler{}
	}
	if permissionChecker == nil {
		permissionChecker = &AllowAny{}
	}
	if serializer == nil {
		serializer = &DefaultSerializer[EntityType]{}
	}
	if formValidator == nil {
		formValidator = &DefaultValidator[EntityType, ValidatedType]{}
	}

	viewSet := &ViewSet[EntityType]{
		BasePath:          basePath,
		ExceptionHandler:  exceptionHandler,
		PermissionChecker: permissionChecker,
		Manager:           manager,
		Serializer:        serializer,
		FormValidator:     formValidator,
	}

	if shouldAddAction(DEFAULT_LIST_ACTION, excludeDefaultActions) {
		viewSet.Actions = append(viewSet.Actions, Route[EntityType]{
			Action:  DEFAULT_LIST_ACTION,
			SubPath: "/",
			Method:  http.MethodGet,
			Handler: List[EntityType],
		})
	}
	if shouldAddAction(DEFAULT_RETRIEVE_ACTION, excludeDefaultActions) {
		viewSet.Actions = append(viewSet.Actions, Route[EntityType]{
			Action:  DEFAULT_RETRIEVE_ACTION,
			SubPath: detailParams,
			Method:  http.MethodGet,
			Handler: Retrieve[EntityType],
		})
	}
	if shouldAddAction(DEFAULT_CREATE_ACTION, excludeDefaultActions) {
		viewSet.Actions = append(viewSet.Actions, Route[EntityType]{
			Action:  DEFAULT_CREATE_ACTION,
			SubPath: "/",
			Method:  http.MethodPost,
			Handler: Create[EntityType],
		})
	}
	if shouldAddAction(DEFAULT_UPDATE_ACTION, excludeDefaultActions) {
		viewSet.Actions = append(
			viewSet.Actions,
			Route[EntityType]{
				Action:  DEFAULT_UPDATE_ACTION,
				SubPath: detailParams,
				Method:  http.MethodPut,
				Handler: Update[EntityType],
			}, Route[EntityType]{
				Action:  DEFAULT_UPDATE_ACTION,
				SubPath: detailParams,
				Method:  http.MethodPatch,
				Handler: Update[EntityType],
			},
		)
	}
	if shouldAddAction(DEFAULT_DELETE_ACTION, excludeDefaultActions) {
		viewSet.Actions = append(viewSet.Actions, Route[EntityType]{
			Action:  DEFAULT_DELETE_ACTION,
			SubPath: detailParams,
			Method:  http.MethodDelete,
			Handler: Delete[EntityType],
		})
	}
	viewSet.Actions = append(viewSet.Actions, additionalActions...)
	return viewSet
}

func (viewSet *ViewSet[_]) Register(handler gin.IRouter) {
	gr := handler.Group(viewSet.BasePath)
	{
		for _, route := range viewSet.Actions {
			gr.Handle(
				route.Method,
				route.SubPath,
				getHandler(route.Action, *viewSet, route.Handler),
			)
		}
	}
}

func shouldAddAction(action string, excludeList []string) bool {
	for _, exclAction := range excludeList {
		if exclAction == action {
			return false
		}
	}
	return true
}

func getHandler[EntityType any](
	action string,
	viewSet ViewSet[EntityType], // copy viewset for each route
	function HandlerWithViewSetFunc[EntityType],
) gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := viewSet.PermissionChecker.Check(action, c); err != nil {
			viewSet.ExceptionHandler.Handle(NewViewSetError(
				err.Error(), http.StatusForbidden, err,
			), c)
			return
		}
		function(action, &viewSet, c)
	}
}

func List[EntityType any](
	action string,
	viewSet *ViewSet[EntityType],
	c *gin.Context,
) {
	paginatedMeta := new(map[string]any)
	entities := make([]*EntityType, 0, 20)
	manyResponse := make([]map[string]any, 0, 20)

	if err := viewSet.Manager.GetObjects(&entities, paginatedMeta, c); err != nil {
		viewSet.ExceptionHandler.Handle(NewViewSetError(
			err.Error(), http.StatusInternalServerError, err,
		), c)
		return
	}
	if err := viewSet.Serializer.ManySerialize(&manyResponse, &entities, c); err != nil {
		viewSet.ExceptionHandler.Handle(NewViewSetError(
			err.Error(), http.StatusInternalServerError, err,
		), c)
		return
	}
	c.JSON(http.StatusOK, map[string]any{
		"meta":    paginatedMeta,
		"results": manyResponse,
	})
}

func Retrieve[EntityType any](
	action string,
	viewSet *ViewSet[EntityType],
	c *gin.Context,
) {
	entity := new(EntityType)
	response := new(map[string]any)

	if err := viewSet.Manager.GetObject(&entity, c); err != nil {
		viewSet.ExceptionHandler.Handle(NewViewSetError(
			err.Error(), http.StatusNotFound, err,
		), c)
		return
	}
	if err := viewSet.Serializer.Serialize(response, entity, c); err != nil {
		viewSet.ExceptionHandler.Handle(NewViewSetError(
			err.Error(), http.StatusInternalServerError, err,
		), c)
		return
	}
	c.JSON(http.StatusOK, response)
}

func Create[EntityType any](
	action string,
	viewSet *ViewSet[EntityType],
	c *gin.Context,
) {
	var entity *EntityType // make nil
	validatedData := new(map[string]any)
	response := new(map[string]any)

	if err := viewSet.FormValidator.Validate(validatedData, entity, c); err != nil {
		viewSet.ExceptionHandler.Handle(NewViewSetError(
			err.Error(), http.StatusBadRequest, err,
		), c)
		return
	}
	if err := viewSet.Manager.Save(&entity, validatedData, c); err != nil {
		viewSet.ExceptionHandler.Handle(NewViewSetError(
			err.Error(), http.StatusBadRequest, err,
		), c)
		return
	}
	if err := viewSet.Serializer.Serialize(response, entity, c); err != nil {
		viewSet.ExceptionHandler.Handle(NewViewSetError(
			err.Error(), http.StatusInternalServerError, err,
		), c)
		return
	}
	c.JSON(http.StatusCreated, response)
}

func Update[EntityType any](
	action string,
	viewSet *ViewSet[EntityType],
	c *gin.Context,
) {
	entity := new(EntityType)
	validatedData := new(map[string]any)
	response := new(map[string]any)

	if err := viewSet.Manager.GetObject(&entity, c); err != nil {
		viewSet.ExceptionHandler.Handle(NewViewSetError(
			err.Error(), http.StatusNotFound, err,
		), c)
		return
	}
	if err := viewSet.FormValidator.Validate(validatedData, entity, c); err != nil {
		viewSet.ExceptionHandler.Handle(NewViewSetError(
			err.Error(), http.StatusBadRequest, err,
		), c)
		return
	}
	if err := viewSet.Manager.Save(&entity, validatedData, c); err != nil {
		viewSet.ExceptionHandler.Handle(NewViewSetError(
			err.Error(), http.StatusBadRequest, err,
		), c)
		return
	}
	if err := viewSet.Serializer.Serialize(response, entity, c); err != nil {
		viewSet.ExceptionHandler.Handle(NewViewSetError(
			err.Error(), http.StatusInternalServerError, err,
		), c)
		return
	}
	c.JSON(http.StatusOK, response)
}

func Delete[EntityType any](
	action string,
	viewSet *ViewSet[EntityType],
	c *gin.Context,
) {
	entity := new(EntityType)

	if err := viewSet.Manager.GetObject(&entity, c); err != nil {
		viewSet.ExceptionHandler.Handle(NewViewSetError(
			err.Error(), http.StatusNotFound, err,
		), c)
		return
	}
	if err := viewSet.Manager.Delete(&entity, c); err != nil {
		viewSet.ExceptionHandler.Handle(NewViewSetError(
			err.Error(), http.StatusBadRequest, err,
		), c)
		return
	}
	c.JSON(http.StatusNoContent, map[string]any{})
}
