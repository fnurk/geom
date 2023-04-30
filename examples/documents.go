package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/fnurk/geom/pkg/handlers"
	"github.com/fnurk/geom/pkg/model"
	"github.com/fnurk/geom/pkg/pubsub"
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

type Document interface {
	Note | Thing
}

type MetaFields struct {
	CreatedBy    string    `json:"createdBy"`
	Created      time.Time `json:"created"`
	LastModified time.Time `json:"lastModified"`
	SharedWith   []string  `json:"sharedWith"`
}

type Note struct {
	MetaFields
	Body string `json:"body" validate:"required`
}

type Thing struct {
	MetaFields
	Id string `json:"id" validate:"required"` //to be indexed
}

var changes *pubsub.Pubsub

var db store.Database

func main() {
	e := echo.New()

	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "${method} ${uri} -> ${status}\n",
	}))

	e.Validator = &CustomValidator{validator: validator.New()}

	boltdb, err := store.NewBoltDb("test.db")

	db = boltdb

	changes = pubsub.New(1000)

	model.RegisterType("note", Note{})
	model.RegisterType("thing", Thing{})

	db.AddPutHook(func(t string, id string, value []byte) {
		changes.Publish(pubsub.Message{
			Topic: fmt.Sprintf("%s.%s", t, id),
			Body:  string(value),
		})
	})

	err = db.Init()
	if err != nil {
		e.Logger.Fatal(err)
	}
	defer db.Close()

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

	e.GET("/"+t+"/:id", handlers.Get(db, t, isOwner, isSharedWith))
	e.POST("/"+t, handlers.Post(db, t, open))
	e.PUT("/"+t+"/:id", handlers.Put(db, t, isOwner, isSharedWith))
	e.DELETE("/"+t+"/:id", handlers.Delete(db, t, isOwner, isSharedWith))
	e.GET("/"+t+"/:id/live", handlers.LiveUpdates(db, t, changes, isOwner, isSharedWith))
}

// TODO: Start using this
func (cv *CustomValidator) Validate(i interface{}) error {
	if err := cv.validator.Struct(i); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return nil
}