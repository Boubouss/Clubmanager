package services

import (
	"clubmanager/internal/adapters/api/grpc/dto"
	"clubmanager/internal/domain"
	"clubmanager/internal/domain/clubs"
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// ── Mocks ─────────────────────────────────────────────────────────────────────

type mockClubRepository struct {
	saveFunc   func(context.Context, *clubs.Club) (*clubs.Club, error)
	findFunc   func(context.Context, string) (*clubs.Club, error)
	searchFunc func(context.Context, *domain.SearchParams) ([]*clubs.Club, error)
	deleteFunc func(context.Context, string) (bool, error)
}

func (m mockClubRepository) Save(ctx context.Context, c *clubs.Club) (*clubs.Club, error) {
	return m.saveFunc(ctx, c)
}
func (m mockClubRepository) Find(ctx context.Context, id string) (*clubs.Club, error) {
	return m.findFunc(ctx, id)
}
func (m mockClubRepository) Search(ctx context.Context, p *domain.SearchParams) ([]*clubs.Club, error) {
	return m.searchFunc(ctx, p)
}
func (m mockClubRepository) Delete(ctx context.Context, id string) (bool, error) {
	return m.deleteFunc(ctx, id)
}

type mockClubStatusRepo struct {
	updateStatusFunc func(context.Context, string, string) error
}

func (m mockClubStatusRepo) UpdateStatus(ctx context.Context, clubId, status string) error {
	return m.updateStatusFunc(ctx, clubId, status)
}

type mockRoleCheckerForClub struct {
	isSuperAdminFunc func(context.Context, string) (bool, error)
	hasRoleFunc      func(context.Context, string, string, ...string) (bool, error)
}

func (m mockRoleCheckerForClub) HasRole(ctx context.Context, userId, clubId string, roles ...string) (bool, error) {
	if m.hasRoleFunc != nil {
		return m.hasRoleFunc(ctx, userId, clubId, roles...)
	}
	return false, nil
}
func (m mockRoleCheckerForClub) IsSuperAdmin(ctx context.Context, userId string) (bool, error) {
	return m.isSuperAdminFunc(ctx, userId)
}

type mockSireneVerifier struct {
	verifyFunc func(context.Context, string) error
}

func (m mockSireneVerifier) Verify(ctx context.Context, siren string) error {
	return m.verifyFunc(ctx, siren)
}

// mockClubStatusChecker is used by member/event/role service tests.
type mockClubStatusChecker struct {
	isActiveFunc func(context.Context, string) (bool, error)
}

func (m mockClubStatusChecker) IsActive(ctx context.Context, clubId string) (bool, error) {
	return m.isActiveFunc(ctx, clubId)
}

// ── Fixtures ──────────────────────────────────────────────────────────────────

var (
	clubUUID = uuid.MustParse("00000000-0000-0000-0000-000000000002")
	testClub = &clubs.Club{
		Id: clubUUID, Siren: "123456789", Name: "Judo Club Paris",
		Address: "1 rue de la Paix", City: "Paris", PostalCode: "75001", Country: "FRANCE",
		Status: clubs.StatusPending,
	}
)

func newTestClubService(repo domain.Repository[clubs.Club, string]) *clubService {
	return NewClubService(ClubServiceConfig{Repository: repo})
}

func newFullTestClubService(
	repo domain.Repository[clubs.Club, string],
	statusRepo ClubStatusRepository,
	checker mockRoleCheckerForClub,
	verifier mockSireneVerifier,
) *clubService {
	return NewClubService(ClubServiceConfig{
		Repository:     repo,
		StatusRepo:     statusRepo,
		RoleChecker:    checker,
		SireneVerifier: verifier,
	})
}

// ── Tests ─────────────────────────────────────────────────────────────────────

func TestCreateClub(t *testing.T) {
	tests := []struct {
		name           string
		req            *dto.CreateClubRequest
		searchResp     []*clubs.Club
		searchErr      error
		sireneErr      error
		wantErrors     bool
		wantErr        bool
		wantStatus     string
	}{
		{
			name:       "Valid club — Sirene down → pending",
			req:        &dto.CreateClubRequest{Siren: "123456789", Name: "Judo Club Paris", Address: "1 rue de la Paix", City: "Paris", PostalCode: "75001"},
			searchResp: []*clubs.Club{},
			sireneErr:  errors.New("api unreachable"),
			wantStatus: clubs.StatusPending,
		},
		{
			name:       "Valid club — Sirene OK → active",
			req:        &dto.CreateClubRequest{Siren: "123456789", Name: "Judo Club Paris", Address: "1 rue de la Paix", City: "Paris", PostalCode: "75001"},
			searchResp: []*clubs.Club{},
			sireneErr:  nil,
			wantStatus: clubs.StatusActive,
		},
		{
			name:       "Invalid SIREN",
			req:        &dto.CreateClubRequest{Siren: "bad", Name: "Judo Club", Address: "1 rue", City: "Paris", PostalCode: "75001"},
			wantErrors: true,
		},
		{
			name:       "SIREN already exists",
			req:        &dto.CreateClubRequest{Siren: "123456789", Name: "Judo Club Paris", Address: "1 rue de la Paix", City: "Paris", PostalCode: "75001"},
			searchResp: []*clubs.Club{testClub},
			wantErrors: true,
		},
		{
			name:      "Repository error on search",
			req:       &dto.CreateClubRequest{Siren: "123456789", Name: "Judo Club Paris", Address: "1 rue de la Paix", City: "Paris", PostalCode: "75001"},
			searchErr: errors.New("db error"),
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := mockClubRepository{
				searchFunc: func(_ context.Context, _ *domain.SearchParams) ([]*clubs.Club, error) {
					return tt.searchResp, tt.searchErr
				},
				saveFunc: func(_ context.Context, c *clubs.Club) (*clubs.Club, error) {
					c.Id = clubUUID
					return c, nil
				},
			}
			statusRepo := mockClubStatusRepo{
				updateStatusFunc: func(_ context.Context, _, _ string) error { return nil },
			}
			checker := mockRoleCheckerForClub{
				isSuperAdminFunc: func(_ context.Context, _ string) (bool, error) { return false, nil },
			}
			verifier := mockSireneVerifier{
				verifyFunc: func(_ context.Context, _ string) error { return tt.sireneErr },
			}

			svc := newFullTestClubService(repo, statusRepo, checker, verifier)
			res, err := svc.CreateClub(context.Background(), tt.req)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			if tt.wantErrors {
				assert.True(t, len(res.Errors) > 0)
				assert.Nil(t, res.Club)
			} else {
				assert.Empty(t, res.Errors)
				assert.NotNil(t, res.Club)
				if tt.wantStatus != "" {
					assert.Equal(t, tt.wantStatus, res.Club.Status)
				}
			}
		})
	}
}

