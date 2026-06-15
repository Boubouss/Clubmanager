package handlers

import (
	"clubmanager/internal/adapters/api/http/views/pages"
	"errors"
	"net/http"

	"github.com/labstack/echo/v5"
)

func HTTPErrorHandler(c *echo.Context, err error) {
	code := http.StatusInternalServerError
	message := "Une erreur inattendue s'est produite."

	var he *echo.HTTPError
	if errors.As(err, &he) {
		code = he.Code
		switch code {
		case http.StatusNotFound:
			message = "La page demandée est introuvable."
		case http.StatusForbidden:
			message = "Vous n'avez pas accès à cette ressource."
		case http.StatusUnauthorized:
			message = "Authentification requise."
		default:
			if he.Message != "" {
				message = he.Message
			}
		}
	}

	c.Response().WriteHeader(code)
	_ = render(c, pages.Error(code, message))
}
