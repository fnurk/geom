package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/fnurk/geom/pkg/auth"
	"github.com/fnurk/geom/pkg/model"
	"github.com/fnurk/geom/pkg/pubsub"
	"github.com/fnurk/geom/pkg/store"
	"github.com/labstack/echo/v4"
	"golang.org/x/net/websocket"
)

type CRUDLAccessCheckers struct {
	GetCheck    auth.AccessFunc
	PostCheck   auth.AccessFunc
	PutCheck    auth.AccessFunc
	DeleteCheck auth.AccessFunc
	LiveCheck   auth.AccessFunc
}

func AddCrudEndpointsForType(e *echo.Echo, db store.Database, pb pubsub.Pubsub, t string, checkers CRUDLAccessCheckers) {
	e.GET("/"+t+"/:id", Get(db, t, checkers.GetCheck))
	e.POST("/"+t, Post(db, t, checkers.PostCheck))
	e.PUT("/"+t+"/:id", Put(db, t, checkers.PutCheck))
	e.DELETE("/"+t+"/:id", Delete(db, t, checkers.DeleteCheck))
	e.GET("/"+t+"/:id/live", LiveUpdates(db, t, pb, checkers.LiveCheck))
}

func AddCrudEndpointsForTypeInGroup(e *echo.Group, db store.Database, pb pubsub.Pubsub, t string, checkers CRUDLAccessCheckers) {
	e.GET("/"+t+"/:id", Get(db, t, checkers.GetCheck))
	e.POST("/"+t, Post(db, t, checkers.PostCheck))
	e.PUT("/"+t+"/:id", Put(db, t, checkers.PutCheck))
	e.DELETE("/"+t+"/:id", Delete(db, t, checkers.DeleteCheck))
	e.GET("/"+t+"/:id/live", LiveUpdates(db, t, pb, checkers.LiveCheck))
}

func Get(store store.Database, t string, accessChecker auth.AccessFunc) func(echo.Context) error {
	return func(c echo.Context) error {
		id := c.Param("id")

		doc, err := store.Get(t, id)
		if err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}
		if doc == nil {
			return c.NoContent(http.StatusNotFound)
		}

		if !accessChecker(c, doc) {
			return c.NoContent(http.StatusForbidden)
		}

		dataType := model.Types[t]
		obj := dataType

		err = json.Unmarshal([]byte(doc), &obj)

		if err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}

		if err != nil {
			return err
		}

		if doc == nil {
			return c.NoContent(http.StatusNotFound)
		}

		return c.JSON(http.StatusOK, obj)
	}
}

func Post(store store.Database, t string, accessChecker auth.AccessFunc) func(echo.Context) error {
	return func(c echo.Context) error {
		dataType := model.Types[t]
		obj := dataType
		if err := c.Bind(&obj); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}

		doc, err := json.Marshal(obj)

		if !accessChecker(c, doc) {
			return c.NoContent(http.StatusForbidden)
		}

		if err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}

		id, err := store.Put(t, "", doc)

		if err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}

		return c.String(http.StatusOK, fmt.Sprintf("%s", id))
	}
}

func Put(store store.Database, t string, accessChecker auth.AccessFunc) func(echo.Context) error {
	return func(c echo.Context) error {
		id := c.Param("id")

		doc, err := store.Get(t, id)
		if err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}
		if doc == nil {
			return c.NoContent(http.StatusNotFound)
		}
		if !accessChecker(c, doc) {
			return c.NoContent(http.StatusForbidden)
		}

		dataType := model.Types[t]
		obj := dataType
		if err := c.Bind(&obj); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}

		bytes, err := json.Marshal(obj)

		if err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}

		_, err = store.Put(t, id, bytes)

		if err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}
		return c.NoContent(http.StatusOK)
	}
}

func Delete(store store.Database, t string, accessChecker auth.AccessFunc) func(echo.Context) error {
	return func(c echo.Context) error {
		id := c.Param("id")

		doc, err := store.Get(t, id)

		if err != nil {
			return err
		}
		if doc == nil {
			return c.NoContent(http.StatusNotFound)
		}

		if !accessChecker(c, doc) {
			return c.NoContent(http.StatusForbidden)
		}

		err = store.Delete(t, id)

		if err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}

		return c.NoContent(http.StatusOK)
	}
}

func LiveUpdates(store store.Database, t string, changes pubsub.Pubsub, accessChecker auth.AccessFunc) func(echo.Context) error {
	return func(c echo.Context) error {
		id := c.Param("id")

		doc, err := store.Get(t, id)

		if err != nil {
			return err
		}

		if doc == nil {
			return c.NoContent(http.StatusNotFound)
		}

		if !accessChecker(c, doc) {
			return c.NoContent(http.StatusForbidden)
		}

		websocket.Handler(func(ws *websocket.Conn) {
			defer ws.Close()
			changes.Subscribe(fmt.Sprintf("%s.%s", t, id), func(m *pubsub.Message, s pubsub.Subscriber) {
				err := websocket.Message.Send(ws, m.Body)
				if err != nil {
					s.Unsubscribe()
					ws.Close()
				}
			}, func() {
				ws.Close()
			})

			for {
				msg := ""
				err := websocket.Message.Receive(ws, &msg)
				if err != nil {
					ws.Close()
					break
				}
			}

		}).ServeHTTP(c.Response(), c.Request())
		return nil
	}
}
