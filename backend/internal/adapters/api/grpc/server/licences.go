package server

import (
	"clubmanager/internal/adapters/api/grpc/dto"
	"clubmanager/internal/adapters/api/grpc/proto"
	"context"
)

func (s ClubManagerServiceServer) CreateLicence(ctx context.Context, req *proto.CreateLicenceRequest) (*proto.CreateLicenceResponse, error) {
	res, err := s.lsvc.CreateLicence(ctx, &dto.CreateLicenceRequest{
		MemberId:      req.MemberId,
		LicenceNumber: req.LicenceNumber,
		ValidFrom:     req.ValidFrom,
		ValidUntil:    req.ValidUntil,
	})
	if err != nil {
		return nil, err
	}
	var l *proto.Licence
	if res.Licence != nil {
		l = licenceProto(res.Licence)
	}
	return &proto.CreateLicenceResponse{Licence: l, Errors: res.Errors}, nil
}

func (s ClubManagerServiceServer) GetMemberLicences(ctx context.Context, req *proto.GetMemberLicencesRequest) (*proto.GetMemberLicencesResponse, error) {
	res, err := s.lsvc.GetMemberLicences(ctx, &dto.GetMemberLicencesRequest{MemberId: req.MemberId})
	if err != nil {
		return nil, err
	}
	return &proto.GetMemberLicencesResponse{Licences: arrayLicenceProto(res.Licences), Errors: res.Errors}, nil
}

func (s ClubManagerServiceServer) UpdateLicenceStatus(ctx context.Context, req *proto.UpdateLicenceStatusRequest) (*proto.UpdateLicenceStatusResponse, error) {
	res, err := s.lsvc.UpdateLicenceStatus(ctx, &dto.UpdateLicenceStatusRequest{
		Id:     req.Id,
		Status: req.Status,
	})
	if err != nil {
		return nil, err
	}
	var l *proto.Licence
	if res.Licence != nil {
		l = licenceProto(res.Licence)
	}
	return &proto.UpdateLicenceStatusResponse{Licence: l, Errors: res.Errors}, nil
}
