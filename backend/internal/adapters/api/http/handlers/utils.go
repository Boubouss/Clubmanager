package handlers

import (
	"github.com/a-h/templ"
	"github.com/labstack/echo/v5"
)

func render(c *echo.Context, component templ.Component) error {
	c.Response().Header().Set(echo.HeaderContentType, "text/html; charset=utf-8")
	return component.Render(c.Request().Context(), c.Response())
}
