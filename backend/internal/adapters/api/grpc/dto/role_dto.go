package dto

import r "clubmanager/internal/domain/roles"

type AssignRoleRequest struct {
	UserId string
	ClubId string
	Role   string
}

type AssignRoleResponse struct {
	Role   *r.Role
	Errors map[string]string
}

type RemoveRoleRequest struct {
	Id string
}

type GetClubRolesRequest struct {
	ClubId string
}

type GetClubRolesResponse struct {
	Roles  []*r.Role
	Errors map[string]string
}

type GetUserRolesRequest struct {
	UserId string
}

type GetUserRolesResponse struct {
	Roles  []*r.Role
	Errors map[string]string
}
