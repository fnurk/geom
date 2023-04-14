package handlers

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/fnurk/geom/pkg/store"
	"github.com/labstack/echo/v4"
)

func GetDoc(docs store.DocStore) echo.HandlerFunc {
	return func(c echo.Context) error {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			return err
		}

		doc, err := docs.Get(id)

		if err != nil {
			return err
		}

		if doc == nil {
			return c.NoContent(http.StatusNotFound)
		}

		return c.String(http.StatusOK, fmt.Sprintf("DOC: %v", doc.Content))
	}
}

func PostDoc(docs store.DocStore) echo.HandlerFunc {
	return func(c echo.Context) error {
		body, err := ioutil.ReadAll(c.Request().Body)
		if err != nil {
			return err
		}

		doc := store.Document{
			Content:     body,
			ContentType: c.Request().Header["Content-Type"][0],
		}

		id, err := docs.Create(doc)

		if err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}

		fmt.Printf("ID: %d", id)

		return c.String(http.StatusOK, fmt.Sprintf("%d", id))
	}
}

func PutDoc(docs store.DocStore) echo.HandlerFunc {
	return func(c echo.Context) error {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			return err
		}

		doc, err := docs.Get(id)
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

		doc.Content = body
		doc.ContentType = c.Request().Header["Content-Type"][0]

		err = docs.Put(id, *doc)

		if err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}

		return c.NoContent(http.StatusOK)

	}
}

func DeleteDoc(docs store.DocStore) echo.HandlerFunc {
	return func(c echo.Context) error {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			return err
		}

		doc, err := docs.Get(id)
		if err != nil {
			return err
		}
		if doc == nil {
			return c.NoContent(http.StatusNotFound)
		}

		err = docs.Delete(id)

		if err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}

		return c.NoContent(http.StatusOK)

	}
}
