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
}

type ListClubEventsResponse struct {
	Events []*e.Event
	Errors map[string]string
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

// Judo categories

type ListJudoCategoriesResponse struct {
	Categories []*e.JudoCategory
}
