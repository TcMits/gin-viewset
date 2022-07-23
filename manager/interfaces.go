package manager

import "github.com/gin-gonic/gin"

type Manager[EntityType any, URIType any] interface {
	GetObject(**EntityType, *gin.Context) error
	GetObjects(*[]*EntityType, *map[string]any, *gin.Context) error
	Save(**EntityType, *map[string]any, *gin.Context) error
	Delete(**EntityType, *gin.Context) error
}
