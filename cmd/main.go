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

	err := store.Init()
	if err != nil {
		e.Logger.Fatal(err)
	}
	defer store.Close()

	e.Use(middleware.Recover())

	e.GET("/:type/:id", handlers.Get())
	e.POST("/:type", handlers.Post())
	e.PUT("/:type/:id", handlers.Put())
	e.DELETE("/:type/:id", handlers.Delete())

	e.Logger.Fatal(e.Start(":8080"))
}
