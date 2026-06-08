package services

import (
	"clubmanager/internal/adapters/api/grpc/dto"
	"clubmanager/internal/domain"
	"clubmanager/internal/domain/roles"
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

type mockRoleRepository struct {
	saveFunc   func(context.Context, *roles.Role) (*roles.Role, error)
	findFunc   func(context.Context, string) (*roles.Role, error)
	searchFunc func(context.Context, *domain.SearchParams) ([]*roles.Role, error)
	deleteFunc func(context.Context, string) (bool, error)
}

func (m mockRoleRepository) Save(ctx context.Context, r *roles.Role) (*roles.Role, error) {
	return m.saveFunc(ctx, r)
}
func (m mockRoleRepository) Find(ctx context.Context, id string) (*roles.Role, error) {
	return m.findFunc(ctx, id)
}
func (m mockRoleRepository) Search(ctx context.Context, p *domain.SearchParams) ([]*roles.Role, error) {
	return m.searchFunc(ctx, p)
}
func (m mockRoleRepository) Delete(ctx context.Context, id string) (bool, error) {
	return m.deleteFunc(ctx, id)
}

type mockRoleChecker struct {
	hasRoleFunc     func(context.Context, string, string, ...string) (bool, error)
	isSuperAdminFunc func(context.Context, string) (bool, error)
}

func (m mockRoleChecker) HasRole(ctx context.Context, userId, clubId string, r ...string) (bool, error) {
	return m.hasRoleFunc(ctx, userId, clubId, r...)
}
func (m mockRoleChecker) IsSuperAdmin(ctx context.Context, userId string) (bool, error) {
	return m.isSuperAdminFunc(ctx, userId)
}

var (
	roleUUID   = uuid.MustParse("00000000-0000-0000-0000-000000000010")
	assignerID = uuid.MustParse("00000000-0000-0000-0000-000000000011")
	targetID   = uuid.MustParse("00000000-0000-0000-0000-000000000012")
	testClubID = uuid.MustParse("00000000-0000-0000-0000-000000000002")
)

func ctxWithUser(userId string) context.Context {
	return context.WithValue(context.Background(), "user_id", userId)
}

func newTestRoleService(repo domain.Repository[roles.Role, string], checker roles.RoleChecker) *roleService {
	return NewRoleService(RoleServiceConfig{Repository: repo, RoleChecker: checker})
}

func TestAssignRole(t *testing.T) {
	presidentRole := &roles.Role{
		Id: roleUUID, UserId: assignerID, ClubId: testClubID, Role: roles.RolePresident,
	}

	tests := []struct {
		name         string
		req          *dto.AssignRoleRequest
		assignerRoles []*roles.Role
		isSuperAdmin bool
		checkerErr   error
		saveErr      error
		wantErrors   bool
		wantErr      bool
	}{
		{
			name:         "President assigns staff",
			req:          &dto.AssignRoleRequest{UserId: targetID.String(), ClubId: testClubID.String(), Role: roles.RoleStaff},
			assignerRoles: []*roles.Role{presidentRole},
		},
		{
			name:         "SuperAdmin assigns any role",
			req:          &dto.AssignRoleRequest{UserId: targetID.String(), ClubId: testClubID.String(), Role: roles.RolePresident},
			isSuperAdmin: true,
		},
		{
			name:       "Invalid role",
			req:        &dto.AssignRoleRequest{UserId: targetID.String(), ClubId: testClubID.String(), Role: "invalid"},
			wantErrors: true,
		},
		{
			name:         "Insufficient permissions",
			req:          &dto.AssignRoleRequest{UserId: targetID.String(), ClubId: testClubID.String(), Role: roles.RoleStaff},
			assignerRoles: []*roles.Role{{UserId: assignerID, ClubId: testClubID, Role: roles.RoleCoach}},
			wantErrors:   true,
		},
		{
			name:       "Checker error",
			req:        &dto.AssignRoleRequest{UserId: targetID.String(), ClubId: testClubID.String(), Role: roles.RoleStaff},
			checkerErr: errors.New("db error"),
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := mockRoleRepository{
				searchFunc: func(_ context.Context, _ *domain.SearchParams) ([]*roles.Role, error) {
					return tt.assignerRoles, nil
				},
				saveFunc: func(_ context.Context, r *roles.Role) (*roles.Role, error) {
					if tt.saveErr != nil {
						return nil, tt.saveErr
					}
					r.Id = roleUUID
					return r, nil
				},
			}
			checker := mockRoleChecker{
				isSuperAdminFunc: func(_ context.Context, _ string) (bool, error) {
					return tt.isSuperAdmin, tt.checkerErr
				},
				hasRoleFunc: func(_ context.Context, _, _ string, _ ...string) (bool, error) {
					return false, nil
				},
			}

			svc := newTestRoleService(repo, checker)
			ctx := ctxWithUser(assignerID.String())
			res, err := svc.AssignRole(ctx, tt.req)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			if tt.wantErrors {
				assert.True(t, len(res.Errors) > 0)
				assert.Nil(t, res.Role)
			} else {
				assert.Empty(t, res.Errors)
				assert.NotNil(t, res.Role)
			}
		})
	}
}

