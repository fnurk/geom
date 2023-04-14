package handlers

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/fnurk/geom/pkg/model"
	"github.com/fnurk/geom/pkg/store"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func Get() echo.HandlerFunc {
	return func(c echo.Context) error {
		id := c.Param("id")
		t := c.Param("type")

		doc, err := store.Get(t, id)

		if err != nil {
			return err
		}

		if doc == nil {
			return c.NoContent(http.StatusNotFound)
		}

		return c.JSON(http.StatusOK, doc)
	}
}

func Post() echo.HandlerFunc {
	return func(c echo.Context) error {
		t := c.Param("type")

		dataType := model.Types[t]
		obj := dataType.Template
		if err := c.Bind(&obj); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}

		var id *string
		switch dataType.IdType {
		case model.UUID:
			uuid := uuid.New().String()
			id = &uuid
		case model.AutoIncr:
			empty := ""
			id = &empty //Make nicer, empty string -> id from db

		}

		id, err := store.Put(t, *id, obj)

		if err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}

		return c.String(http.StatusOK, fmt.Sprintf("%s", *id))
	}
}

func Put() echo.HandlerFunc {
	return func(c echo.Context) error {
		id := c.Param("id")
		t := c.Param("type")

		doc, err := store.Get(t, id)
		if err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}
		if doc == nil {
			return c.NoContent(http.StatusNotFound)
		}

		body, err := ioutil.ReadAll(c.Request().Body)
		if err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}

		_, err = store.Put(t, id, body)

		if err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}

		return c.NoContent(http.StatusOK)

	}
}

func Delete() echo.HandlerFunc {
	return func(c echo.Context) error {
		id := c.Param("id")
		t := c.Param("type")

		doc, err := store.Get(t, id)
		if err != nil {
			return err
		}
		if doc == nil {
			return c.NoContent(http.StatusNotFound)
		}

		err = store.Delete(t, id)

		if err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}

		return c.NoContent(http.StatusOK)

	}
}
