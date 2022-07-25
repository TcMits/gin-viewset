package manager

import (
	"github.com/gin-gonic/gin"
)

type Manager[EntityType, ValidateType any] interface {
	GetObject(**EntityType, *gin.Context) error
	GetObjects(*[]*EntityType, *map[string]any, *gin.Context) error
	Save(**EntityType, *ValidateType, *gin.Context) error
	Delete(**EntityType, *gin.Context) error
}
