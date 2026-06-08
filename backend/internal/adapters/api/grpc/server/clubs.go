package server

import (
	"clubmanager/internal/adapters/api/grpc/dto"
	"clubmanager/internal/adapters/api/grpc/proto"
	"context"
	"fmt"
)

func (s ClubManagerServiceServer) CreateClub(ctx context.Context, req *proto.CreateClubRequest) (*proto.CreateClubResponse, error) {
	res, err := s.csvc.CreateClub(ctx, &dto.CreateClubRequest{
		Siren:       req.Siren,
		Name:        req.Name,
		Address:     req.Address,
		City:        req.City,
		PostalCode:  req.PostalCode,
		Country:     req.Country,
		Phonenumber: req.Phonenumber,
	})
	if err != nil {
		return nil, err
	}

	if res.Club != nil {
		userId, ok := ctx.Value("user_id").(string)
		if !ok || userId == "" {
			return nil, fmt.Errorf("unauthorized: missing user context")
		}
		if _, err := s.rsvc.InitPresidentRole(ctx, userId, res.Club.Id.String()); err != nil {
			return nil, err
		}
		return &proto.CreateClubResponse{Club: clubProto(res.Club), Errors: res.Errors}, nil
	}
	return &proto.CreateClubResponse{Errors: res.Errors}, nil
}

func (s ClubManagerServiceServer) ReadClub(ctx context.Context, req *proto.ReadClubRequest) (*proto.ReadClubResponse, error) {
	params := make(map[string]any, len(req.Params))
	for k, v := range req.Params {
		params[k] = v
	}

	res, err := s.csvc.ReadClub(ctx, &dto.ReadClubRequest{Params: params})
	if err != nil {
		return nil, err
	}
	return &proto.ReadClubResponse{Clubs: arrayClubProto(res.Clubs), Errors: res.Errors}, nil
}

func (s ClubManagerServiceServer) UpdateClub(ctx context.Context, req *proto.UpdateClubRequest) (*proto.UpdateClubResponse, error) {
	res, err := s.csvc.UpdateClub(ctx, &dto.UpdateClubRequest{
		Id:          req.Id,
		Name:        req.Name,
		Address:     req.Address,
		City:        req.City,
		PostalCode:  req.PostalCode,
		Country:     req.Country,
		Phonenumber: req.Phonenumber,
	})
	if err != nil {
		return nil, err
	}

	var club *proto.Club
	if res.Club != nil {
		club = clubProto(res.Club)
	}
	return &proto.UpdateClubResponse{Club: club, Errors: res.Errors}, nil
}

func (s ClubManagerServiceServer) DeleteClub(ctx context.Context, req *proto.DeleteClubRequest) (*proto.DeleteClubResponse, error) {
	ok, err := s.csvc.DeleteClub(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &proto.DeleteClubResponse{Ok: ok}, nil
}
