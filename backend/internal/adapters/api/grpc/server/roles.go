package server

import (
	"clubmanager/internal/adapters/api/grpc/dto"
	"clubmanager/internal/adapters/api/grpc/proto"
	"context"
)

func (s ClubManagerServiceServer) AssignRole(ctx context.Context, req *proto.AssignRoleRequest) (*proto.AssignRoleResponse, error) {
	res, err := s.rsvc.AssignRole(ctx, &dto.AssignRoleRequest{
		UserId: req.UserId,
		ClubId: req.ClubId,
		Role:   req.Role,
	})
	if err != nil {
		return nil, err
	}

	var role *proto.Role
	if res.Role != nil {
		role = roleProto(res.Role)
	}
	return &proto.AssignRoleResponse{Role: role, Errors: res.Errors}, nil
}

func (s ClubManagerServiceServer) RemoveRole(ctx context.Context, req *proto.RemoveRoleRequest) (*proto.RemoveRoleResponse, error) {
	ok, err := s.rsvc.RemoveRole(ctx, &dto.RemoveRoleRequest{Id: req.Id})
	if err != nil {
		return nil, err
	}
	return &proto.RemoveRoleResponse{Ok: ok}, nil
}

func (s ClubManagerServiceServer) GetClubRoles(ctx context.Context, req *proto.GetClubRolesRequest) (*proto.GetClubRolesResponse, error) {
	res, err := s.rsvc.GetClubRoles(ctx, &dto.GetClubRolesRequest{ClubId: req.ClubId})
	if err != nil {
		return nil, err
	}
	return &proto.GetClubRolesResponse{Roles: arrayRoleProto(res.Roles), Errors: res.Errors}, nil
}

func (s ClubManagerServiceServer) GetUserRoles(ctx context.Context, req *proto.GetUserRolesRequest) (*proto.GetUserRolesResponse, error) {
	res, err := s.rsvc.GetUserRoles(ctx, &dto.GetUserRolesRequest{UserId: req.UserId})
	if err != nil {
		return nil, err
	}
	return &proto.GetUserRolesResponse{Roles: arrayRoleProto(res.Roles), Errors: res.Errors}, nil
}
