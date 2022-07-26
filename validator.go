package viewset

import (
	"github.com/gin-gonic/gin"
)

var _ FormValidator[any, any] = &DefaultValidator[any, any]{}

type DefaultValidator[EntityType any, ValidateType any] struct{}

func (_ *DefaultValidator[EntityType, ValidateType]) Validate(
	dest *ValidateType, entity *EntityType, c *gin.Context,
) error {
	if err := c.ShouldBind(dest); err != nil {
		return err
	}
	return nil
}
