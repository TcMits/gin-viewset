package viewset

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	DEFAULT_LIST_ACTION     = "list"
	DEFAULT_RETRIEVE_ACTION = "retrieve"
	DEFAULT_CREATE_ACTION   = "create"
	DEFAULT_UPDATE_ACTION   = "update"
	DEFAULT_DELETE_ACTION   = "delete"
)

type HandlerWithViewSetFunc[EntityType any, ValidatedType any] func(
	string,
	*ViewSet[EntityType, ValidatedType],
	*gin.Context,
)

type Route[EntityType any, ValidatedType any] struct {
	Action  string
	SubPath string
	Method  string
	Handler HandlerWithViewSetFunc[EntityType, ValidatedType]
}

type ViewSet[EntityType any, ValidatedType any] struct {
	BasePath string
	Actions  []Route[EntityType, ValidatedType]

	ExceptionHandler
	PermissionChecker
	ObjectManager ObjectManager[EntityType, ValidatedType]
	Serializer    Serializer[EntityType]
	FormValidator FormValidator[EntityType, ValidatedType]
}

func NewViewSet[EntityType any, ValidatedType any](
	basePath string,
	detailParams string,
	excludeDefaultActions []string,
	additionalActions []Route[EntityType, ValidatedType],

	objectManager ObjectManager[EntityType, ValidatedType],
	exceptionHandler ExceptionHandler,
	permissionChecker PermissionChecker,
	serializer Serializer[EntityType],
	formValidator FormValidator[EntityType, ValidatedType],
) *ViewSet[EntityType, ValidatedType] {
	if objectManager == nil {
		panic("objectManager is required")
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

	viewSet := &ViewSet[EntityType, ValidatedType]{
		BasePath:          basePath,
		ExceptionHandler:  exceptionHandler,
		PermissionChecker: permissionChecker,
		ObjectManager:     objectManager,
		Serializer:        serializer,
		FormValidator:     formValidator,
	}

	if shouldAddAction(DEFAULT_LIST_ACTION, excludeDefaultActions) {
		viewSet.Actions = append(viewSet.Actions, Route[EntityType, ValidatedType]{
			Action:  DEFAULT_LIST_ACTION,
			SubPath: "/",
			Method:  http.MethodGet,
			Handler: List[EntityType, ValidatedType],
		})
	}
	if shouldAddAction(DEFAULT_RETRIEVE_ACTION, excludeDefaultActions) {
		viewSet.Actions = append(viewSet.Actions, Route[EntityType, ValidatedType]{
			Action:  DEFAULT_RETRIEVE_ACTION,
			SubPath: detailParams,
			Method:  http.MethodGet,
			Handler: Retrieve[EntityType, ValidatedType],
		})
	}
	if shouldAddAction(DEFAULT_CREATE_ACTION, excludeDefaultActions) {
		viewSet.Actions = append(viewSet.Actions, Route[EntityType, ValidatedType]{
			Action:  DEFAULT_CREATE_ACTION,
			SubPath: "/",
			Method:  http.MethodPost,
			Handler: Create[EntityType, ValidatedType],
		})
	}
	if shouldAddAction(DEFAULT_UPDATE_ACTION, excludeDefaultActions) {
		viewSet.Actions = append(
			viewSet.Actions,
			Route[EntityType, ValidatedType]{
				Action:  DEFAULT_UPDATE_ACTION,
				SubPath: detailParams,
				Method:  http.MethodPut,
				Handler: Update[EntityType, ValidatedType],
			}, Route[EntityType, ValidatedType]{
				Action:  DEFAULT_UPDATE_ACTION,
				SubPath: detailParams,
				Method:  http.MethodPatch,
				Handler: Update[EntityType, ValidatedType],
			},
		)
	}
	if shouldAddAction(DEFAULT_DELETE_ACTION, excludeDefaultActions) {
		viewSet.Actions = append(viewSet.Actions, Route[EntityType, ValidatedType]{
			Action:  DEFAULT_DELETE_ACTION,
			SubPath: detailParams,
			Method:  http.MethodDelete,
			Handler: Delete[EntityType, ValidatedType],
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
				getHandler(route.Action, viewSet, route.Handler),
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

func getHandler[EntityType any, ValidatedType any](
	action string,
	viewSet *ViewSet[EntityType, ValidatedType],
	function HandlerWithViewSetFunc[EntityType, ValidatedType],
) gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := viewSet.PermissionChecker.Check(action, c); err != nil {
			viewSet.ExceptionHandler.Handle(NewViewSetError(
				err.Error(), http.StatusForbidden, err,
			), c)
			return
		}
		function(action, viewSet, c)
	}
}

func List[EntityType any, ValidatedType any](
	action string,
	viewSet *ViewSet[EntityType, ValidatedType],
	c *gin.Context,
) {
	paginatedMeta := gin.H{}
	entities := make([]*EntityType, 0, 20)
	manyResponse := make([]gin.H, 0, 20)

	if err := viewSet.ObjectManager.GetObjects(&entities, &paginatedMeta, c); err != nil {
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
	c.JSON(http.StatusOK, gin.H{
		"meta":    paginatedMeta,
		"results": manyResponse,
	})
}

func Retrieve[EntityType any, ValidatedType any](
	action string,
	viewSet *ViewSet[EntityType, ValidatedType],
	c *gin.Context,
) {
	entity := new(EntityType)
	response := gin.H{}

	if err := viewSet.ObjectManager.GetObject(&entity, c); err != nil {
		viewSet.ExceptionHandler.Handle(NewViewSetError(
			err.Error(), http.StatusNotFound, err,
		), c)
		return
	}
	if err := viewSet.Serializer.Serialize(&response, entity, c); err != nil {
		viewSet.ExceptionHandler.Handle(NewViewSetError(
			err.Error(), http.StatusInternalServerError, err,
		), c)
		return
	}
	c.JSON(http.StatusOK, response)
}

func Create[EntityType any, ValidatedType any](
	action string,
	viewSet *ViewSet[EntityType, ValidatedType],
	c *gin.Context,
) {
	var entity *EntityType // make nil
	entityRequest := new(ValidatedType)
	response := gin.H{}

	if err := viewSet.FormValidator.Validate(entityRequest, entity, c); err != nil {
		viewSet.ExceptionHandler.Handle(NewViewSetError(
			err.Error(), http.StatusBadRequest, err,
		), c)
		return
	}
	if err := viewSet.ObjectManager.Save(&entity, entityRequest, c); err != nil {
		viewSet.ExceptionHandler.Handle(NewViewSetError(
			err.Error(), http.StatusBadRequest, err,
		), c)
		return
	}
	if err := viewSet.Serializer.Serialize(&response, entity, c); err != nil {
		viewSet.ExceptionHandler.Handle(NewViewSetError(
			err.Error(), http.StatusInternalServerError, err,
		), c)
		return
	}
	c.JSON(http.StatusCreated, response)
}

func Update[EntityType any, ValidatedType any](
	action string,
	viewSet *ViewSet[EntityType, ValidatedType],
	c *gin.Context,
) {
	entity := new(EntityType)
	entityRequest := new(ValidatedType)
	response := gin.H{}

	if err := viewSet.ObjectManager.GetObject(&entity, c); err != nil {
		viewSet.ExceptionHandler.Handle(NewViewSetError(
			err.Error(), http.StatusNotFound, err,
		), c)
		return
	}
	if err := viewSet.FormValidator.Validate(entityRequest, entity, c); err != nil {
		viewSet.ExceptionHandler.Handle(NewViewSetError(
			err.Error(), http.StatusBadRequest, err,
		), c)
		return
	}
	if err := viewSet.ObjectManager.Save(&entity, entityRequest, c); err != nil {
		viewSet.ExceptionHandler.Handle(NewViewSetError(
			err.Error(), http.StatusBadRequest, err,
		), c)
		return
	}
	if err := viewSet.Serializer.Serialize(&response, entity, c); err != nil {
		viewSet.ExceptionHandler.Handle(NewViewSetError(
			err.Error(), http.StatusInternalServerError, err,
		), c)
		return
	}
	c.JSON(http.StatusOK, response)
}

func Delete[EntityType any, ValidatedType any](
	action string,
	viewSet *ViewSet[EntityType, ValidatedType],
	c *gin.Context,
) {
	entity := new(EntityType)

	if err := viewSet.ObjectManager.GetObject(&entity, c); err != nil {
		viewSet.ExceptionHandler.Handle(NewViewSetError(
			err.Error(), http.StatusNotFound, err,
		), c)
		return
	}
	if err := viewSet.ObjectManager.Delete(&entity, c); err != nil {
		viewSet.ExceptionHandler.Handle(NewViewSetError(
			err.Error(), http.StatusBadRequest, err,
		), c)
		return
	}
	c.JSON(http.StatusNoContent, gin.H{})
}
