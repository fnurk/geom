package main

import (
	"fmt"
	"time"

	"github.com/fnurk/geom/pkg/auth"
	"github.com/fnurk/geom/pkg/handlers"
	"github.com/fnurk/geom/pkg/model"

	"github.com/fnurk/geom/pkg/pubsub"
	"github.com/fnurk/geom/pkg/store"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/tidwall/gjson"
)

type MetaFields struct {
	CreatedBy    string    `json:"createdBy"`
	Created      time.Time `json:"created"`
	LastModified time.Time `json:"lastModified"`
	SharedWith   []string  `json:"sharedWith"`
}

type Note struct {
	MetaFields
	Body string `json:"body"`
}

type Thing struct {
	MetaFields
	Id      string `json:"id" index:"inmem"`        //to be indexed
	OtherId string `json:"otherId" index:"persist"` //to be indexed
}

var changes pubsub.Pubsub

var ds *store.Datastore

func main() {
	e := echo.New()

	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "${time_rfc3339}: ${method} ${uri} -> ${status}\n",
	}))

	boltdb, err := store.NewBoltDb("test.db")
	cache := store.NewInMemKV()

	ds = store.NewDatastore(boltdb, cache)

	changes = pubsub.NewChanPubsub()

	model.RegisterType("note", Note{})
	model.RegisterType("thing", Thing{})

	ds.AddPutHook(func(t string, id string, value []byte) {
		changes.Publish(&pubsub.Message{
			Topic: fmt.Sprintf("%s.%s", t, id),
			Body:  string(value),
		})
	})

	err = ds.Init()
	if err != nil {
		e.Logger.Fatal(err)
	}
	defer ds.Close()

	e.Use(middleware.Recover())

	handlers.AddCrudEndpointsForType(e, ds, changes, "note", handlers.CRUDLAccessCheckers{
		GetCheck:    auth.Any(isOwner, isSharedWith),
		PostCheck:   open,
		PutCheck:    auth.Any(isOwner, isSharedWith),
		DeleteCheck: auth.Any(isOwner, isSharedWith),
		LiveCheck:   auth.Any(isOwner, isSharedWith),
	})

	//use echo groups - maybe custom middleware for just these endpoints?
	docGroup := e.Group("/documents")

	//the access checkers could be reused here since they're the same
	handlers.AddCrudEndpointsForTypeInGroup(docGroup, ds, changes, "thing", handlers.CRUDLAccessCheckers{
		GetCheck:    auth.Any(isOwner, isSharedWith),
		PostCheck:   open,
		PutCheck:    auth.Any(isOwner, isSharedWith),
		DeleteCheck: auth.Any(isOwner, isSharedWith),
		LiveCheck:   auth.Any(isOwner, isSharedWith),
	})

	//Serve the dummy index.html
	e.Static("/", ".")

	e.Logger.Fatal(e.Start(":8080"))
}

var open = auth.AccessFunc(func(c echo.Context, doc []byte) bool {
	// Allow all reqs, assume middleware handles unauthenticated users.
	return true
})

var isOwner = auth.AccessFunc(func(c echo.Context, doc []byte) bool {
	// Get logged in user(in this case userId 124) from context, maybe middleware already populated it?
	return gjson.GetBytes(doc, "createdBy").String() == "124"
})

var isSharedWith = auth.AccessFunc(func(c echo.Context, doc []byte) bool {
	// Get logged in user(in this case userId 124) from context, maybe middleware already populated it?
	shares := gjson.GetBytes(doc, "sharedWith").Array()
	for _, v := range shares {
		if v.String() == "124" {
			return true
		}
	}
	return false
})
