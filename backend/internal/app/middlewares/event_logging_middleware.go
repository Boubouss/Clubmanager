package middlewares

import (
	"clubmanager/internal/adapters/api/grpc/dto"
	"clubmanager/internal/app/services"
	events "clubmanager/internal/domain/events"
	"context"
	"fmt"
	"time"
)

type eventLoggingService struct {
	next services.EventService
}

func NewEventLoggingService(next services.EventService) services.EventService {
	return &eventLoggingService{next: next}
}

func (s *eventLoggingService) CreateEvent(ctx context.Context, data *dto.CreateEventRequest) (res *dto.CreateEventResponse, err error) {
	defer func(begin time.Time) {
		fmt.Printf("=> type: '%s'; took: '%v'; err: '%v'.\n", "CreateEvent", time.Since(begin), err)
	}(time.Now())
	return s.next.CreateEvent(ctx, data)
}

func (s *eventLoggingService) GetEvent(ctx context.Context, id string) (res *dto.GetEventResponse, err error) {
	defer func(begin time.Time) {
		fmt.Printf("=> type: '%s'; took: '%v'; err: '%v'.\n", "GetEvent", time.Since(begin), err)
	}(time.Now())
	return s.next.GetEvent(ctx, id)
}

func (s *eventLoggingService) ListClubEvents(ctx context.Context, req *dto.ListClubEventsRequest) (res *dto.ListClubEventsResponse, err error) {
	defer func(begin time.Time) {
		fmt.Printf("=> type: '%s'; took: '%v'; err: '%v'.\n", "ListClubEvents", time.Since(begin), err)
	}(time.Now())
	return s.next.ListClubEvents(ctx, req)
}

func (s *eventLoggingService) UpdateEvent(ctx context.Context, data *dto.UpdateEventRequest) (res *dto.UpdateEventResponse, err error) {
	defer func(begin time.Time) {
		fmt.Printf("=> type: '%s'; took: '%v'; err: '%v'.\n", "UpdateEvent", time.Since(begin), err)
	}(time.Now())
	return s.next.UpdateEvent(ctx, data)
}

func (s *eventLoggingService) CancelEvent(ctx context.Context, id string) (ok bool, err error) {
	defer func(begin time.Time) {
		fmt.Printf("=> type: '%s'; took: '%v'; err: '%v'.\n", "CancelEvent", time.Since(begin), err)
	}(time.Now())
	return s.next.CancelEvent(ctx, id)
}

func (s *eventLoggingService) ReopenEvent(ctx context.Context, id string) (ok bool, err error) {
	defer func(begin time.Time) {
		fmt.Printf("=> type: '%s'; took: '%v'; err: '%v'.\n", "ReopenEvent", time.Since(begin), err)
	}(time.Now())
	return s.next.ReopenEvent(ctx, id)
}

func (s *eventLoggingService) DeleteEvent(ctx context.Context, id string) (ok bool, err error) {
	defer func(begin time.Time) {
		fmt.Printf("=> type: '%s'; took: '%v'; err: '%v'.\n", "DeleteEvent", time.Since(begin), err)
	}(time.Now())
	return s.next.DeleteEvent(ctx, id)
}

func (s *eventLoggingService) OpenEvent(ctx context.Context, id string) (ok bool, err error) {
	defer func(begin time.Time) {
		fmt.Printf("=> type: '%s'; took: '%v'; err: '%v'.\n", "OpenEvent", time.Since(begin), err)
	}(time.Now())
	return s.next.OpenEvent(ctx, id)
}

func (s *eventLoggingService) RegisterParticipant(ctx context.Context, data *dto.RegisterParticipantRequest) (res *dto.RegisterParticipantResponse, err error) {
	defer func(begin time.Time) {
		fmt.Printf("=> type: '%s'; took: '%v'; err: '%v'.\n", "RegisterParticipant", time.Since(begin), err)
	}(time.Now())
	return s.next.RegisterParticipant(ctx, data)
}

func (s *eventLoggingService) UnregisterParticipant(ctx context.Context, data *dto.UnregisterParticipantRequest) (ok bool, err error) {
	defer func(begin time.Time) {
		fmt.Printf("=> type: '%s'; took: '%v'; err: '%v'.\n", "UnregisterParticipant", time.Since(begin), err)
	}(time.Now())
	return s.next.UnregisterParticipant(ctx, data)
}