func TestValidateClub(t *testing.T) {
	superAdminCtx := context.WithValue(context.Background(), "user_id", "super-user-id")
	normalCtx := context.WithValue(context.Background(), "user_id", "normal-user-id")

	tests := []struct {
		name           string
		ctx            context.Context
		club           *clubs.Club
		isSuperAdmin   bool
		updateErr      error
		wantErrors     bool
		wantErr        bool
	}{
		{
			name:         "Superadmin validates pending club",
			ctx:          superAdminCtx,
			club:         testClub, // StatusPending
			isSuperAdmin: true,
		},
		{
			name:         "Non-superadmin rejected",
			ctx:          normalCtx,
			club:         testClub,
			isSuperAdmin: false,
			wantErrors:   true,
		},
		{
			name:         "Club not found",
			ctx:          superAdminCtx,
			club:         nil,
			isSuperAdmin: true,
			wantErrors:   true,
		},
		{
			name:         "Club already active",
			ctx:          superAdminCtx,
			club:         &clubs.Club{Id: clubUUID, Status: clubs.StatusActive},
			isSuperAdmin: true,
			wantErrors:   true,
		},
		{
			name:         "UpdateStatus error",
			ctx:          superAdminCtx,
			club:         testClub,
			isSuperAdmin: true,
			updateErr:    errors.New("db error"),
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture a copy to avoid mutation across test cases.
			club := tt.club
			var clubCopy *clubs.Club
			if club != nil {
				c := *club
				clubCopy = &c
			}
			repo := mockClubRepository{
				findFunc: func(_ context.Context, _ string) (*clubs.Club, error) {
					return clubCopy, nil
				},
			}
			statusRepo := mockClubStatusRepo{
				updateStatusFunc: func(_ context.Context, _, _ string) error { return tt.updateErr },
			}
			checker := mockRoleCheckerForClub{
				isSuperAdminFunc: func(_ context.Context, _ string) (bool, error) {
					return tt.isSuperAdmin, nil
				},
			}
			verifier := mockSireneVerifier{
				verifyFunc: func(_ context.Context, _ string) error { return nil },
			}

			svc := newFullTestClubService(repo, statusRepo, checker, verifier)
			res, err := svc.ValidateClub(tt.ctx, &dto.ValidateClubRequest{ClubId: clubUUID.String()})

			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			if tt.wantErrors {
				assert.True(t, len(res.Errors) > 0)
			} else {
				assert.Empty(t, res.Errors)
				assert.NotNil(t, res.Club)
				assert.Equal(t, clubs.StatusActive, res.Club.Status)
			}
		})
	}
}

