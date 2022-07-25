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
