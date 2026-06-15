package main

import (
	"clubmanager/internal/adapters/api/grpc/server"
	"clubmanager/internal/adapters/auth"
	"clubmanager/internal/app/bootstrap"
	"clubmanager/internal/config"
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func connectWithRetry(cfg *config.Config) (*pgxpool.Pool, error) {
	var (
		db  *pgxpool.Pool
		err error
	)
	for i := 0; i < cfg.DBMaxRetries; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		db, err = pgxpool.New(ctx, cfg.DBURL)
		cancel()
		if err == nil {
			pingCtx, pingCancel := context.WithTimeout(context.Background(), 2*time.Second)
			err = db.Ping(pingCtx)
			pingCancel()
			if err == nil {
				return db, nil
			}
			db.Close()
		}
		fmt.Printf("Database not ready (attempt %d/%d): %v\n", i+1, cfg.DBMaxRetries, err)
		if i < cfg.DBMaxRetries-1 {
			time.Sleep(time.Duration(i+1) * 2 * time.Second)
		}
	}
	return nil, fmt.Errorf("could not connect to database after %d attempts: %w", cfg.DBMaxRetries, err)
}

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Println("Configuration error:", err)
		os.Exit(1)
	}

	fmt.Println("Connecting to database...")
	db, err := connectWithRetry(cfg)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer db.Close()
	fmt.Println("Database connected.")

	tkm, err := auth.NewJwtTokenManager(cfg.JWTSecret, cfg.JWTTTLDays)
	if err != nil {
		fmt.Println("JWT configuration error:", err)
		os.Exit(1)
	}

	svc := bootstrap.Build(db, tkm)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	fmt.Printf("Starting gRPC server on %s...\n", cfg.GRPCPort)
	if err := server.MakeServerAndRun(ctx, cfg.GRPCPort, svc, tkm); err != nil {
		fmt.Println("Server error:", err)
		os.Exit(1)
	}
	fmt.Println("Server stopped gracefully.")
}
