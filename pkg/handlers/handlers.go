package handlers

import (
	"fmt"
	"net/http"

	"github.com/fnurk/geom/pkg/model"
	"github.com/fnurk/geom/pkg/store"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func Get(c echo.Context) error {
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

func Post(c echo.Context) error {
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
		id = &empty //TODO: Make nicer, now empty string -> id from db
	}

	id, err := store.Put(t, *id, obj)

	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	return c.String(http.StatusOK, fmt.Sprintf("%s", *id))
}

func Put(c echo.Context) error {
	id := c.Param("id")
	t := c.Param("type")

	doc, err := store.Get(t, id)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	if doc == nil {
		return c.NoContent(http.StatusNotFound)
	}

	dataType := model.Types[t]
	obj := dataType.Template
	if err := c.Bind(&obj); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	_, err = store.Put(t, id, obj)

	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	return c.NoContent(http.StatusOK)
}

func Delete(c echo.Context) error {
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
