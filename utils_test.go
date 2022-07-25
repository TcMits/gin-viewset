package viewset

import (
	"errors"

	"github.com/gin-gonic/gin"
)

type testObject struct {
	Pk   int    `mapstructure:"-"`
	Name string `mapstructure:"name"`
	Age  int    `mapstructure:"age"`
}

type testObjectRequest struct {
	Name string `json:"name" form:"name" binding:"required"`
	Age  int    `json:"age" form:"age" binding:"required"`
}

type testObjectURI struct {
	Pk int `uri:"pk"`
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
	uri := new(testObjectURI)
	if err := c.ShouldBindUri(uri); err != nil {
		return err
	}
	pk := uri.Pk
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
	validatedData *testObjectRequest,
	c *gin.Context,
) error {
	if om.RaiseError {
		return errors.New("Saving error")
	}
	if *dest == nil {
		// create
		newObject := testObject{
			Pk:   len(om.Database) + 1,
			Name: validatedData.Name,
			Age:  validatedData.Age,
		}
		om.Database = append(om.Database, newObject)
		*dest = &newObject
		return nil
	}
	(*dest).Name = validatedData.Name
	(*dest).Age = validatedData.Age
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
