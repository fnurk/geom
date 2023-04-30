package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/fnurk/geom/pkg/model"
	"github.com/fnurk/geom/pkg/pubsub"
	"github.com/fnurk/geom/pkg/store"
	"github.com/labstack/echo/v4"
	"golang.org/x/net/websocket"
)

type AccessFunc func([]byte) bool

func checkAccess(doc []byte, checks ...AccessFunc) bool {
	allowed := false
	for _, check := range checks {
		if check(doc) {
			allowed = true
			break
		}
	}
	return allowed
}

func Get(store store.Database, t string, accessCheckers ...AccessFunc) func(echo.Context) error {
	return func(c echo.Context) error {
		id := c.Param("id")

		doc, err := store.Get(t, id)
		if err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}
		if doc == nil {
			return c.NoContent(http.StatusNotFound)
		}

		if !checkAccess(doc, accessCheckers...) {
			return c.NoContent(http.StatusForbidden)
		}

		dataType := model.Types[t]
		obj := dataType.Template

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

func Post(store store.Database, t string, accessCheckers ...AccessFunc) func(echo.Context) error {
	return func(c echo.Context) error {
		dataType := model.Types[t]
		obj := dataType.Template
		if err := c.Bind(&obj); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}

		doc, err := json.Marshal(obj)

		if !checkAccess(doc, accessCheckers...) {
			return c.NoContent(http.StatusForbidden)
		}

		if err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}

		id, err := store.Put(t, "", doc)

		if err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}

		return c.String(http.StatusOK, fmt.Sprintf("%d", id))
	}
}

func Put(store store.Database, t string, accessCheckers ...AccessFunc) func(echo.Context) error {
	return func(c echo.Context) error {
		id := c.Param("id")

		doc, err := store.Get(t, id)
		if err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}
		if doc == nil {
			return c.NoContent(http.StatusNotFound)
		}
		if !checkAccess(doc, accessCheckers...) {
			return c.NoContent(http.StatusForbidden)
		}

		dataType := model.Types[t]
		obj := dataType.Template
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

func Delete(store store.Database, t string, accessCheckers ...AccessFunc) func(echo.Context) error {
	return func(c echo.Context) error {
		id := c.Param("id")

		doc, err := store.Get(t, id)

		if err != nil {
			return err
		}
		if doc == nil {
			return c.NoContent(http.StatusNotFound)
		}

		if !checkAccess(doc, accessCheckers...) {
			return c.NoContent(http.StatusForbidden)
		}

		err = store.Delete(t, id)

		if err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}

		return c.NoContent(http.StatusOK)
	}
}

func LiveUpdates(store store.Database, t string, changes *pubsub.Pubsub, accessCheckers ...AccessFunc) func(echo.Context) error {
	return func(c echo.Context) error {
		id := c.Param("id")

		doc, err := store.Get(t, id)

		if err != nil {
			return err
		}

		if doc == nil {
			return c.NoContent(http.StatusNotFound)
		}

		if !checkAccess(doc, accessCheckers...) {
			return c.NoContent(http.StatusForbidden)
		}

		websocket.Handler(func(ws *websocket.Conn) {
			defer ws.Close()
			changes.Subscribe(fmt.Sprintf("%s.%s", t, id), func(m *pubsub.Message, s *pubsub.Subscriber) {
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
