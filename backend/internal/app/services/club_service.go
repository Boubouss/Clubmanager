package services

import (
	"clubmanager/internal/adapters/api/grpc/dto"
	"clubmanager/internal/adapters/external/sirene"
	"clubmanager/internal/domain"
	"clubmanager/internal/domain/clubs"
	"clubmanager/internal/domain/roles"
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// ClubStatusRepository extends the base repository with status-specific mutations.
type ClubStatusRepository interface {
	UpdateStatus(ctx context.Context, clubId, status string) error
}

type ClubService interface {
	CreateClub(context.Context, *dto.CreateClubRequest) (*dto.CreateClubResponse, error)
	GetClub(context.Context, string) (*dto.GetClubResponse, error)
	ReadClub(context.Context, *dto.ReadClubRequest) (*dto.ReadClubResponse, error)
	UpdateClub(context.Context, *dto.UpdateClubRequest) (*dto.UpdateClubResponse, error)
	DeleteClub(context.Context, string) (bool, error)
	ValidateClub(context.Context, *dto.ValidateClubRequest) (*dto.ValidateClubResponse, error)
	SuspendClub(context.Context, *dto.SuspendClubRequest) (*dto.SuspendClubResponse, error)
}

type ClubServiceConfig struct {
	Repository     domain.Repository[clubs.Club, string]
	StatusRepo     ClubStatusRepository
	RoleChecker    roles.RoleChecker
	SireneVerifier sirene.SireneVerifier
}

type clubService struct {
	repo           domain.Repository[clubs.Club, string]
	statusRepo     ClubStatusRepository
	roleChecker    roles.RoleChecker
	sireneVerifier sirene.SireneVerifier
}

func NewClubService(config ClubServiceConfig) *clubService {
	return &clubService{
		repo:           config.Repository,
		statusRepo:     config.StatusRepo,
		roleChecker:    config.RoleChecker,
		sireneVerifier: config.SireneVerifier,
	}
}

func (s *clubService) CreateClub(ctx context.Context, data *dto.CreateClubRequest) (*dto.CreateClubResponse, error) {
	c, errs := clubs.NewClub(data.Map())
	if len(errs) > 0 {
		return &dto.CreateClubResponse{Errors: errs}, nil
	}

	list, err := s.repo.Search(ctx, &domain.SearchParams{
		Fields:    map[string]any{"siren": c.Siren},
		Keys:      map[string]bool{"siren": true},
		Connector: "AND",
	})
	if err != nil {
		return nil, err
	}
	if len(list) > 0 {
		return &dto.CreateClubResponse{
			Errors: map[string]string{"siren": "A club with this SIREN already exists."},
		}, nil
	}

	// Sirene verification: soft failure — club is created as pending regardless.
	// If the API confirms it's a valid sporting association, activate immediately.
	if s.sireneVerifier != nil {
		verifyCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
		defer cancel()
		if err := s.sireneVerifier.Verify(verifyCtx, c.Siren); err == nil {
			c.Status = clubs.StatusActive
		}
		// If Verify returned an error (API down or invalid SIREN), keep StatusPending.
	}

	if creatorId, err := uuid.Parse(data.UserId); err == nil {
		c.CreatorId = creatorId
	}

	created, err := s.repo.Save(ctx, c)
	if err != nil {
		return nil, err
	}
	return &dto.CreateClubResponse{Club: created, Errors: make(map[string]string)}, nil
}

func (s *clubService) GetClub(ctx context.Context, id string) (*dto.GetClubResponse, error) {
	club, err := s.repo.Find(ctx, id)
	if err != nil {
		return nil, err
	}
	return &dto.GetClubResponse{Club: club, Errors: make(map[string]string)}, nil
}

func (s *clubService) ReadClub(ctx context.Context, data *dto.ReadClubRequest) (*dto.ReadClubResponse, error) {
	list, err := s.repo.Search(ctx, &domain.SearchParams{
		Fields:    data.Params,
		Keys:      map[string]bool{"siren": true, "name": true, "city": true, "postal_code": true},
		Connector: "AND",
	})
	if err != nil {
		return nil, err
	}
	return &dto.ReadClubResponse{Clubs: list, Errors: make(map[string]string)}, nil
}

func (s *clubService) UpdateClub(ctx context.Context, data *dto.UpdateClubRequest) (*dto.UpdateClubResponse, error) {
	current, err := s.repo.Find(ctx, data.Id)
	if err != nil {
		return nil, err
	}

	c, errs := current.Update(data.Map())
	if len(errs) > 0 {
		return &dto.UpdateClubResponse{Errors: errs}, nil
	}

	updated, err := s.repo.Save(ctx, c)
	if err != nil {
		return nil, err
	}
	return &dto.UpdateClubResponse{Club: updated, Errors: make(map[string]string)}, nil
}

func (s *clubService) DeleteClub(ctx context.Context, id string) (bool, error) {
	return s.repo.Delete(ctx, id)
}

// ValidateClub sets a pending club to active. Only superadmins may call this.
func (s *clubService) ValidateClub(ctx context.Context, data *dto.ValidateClubRequest) (*dto.ValidateClubResponse, error) {
	userId, ok := ctx.Value("user_id").(string)
	if !ok || userId == "" {
		return &dto.ValidateClubResponse{Errors: map[string]string{"auth": "Missing user context."}}, nil
	}

	isSuperAdmin, err := s.roleChecker.IsSuperAdmin(ctx, userId)
	if err != nil {
		return nil, err
	}
	if !isSuperAdmin {
		return &dto.ValidateClubResponse{Errors: map[string]string{"auth": "Only superadmins can validate clubs."}}, nil
	}

	club, err := s.repo.Find(ctx, data.ClubId)
	if err != nil {
		return nil, err
	}
	if club == nil {
		return &dto.ValidateClubResponse{Errors: map[string]string{"club": "Club not found."}}, nil
	}
	if club.Status != clubs.StatusPending {
		return &dto.ValidateClubResponse{
			Errors: map[string]string{"club": fmt.Sprintf("Club is not pending (current status: %s).", club.Status)},
		}, nil
	}

	if err := s.statusRepo.UpdateStatus(ctx, data.ClubId, clubs.StatusActive); err != nil {
		return nil, err
	}
	club.Status = clubs.StatusActive
	return &dto.ValidateClubResponse{Club: club, Errors: make(map[string]string)}, nil
}

// SuspendClub sets a club to suspended. Only superadmins may call this.
func (s *clubService) SuspendClub(ctx context.Context, data *dto.SuspendClubRequest) (*dto.SuspendClubResponse, error) {
	userId, ok := ctx.Value("user_id").(string)
	if !ok || userId == "" {
		return &dto.SuspendClubResponse{Errors: map[string]string{"auth": "Missing user context."}}, nil
	}

	isSuperAdmin, err := s.roleChecker.IsSuperAdmin(ctx, userId)
	if err != nil {
		return nil, err
	}
	if !isSuperAdmin {
		return &dto.SuspendClubResponse{Errors: map[string]string{"auth": "Only superadmins can suspend clubs."}}, nil
	}

	club, err := s.repo.Find(ctx, data.ClubId)
	if err != nil {
		return nil, err
	}
	if club == nil {
		return &dto.SuspendClubResponse{Errors: map[string]string{"club": "Club not found."}}, nil
	}

	if err := s.statusRepo.UpdateStatus(ctx, data.ClubId, clubs.StatusSuspended); err != nil {
		return nil, err
	}
	club.Status = clubs.StatusSuspended
	return &dto.SuspendClubResponse{Club: club, Errors: make(map[string]string)}, nil
}
