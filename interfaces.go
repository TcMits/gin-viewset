package viewset

import "github.com/gin-gonic/gin"

type Field[EntityType any] interface {
	Serialize(*EntityType, *gin.Context) (any, error)
}

type SingleSerializer[EntityType any] interface {
	Serialize(*gin.H, *EntityType, *gin.Context) error
}

type ManySerializer[EntityType any] interface {
	ManySerialize(*[]gin.H, *[]*EntityType, *gin.Context) error
}

type Serializer[EntityType any] interface {
	SingleSerializer[EntityType]
	ManySerializer[EntityType]
}

type FormValidator[EntityType any, ValidatedType any] interface {
	Validate(*ValidatedType, *EntityType, *gin.Context) error
}

type ObjectManager[EntityType any, ValidatedType any] interface {
	GetObject(**EntityType, *gin.Context) error
	GetObjects(*[]*EntityType, *gin.H, *gin.Context) error
	Save(**EntityType, *ValidatedType, *gin.Context) error
	Delete(**EntityType, *gin.Context) error
}

type PermissionChecker interface {
	Check(string, *gin.Context) error
	// check action
}

type ExceptionHandler interface {
	Handle(error, *gin.Context)
	// use something like ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"Message": "Unauthorized"}) to abort request
}
