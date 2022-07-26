package viewset

import (
	"github.com/gin-gonic/gin"
	"github.com/mitchellh/mapstructure"
)

var _ Serializer[any] = &DefaultSerializer[any]{}

type DefaultSerializer[EntityType any] struct {
	AdditionalField map[string]Field[EntityType]
}

func (s *DefaultSerializer[EntityType]) Serialize(
	dest *map[string]any, entity *EntityType, c *gin.Context,
) error {
	if err := mapstructure.Decode(entity, dest); err != nil {
		return err
	}
	for k, field := range s.AdditionalField {
		fieldValue, err := field.Serialize(entity, c)
		if err != nil {
			return err
		}
		(*dest)[k] = fieldValue

	}
	return nil
}

func (s *DefaultSerializer[EntityType]) ManySerialize(
	dest *[]map[string]any, entities *[]*EntityType, c *gin.Context,
) error {
	for _, entity := range *entities {
		var destObject map[string]any
		err := s.Serialize(&destObject, entity, c)
		if err != nil {
			return err
		}
		*dest = append(*dest, destObject)
	}
	return nil
}