func TestRemoveRole(t *testing.T) {
	testRole := &roles.Role{
		Id: roleUUID, UserId: targetID, ClubId: testClubID, Role: roles.RoleCoach,
	}

	tests := []struct {
		name         string
		findResp     *roles.Role
		findErr      error
		isSuperAdmin bool
		isPresident  bool
		checkerErr   error
		deleteErr    error
		wantOk       bool
		wantErr      bool
	}{
		{"SuperAdmin removes role", testRole, nil, true, false, nil, nil, true, false},
		{"President removes role", testRole, nil, false, true, nil, nil, true, false},
		{"Insufficient permissions", testRole, nil, false, false, nil, nil, false, true},
		{"Role not found", nil, errors.New("not found"), false, false, nil, nil, false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := mockRoleRepository{
				findFunc: func(_ context.Context, _ string) (*roles.Role, error) {
					return tt.findResp, tt.findErr
				},
				deleteFunc: func(_ context.Context, _ string) (bool, error) {
					return true, tt.deleteErr
				},
			}
			checker := mockRoleChecker{
				isSuperAdminFunc: func(_ context.Context, _ string) (bool, error) {
					return tt.isSuperAdmin, tt.checkerErr
				},
				hasRoleFunc: func(_ context.Context, _, _ string, _ ...string) (bool, error) {
					return tt.isPresident, nil
				},
			}

			svc := newTestRoleService(repo, checker)
			ctx := ctxWithUser(assignerID.String())
			ok, err := svc.RemoveRole(ctx, &dto.RemoveRoleRequest{Id: roleUUID.String()})

			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.wantOk, ok)
		})
	}
}

func TestInitPresidentRole(t *testing.T) {
	tests := []struct {
		name    string
		userId  string
		clubId  string
		saveErr error
		wantErr bool
	}{
		{"Valid", assignerID.String(), testClubID.String(), nil, false},
		{"Invalid userId", "not-a-uuid", testClubID.String(), nil, true},
		{"Repository error", assignerID.String(), testClubID.String(), errors.New("db error"), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := mockRoleRepository{
				saveFunc: func(_ context.Context, r *roles.Role) (*roles.Role, error) {
					if tt.saveErr != nil {
						return nil, tt.saveErr
					}
					r.Id = roleUUID
					return r, nil
				},
			}
			checker := mockRoleChecker{}

			svc := newTestRoleService(repo, checker)
			role, err := svc.InitPresidentRole(context.Background(), tt.userId, tt.clubId)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.NotNil(t, role)
			assert.Equal(t, roles.RolePresident, role.Role)
		})
	}
}
