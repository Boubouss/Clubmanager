package server

import (
	"clubmanager/internal/adapters/api/grpc/dto"
	"clubmanager/internal/adapters/api/grpc/proto"
	"context"
)

func (s ClubManagerServiceServer) CreateEvent(ctx context.Context, req *proto.CreateEventRequest) (*proto.CreateEventResponse, error) {
	cats := make([]*dto.CreateEventCategoryRequest, 0, len(req.Categories))
	for _, c := range req.Categories {
		cats = append(cats, &dto.CreateEventCategoryRequest{
			JudoCategoryId: c.JudoCategoryId,
			WeighInAt:      c.WeighInAt,
			StartsAt:       c.StartsAt,
		})
	}

	res, err := s.esvc.CreateEvent(ctx, &dto.CreateEventRequest{
		ClubId:              req.ClubId,
		Title:               req.Title,
		Description:         req.Description,
		Type:                req.Type,
		Location:            req.Location,
		RegistrationOpenAt:  req.RegistrationOpenAt,
		RegistrationCloseAt: req.RegistrationCloseAt,
		Date:                req.Date,
		MaxParticipants:     int(req.MaxParticipants),
		Categories:          cats,
	})
	if err != nil {
		return nil, err
	}
	return &proto.CreateEventResponse{
		Event:      eventProto(res.Event),
		Categories: arrayEventCategoryProto(res.Categories),
		Errors:     res.Errors,
	}, nil
}

