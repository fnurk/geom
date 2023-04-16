package main

import (
	"github.com/fnurk/geom/pkg/handlers"
	"github.com/fnurk/geom/pkg/pubsub"
	"github.com/fnurk/geom/pkg/store"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	e := echo.New()

	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "${method} ${uri} -> ${status}\n",
	}))

	store.AddIndex("thing", "id", "thing.id")

	err := store.Init()
	if err != nil {
		e.Logger.Fatal(err)
	}
	defer store.Close()

	store.Changes.Subscribe("*", func(m *pubsub.Message) {
		e.Logger.Printf("Got change on %s, body: %s", m.Topic, m.Body)

	}, func() {})

	e.Use(middleware.Recover())

	e.GET("/:type/:id", handlers.Get)
	e.POST("/:type", handlers.Post)
	e.PUT("/:type/:id", handlers.Put)
	e.DELETE("/:type/:id", handlers.Delete)

	e.Logger.Fatal(e.Start(":8080"))
}
