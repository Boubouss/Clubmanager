package main

import (
	"clubmanager/internal/adapters/api/grpc/server"
	"clubmanager/internal/adapters/auth"
	"clubmanager/internal/adapters/db/postgres"
	"clubmanager/internal/app/middlewares"
	"clubmanager/internal/app/services"
	"clubmanager/internal/config"

	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func getServices(db *pgxpool.Pool, tkm services.TokenManager) *server.ClubManagerServices {
	userConfig := services.UserServiceConfig{
		Repository:   postgres.NewUserRepository(db),
		Hasher:       auth.NewBcryptHasher(),
		TokenManager: tkm,
	}
	usvc := middlewares.NewUserLoggingService(services.NewUserService(userConfig))

	clubConfig := services.ClubServiceConfig{
		Repository: postgres.NewClubRepository(db),
	}
	csvc := middlewares.NewClubLoggingService(services.NewClubService(clubConfig))

	memberConfig := services.MemberServiceConfig{
		Repository: postgres.NewMemberRepository(db),
	}
	msvc := middlewares.NewMemberLoggingService(services.NewMemberService(memberConfig))

	licenceConfig := services.LicenceServiceConfig{
		Repository: postgres.NewLicenceRepository(db),
	}
	lsvc := middlewares.NewLicenceLoggingService(services.NewLicenceService(licenceConfig))

	roleRepo := postgres.NewRoleRepository(db)
	roleConfig := services.RoleServiceConfig{
		Repository:  roleRepo,
		RoleChecker: roleRepo,
	}
	rsvc := middlewares.NewRoleLoggingService(services.NewRoleService(roleConfig))

	eventConfig := services.EventServiceConfig{
		EventRepo:       postgres.NewEventRepository(db),
		CategoryRepo:    postgres.NewEventCategoryRepository(db),
		ParticipantRepo: postgres.NewEventParticipantRepository(db),
		CarpoolRepo:     postgres.NewCarpoolRepository(db),
		JudoCatRepo:     postgres.NewJudoCategoryRepository(db),
		MemberRepo:      postgres.NewMemberRepository(db),
		LicenceRepo:     postgres.NewLicenceRepository(db),
		RoleChecker:     roleRepo,
	}
	esvc := middlewares.NewEventLoggingService(services.NewEventService(eventConfig))

	return &server.ClubManagerServices{
		UserService:    usvc,
		ClubService:    csvc,
		MemberService:  msvc,
		LicenceService: lsvc,
		RoleService:    rsvc,
		EventService:   esvc,
	}
}

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
			// Verify the connection is actually alive.
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

	svc := getServices(db, tkm)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	fmt.Printf("Starting gRPC server on %s...\n", cfg.GRPCPort)
	if err := server.MakeServerAndRun(ctx, cfg.GRPCPort, svc, tkm); err != nil {
		fmt.Println("Server error:", err)
		os.Exit(1)
	}
	fmt.Println("Server stopped gracefully.")
}
