package main

import (
	"clubmanager/internal/handler"

	"github.com/labstack/echo/v5"
)

func main() {
  app := echo.New()

  homeHandler := handler.NewHomeHandler()
  userHandler := handler.NewUserHandler()

  defer userHandler.CloseClient()

  app.GET("/", homeHandler.HandleLandingPage)
  app.GET("/home", homeHandler.HandleHomePage)

  auth := app.Group("/user")

  auth.GET("/connexion", userHandler.HandleConnexion)
  auth.GET("/register", userHandler.HandleRegisterForm)
  auth.GET("/login", userHandler.HandleLoginForm)
  auth.POST("/login", userHandler.HandleLoginUser)
  auth.POST("/register", userHandler.HandleRegisterUser)
  
  if err := app.Start(":8080"); err != nil {
    app.Logger.Error("Failed to start server", "error", err)
  }
}
