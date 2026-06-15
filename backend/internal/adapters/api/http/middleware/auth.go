package middleware

import (
	"clubmanager/internal/app/services"
	"context"
	"net/http"

	"github.com/labstack/echo/v5"
)

const cookieName = "ClubManagerAuth"

// RequireAuth verifies the JWT cookie. On success it injects "userId" into the
// Echo context AND "user_id" into the request context.Context so that service
// methods (which use ctx.Value("user_id")) work identically to the gRPC path.
func RequireAuth(tkm services.TokenManager) func(echo.HandlerFunc) echo.HandlerFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			cookie, err := c.Cookie(cookieName)
			if err != nil || cookie.Value == "" {
				return c.Redirect(http.StatusSeeOther, "/user/connexion")
			}

			userId, err := tkm.ParseToken(cookie.Value)
			if err != nil {
				return c.Redirect(http.StatusSeeOther, "/user/connexion")
			}

			// Echo context — for handlers
			c.Set("userId", userId)

			// request context.Context — for service methods
			ctx := context.WithValue(c.Request().Context(), "user_id", userId)
			c.SetRequest(c.Request().WithContext(ctx))

			return next(c)
		}
	}
}
