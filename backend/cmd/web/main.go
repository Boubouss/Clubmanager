package main

import (
	"clubmanager/internal/adapters/api/http/handlers"
	httpmw "clubmanager/internal/adapters/api/http/middleware"
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
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
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

	e := echo.New()
	e.HTTPErrorHandler = handlers.HTTPErrorHandler

	// CSRF — reads token from form field "_csrf", stores it in context key "csrf"
	csrfMW := middleware.CSRFWithConfig(middleware.CSRFConfig{
		TokenLookup: "form:_csrf",
		ContextKey:  "csrf",
		CookieName:  "_csrf",
	})

	// Static files
	e.Static("/public", "public")

	// Public routes (with CSRF)
	homeHandler := handlers.NewHomeHandler(svc.MemberService, svc.PostService)
	userHandler := handlers.NewUserHandler(svc.UserService, cfg.AppEnv == "production")

	e.GET("/", homeHandler.HandleLandingPage)
	e.GET("/user/connexion", userHandler.HandleLoginPage, csrfMW)
	e.POST("/user/connexion", userHandler.HandleLoginUser, csrfMW)
	e.GET("/user/inscription", userHandler.HandleRegisterPage, csrfMW)
	e.POST("/user/inscription", userHandler.HandleRegisterUser, csrfMW)
	e.POST("/user/deconnexion", userHandler.HandleLogout, csrfMW)

	// Protected routes
	protected := e.Group("")
	protected.Use(httpmw.RequireAuth(tkm))
	protected.GET("/home", homeHandler.HandleHomePage)

	clubHandler := handlers.NewClubHandler(svc.ClubService, svc.MemberService, svc.RoleService, svc.EventService, svc.PostService)
	protected.GET("/clubs", clubHandler.HandleClubList)
	protected.GET("/clubs/create", clubHandler.HandleCreateClubPage, csrfMW)
	protected.GET("/clubs/create/member-fields", clubHandler.HandleMemberFieldsFragment)
	protected.POST("/clubs/create", clubHandler.HandleCreateClub, csrfMW)
	protected.GET("/clubs/:id", clubHandler.HandleClubDetail)

	memberHandler := handlers.NewMemberHandler(svc.MemberService, svc.ClubService, svc.RoleService, svc.RoleChecker)
	protected.GET("/members", memberHandler.HandleMyMembersPage)
	protected.GET("/members/new", memberHandler.HandleCreateMemberPage, csrfMW)
	protected.POST("/members/new", memberHandler.HandleCreateMember, csrfMW)
	protected.GET("/members/:id/edit", memberHandler.HandleEditMemberPage, csrfMW)
	protected.POST("/members/:id/edit", memberHandler.HandleEditMember, csrfMW)
	protected.POST("/members/:id/delete", memberHandler.HandleDeleteMember, csrfMW)
	protected.GET("/clubs/:id/members", memberHandler.HandleMemberList, csrfMW)
	protected.GET("/clubs/:id/join", memberHandler.HandleRequestMembershipPage, csrfMW)
	protected.POST("/clubs/:id/join", memberHandler.HandleRequestMembership, csrfMW)
	protected.GET("/clubs/:id/members/add", memberHandler.HandleMemberAddPage, csrfMW)
	protected.POST("/clubs/:id/members/add", memberHandler.HandleMemberAdd, csrfMW)
	protected.POST("/clubs/:id/members/:membershipId/validate", memberHandler.HandleValidateMember, csrfMW)
	protected.POST("/clubs/:id/members/:membershipId/refuse", memberHandler.HandleRefuseMember, csrfMW)
	protected.POST("/clubs/:id/members/:membershipId/remove", memberHandler.HandleRemoveMember, csrfMW)
	protected.POST("/clubs/:id/roles/assign", memberHandler.HandleAssignRole, csrfMW)
	protected.POST("/clubs/:id/roles/:roleId/remove", memberHandler.HandleRemoveRole, csrfMW)

	postHandler := handlers.NewPostHandler(svc.PostService, svc.ClubService, svc.RoleChecker)
	protected.GET("/clubs/:id/posts", postHandler.HandlePostList)
	protected.GET("/clubs/:id/posts/new", postHandler.HandleNewPostPage, csrfMW)
	protected.POST("/clubs/:id/posts", postHandler.HandleCreatePost, csrfMW)
	protected.GET("/posts/:id", postHandler.HandlePostDetail)
	protected.GET("/posts/:id/edit", postHandler.HandleEditPostPage, csrfMW)
	protected.POST("/posts/:id", postHandler.HandleEditPost, csrfMW)
	protected.POST("/posts/:id/publish", postHandler.HandlePublishPost, csrfMW)
	protected.POST("/posts/:id/unpublish", postHandler.HandleUnpublishPost, csrfMW)
	protected.POST("/posts/:id/delete", postHandler.HandleDeletePost, csrfMW)

	eventHandler := handlers.NewEventHandler(svc.EventService, svc.MemberService, svc.ClubService, svc.RoleChecker, svc.PostService)
	protected.GET("/clubs/:id/events", eventHandler.HandleEventList)
	protected.GET("/clubs/:id/events/create", eventHandler.HandleCreateEventPage, csrfMW)
	protected.POST("/clubs/:id/events/create", eventHandler.HandleCreateEvent, csrfMW)
	protected.GET("/events/:id", eventHandler.HandleEventDetail)
	protected.POST("/events/:id/open", eventHandler.HandleOpenEvent, csrfMW)
	protected.POST("/events/:id/cancel", eventHandler.HandleCancelEvent, csrfMW)
	protected.POST("/events/:id/delete", eventHandler.HandleDeleteEvent, csrfMW)
	protected.POST("/events/:id/reopen", eventHandler.HandleReopenEvent, csrfMW)
	protected.GET("/events/:id/edit", eventHandler.HandleEditEventPage, csrfMW)
	protected.POST("/events/:id/edit", eventHandler.HandleEditEvent, csrfMW)
	protected.GET("/events/:id/join", eventHandler.HandleJoinEventPage, csrfMW)
	protected.POST("/events/:id/join", eventHandler.HandleJoinEvent, csrfMW)
	protected.POST("/events/:id/leave", eventHandler.HandleLeaveEvent, csrfMW)
	protected.POST("/events/:id/remove-participant", eventHandler.HandleRemoveParticipant, csrfMW)
	protected.GET("/events/:id/participants", eventHandler.HandleAgeGroupParticipants)
	protected.GET("/events/:id/carpools/new", eventHandler.HandleCarpoolOfferPage, csrfMW)
	protected.POST("/events/:id/carpools", eventHandler.HandleCreateCarpoolOffer, csrfMW)
	protected.POST("/events/:id/carpools/:offerId/join", eventHandler.HandleJoinCarpoolOffer, csrfMW)
	protected.POST("/carpools/:offerId/leave", eventHandler.HandleLeaveCarpoolOffer, csrfMW)
	protected.POST("/carpools/:offerId/cancel", eventHandler.HandleCancelCarpoolOffer, csrfMW)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	sc := echo.StartConfig{
		Address:    cfg.HTTPPort,
		HideBanner: true,
	}

	fmt.Printf("Starting HTTP server on %s...\n", cfg.HTTPPort)
	if err := sc.Start(ctx, e); err != nil {
		fmt.Println("HTTP server error:", err)
		os.Exit(1)
	}
	fmt.Println("HTTP server stopped gracefully.")
}
