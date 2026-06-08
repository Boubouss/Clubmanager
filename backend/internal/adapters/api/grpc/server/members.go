package server

import (
	"clubmanager/internal/adapters/api/grpc/dto"
	"clubmanager/internal/adapters/api/grpc/proto"
	"context"
)

func (s ClubManagerServiceServer) AddMember(ctx context.Context, req *proto.AddMemberRequest) (*proto.AddMemberResponse, error) {
	res, err := s.msvc.AddMember(ctx, &dto.AddMemberRequest{
		UserId:    req.UserId,
		ClubId:    req.ClubId,
		Firstname: req.Firstname,
		Lastname:  req.Lastname,
		Birthdate: req.Birthdate,
		Gender:    req.Gender,
	})
	if err != nil {
		return nil, err
	}
	var m *proto.Member
	if res.Member != nil {
		m = memberProto(res.Member)
	}
	return &proto.AddMemberResponse{Member: m, Errors: res.Errors}, nil
}

func (s ClubManagerServiceServer) ValidateMember(ctx context.Context, req *proto.ValidateMemberRequest) (*proto.ValidateMemberResponse, error) {
	res, err := s.msvc.ValidateMember(ctx, &dto.ValidateMemberRequest{Id: req.Id})
	if err != nil {
		return nil, err
	}
	var m *proto.Member
	if res.Member != nil {
		m = memberProto(res.Member)
	}
	return &proto.ValidateMemberResponse{Member: m, Errors: res.Errors}, nil
}

func (s ClubManagerServiceServer) GetMembersByUser(ctx context.Context, req *proto.GetMembersByUserRequest) (*proto.GetMembersByUserResponse, error) {
	res, err := s.msvc.GetMembersByUser(ctx, &dto.GetMembersByUserRequest{
		UserId: req.UserId,
		ClubId: req.ClubId,
	})
	if err != nil {
		return nil, err
	}
	return &proto.GetMembersByUserResponse{Members: arrayMemberProto(res.Members), Errors: res.Errors}, nil
}

func (s ClubManagerServiceServer) UpdateMember(ctx context.Context, req *proto.UpdateMemberRequest) (*proto.UpdateMemberResponse, error) {
	res, err := s.msvc.UpdateMember(ctx, &dto.UpdateMemberRequest{
		Id:        req.Id,
		Firstname: req.Firstname,
		Lastname:  req.Lastname,
		Birthdate: req.Birthdate,
		Gender:    req.Gender,
	})
	if err != nil {
		return nil, err
	}
	var m *proto.Member
	if res.Member != nil {
		m = memberProto(res.Member)
	}
	return &proto.UpdateMemberResponse{Member: m, Errors: res.Errors}, nil
}

func (s ClubManagerServiceServer) RemoveMember(ctx context.Context, req *proto.RemoveMemberRequest) (*proto.RemoveMemberResponse, error) {
	ok, err := s.msvc.RemoveMember(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &proto.RemoveMemberResponse{Ok: ok}, nil
}
