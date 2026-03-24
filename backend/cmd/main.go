package main

import (
	"clubmanager/internal/adapters/api/grpc/server"
	"clubmanager/internal/adapters/auth"
	"clubmanager/internal/adapters/db/postgres"
	"clubmanager/internal/app/middlewares"
	"clubmanager/internal/app/services"

	"context"
	"flag"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5"
)

func getServices(db *pgx.Conn) *server.ClubManagerServices {
  userConfig := services.UserServiceConfig{
    Repository: postgres.NewUserRepository(db),
    Hasher: auth.NewBcryptHasher(),
    TokenManager: auth.NewJwtTokenManager(),
  }
  usvc := middlewares.NewUserLoggingService(services.NewUserService(userConfig))

  svc := server.ClubManagerServices{
    UserService: usvc,
  }
  
  return &svc
}

func main() {
  port := flag.String("port", ":50051", "gRPC port for user service")
  flag.Parse()

  ctx := context.Background()

  fmt.Println("Connection to database...")
  db, err := pgx.Connect(ctx, os.Getenv("DB_URL"))
  if err != nil {
    fmt.Println("Connection to database failed")
    fmt.Println(os.Getenv("DB_URL"))
    return
  }

  fmt.Println("Start server...")
  svc := getServices(db)

  if err := server.MakeServerAndRun(*port, svc); err != nil {
    fmt.Println("Server failed to start")
  }
}
