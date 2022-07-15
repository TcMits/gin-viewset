package viewset

import (
	"errors"
	"strconv"

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
	pagiantedMeta *gin.H,
	c *gin.Context,
) error {
	if om.RaiseError {
		return errors.New("Get objects error")
	}
	(*pagiantedMeta)["count"] = len(om.Database)
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
	validatedObject *testObjectRequest,
	c *gin.Context,
) error {
	if om.RaiseError {
		return errors.New("Saving error")
	}
	if *dest == nil {
		// create
		newObject := testObject{
			Pk:   len(om.Database) + 1,
			Name: validatedObject.Name,
			Age:  validatedObject.Age,
		}
		om.Database = append(om.Database, newObject)
		*dest = &newObject
		return nil
	}
	(*dest).Name = validatedObject.Name
	(*dest).Age = validatedObject.Age
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
