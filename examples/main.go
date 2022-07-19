package main

import (
	"errors"
	"strconv"

	"github.com/TcMits/viewset"
	"github.com/gin-gonic/gin"
)

type Person struct {
	Pk   int    `mapstructure:"pk"`
	Name string `mapstructure:"name"`
	Age  int    `mapstructure:"age"`
}

type PersonRequest struct {
	Name string `json:"name" form:"name" binding:"required"`
	Age  int    `json:"age" form:"age" binding:"required"`
}

type PersonManager struct {
	Database []Person
}

func (pm *PersonManager) GetObjects(
	dest *[]*Person,
	paginatedMeta *gin.H,
	c *gin.Context,
) error {
	(*paginatedMeta)["count"] = len(pm.Database)
	for i := range pm.Database {
		*dest = append(*dest, &pm.Database[i])
	}
	return nil
}

func (pm *PersonManager) GetObject(
	dest **Person,
	c *gin.Context,
) error {
	pk, err := strconv.Atoi(c.Param("pk"))
	if err != nil {
		return errors.New("Object not found")
	}
	for i, object := range pm.Database {
		if object.Pk == pk {
			*dest = &pm.Database[i]
			return nil
		}
	}
	return errors.New("Object not found")
}

func (pm *PersonManager) Save(
	dest **Person,
	validatedObject *PersonRequest,
	c *gin.Context,
) error {
	if *dest == nil {
		// create
		newObject := Person{
			Pk:   len(pm.Database) + 1,
			Name: validatedObject.Name,
			Age:  validatedObject.Age,
		}
		pm.Database = append(pm.Database, newObject)
		*dest = &newObject
		return nil
	}
	(*dest).Name = validatedObject.Name
	(*dest).Age = validatedObject.Age
	return nil
}

func (pm *PersonManager) Delete(
	dest **Person,
	c *gin.Context,
) error {
	for i, object := range pm.Database {
		if object.Pk == (*dest).Pk {
			pm.Database = append(pm.Database[:i], pm.Database[i+1:]...)
			return nil
		}
	}
	return errors.New("Object not found")
}

func main() {
	r := gin.Default()
	basePath := "/users"

	personManager := &PersonManager{}
	personManager.Database = append(
		personManager.Database,
		Person{Pk: 1, Name: "test", Age: 20},
	)
	viewSet := viewset.NewViewSet[Person, PersonRequest](
		basePath, "/:pk", nil, nil, personManager, nil, nil, nil, nil,
	)
	viewSet.Register(r)
	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}
