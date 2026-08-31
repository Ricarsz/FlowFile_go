package handler

import (
	"io"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
)

type FileServer interface {
	Store(key string, r io.Reader) error
	Get(key string) error
	Has(key string) bool
	Read(key string) (int64, io.ReadCloser, error)
	Delete(key string) error
}

type Handler struct {
	fs FileServer
}

func New(fs FileServer) *Handler {
	return &Handler{fs: fs}
}

func (h *Handler) Register(e *echo.Echo) {
	e.POST("/store/:key", h.Store)
	e.GET("/store/:key", h.Get)
	e.DELETE("/store/:key", h.Delete)
}

func (h *Handler) Store(c echo.Context) error {
	key := c.Param("key")
	if key == "" {
		return c.String(http.StatusBadRequest, "missing key")
	}
	if err := h.fs.Store(key, c.Request().Body); err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	return c.String(http.StatusCreated, "stored "+key)
}

func (h *Handler) Get(c echo.Context) error {
	key := c.Param("key")
	if key == "" {
		return c.String(http.StatusBadRequest, "missing key")
	}
	if !h.fs.Has(key) {
		if err := h.fs.Get(key); err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}
		if !h.fs.Has(key) {
			return c.String(http.StatusNotFound, "not found "+key)
		}
	}
	size, r, err := h.fs.Read(key)
	if err != nil {
		return c.String(http.StatusNotFound, err.Error())
	}
	defer r.Close()
	c.Response().Header().Set("Content-Length", strconv.FormatInt(size, 10))
	return c.Stream(http.StatusOK, "application/octet-stream", r)
}

func (h *Handler) Delete(c echo.Context) error {
	key := c.Param("key")
	if key == "" {
		return c.String(http.StatusBadRequest, "missing key")
	}
	if !h.fs.Has(key) {
		return c.String(http.StatusNotFound, "not found"+key)
	}
	if err:=h.fs.Delete(key);err!=nil{
		return c.String(http.StatusBadRequest, "err")
	}
	return c.String(http.StatusNotImplemented, "todo")
}
