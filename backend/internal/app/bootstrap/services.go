package bootstrap

import (
	"clubmanager/internal/adapters/auth"
	"clubmanager/internal/adapters/db/postgres"
	"clubmanager/internal/adapters/external/sirene"
	"clubmanager/internal/app/middlewares"
	"clubmanager/internal/app/services"
	"clubmanager/internal/domain/roles"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Services holds all application services shared between the gRPC and HTTP servers.
type Services struct {
	UserService    services.UserService
	ClubService    services.ClubService
	MemberService  services.MemberService
	LicenceService services.LicenceService
	RoleService    services.RoleService
	EventService   services.EventService
	PostService    services.PostService
	RoleChecker    roles.RoleChecker
}

func Build(db *pgxpool.Pool, tkm services.TokenManager) *Services {
	userConfig := services.UserServiceConfig{
		Repository:   postgres.NewUserRepository(db),
		Hasher:       auth.NewBcryptHasher(),
		TokenManager: tkm,
	}
	usvc := middlewares.NewUserLoggingService(services.NewUserService(userConfig))

	clubRepo := postgres.NewClubRepository(db)
	clubConfig := services.ClubServiceConfig{
		Repository:     clubRepo,
		StatusRepo:     clubRepo,
		RoleChecker:    postgres.NewRoleRepository(db),
		SireneVerifier: sirene.NewHTTPVerifier(),
	}
	csvc := middlewares.NewClubLoggingService(services.NewClubService(clubConfig))

	membershipRepo := postgres.NewClubMembershipRepository(db)
	memberConfig := services.MemberServiceConfig{
		Repository:           postgres.NewMemberRepository(db),
		MembershipRepository: membershipRepo,
		UserClubsPort:        membershipRepo,
		StaffContactsPort:    postgres.NewRoleRepository(db),
		ClubStatusChecker:    clubRepo,
	}
	msvc := middlewares.NewMemberLoggingService(services.NewMemberService(memberConfig))

	licenceConfig := services.LicenceServiceConfig{
		Repository: postgres.NewLicenceRepository(db),
	}
	lsvc := middlewares.NewLicenceLoggingService(services.NewLicenceService(licenceConfig))

	roleRepo := postgres.NewRoleRepository(db)
	roleConfig := services.RoleServiceConfig{
		Repository:        roleRepo,
		RoleChecker:       roleRepo,
		ClubStatusChecker: clubRepo,
	}
	rsvc := middlewares.NewRoleLoggingService(services.NewRoleService(roleConfig))

	eventRepo := postgres.NewEventRepository(db)
	eventConfig := services.EventServiceConfig{
		EventRepo:         eventRepo,
		EventUpcomingPort: eventRepo,
		EventClubListPort: eventRepo,
		CategoryRepo:      postgres.NewEventCategoryRepository(db),
		ParticipantRepo:   postgres.NewEventParticipantRepository(db),
		CarpoolRepo:       postgres.NewCarpoolRepository(db),
		JudoCatRepo:       postgres.NewJudoCategoryRepository(db),
		MemberRepo:        postgres.NewMemberRepository(db),
		LicenceRepo:       postgres.NewLicenceRepository(db),
		RoleChecker:       roleRepo,
		ClubStatusChecker: clubRepo,
	}
	esvc := middlewares.NewEventLoggingService(services.NewEventService(eventConfig))

	postRepo := postgres.NewPostRepository(db)
	postConfig := services.PostServiceConfig{
		Repository:        postRepo,
		RoleChecker:       roleRepo,
		ClubStatusChecker: clubRepo,
		MembershipChecker: postRepo,
	}
	psvc := middlewares.NewPostLoggingService(services.NewPostService(postConfig))

	return &Services{
		UserService:    usvc,
		ClubService:    csvc,
		MemberService:  msvc,
		LicenceService: lsvc,
		RoleService:    rsvc,
		EventService:   esvc,
		PostService:    psvc,
		RoleChecker:    roleRepo,
	}
}
