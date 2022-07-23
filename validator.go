package viewset

import (
	"github.com/gin-gonic/gin"
	"github.com/mitchellh/mapstructure"
)

type DefaultValidator[EntityType any, ValidatedType any] struct{}

func (_ *DefaultValidator[EntityType, ValidatedType]) Validate(
	dest *map[string]any, entity *EntityType, c *gin.Context,
) error {
	validator := new(ValidatedType)
	if err := c.ShouldBind(validator); err != nil {
		return err
	}
	if err := mapstructure.Decode(validator, dest); err != nil {
		return err
	}
	return nil
}