func (s ClubManagerServiceServer) GetEvent(ctx context.Context, req *proto.GetEventRequest) (*proto.GetEventResponse, error) {
	res, err := s.esvc.GetEvent(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &proto.GetEventResponse{
		Event:      eventProto(res.Event),
		Categories: arrayEventCategoryProto(res.Categories),
		Errors:     res.Errors,
	}, nil
}

func (s ClubManagerServiceServer) ListClubEvents(ctx context.Context, req *proto.ListClubEventsRequest) (*proto.ListClubEventsResponse, error) {
	res, err := s.esvc.ListClubEvents(ctx, req.ClubId)
	if err != nil {
		return nil, err
	}
	return &proto.ListClubEventsResponse{Events: arrayEventProto(res.Events), Errors: res.Errors}, nil
}

func (s ClubManagerServiceServer) UpdateEvent(ctx context.Context, req *proto.UpdateEventRequest) (*proto.UpdateEventResponse, error) {
	res, err := s.esvc.UpdateEvent(ctx, &dto.UpdateEventRequest{
		Id:                  req.Id,
		Title:               req.Title,
		Description:         req.Description,
		Location:            req.Location,
		RegistrationOpenAt:  req.RegistrationOpenAt,
		RegistrationCloseAt: req.RegistrationCloseAt,
		Date:                req.Date,
		MaxParticipants:     int(req.MaxParticipants),
	})
	if err != nil {
		return nil, err
	}
	return &proto.UpdateEventResponse{Event: eventProto(res.Event), Errors: res.Errors}, nil
}

func (s ClubManagerServiceServer) CancelEvent(ctx context.Context, req *proto.CancelEventRequest) (*proto.CancelEventResponse, error) {
	ok, err := s.esvc.CancelEvent(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &proto.CancelEventResponse{Ok: ok}, nil
}

func (s ClubManagerServiceServer) OpenEvent(ctx context.Context, req *proto.OpenEventRequest) (*proto.OpenEventResponse, error) {
	ok, err := s.esvc.OpenEvent(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &proto.OpenEventResponse{Ok: ok}, nil
}

func (s ClubManagerServiceServer) RegisterParticipant(ctx context.Context, req *proto.RegisterParticipantRequest) (*proto.RegisterParticipantResponse, error) {
	res, err := s.esvc.RegisterParticipant(ctx, &dto.RegisterParticipantRequest{
		EventId:         req.EventId,
		MemberId:        req.MemberId,
		Role:            req.Role,
		EventCategoryId: req.EventCategoryId,
	})
	if err != nil {
		return nil, err
	}
	var p *proto.EventParticipant
	if res.Participant != nil {
		p = participantProto(res.Participant)
	}
	return &proto.RegisterParticipantResponse{Participant: p, Errors: res.Errors}, nil
}

func (s ClubManagerServiceServer) UnregisterParticipant(ctx context.Context, req *proto.UnregisterParticipantRequest) (*proto.UnregisterParticipantResponse, error) {
	ok, err := s.esvc.UnregisterParticipant(ctx, &dto.UnregisterParticipantRequest{
		EventId:  req.EventId,
		MemberId: req.MemberId,
	})
	if err != nil {
		return nil, err
	}
	return &proto.UnregisterParticipantResponse{Ok: ok}, nil
}

func (s ClubManagerServiceServer) GetParticipants(ctx context.Context, req *proto.GetParticipantsRequest) (*proto.GetParticipantsResponse, error) {
	res, err := s.esvc.GetParticipants(ctx, req.EventId)
	if err != nil {
		return nil, err
	}
	return &proto.GetParticipantsResponse{
		Participants: arrayParticipantProto(res.Participants),
		Errors:       res.Errors,
	}, nil
}

func (s ClubManagerServiceServer) ExportParticipants(ctx context.Context, req *proto.ExportParticipantsRequest) (*proto.ExportParticipantsResponse, error) {
	data, err := s.esvc.ExportParticipants(ctx, req.EventId)
	if err != nil {
		return nil, err
	}
	return &proto.ExportParticipantsResponse{CsvData: data}, nil
}

func (s ClubManagerServiceServer) UpdateCategoryStatus(ctx context.Context, req *proto.UpdateCategoryStatusRequest) (*proto.UpdateCategoryStatusResponse, error) {
	res, err := s.esvc.UpdateCategoryStatus(ctx, &dto.UpdateCategoryStatusRequest{
		CategoryId: req.CategoryId,
		Status:     req.Status,
	})
	if err != nil {
		return nil, err
	}
	var cat *proto.EventCategory
	if res.Category != nil {
		cat = eventCategoryProto(res.Category)
	}
	return &proto.UpdateCategoryStatusResponse{Category: cat, Errors: res.Errors}, nil
}

func (s ClubManagerServiceServer) CreateCarpoolOffer(ctx context.Context, req *proto.CreateCarpoolOfferRequest) (*proto.CreateCarpoolOfferResponse, error) {
	res, err := s.esvc.CreateCarpoolOffer(ctx, &dto.CreateCarpoolOfferRequest{
		EventId:          req.EventId,
		MemberId:         req.MemberId,
		DepartureAddress: req.DepartureAddress,
		AvailableSeats:   int(req.AvailableSeats),
		DepartureAt:      req.DepartureAt,
	})
	if err != nil {
		return nil, err
	}
	var offer *proto.CarpoolOffer
	if res.Offer != nil {
		offer = carpoolOfferProto(res.Offer)
	}
	return &proto.CreateCarpoolOfferResponse{Offer: offer, Errors: res.Errors}, nil
}

func (s ClubManagerServiceServer) JoinCarpoolOffer(ctx context.Context, req *proto.JoinCarpoolOfferRequest) (*proto.JoinCarpoolOfferResponse, error) {
	ok, err := s.esvc.JoinCarpoolOffer(ctx, req.OfferId, req.MemberId)
	if err != nil {
		return nil, err
	}
	return &proto.JoinCarpoolOfferResponse{Ok: ok}, nil
}

func (s ClubManagerServiceServer) LeaveCarpoolOffer(ctx context.Context, req *proto.LeaveCarpoolOfferRequest) (*proto.LeaveCarpoolOfferResponse, error) {
	ok, err := s.esvc.LeaveCarpoolOffer(ctx, req.OfferId, req.MemberId)
	if err != nil {
		return nil, err
	}
	return &proto.LeaveCarpoolOfferResponse{Ok: ok}, nil
}

func (s ClubManagerServiceServer) GetEventCarpools(ctx context.Context, req *proto.GetEventCarpoolsRequest) (*proto.GetEventCarpoolsResponse, error) {
	res, err := s.esvc.GetEventCarpools(ctx, req.EventId)
	if err != nil {
		return nil, err
	}
	return &proto.GetEventCarpoolsResponse{Offers: arrayCarpoolOfferProto(res.Offers), Errors: res.Errors}, nil
}

func (s ClubManagerServiceServer) CancelCarpoolOffer(ctx context.Context, req *proto.CancelCarpoolOfferRequest) (*proto.CancelCarpoolOfferResponse, error) {
	ok, err := s.esvc.CancelCarpoolOffer(ctx, req.OfferId)
	if err != nil {
		return nil, err
	}
	return &proto.CancelCarpoolOfferResponse{Ok: ok}, nil
}

func (s ClubManagerServiceServer) ListJudoCategories(ctx context.Context, req *proto.ListJudoCategoriesRequest) (*proto.ListJudoCategoriesResponse, error) {
	res, err := s.esvc.ListJudoCategories(ctx)
	if err != nil {
		return nil, err
	}
	return &proto.ListJudoCategoriesResponse{Categories: arrayJudoCategoryProto(res.Categories)}, nil
}
