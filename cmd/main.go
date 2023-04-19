package main

import (
	"net/http"

	"github.com/fnurk/geom/pkg/handlers"
	"github.com/fnurk/geom/pkg/model"
	"github.com/fnurk/geom/pkg/store"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/tidwall/gjson"
)

type (
	CustomValidator struct {
		validator *validator.Validate
	}
)

func main() {
	e := echo.New()

	e.Validator = &CustomValidator{validator: validator.New()}

	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "${method} ${uri} -> ${status}\n",
	}))

	err := store.Init()
	if err != nil {
		e.Logger.Fatal(err)
	}
	defer store.Close()

	//For testing subscriptions
	// store.Changes.Subscribe("*", func(m *pubsub.Message) {
	// 	e.Logger.Printf("Got change on %s, body: %s", m.Topic, m.Body)
	// }, func() {})

	e.Use(middleware.Recover())

	//Set up CRUD+live updates for all types registered in model package
	for k := range model.Types {
		AddCrudEndpointsForType(e, k)
	}
	e.Static("/", ".")

	e.Logger.Fatal(e.Start(":8080"))
}

func AddCrudEndpointsForType(e *echo.Echo, t string) {
	open := func(doc []byte) bool {
		return true
	}

	isOwner := func(doc []byte) bool {
		return gjson.GetBytes(doc, "createdBy").String() == "124"
	}

	isSharedWith := func(doc []byte) bool {
		shares := gjson.GetBytes(doc, "sharedWith").Array()
		for _, v := range shares {
			if v.String() == "124" {
				return true
			}
		}
		return false
	}

	e.GET("/"+t+"/:id", handlers.Get(t, isOwner, isSharedWith))
	e.POST("/"+t, handlers.Post(t, open))
	e.PUT("/"+t+"/:id", handlers.Put(t, isOwner, isSharedWith))
	e.DELETE("/"+t+"/:id", handlers.Delete(t, isOwner, isSharedWith))
	e.GET("/"+t+"/:id/live", handlers.LiveUpdates(t, isOwner, isSharedWith))
}

//TODO: Start using this
func (cv *CustomValidator) Validate(i interface{}) error {
	if err := cv.validator.Struct(i); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return nil
}
