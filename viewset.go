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

type HandlerWithViewSetFunc[EntityType, ValidateType any] func(
	string,
	*ViewSet[EntityType, ValidateType],
	*gin.Context,
)

type Route[EntityType, ValidateType any] struct {
	Action  string
	SubPath string
	Method  string
	Handler HandlerWithViewSetFunc[EntityType, ValidateType]
}

type ViewSet[EntityType, ValidateType any] struct {
	BasePath string
	Actions  []Route[EntityType, ValidateType]

	ExceptionHandler
	PermissionChecker
	Manager       manager.Manager[EntityType, ValidateType]
	Serializer    Serializer[EntityType]
	FormValidator FormValidator[EntityType, ValidateType]
}

func NewViewSet[EntityType, ValidateType any](
	basePath string,
	detailParams string,
	excludeDefaultActions []string,
	additionalActions []Route[EntityType, ValidateType],

	manager manager.Manager[EntityType, ValidateType],
	exceptionHandler ExceptionHandler,
	permissionChecker PermissionChecker,
	serializer Serializer[EntityType],
	formValidator FormValidator[EntityType, ValidateType],
) *ViewSet[EntityType, ValidateType] {
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
		formValidator = &DefaultValidator[EntityType, ValidateType]{}
	}

	viewSet := &ViewSet[EntityType, ValidateType]{
		BasePath:          basePath,
		ExceptionHandler:  exceptionHandler,
		PermissionChecker: permissionChecker,
		Manager:           manager,
		Serializer:        serializer,
		FormValidator:     formValidator,
	}

	if shouldAddAction(DEFAULT_LIST_ACTION, excludeDefaultActions) {
		viewSet.Actions = append(viewSet.Actions, Route[EntityType, ValidateType]{
			Action:  DEFAULT_LIST_ACTION,
			SubPath: "/",
			Method:  http.MethodGet,
			Handler: List[EntityType, ValidateType],
		})
	}
	if shouldAddAction(DEFAULT_RETRIEVE_ACTION, excludeDefaultActions) {
		viewSet.Actions = append(viewSet.Actions, Route[EntityType, ValidateType]{
			Action:  DEFAULT_RETRIEVE_ACTION,
			SubPath: detailParams,
			Method:  http.MethodGet,
			Handler: Retrieve[EntityType, ValidateType],
		})
	}
	if shouldAddAction(DEFAULT_CREATE_ACTION, excludeDefaultActions) {
		viewSet.Actions = append(viewSet.Actions, Route[EntityType, ValidateType]{
			Action:  DEFAULT_CREATE_ACTION,
			SubPath: "/",
			Method:  http.MethodPost,
			Handler: Create[EntityType, ValidateType],
		})
	}
	if shouldAddAction(DEFAULT_UPDATE_ACTION, excludeDefaultActions) {
		viewSet.Actions = append(
			viewSet.Actions,
			Route[EntityType, ValidateType]{
				Action:  DEFAULT_UPDATE_ACTION,
				SubPath: detailParams,
				Method:  http.MethodPut,
				Handler: Update[EntityType, ValidateType],
			}, Route[EntityType, ValidateType]{
				Action:  DEFAULT_UPDATE_ACTION,
				SubPath: detailParams,
				Method:  http.MethodPatch,
				Handler: Update[EntityType, ValidateType],
			},
		)
	}
	if shouldAddAction(DEFAULT_DELETE_ACTION, excludeDefaultActions) {
		viewSet.Actions = append(viewSet.Actions, Route[EntityType, ValidateType]{
			Action:  DEFAULT_DELETE_ACTION,
			SubPath: detailParams,
			Method:  http.MethodDelete,
			Handler: Delete[EntityType, ValidateType],
		})
	}
	viewSet.Actions = append(viewSet.Actions, additionalActions...)
	return viewSet
}

func (viewSet *ViewSet[_, _]) Register(handler gin.IRouter) {
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

func getHandler[EntityType, ValidateType any](
	action string,
	viewSet ViewSet[EntityType, ValidateType], // copy viewset for each route
	function HandlerWithViewSetFunc[EntityType, ValidateType],
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

func List[EntityType, ValidateType any](
	action string,
	viewSet *ViewSet[EntityType, ValidateType],
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

func Retrieve[EntityType, ValidateType any](
	action string,
	viewSet *ViewSet[EntityType, ValidateType],
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

func Create[EntityType, ValidateType any](
	action string,
	viewSet *ViewSet[EntityType, ValidateType],
	c *gin.Context,
) {
	var entity *EntityType // make nil
	validatedData := new(ValidateType)
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

func Update[EntityType, ValidateType any](
	action string,
	viewSet *ViewSet[EntityType, ValidateType],
	c *gin.Context,
) {
	entity := new(EntityType)
	validatedData := new(ValidateType)
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

func Delete[EntityType, ValidateType any](
	action string,
	viewSet *ViewSet[EntityType, ValidateType],
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
