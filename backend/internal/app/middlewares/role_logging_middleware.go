package middlewares

import (
	"clubmanager/internal/adapters/api/grpc/dto"
	"clubmanager/internal/app/services"
	"clubmanager/internal/domain/roles"
	"context"
	"fmt"
	"time"
)

type roleLoggingService struct {
	next services.RoleService
}

func NewRoleLoggingService(next services.RoleService) services.RoleService {
	return &roleLoggingService{next: next}
}

func (s *roleLoggingService) AssignRole(ctx context.Context, data *dto.AssignRoleRequest) (res *dto.AssignRoleResponse, err error) {
	defer func(begin time.Time) {
		fmt.Printf("=> type: '%s'; took: '%v'; err: '%v'.\n", "AssignRole", time.Since(begin), err)
	}(time.Now())
	return s.next.AssignRole(ctx, data)
}

func (s *roleLoggingService) RemoveRole(ctx context.Context, data *dto.RemoveRoleRequest) (ok bool, err error) {
	defer func(begin time.Time) {
		fmt.Printf("=> type: '%s'; took: '%v'; err: '%v'.\n", "RemoveRole", time.Since(begin), err)
	}(time.Now())
	return s.next.RemoveRole(ctx, data)
}

func (s *roleLoggingService) GetClubRoles(ctx context.Context, data *dto.GetClubRolesRequest) (res *dto.GetClubRolesResponse, err error) {
	defer func(begin time.Time) {
		fmt.Printf("=> type: '%s'; took: '%v'; err: '%v'.\n", "GetClubRoles", time.Since(begin), err)
	}(time.Now())
	return s.next.GetClubRoles(ctx, data)
}

func (s *roleLoggingService) GetUserRoles(ctx context.Context, data *dto.GetUserRolesRequest) (res *dto.GetUserRolesResponse, err error) {
	defer func(begin time.Time) {
		fmt.Printf("=> type: '%s'; took: '%v'; err: '%v'.\n", "GetUserRoles", time.Since(begin), err)
	}(time.Now())
	return s.next.GetUserRoles(ctx, data)
}

func (s *roleLoggingService) InitPresidentRole(ctx context.Context, userId, clubId string) (role *roles.Role, err error) {
	defer func(begin time.Time) {
		fmt.Printf("=> type: '%s'; took: '%v'; err: '%v'.\n", "InitPresidentRole", time.Since(begin), err)
	}(time.Now())
	return s.next.InitPresidentRole(ctx, userId, clubId)
}
