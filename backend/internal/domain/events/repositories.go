package events

import "context"

type EventCategoryRepository interface {
	Save(ctx context.Context, c *EventCategory) (*EventCategory, error)
	FindByEvent(ctx context.Context, eventId string) ([]*EventCategory, error)
	FindById(ctx context.Context, id string) (*EventCategory, error)
	UpdateStatus(ctx context.Context, id, status string) (*EventCategory, error)
	DeleteByEvent(ctx context.Context, eventId string) error
}

type ParticipantRepository interface {
	// Register inserts a participant atomically. maxParticipants=0 means unlimited.
	// Returns (nil, nil) if the insert was blocked (duplicate or capacity exceeded).
	Register(ctx context.Context, p *EventParticipant, maxParticipants int) (*EventParticipant, error)
	Unregister(ctx context.Context, eventId, memberId string) (bool, error)
	FindByEvent(ctx context.Context, eventId string) ([]*EventParticipant, error)
	FindByMember(ctx context.Context, eventId, memberId string) (*EventParticipant, error)
	CountByEvent(ctx context.Context, eventId string) (int, error)
}

type CarpoolRepository interface {
	CreateOffer(ctx context.Context, o *CarpoolOffer) (*CarpoolOffer, error)
	FindOfferById(ctx context.Context, offerId string) (*CarpoolOffer, error)
	FindOffersByEvent(ctx context.Context, eventId string) ([]*CarpoolOffer, error)
	FindOfferByMember(ctx context.Context, eventId, memberId string) (*CarpoolOffer, error)
	DeleteOffer(ctx context.Context, offerId string) (bool, error)
	AddPassenger(ctx context.Context, offerId, memberId string) (bool, error)
	RemovePassenger(ctx context.Context, offerId, memberId string) (bool, error)
	CountPassengers(ctx context.Context, offerId string) (int, error)
	IsPassenger(ctx context.Context, eventId, memberId string) (bool, error)
}

type JudoCategoryRepository interface {
	FindAll(ctx context.Context) ([]*JudoCategory, error)
	FindById(ctx context.Context, id string) (*JudoCategory, error)
}
