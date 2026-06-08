package services

import (
	"clubmanager/internal/adapters/api/grpc/dto"
	"clubmanager/internal/domain"
	"clubmanager/internal/domain/licences"
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

type mockLicenceRepository struct {
	saveFunc   func(context.Context, *licences.Licence) (*licences.Licence, error)
	findFunc   func(context.Context, string) (*licences.Licence, error)
	searchFunc func(context.Context, *domain.SearchParams) ([]*licences.Licence, error)
	deleteFunc func(context.Context, string) (bool, error)
}

func (m mockLicenceRepository) Save(ctx context.Context, l *licences.Licence) (*licences.Licence, error) {
	return m.saveFunc(ctx, l)
}
func (m mockLicenceRepository) Find(ctx context.Context, id string) (*licences.Licence, error) {
	return m.findFunc(ctx, id)
}
func (m mockLicenceRepository) Search(ctx context.Context, p *domain.SearchParams) ([]*licences.Licence, error) {
	return m.searchFunc(ctx, p)
}
func (m mockLicenceRepository) Delete(ctx context.Context, id string) (bool, error) {
	return m.deleteFunc(ctx, id)
}

var (
	licenceUUID  = uuid.MustParse("00000000-0000-0000-0000-000000000030")
	licenceMember = uuid.MustParse("00000000-0000-0000-0000-000000000020")
	testLicence  = &licences.Licence{
		Id:            licenceUUID,
		MemberId:      licenceMember,
		LicenceNumber: "LIC-2026-001",
		ValidFrom:     "2026-01-01",
		ValidUntil:    "2026-12-31",
		Status:        "pending",
	}
)

func newTestLicenceService(repo domain.Repository[licences.Licence, string]) *licenceService {
	return NewLicenceService(LicenceServiceConfig{Repository: repo})
}

func TestCreateLicence(t *testing.T) {
	tests := []struct {
		name       string
		req        *dto.CreateLicenceRequest
		saveErr    error
		wantErrors bool
		wantErr    bool
	}{
		{
			name: "Valid licence",
			req: &dto.CreateLicenceRequest{
				MemberId: licenceMember.String(), LicenceNumber: "LIC-2026-001",
				ValidFrom: "2026-01-01", ValidUntil: "2026-12-31",
			},
		},
		{
			name: "Missing licence number",
			req: &dto.CreateLicenceRequest{
				MemberId: licenceMember.String(),
				ValidFrom: "2026-01-01", ValidUntil: "2026-12-31",
			},
			wantErrors: true,
		},
		{
			name: "ValidUntil before ValidFrom",
			req: &dto.CreateLicenceRequest{
				MemberId: licenceMember.String(), LicenceNumber: "LIC-2026-001",
				ValidFrom: "2026-12-01", ValidUntil: "2026-01-01",
			},
			wantErrors: true,
		},
		{
			name: "Repository error",
			req: &dto.CreateLicenceRequest{
				MemberId: licenceMember.String(), LicenceNumber: "LIC-2026-001",
				ValidFrom: "2026-01-01", ValidUntil: "2026-12-31",
			},
			saveErr: errors.New("db error"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := mockLicenceRepository{
				saveFunc: func(_ context.Context, l *licences.Licence) (*licences.Licence, error) {
					if tt.saveErr != nil {
						return nil, tt.saveErr
					}
					l.Id = licenceUUID
					return l, nil
				},
			}

			svc := newTestLicenceService(repo)
			res, err := svc.CreateLicence(context.Background(), tt.req)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			if tt.wantErrors {
				assert.True(t, len(res.Errors) > 0)
				assert.Nil(t, res.Licence)
			} else {
				assert.Empty(t, res.Errors)
				assert.NotNil(t, res.Licence)
				assert.Equal(t, "pending", res.Licence.Status)
			}
		})
	}
}

func TestUpdateLicenceStatus(t *testing.T) {
	tests := []struct {
		name       string
		req        *dto.UpdateLicenceStatusRequest
		findResp   *licences.Licence
		findErr    error
		saveErr    error
		wantErrors bool
		wantErr    bool
	}{
		{
			name:     "Valid transition pending → active",
			req:      &dto.UpdateLicenceStatusRequest{Id: licenceUUID.String(), Status: "active"},
			findResp: testLicence,
		},
		{
			name:     "Valid transition active → expired",
			req:      &dto.UpdateLicenceStatusRequest{Id: licenceUUID.String(), Status: "expired"},
			findResp: &licences.Licence{Id: licenceUUID, MemberId: licenceMember, LicenceNumber: "L1", ValidFrom: "2026-01-01", ValidUntil: "2026-12-31", Status: "active"},
		},
		{
			name:       "Invalid status",
			req:        &dto.UpdateLicenceStatusRequest{Id: licenceUUID.String(), Status: "invalid"},
			findResp:   testLicence,
			wantErrors: true,
		},
		{
			name:       "Licence not found",
			req:        &dto.UpdateLicenceStatusRequest{Id: licenceUUID.String(), Status: "active"},
			findResp:   nil,
			wantErrors: true,
		},
		{
			name:    "Repository error on Find",
			req:     &dto.UpdateLicenceStatusRequest{Id: licenceUUID.String(), Status: "active"},
			findErr: errors.New("db error"),
			wantErr: true,
		},
		{
			name:     "Repository error on Save",
			req:      &dto.UpdateLicenceStatusRequest{Id: licenceUUID.String(), Status: "active"},
			findResp: testLicence,
			saveErr:  errors.New("db error"),
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := mockLicenceRepository{
				findFunc: func(_ context.Context, _ string) (*licences.Licence, error) {
					return tt.findResp, tt.findErr
				},
				saveFunc: func(_ context.Context, l *licences.Licence) (*licences.Licence, error) {
					if tt.saveErr != nil {
						return nil, tt.saveErr
					}
					return l, nil
				},
			}

			svc := newTestLicenceService(repo)
			res, err := svc.UpdateLicenceStatus(context.Background(), tt.req)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			if tt.wantErrors {
				assert.True(t, len(res.Errors) > 0)
			} else {
				assert.Empty(t, res.Errors)
				assert.NotNil(t, res.Licence)
				assert.Equal(t, tt.req.Status, res.Licence.Status)
			}
		})
	}
}

func TestGetMemberLicences(t *testing.T) {
	tests := []struct {
		name       string
		searchResp []*licences.Licence
		searchErr  error
		wantCount  int
		wantErr    bool
	}{
		{
			name:       "Returns licences",
			searchResp: []*licences.Licence{testLicence},
			wantCount:  1,
		},
		{
			name:       "No licences",
			searchResp: []*licences.Licence{},
			wantCount:  0,
		},
		{
			name:      "Repository error",
			searchErr: errors.New("db error"),
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := mockLicenceRepository{
				searchFunc: func(_ context.Context, _ *domain.SearchParams) ([]*licences.Licence, error) {
					return tt.searchResp, tt.searchErr
				},
			}

			svc := newTestLicenceService(repo)
			res, err := svc.GetMemberLicences(context.Background(), &dto.GetMemberLicencesRequest{MemberId: licenceMember.String()})

			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Len(t, res.Licences, tt.wantCount)
		})
	}
}