func (s *eventLoggingService) GetParticipants(ctx context.Context, eventId string) (res *dto.GetParticipantsResponse, err error) {
	defer func(begin time.Time) {
		fmt.Printf("=> type: '%s'; took: '%v'; err: '%v'.\n", "GetParticipants", time.Since(begin), err)
	}(time.Now())
	return s.next.GetParticipants(ctx, eventId)
}

func (s *eventLoggingService) ExportParticipants(ctx context.Context, eventId string) (data []byte, err error) {
	defer func(begin time.Time) {
		fmt.Printf("=> type: '%s'; took: '%v'; err: '%v'.\n", "ExportParticipants", time.Since(begin), err)
	}(time.Now())
	return s.next.ExportParticipants(ctx, eventId)
}

func (s *eventLoggingService) UpdateCategoryStatus(ctx context.Context, data *dto.UpdateCategoryStatusRequest) (res *dto.UpdateCategoryStatusResponse, err error) {
	defer func(begin time.Time) {
		fmt.Printf("=> type: '%s'; took: '%v'; err: '%v'.\n", "UpdateCategoryStatus", time.Since(begin), err)
	}(time.Now())
	return s.next.UpdateCategoryStatus(ctx, data)
}

func (s *eventLoggingService) CreateCarpoolOffer(ctx context.Context, data *dto.CreateCarpoolOfferRequest) (res *dto.CreateCarpoolOfferResponse, err error) {
	defer func(begin time.Time) {
		fmt.Printf("=> type: '%s'; took: '%v'; err: '%v'.\n", "CreateCarpoolOffer", time.Since(begin), err)
	}(time.Now())
	return s.next.CreateCarpoolOffer(ctx, data)
}

func (s *eventLoggingService) JoinCarpoolOffer(ctx context.Context, offerId, memberId string) (ok bool, err error) {
	defer func(begin time.Time) {
		fmt.Printf("=> type: '%s'; took: '%v'; err: '%v'.\n", "JoinCarpoolOffer", time.Since(begin), err)
	}(time.Now())
	return s.next.JoinCarpoolOffer(ctx, offerId, memberId)
}

func (s *eventLoggingService) LeaveCarpoolOffer(ctx context.Context, offerId, memberId string) (ok bool, err error) {
	defer func(begin time.Time) {
		fmt.Printf("=> type: '%s'; took: '%v'; err: '%v'.\n", "LeaveCarpoolOffer", time.Since(begin), err)
	}(time.Now())
	return s.next.LeaveCarpoolOffer(ctx, offerId, memberId)
}

func (s *eventLoggingService) GetEventCarpoolPassengers(ctx context.Context, eventId string) (res []*events.CarpoolPassenger, err error) {
	defer func(begin time.Time) {
		fmt.Printf("=> type: '%s'; took: '%v'; err: '%v'.\n", "GetEventCarpoolPassengers", time.Since(begin), err)
	}(time.Now())
	return s.next.GetEventCarpoolPassengers(ctx, eventId)
}

func (s *eventLoggingService) GetEventCarpools(ctx context.Context, eventId string) (res *dto.GetEventCarpoolsResponse, err error) {
	defer func(begin time.Time) {
		fmt.Printf("=> type: '%s'; took: '%v'; err: '%v'.\n", "GetEventCarpools", time.Since(begin), err)
	}(time.Now())
	return s.next.GetEventCarpools(ctx, eventId)
}

func (s *eventLoggingService) CancelCarpoolOffer(ctx context.Context, offerId string) (ok bool, err error) {
	defer func(begin time.Time) {
		fmt.Printf("=> type: '%s'; took: '%v'; err: '%v'.\n", "CancelCarpoolOffer", time.Since(begin), err)
	}(time.Now())
	return s.next.CancelCarpoolOffer(ctx, offerId)
}

func (s *eventLoggingService) ListJudoCategories(ctx context.Context) (res *dto.ListJudoCategoriesResponse, err error) {
	defer func(begin time.Time) {
		fmt.Printf("=> type: '%s'; took: '%v'; err: '%v'.\n", "ListJudoCategories", time.Since(begin), err)
	}(time.Now())
	return s.next.ListJudoCategories(ctx)
}
