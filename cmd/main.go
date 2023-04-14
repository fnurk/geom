package main

import (
	"github.com/fnurk/geom/pkg/handlers"
	"github.com/fnurk/geom/pkg/store"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	e := echo.New()

	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "${method} ${uri} -> ${status}\n",
	}))

	docStore, err := store.Init()
	if err != nil {
		e.Logger.Fatal(err)
	}
	defer store.Close()

	e.Use(middleware.Recover())

	e.GET("/doc/:id", handlers.GetDoc(docStore))
	e.POST("/doc", handlers.PostDoc(docStore))
	e.PUT("/doc/:id", handlers.PutDoc(docStore))
	e.DELETE("/doc/:id", handlers.DeleteDoc(docStore))

	e.Logger.Fatal(e.Start(":8080"))
}
