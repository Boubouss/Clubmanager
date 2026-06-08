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

var (
	clubUUID = uuid.MustParse("00000000-0000-0000-0000-000000000002")
	testClub = &clubs.Club{
		Id: clubUUID, Siren: "123456789", Name: "Judo Club Paris",
		Address: "1 rue de la Paix", City: "Paris", PostalCode: "75001", Country: "FRANCE",
	}
)

func newTestClubService(repo domain.Repository[clubs.Club, string]) *clubService {
	return NewClubService(ClubServiceConfig{Repository: repo})
}

func TestCreateClub(t *testing.T) {
	tests := []struct {
		name       string
		req        *dto.CreateClubRequest
		searchResp []*clubs.Club
		searchErr  error
		wantErrors bool
		wantErr    bool
	}{
		{
			name:       "Valid club",
			req:        &dto.CreateClubRequest{Siren: "123456789", Name: "Judo Club Paris", Address: "1 rue de la Paix", City: "Paris", PostalCode: "75001"},
			searchResp: []*clubs.Club{},
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

			svc := newTestClubService(repo)
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
