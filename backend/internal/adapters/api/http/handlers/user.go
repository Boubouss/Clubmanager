package handlers

import (
	"clubmanager/internal/adapters/api/grpc/dto"
	"clubmanager/internal/adapters/api/http/views/pages"
	"clubmanager/internal/app/services"
	"net/http"
	"time"

	"github.com/labstack/echo/v5"
)

const cookieName = "ClubManagerAuth"

type UserHandler struct {
	userSvc services.UserService
	isProd  bool
}

func NewUserHandler(userSvc services.UserService, isProd bool) *UserHandler {
	return &UserHandler{userSvc: userSvc, isProd: isProd}
}

func (h *UserHandler) HandleLoginPage(c *echo.Context) error {
	return render(c, pages.Connexion(nil, csrf(c)))
}

func (h *UserHandler) HandleLoginUser(c *echo.Context) error {
	req := &dto.LoginUserRequest{
		Email:    c.FormValue("email"),
		Password: c.FormValue("password"),
	}

	resp, err := h.userSvc.LoginUser(c.Request().Context(), req)
	if err != nil {
		return err
	}

	if len(resp.Errors) > 0 {
		return render(c, pages.Connexion(resp.Errors, csrf(c)))
	}

	h.setAuthCookie(c, resp.Token)
	return c.Redirect(http.StatusSeeOther, "/home")
}

func (h *UserHandler) HandleRegisterPage(c *echo.Context) error {
	return render(c, pages.Inscription(nil, csrf(c)))
}

func (h *UserHandler) HandleRegisterUser(c *echo.Context) error {
	req := &dto.CreateUserRequest{
		Username:    c.FormValue("username"),
		Email:       c.FormValue("email"),
		Phonenumber: c.FormValue("phonenumber"),
		Password:    c.FormValue("password"),
	}

	resp, err := h.userSvc.CreateUser(c.Request().Context(), req)
	if err != nil {
		return err
	}

	if len(resp.Errors) > 0 {
		return render(c, pages.Inscription(resp.Errors, csrf(c)))
	}

	h.setAuthCookie(c, resp.Token)
	return c.Redirect(http.StatusSeeOther, "/home")
}

func (h *UserHandler) HandleLogout(c *echo.Context) error {
	cookie := new(http.Cookie)
	cookie.Name = cookieName
	cookie.Value = ""
	cookie.MaxAge = -1
	cookie.Expires = time.Unix(0, 0)
	cookie.Path = "/"
	cookie.HttpOnly = true
	c.SetCookie(cookie)
	return c.Redirect(http.StatusSeeOther, "/")
}

func (h *UserHandler) setAuthCookie(c *echo.Context, token string) {
	cookie := new(http.Cookie)
	cookie.Name = cookieName
	cookie.Value = token
	cookie.Expires = time.Now().Add(30 * 24 * time.Hour)
	cookie.Path = "/"
	cookie.HttpOnly = true
	cookie.SameSite = http.SameSiteLaxMode
	if h.isProd {
		cookie.Secure = true
	}
	c.SetCookie(cookie)
}

// csrf extracts the CSRF token from the Echo context (set by the CSRF middleware).
func csrf(c *echo.Context) string {
	token, _ := c.Get("csrf").(string)
	return token
}
