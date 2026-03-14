package main

import (
	"clubmanager/api/grpc/server"
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5"
)


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
  if err := server.MakeServerAndRun(*port, db); err != nil {
    fmt.Println("Server failed to start")
  }

  fmt.Println("Stop server...")
}
