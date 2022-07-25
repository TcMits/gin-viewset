package viewset

import "github.com/gin-gonic/gin"

type Field[EntityType any] interface {
	Serialize(*EntityType, *gin.Context) (any, error)
}

type SingleSerializer[EntityType any] interface {
	Serialize(*map[string]any, *EntityType, *gin.Context) error
}

type ManySerializer[EntityType any] interface {
	ManySerialize(*[]map[string]any, *[]*EntityType, *gin.Context) error
}

type Serializer[EntityType any] interface {
	SingleSerializer[EntityType]
	ManySerializer[EntityType]
}

type FormValidator[EntityType any, ValidateType any] interface {
	Validate(*ValidateType, *EntityType, *gin.Context) error
}

type PermissionChecker interface {
	Check(string, *gin.Context) error
	// check action
}

type ExceptionHandler interface {
	Handle(error, *gin.Context)
	// use something like ctx.AbortWithStatusJSON(http.StatusUnauthorized, map[string]any{"Message": "Unauthorized"}) to abort request
}
