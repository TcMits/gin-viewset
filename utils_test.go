package viewset

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"
)

type testObject struct {
	Pk   int    `mapstructure:"-" uri:"pk"`
	Name string `mapstructure:"name" json:"name" form:"name" binding:"required"`
	Age  int    `mapstructure:"age" json:"age" form:"age" binding:"required"`
}

type testObjectManager struct {
	Database   []testObject
	RaiseError bool
}

type testField struct{}

func (tf testField) Serialize(object *testObject, c *gin.Context) (any, error) {
	if object.Age < 18 {
		return nil, errors.New("Fail")
	}
	return "testing", nil
}

func (om *testObjectManager) GetObjects(
	dest *[]*testObject,
	paginatedMeta *map[string]any,
	c *gin.Context,
) error {
	if om.RaiseError {
		return errors.New("Get objects error")
	}
	(*paginatedMeta) = map[string]any{"count": len(om.Database)}
	for i := range om.Database {
		*dest = append(*dest, &om.Database[i])
	}
	return nil
}

func (om *testObjectManager) GetObject(
	dest **testObject,
	c *gin.Context,
) error {
	pk, err := strconv.Atoi(c.Param("pk"))
	if err != nil {
		return errors.New("Object not found")
	}
	for i, object := range om.Database {
		if object.Pk == pk {
			*dest = &om.Database[i]
			return nil
		}
	}
	return errors.New("Object not found")
}

func (om *testObjectManager) Save(
	dest **testObject,
	validatedData *map[string]any,
	c *gin.Context,
) error {
	if om.RaiseError {
		return errors.New("Saving error")
	}
	if *dest == nil {
		// create
		newObject := testObject{
			Pk:   len(om.Database) + 1,
			Name: (*validatedData)["name"].(string),
			Age:  (*validatedData)["age"].(int),
		}
		om.Database = append(om.Database, newObject)
		*dest = &newObject
		return nil
	}
	(*dest).Name = (*validatedData)["name"].(string)
	(*dest).Age = (*validatedData)["age"].(int)
	return nil
}

func (om *testObjectManager) Delete(
	dest **testObject,
	c *gin.Context,
) error {
	if om.RaiseError {
		return errors.New("Deleting error")
	}
	for i, object := range om.Database {
		if object.Pk == (*dest).Pk {
			om.Database = append(om.Database[:i], om.Database[i+1:]...)
			return nil
		}
	}
	return errors.New("Object not found")
}

func SetUpRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	return router
}
