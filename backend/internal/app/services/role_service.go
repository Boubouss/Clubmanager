package services

import (
	"clubmanager/internal/adapters/api/grpc/dto"
	"clubmanager/internal/domain"
	"clubmanager/internal/domain/roles"
	"context"
	"fmt"

	"github.com/google/uuid"
)

type RoleService interface {
	AssignRole(context.Context, *dto.AssignRoleRequest) (*dto.AssignRoleResponse, error)
	RemoveRole(context.Context, *dto.RemoveRoleRequest) (bool, error)
	GetClubRoles(context.Context, *dto.GetClubRolesRequest) (*dto.GetClubRolesResponse, error)
	GetUserRoles(context.Context, *dto.GetUserRolesRequest) (*dto.GetUserRolesResponse, error)
	InitPresidentRole(ctx context.Context, userId, clubId string) (*roles.Role, error)
}

type RoleServiceConfig struct {
	Repository  domain.Repository[roles.Role, string]
	RoleChecker roles.RoleChecker
}

type roleService struct {
	repo    domain.Repository[roles.Role, string]
	checker roles.RoleChecker
}

func NewRoleService(config RoleServiceConfig) *roleService {
	return &roleService{repo: config.Repository, checker: config.RoleChecker}
}

func (s *roleService) AssignRole(ctx context.Context, data *dto.AssignRoleRequest) (*dto.AssignRoleResponse, error) {
	if !roles.IsValid(data.Role) {
		return &dto.AssignRoleResponse{Errors: map[string]string{"role": "Invalid role."}}, nil
	}

	assignerUserId, ok := ctx.Value("user_id").(string)
	if !ok || assignerUserId == "" {
		return &dto.AssignRoleResponse{
			Errors: map[string]string{"auth": "Missing or invalid user context."},
		}, nil
	}

	isSuperAdmin, err := s.checker.IsSuperAdmin(ctx, assignerUserId)
	if err != nil {
		return nil, err
	}

	if !isSuperAdmin {
		assignerRoles, err := s.repo.Search(ctx, &domain.SearchParams{
			Fields:    map[string]any{"user_id": assignerUserId, "club_id": data.ClubId},
			Keys:      map[string]bool{"user_id": true, "club_id": true},
			Connector: "AND",
		})
		if err != nil {
			return nil, err
		}

		canAssign := false
		for _, r := range assignerRoles {
			if roles.CanAssign(r.Role, data.Role) {
				canAssign = true
				break
			}
		}
		if !canAssign {
			return &dto.AssignRoleResponse{
				Errors: map[string]string{"role": "Insufficient permissions to assign this role."},
			}, nil
		}
	}

	userId, err := uuid.Parse(data.UserId)
	if err != nil {
		return nil, fmt.Errorf("invalid user_id: %w", err)
	}
	clubId, err := uuid.Parse(data.ClubId)
	if err != nil {
		return nil, fmt.Errorf("invalid club_id: %w", err)
	}

	role := &roles.Role{
		UserId: userId,
		ClubId: clubId,
		Role:   data.Role,
	}

	created, err := s.repo.Save(ctx, role)
	if err != nil {
		return nil, err
	}
	return &dto.AssignRoleResponse{Role: created, Errors: make(map[string]string)}, nil
}

func (s *roleService) RemoveRole(ctx context.Context, data *dto.RemoveRoleRequest) (bool, error) {
	callerUserId, ok := ctx.Value("user_id").(string)
	if !ok || callerUserId == "" {
		return false, fmt.Errorf("missing or invalid user context")
	}

	role, err := s.repo.Find(ctx, data.Id)
	if err != nil {
		return false, err
	}

	isSuperAdmin, err := s.checker.IsSuperAdmin(ctx, callerUserId)
	if err != nil {
		return false, err
	}

	if !isSuperAdmin {
		isPresident, err := s.checker.HasRole(ctx, callerUserId, role.ClubId.String(), roles.RolePresident)
		if err != nil {
			return false, err
		}
		if !isPresident {
			return false, fmt.Errorf("insufficient permissions to remove this role")
		}
	}

	return s.repo.Delete(ctx, data.Id)
}

func (s *roleService) GetClubRoles(ctx context.Context, data *dto.GetClubRolesRequest) (*dto.GetClubRolesResponse, error) {
	list, err := s.repo.Search(ctx, &domain.SearchParams{
		Fields:    map[string]any{"club_id": data.ClubId},
		Keys:      map[string]bool{"club_id": true},
		Connector: "AND",
	})
	if err != nil {
		return nil, err
	}
	return &dto.GetClubRolesResponse{Roles: list, Errors: make(map[string]string)}, nil
}

func (s *roleService) GetUserRoles(ctx context.Context, data *dto.GetUserRolesRequest) (*dto.GetUserRolesResponse, error) {
	list, err := s.repo.Search(ctx, &domain.SearchParams{
		Fields:    map[string]any{"user_id": data.UserId},
		Keys:      map[string]bool{"user_id": true},
		Connector: "AND",
	})
	if err != nil {
		return nil, err
	}
	return &dto.GetUserRolesResponse{Roles: list, Errors: make(map[string]string)}, nil
}

func (s *roleService) InitPresidentRole(ctx context.Context, userId, clubId string) (*roles.Role, error) {
	uid, err := uuid.Parse(userId)
	if err != nil {
		return nil, fmt.Errorf("invalid user_id: %w", err)
	}
	cid, err := uuid.Parse(clubId)
	if err != nil {
		return nil, fmt.Errorf("invalid club_id: %w", err)
	}

	role := &roles.Role{
		UserId: uid,
		ClubId: cid,
		Role:   roles.RolePresident,
	}
	return s.repo.Save(ctx, role)
}
