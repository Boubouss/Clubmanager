package handler

import (
	"clubmanager/api/grpc/client"
	"clubmanager/api/grpc/proto"
	"fmt"

	"clubmanager/internal/views/components"
	"clubmanager/internal/views/pages"
	"net/http"
	"time"

	"github.com/labstack/echo/v5"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type UserHandler struct {
  client proto.ClubManagerServiceClient
  conn *grpc.ClientConn
}

func NewUserHandler() UserHandler {
  // Ajouter la target !
  client, conn, err := client.NewClubManagerServiceClient("backend:50051")
  if err != nil {
    panic(err)
  } 
  return UserHandler{
    client: client,
    conn: conn,
  }
}

func (h *UserHandler) CloseClient() {
  h.conn.Close()
}

func (h *UserHandler) HandleConnexion(c *echo.Context) error {
  return render(c, pages.Connexion())
}

func (h *UserHandler) HandleRegisterForm(c *echo.Context) error {
  return render(c, components.RegisterForm())
}

func (h *UserHandler) HandleRegisterUser(c *echo.Context) error {
  // Add metadata for sql logs
  md := metadata.New(map[string]string{
    "client-ip": c.RealIP(),
    "user-agent": c.Request().UserAgent(),
  })

  fmt.Println(c.RealIP())
  fmt.Println(c.Request().UserAgent())

  // gRPC call 
  res, err := h.client.CreateUser(metadata.NewOutgoingContext(c.Request().Context(), md), &proto.CreateUserRequest{
    Username: c.FormValue("username"),
    Email: c.FormValue("email"),
    Password: c.FormValue("password"),
    Phonenumber: c.FormValue("phonenumber"),
  })

  if err != nil {
    return render(c, pages.Connexion())
  }
  
  // Check validation errors
  if len(res.Errors) > 0 {
    return render(c, pages.Connexion())
  }

  // Add JWT token as cookie 
  c.SetCookie(&http.Cookie{
    Name: "ClubManagerAuth",
    Value: res.Token,
    Expires: time.Now().Add(time.Hour * 24 * 30),
  })

  return render(c, pages.Home(res.Username))
}

func (h *UserHandler) HandleLoginForm(c *echo.Context) error {
  return render(c, components.LoginForm())
}

func (h *UserHandler) HandleLoginUser(c *echo.Context) error {
  return render(c, pages.Home("Red"))
}
