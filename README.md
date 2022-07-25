# viewset

## About the project

Gin basic CRUD endpoint/handler for objects

### But Why?!

Lazy <3.

### Layout

```tree
├── README.md
├── error.go
├── error_test.go
├── examples
│   └── gorm
│       └── main.go
├── go.mod
├── go.sum
├── interfaces.go
├── manager
│   ├── gorm.go
│   ├── gorm_test.go
│   └── interfaces.go
├── mock_test.go
├── permission.go
├── permission_test.go
├── pkg
│   └── urlclone
│       └── urlclone.go
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
go get github.com/TcMits/viewset
```
Import it in your code:<br />
```go
import "github.com/TcMits/viewset"
```

### Example

```go
package main

import (
	"github.com/TcMits/viewset"
	"github.com/TcMits/viewset/manager"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Book struct {
	ID     uint   `mapstructure:"id" gorm:"primary_key"`
	Title  string `mapstructure:"title"`
	Author string `mapstructure:"author"`
}

type BookRequest struct {
	Title  string `mapstructure:"title" form:"title" binding:"required"`
	Author string `mapstructure:"author" form:"author" binding:"required"`
}

type BookURI struct {
	ID uint `mapstructure:"id" uri:"pk" binding:"required"`
}

func main() {
	r := gin.Default()
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		panic("Failed to connect to database!")
	}
	db.AutoMigrate(&Book{})

	bookManager := manager.NewGormManager[Book, BookRequest, BookURI](
		db.Model(&Book{}), nil, nil, nil, nil, "db",
	)
	bookViewSet := viewset.NewViewSet[Book, BookRequest](
		"/books", "/:pk", nil, nil, bookManager, nil, nil, nil, nil,
	)
	bookViewSet.Register(r)
	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}
```
```
$ go run example.go
```
See the logs
```
[GIN-debug] GET    /books/                   --> github.com/TcMits/viewset.getHandler[...].func1 (3 handlers)
[GIN-debug] GET    /books/:pk                --> github.com/TcMits/viewset.getHandler[...].func1 (3 handlers)
[GIN-debug] POST   /books/                   --> github.com/TcMits/viewset.getHandler[...].func1 (3 handlers)
[GIN-debug] PUT    /books/:pk                --> github.com/TcMits/viewset.getHandler[...].func1 (3 handlers)
[GIN-debug] PATCH  /books/:pk                --> github.com/TcMits/viewset.getHandler[...].func1 (3 handlers)
[GIN-debug] DELETE /books/:pk                --> github.com/TcMits/viewset.getHandler[...].func1 (3 handlers)
```


Example listing view:
```json
{
  "meta": {
    "next": null,
    "previous": null
  },
  "results": [
    {
      "author": "phuc 2",
      "id": 1,
      "title": "first book 2"
    },
    {
      "author": "phuc",
      "id": 2,
      "title": "second book"
    }
  ]
}
```


Example detail view:
```json
{
  "author": "phuc 2",
  "id": 1,
  "title": "first book 2"
}
```