func TestSuspendClub(t *testing.T) {
	superAdminCtx := context.WithValue(context.Background(), "user_id", "super-user-id")
	normalCtx := context.WithValue(context.Background(), "user_id", "normal-user-id")

	tests := []struct {
		name         string
		ctx          context.Context
		club         *clubs.Club
		isSuperAdmin bool
		updateErr    error
		wantErrors   bool
		wantErr      bool
	}{
		{
			name:         "Superadmin suspends active club",
			ctx:          superAdminCtx,
			club:         &clubs.Club{Id: clubUUID, Status: clubs.StatusActive},
			isSuperAdmin: true,
		},
		{
			name:         "Non-superadmin rejected",
			ctx:          normalCtx,
			club:         &clubs.Club{Id: clubUUID, Status: clubs.StatusActive},
			isSuperAdmin: false,
			wantErrors:   true,
		},
		{
			name:         "Club not found",
			ctx:          superAdminCtx,
			club:         nil,
			isSuperAdmin: true,
			wantErrors:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			club := tt.club
			var clubCopy *clubs.Club
			if club != nil {
				c := *club
				clubCopy = &c
			}
			repo := mockClubRepository{
				findFunc: func(_ context.Context, _ string) (*clubs.Club, error) {
					return clubCopy, nil
				},
			}
			statusRepo := mockClubStatusRepo{
				updateStatusFunc: func(_ context.Context, _, _ string) error { return tt.updateErr },
			}
			checker := mockRoleCheckerForClub{
				isSuperAdminFunc: func(_ context.Context, _ string) (bool, error) {
					return tt.isSuperAdmin, nil
				},
			}
			verifier := mockSireneVerifier{
				verifyFunc: func(_ context.Context, _ string) error { return nil },
			}

			svc := newFullTestClubService(repo, statusRepo, checker, verifier)
			res, err := svc.SuspendClub(tt.ctx, &dto.SuspendClubRequest{ClubId: clubUUID.String()})

			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			if tt.wantErrors {
				assert.True(t, len(res.Errors) > 0)
			} else {
				assert.Empty(t, res.Errors)
				assert.NotNil(t, res.Club)
				assert.Equal(t, clubs.StatusSuspended, res.Club.Status)
			}
		})
	}
}

func TestUpdateClub(t *testing.T) {
	tests := []struct {
		name       string
		req        *dto.UpdateClubRequest
		findResp   *clubs.Club
		findErr    error
		wantErrors bool
		wantErr    bool
	}{
		{
			name:     "Valid update",
			req:      &dto.UpdateClubRequest{Id: clubUUID.String(), Name: "New Name"},
			findResp: testClub,
		},
		{
			name:       "Invalid postal code",
			req:        &dto.UpdateClubRequest{Id: clubUUID.String(), PostalCode: "bad"},
			findResp:   testClub,
			wantErrors: true,
		},
		{
			name:    "Club not found",
			req:     &dto.UpdateClubRequest{Id: clubUUID.String()},
			findErr: errors.New("not found"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := mockClubRepository{
				findFunc: func(_ context.Context, _ string) (*clubs.Club, error) {
					return tt.findResp, tt.findErr
				},
				saveFunc: func(_ context.Context, c *clubs.Club) (*clubs.Club, error) {
					return c, nil
				},
			}

			svc := newTestClubService(repo)
			res, err := svc.UpdateClub(context.Background(), tt.req)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			if tt.wantErrors {
				assert.True(t, len(res.Errors) > 0)
			} else {
				assert.Empty(t, res.Errors)
				assert.NotNil(t, res.Club)
			}
		})
	}
}

func TestDeleteClub(t *testing.T) {
	tests := []struct {
		name       string
		deleteResp bool
		deleteErr  error
		wantOk     bool
		wantErr    bool
	}{
		{"Valid delete", true, nil, true, false},
		{"Repository error", false, errors.New("db error"), false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := mockClubRepository{
				deleteFunc: func(_ context.Context, _ string) (bool, error) {
					return tt.deleteResp, tt.deleteErr
				},
			}

			svc := newTestClubService(repo)
			ok, err := svc.DeleteClub(context.Background(), clubUUID.String())

			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.wantOk, ok)
		})
	}
}
