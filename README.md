# gin-viewset

## About the project

Gin basic GRUD endpoint/handler for objects

### But Why?!

Lazy <3.

### Layout

```tree
├── README.md
├── error.go
├── error_test.go
├── examples
│   └── main.go
├── go.mod
├── go.sum
├── interfaces.go
├── mock_test.go
├── permission.go
├── permission_test.go
├── serializer.go
├── serializer_test.go
├── utils_test.go
├── validator.go
├── validator_test.go
├── viewset.go
└── viewset_test.go
```

### Design

Inspired by [Django Rest Framework](https://www.django-rest-framework.org/)

## Usage

### Start using it

Download and install it:<br />
```
go get github.com/TcMits/gin-viewset
```
Import it in your code:<br />
```go
import viewset "github.com/TcMits/gin-viewset"
```

### Example

[File](https://github.com/TcMits/gin-viewset/blob/main/examples/main.go)

```go
package main

import (
	"errors"
	"strconv"

	viewset "github.com/TcMits/gin-viewset"
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
	pagiantedMeta *gin.H,
	c *gin.Context,
) error {
	(*pagiantedMeta)["count"] = len(pm.Database)
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
```
```
$ go run example.go
```
See the logs
```
 - using env:   export GIN_MODE=release
 - using code:  gin.SetMode(gin.ReleaseMode)

[GIN-debug] GET    /users/                   --> github.com/TcMits/gin-viewset.getHandler[...].func1 (3 handlers)
[GIN-debug] GET    /users/:pk                --> github.com/TcMits/gin-viewset.getHandler[...].func1 (3 handlers)
[GIN-debug] POST   /users/                   --> github.com/TcMits/gin-viewset.getHandler[...].func1 (3 handlers)
[GIN-debug] PUT    /users/:pk                --> github.com/TcMits/gin-viewset.getHandler[...].func1 (3 handlers)
[GIN-debug] PATCH  /users/:pk                --> github.com/TcMits/gin-viewset.getHandler[...].func1 (3 handlers)
[GIN-debug] DELETE /users/:pk                --> github.com/TcMits/gin-viewset.getHandler[...].func1 (3 handlers)
[GIN-debug] [WARNING] You trusted all proxies, this is NOT safe. We recommend you to set a value.
Please check https://pkg.go.dev/github.com/gin-gonic/gin#readme-don-t-trust-all-proxies for details.
[GIN-debug] Environment variable PORT is undefined. Using port :8080 by default
```


Example listing view:
```json
{
  "meta": {
    "count": 1
  },
  "results": [
    {
      "age": 20,
      "name": "test",
      "pk": 1
    }
  ]
}
```


Example detail view:
```json
{
  "age": 20,
  "name": "test",
  "pk": 1
}
```

