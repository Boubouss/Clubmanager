package events

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

const (
	TypeCompetition = "competition"
	TypeStage       = "stage"
	TypeAutre       = "autre"

	StatusDraft     = "draft"
	StatusOpen      = "open"
	StatusCancelled = "cancelled"

	RoleCompetitor  = "competitor"
	RoleCoach       = "coach"
	RoleReferee     = "referee"
	RoleAccompanist = "accompanist"

	CategoryStatusPending    = "pending"
	CategoryStatusWeighIn    = "weigh_in"
	CategoryStatusInProgress = "in_progress"
	CategoryStatusCompleted  = "completed"
)

var allowedRolesByType = map[string]map[string]bool{
	TypeCompetition: {RoleCompetitor: true, RoleCoach: true, RoleReferee: true, RoleAccompanist: true},
	TypeStage:       {RoleCompetitor: true, RoleCoach: true, RoleAccompanist: true},
	TypeAutre:       {RoleAccompanist: true},
}

// IsRoleAllowed returns true if the role is valid for the given event type.
func IsRoleAllowed(eventType, role string) bool {
	allowed, ok := allowedRolesByType[eventType]
	if !ok {
		return false
	}
	return allowed[role]
}

// RequiresLicence returns true if the role requires an active licence.
func RequiresLicence(role string) bool {
	return role == RoleCompetitor || role == RoleReferee
}

type Event struct {
	Id                  uuid.UUID
	ClubId              uuid.UUID
	CreatedBy           uuid.UUID
	Title               string
	Description         string
	Type                string
	Status              string
	Location            string
	RegistrationOpenAt  string
	RegistrationCloseAt string
	Date                string
	MaxParticipants     int
}

type EventCategory struct {
	Id             uuid.UUID
	EventId        uuid.UUID
	JudoCategoryId uuid.UUID
	WeighInAt      string
	StartsAt       string
	Status         string
}

type JudoCategory struct {
	Id          uuid.UUID
	AgeGroup    string
	Gender      string
	AgeMin      int
	AgeMax      *int
	WeightLabel string
	WeightMax   *float64
}

type EventParticipant struct {
	Id              uuid.UUID
	EventId         uuid.UUID
	MemberId        uuid.UUID
	Role            string
	EventCategoryId *uuid.UUID
	RegisteredAt    string
}

type CarpoolOffer struct {
	Id               uuid.UUID
	EventId          uuid.UUID
	MemberId         uuid.UUID
	DepartureAddress string
	AvailableSeats   int
	DepartureAt      string
}

type CarpoolPassenger struct {
	Id       uuid.UUID
	OfferId  uuid.UUID
	MemberId uuid.UUID
}

func NewEvent(clubId, createdBy string, data map[string]string) (*Event, map[string]string) {
	errs := make(map[string]string)

	cid, err := uuid.Parse(clubId)
	if err != nil {
		errs["club_id"] = "Invalid club ID."
	}
	uid, err := uuid.Parse(createdBy)
	if err != nil {
		errs["created_by"] = "Invalid user ID."
	}

	title := data["title"]
	if title == "" {
		errs["title"] = "Title is required."
	}

	location := data["location"]
	if location == "" {
		errs["location"] = "Location is required."
	}

	eventType := data["type"]
	if _, ok := allowedRolesByType[eventType]; !ok {
		errs["type"] = fmt.Sprintf("Invalid event type. Must be one of: %s, %s, %s.", TypeCompetition, TypeStage, TypeAutre)
	}

	regOpen := data["registration_open_at"]
	regClose := data["registration_close_at"]
	eventDate := data["date"]

	if err := validateEventDates(regOpen, regClose, eventDate, errs); err != nil {
		// errors already added to errs
	}

	if len(errs) > 0 {
		return nil, errs
	}

	return &Event{
		ClubId:              cid,
		CreatedBy:           uid,
		Title:               title,
		Description:         data["description"],
		Type:                eventType,
		Status:              StatusDraft,
		Location:            location,
		RegistrationOpenAt:  regOpen,
		RegistrationCloseAt: regClose,
		Date:                eventDate,
	}, errs
}

func (e Event) Update(data map[string]string) (*Event, map[string]string) {
	errs := make(map[string]string)
	updated := e

	if v, ok := data["title"]; ok && v != "" {
		updated.Title = v
	}
	if v, ok := data["description"]; ok {
		updated.Description = v
	}
	if v, ok := data["location"]; ok && v != "" {
		updated.Location = v
	}
	if v, ok := data["type"]; ok && v != "" {
		if _, valid := allowedRolesByType[v]; !valid {
			errs["type"] = fmt.Sprintf("Invalid event type. Must be one of: %s, %s, %s.", TypeCompetition, TypeStage, TypeAutre)
		} else {
			updated.Type = v
		}
	}
	if v, ok := data["max_participants"]; ok && v != "" {
		var n int
		if _, err := fmt.Sscanf(v, "%d", &n); err != nil || n <= 0 {
			errs["max_participants"] = "max_participants must be a positive integer."
		} else {
			updated.MaxParticipants = n
		}
	}

	regOpen := firstNonEmpty(data["registration_open_at"], e.RegistrationOpenAt)
	regClose := firstNonEmpty(data["registration_close_at"], e.RegistrationCloseAt)
	eventDate := firstNonEmpty(data["date"], e.Date)

	if err := validateEventDates(regOpen, regClose, eventDate, errs); err == nil {
		updated.RegistrationOpenAt = regOpen
		updated.RegistrationCloseAt = regClose
		updated.Date = eventDate
	}

	if len(errs) > 0 {
		return nil, errs
	}
	return &updated, errs
}

func validateEventDates(regOpen, regClose, eventDate string, errs map[string]string) error {
	if regOpen == "" {
		errs["registration_open_at"] = "Registration open date is required."
		return fmt.Errorf("missing date")
	}
	if regClose == "" {
		errs["registration_close_at"] = "Registration close date is required."
		return fmt.Errorf("missing date")
	}
	if eventDate == "" {
		errs["date"] = "Event date is required."
		return fmt.Errorf("missing date")
	}

	tOpen, err := time.Parse(time.RFC3339, regOpen)
	if err != nil {
		errs["registration_open_at"] = "Invalid date format. Use RFC3339."
		return err
	}
	tClose, err := time.Parse(time.RFC3339, regClose)
	if err != nil {
		errs["registration_close_at"] = "Invalid date format. Use RFC3339."
		return err
	}
	tDate, err := time.Parse(time.RFC3339, eventDate)
	if err != nil {
		errs["date"] = "Invalid date format. Use RFC3339."
		return err
	}

	hasErr := false
	if !tClose.After(tOpen) {
		errs["registration_close_at"] = "Registration close date must be after open date."
		hasErr = true
	}
	if !tDate.After(tOpen) {
		errs["date"] = "Event date must be after registration open date."
		hasErr = true
	}
	if hasErr {
		return fmt.Errorf("invalid dates")
	}
	return nil
}

func firstNonEmpty(a, b string) string {
	if a != "" {
		return a
	}
	return b
}
