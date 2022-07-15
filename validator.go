package viewset

import "github.com/gin-gonic/gin"

type DefaultValidator[EntityType any, ValidatedType any] struct{}

func (_ *DefaultValidator[EntityType, ValidatedType]) Validate(
	dest *ValidatedType, entity *EntityType, c *gin.Context,
) error {
	if err := c.ShouldBind(&dest); err != nil {
		return err
	}
	return nil
}
