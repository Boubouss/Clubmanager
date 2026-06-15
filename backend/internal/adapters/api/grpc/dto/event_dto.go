package dto

import e "clubmanager/internal/domain/events"

// Event

type CreateEventRequest struct {
	ClubId              string
	Title               string
	Description         string
	Type                string
	Location            string
	RegistrationOpenAt  string
	RegistrationCloseAt string
	Date                string
	MaxParticipants     int
	Categories          []*CreateEventCategoryRequest // only for competition
}

type CreateEventCategoryRequest struct {
	JudoCategoryId string
	WeighInAt      string
	StartsAt       string
}

type CreateEventResponse struct {
	Event      *e.Event
	Categories []*e.EventCategory
	Errors     map[string]string
}

type GetEventRequest struct {
	Id string
}

type GetEventResponse struct {
	Event      *e.Event
	Categories []*e.EventCategory
	Errors     map[string]string
}

type ListClubEventsRequest struct {
	ClubId string
	// From filters events with date >= From. Zero value means no filter.
	From     string
	Limit    int
	Page     int
	PageSize int
}

type ListClubEventsResponse struct {
	Events []*e.Event
	Total  int
	Errors map[string]string
}

type UpdateEventCategoryRequest struct {
	JudoCategoryId string
	WeighInAt      string
}

type UpdateEventRequest struct {
	Id                  string
	Title               string
	Description         string
	Location            string
	RegistrationOpenAt  string
	RegistrationCloseAt string
	Date                string
	MaxParticipants     int
	// Categories — when non-nil, replaces all existing event categories.
	// Pass nil to leave categories unchanged.
	Categories []*UpdateEventCategoryRequest
}

func (r UpdateEventRequest) Map() map[string]string {
	m := make(map[string]string)
	if r.Title != "" {
		m["title"] = r.Title
	}
	if r.Description != "" {
		m["description"] = r.Description
	}
	if r.Location != "" {
		m["location"] = r.Location
	}
	if r.RegistrationOpenAt != "" {
		m["registration_open_at"] = r.RegistrationOpenAt
	}
	if r.RegistrationCloseAt != "" {
		m["registration_close_at"] = r.RegistrationCloseAt
	}
	if r.Date != "" {
		m["date"] = r.Date
	}
	return m
}

type UpdateEventResponse struct {
	Event  *e.Event
	Errors map[string]string
}

// Participants

type RegisterParticipantRequest struct {
	EventId         string
	MemberId        string
	Role            string
	EventCategoryId string // required for competitor
}

type RegisterParticipantResponse struct {
	Participant *e.EventParticipant
	Errors      map[string]string
}

type UnregisterParticipantRequest struct {
	EventId  string
	MemberId string
	// Force bypasses the registration period check — used by managers.
	Force bool
}

type GetParticipantsRequest struct {
	EventId string
}

type GetParticipantsResponse struct {
	Participants []*e.EventParticipant
	Errors       map[string]string
}

type UpdateCategoryStatusRequest struct {
	CategoryId string
	Status     string
}

type UpdateCategoryStatusResponse struct {
	Category *e.EventCategory
	Errors   map[string]string
}

// Carpool

type CreateCarpoolOfferRequest struct {
	EventId          string
	MemberId         string
	DepartureAddress string
	AvailableSeats   int
	DepartureAt      string
}

type CreateCarpoolOfferResponse struct {
	Offer  *e.CarpoolOffer
	Errors map[string]string
}

type GetEventCarpoolsRequest struct {
	EventId string
}

type GetEventCarpoolsResponse struct {
	Offers []*e.CarpoolOffer
	Errors map[string]string
}

type MyParticipationRow struct {
	MemberName   string
	Role         string
	CategoryName string // empty for non-competitors
	WeighInAt    string // empty if no weigh-in scheduled
	StartsAt     string // empty if no start time scheduled
}

type CarpoolOfferRow struct {
	Offer          *e.CarpoolOffer
	DriverName     string
	PassengerCount int
	IsOwnOffer     bool // one of the current user's members is the driver
	IsPassenger    bool // one of the current user's members is a passenger in this offer
}

// Judo categories

type ListJudoCategoriesResponse struct {
	Categories []*e.JudoCategory
}

// View DTOs

type EnrichedEventCategory struct {
	EventCat *e.EventCategory
	JudoCat  *e.JudoCategory
}

type AgeGroupRow struct {
	AgeGroup         string
	Gender           string // "man", "woman", "mixed"
	ParticipantCount int
}

type ParticipantRow struct {
	Participant   *e.EventParticipant
	MemberName    string
	CategoryLabel string
}
