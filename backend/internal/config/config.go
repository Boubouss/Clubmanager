package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	GRPCPort     string
	HTTPPort     string
	DBURL        string
	JWTSecret    string
	JWTTTLDays   int
	DBMaxRetries int
}

func Load() (*Config, error) {
	port := os.Getenv("GRPC_PORT")
	if port == "" {
		port = ":50051"
	}

	httpPort := os.Getenv("HTTP_PORT")
	if httpPort == "" {
		httpPort = ":8080"
	}

	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		return nil, fmt.Errorf("DB_URL environment variable is required")
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if len(jwtSecret) < 32 {
		return nil, fmt.Errorf("JWT_SECRET must be at least 32 characters (got %d)", len(jwtSecret))
	}

	jwtTTLDays := 30
	if v := os.Getenv("JWT_TTL_DAYS"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n <= 0 {
			return nil, fmt.Errorf("JWT_TTL_DAYS must be a positive integer")
		}
		jwtTTLDays = n
	}

	dbMaxRetries := 5
	if v := os.Getenv("DB_MAX_RETRIES"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n <= 0 {
			return nil, fmt.Errorf("DB_MAX_RETRIES must be a positive integer")
		}
		dbMaxRetries = n
	}

	return &Config{
		GRPCPort:     port,
		HTTPPort:     httpPort,
		DBURL:        dbURL,
		JWTSecret:    jwtSecret,
		JWTTTLDays:   jwtTTLDays,
		DBMaxRetries: dbMaxRetries,
	}, nil
}
