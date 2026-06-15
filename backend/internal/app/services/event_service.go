package services

import (
	"bytes"
	"clubmanager/internal/adapters/api/grpc/dto"
	"clubmanager/internal/domain"
	"clubmanager/internal/domain/clubs"
	"clubmanager/internal/domain/events"
	"clubmanager/internal/domain/licences"
	"clubmanager/internal/domain/members"
	"clubmanager/internal/domain/roles"
	"context"
	"encoding/csv"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// ErrValidation is returned by service methods when a business rule is violated.
// HTTP handlers should treat it as a user-facing error, not a 500.
var ErrValidation = fmt.Errorf("validation error")

// EventUpcomingPort is the port for fetching upcoming events for a club.
type EventUpcomingPort interface {
	FindUpcoming(ctx context.Context, clubId string, from time.Time, limit int) ([]*events.Event, error)
}

// EventClubListPort is the port for paginated club event listing.
type EventClubListPort interface {
	FindByClub(ctx context.Context, clubId string, page, pageSize int) ([]*events.Event, int, error)
}

type EventService interface {
	CreateEvent(context.Context, *dto.CreateEventRequest) (*dto.CreateEventResponse, error)
	GetEvent(context.Context, string) (*dto.GetEventResponse, error)
	ListClubEvents(context.Context, *dto.ListClubEventsRequest) (*dto.ListClubEventsResponse, error)
	UpdateEvent(context.Context, *dto.UpdateEventRequest) (*dto.UpdateEventResponse, error)
	CancelEvent(context.Context, string) (bool, error)
	DeleteEvent(context.Context, string) (bool, error)
	OpenEvent(context.Context, string) (bool, error)
	ReopenEvent(context.Context, string) (bool, error)

	RegisterParticipant(context.Context, *dto.RegisterParticipantRequest) (*dto.RegisterParticipantResponse, error)
	UnregisterParticipant(context.Context, *dto.UnregisterParticipantRequest) (bool, error)
	GetParticipants(context.Context, string) (*dto.GetParticipantsResponse, error)
	ExportParticipants(context.Context, string) ([]byte, error)
	UpdateCategoryStatus(context.Context, *dto.UpdateCategoryStatusRequest) (*dto.UpdateCategoryStatusResponse, error)

	CreateCarpoolOffer(context.Context, *dto.CreateCarpoolOfferRequest) (*dto.CreateCarpoolOfferResponse, error)
	JoinCarpoolOffer(context.Context, string, string) (bool, error)
	LeaveCarpoolOffer(context.Context, string, string) (bool, error)
	GetEventCarpools(context.Context, string) (*dto.GetEventCarpoolsResponse, error)
	GetEventCarpoolPassengers(context.Context, string) ([]*events.CarpoolPassenger, error)
	CancelCarpoolOffer(context.Context, string) (bool, error)

	ListJudoCategories(context.Context) (*dto.ListJudoCategoriesResponse, error)
}

type EventServiceConfig struct {
	EventRepo          domain.Repository[events.Event, string]
	EventUpcomingPort  EventUpcomingPort
	EventClubListPort  EventClubListPort
	CategoryRepo       events.EventCategoryRepository
	ParticipantRepo    events.ParticipantRepository
	CarpoolRepo        events.CarpoolRepository
	JudoCatRepo        events.JudoCategoryRepository
	MemberRepo         domain.Repository[members.Member, string]
	LicenceRepo        domain.Repository[licences.Licence, string]
	RoleChecker        roles.RoleChecker
	ClubStatusChecker  clubs.ClubStatusChecker
}

type eventService struct {
	eventRepo         domain.Repository[events.Event, string]
	eventUpcomingPort EventUpcomingPort
	eventClubListPort EventClubListPort
	categoryRepo      events.EventCategoryRepository
	participantRepo   events.ParticipantRepository
	carpoolRepo       events.CarpoolRepository
	judoCatRepo       events.JudoCategoryRepository
	memberRepo        domain.Repository[members.Member, string]
	licenceRepo       domain.Repository[licences.Licence, string]
	roleChecker       roles.RoleChecker
	clubStatusChecker clubs.ClubStatusChecker
}

func NewEventService(cfg EventServiceConfig) *eventService {
	return &eventService{
		eventRepo:         cfg.EventRepo,
		eventUpcomingPort: cfg.EventUpcomingPort,
		eventClubListPort: cfg.EventClubListPort,
		categoryRepo:      cfg.CategoryRepo,
		participantRepo:   cfg.ParticipantRepo,
		carpoolRepo:       cfg.CarpoolRepo,
		judoCatRepo:       cfg.JudoCatRepo,
		memberRepo:        cfg.MemberRepo,
		licenceRepo:       cfg.LicenceRepo,
		roleChecker:       cfg.RoleChecker,
		clubStatusChecker: cfg.ClubStatusChecker,
	}
}

// parseTimestamp handles both RFC3339 ("2006-01-02T15:04:05Z07:00") and the
// PostgreSQL text representation ("2006-01-02 15:04:05-07").
func parseTimestamp(s string) (time.Time, error) {
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t, nil
	}
	return time.Parse("2006-01-02 15:04:05-07", s)
}

// ── Event CRUD ──────────────────────────────────────────────────────────────

func (s *eventService) CreateEvent(ctx context.Context, data *dto.CreateEventRequest) (*dto.CreateEventResponse, error) {
	userId, ok := ctx.Value("user_id").(string)
	if !ok || userId == "" {
		return nil, fmt.Errorf("unauthorized")
	}

	params := map[string]string{
		"title":                data.Title,
		"description":          data.Description,
		"type":                 data.Type,
		"location":             data.Location,
		"registration_open_at": data.RegistrationOpenAt,
		"registration_close_at": data.RegistrationCloseAt,
		"date":                 data.Date,
	}

	event, errs := events.NewEvent(data.ClubId, userId, params)
	if len(errs) > 0 {
		return &dto.CreateEventResponse{Errors: errs}, nil
	}

	// Block event creation if the club is not active.
	if s.clubStatusChecker != nil {
		active, err := s.clubStatusChecker.IsActive(ctx, data.ClubId)
		if err != nil {
			return nil, fmt.Errorf("check club status: %w", err)
		}
		if !active {
			return &dto.CreateEventResponse{
				Errors: map[string]string{"club": "Club is not active. Events cannot be created for a pending or suspended club."},
			}, nil
		}
	}

	if data.MaxParticipants > 0 {
		event.MaxParticipants = data.MaxParticipants
	}

	created, err := s.eventRepo.Save(ctx, event)
	if err != nil {
		return nil, err
	}

	var cats []*events.EventCategory
	for _, cr := range data.Categories {
		jcId, err := uuid.Parse(cr.JudoCategoryId)
		if err != nil {
			return nil, fmt.Errorf("invalid judo_category_id: %w", err)
		}
		cat := &events.EventCategory{
			EventId:        created.Id,
			JudoCategoryId: jcId,
			WeighInAt:      cr.WeighInAt,
			StartsAt:       cr.StartsAt,
			Status:         events.CategoryStatusPending,
		}
		saved, err := s.categoryRepo.Save(ctx, cat)
		if err != nil {
			return nil, err
		}
		cats = append(cats, saved)
	}

	return &dto.CreateEventResponse{Event: created, Categories: cats, Errors: make(map[string]string)}, nil
}

func (s *eventService) GetEvent(ctx context.Context, id string) (*dto.GetEventResponse, error) {
	event, err := s.eventRepo.Find(ctx, id)
	if err != nil {
		return nil, err
	}
	cats, err := s.categoryRepo.FindByEvent(ctx, id)
	if err != nil {
		return nil, err
	}
	return &dto.GetEventResponse{Event: event, Categories: cats, Errors: make(map[string]string)}, nil
}

func (s *eventService) ListClubEvents(ctx context.Context, req *dto.ListClubEventsRequest) (*dto.ListClubEventsResponse, error) {
	if req.From != "" && s.eventUpcomingPort != nil {
		from, err := time.Parse("2006-01-02", req.From)
		if err != nil {
			return &dto.ListClubEventsResponse{
				Errors: map[string]string{"from": "Invalid date format, expected YYYY-MM-DD."},
			}, nil
		}
		limit := req.Limit
		if limit <= 0 {
			limit = 20
		}
		list, err := s.eventUpcomingPort.FindUpcoming(ctx, req.ClubId, from, limit)
		if err != nil {
			return nil, err
		}
		return &dto.ListClubEventsResponse{Events: list, Total: len(list), Errors: make(map[string]string)}, nil
	}

	if s.eventClubListPort != nil {
		page := req.Page
		if page <= 0 {
			page = 1
		}
		pageSize := req.PageSize
		if pageSize <= 0 {
			pageSize = 20
		}
		list, total, err := s.eventClubListPort.FindByClub(ctx, req.ClubId, page, pageSize)
		if err != nil {
			return nil, err
		}
		return &dto.ListClubEventsResponse{Events: list, Total: total, Errors: make(map[string]string)}, nil
	}

	list, err := s.eventRepo.Search(ctx, &domain.SearchParams{
		Fields:    map[string]any{"club_id": req.ClubId},
		Keys:      map[string]bool{"club_id": true},
		Connector: "AND",
	})
	if err != nil {
		return nil, err
	}
	return &dto.ListClubEventsResponse{Events: list, Total: len(list), Errors: make(map[string]string)}, nil
}

func (s *eventService) UpdateEvent(ctx context.Context, data *dto.UpdateEventRequest) (*dto.UpdateEventResponse, error) {
	userId, ok := ctx.Value("user_id").(string)
	if !ok || userId == "" {
		return nil, fmt.Errorf("unauthorized")
	}

	current, err := s.eventRepo.Find(ctx, data.Id)
	if err != nil {
		return nil, err
	}
	if current == nil {
		return &dto.UpdateEventResponse{Errors: map[string]string{"event": "Event not found."}}, nil
	}

	if err := s.requireEventManager(ctx, userId, current); err != nil {
		return nil, err
	}

	updated, errs := current.Update(data.Map())
	if len(errs) > 0 {
		return &dto.UpdateEventResponse{Errors: errs}, nil
	}

	saved, err := s.eventRepo.Save(ctx, updated)
	if err != nil {
		return nil, err
	}

	// Replace categories when provided.
	if data.Categories != nil {
		if err := s.categoryRepo.DeleteByEvent(ctx, data.Id); err != nil {
			return nil, err
		}
		for _, cr := range data.Categories {
			jcId, err := uuid.Parse(cr.JudoCategoryId)
			if err != nil {
				return nil, fmt.Errorf("invalid judo_category_id: %w", err)
			}
			cat := &events.EventCategory{
				EventId:        saved.Id,
				JudoCategoryId: jcId,
				WeighInAt:      cr.WeighInAt,
				Status:         events.CategoryStatusPending,
			}
			if _, err := s.categoryRepo.Save(ctx, cat); err != nil {
				return nil, err
			}
		}
	}

	return &dto.UpdateEventResponse{Event: saved, Errors: make(map[string]string)}, nil
}

func (s *eventService) CancelEvent(ctx context.Context, id string) (bool, error) {
	userId, ok := ctx.Value("user_id").(string)
	if !ok || userId == "" {
		return false, fmt.Errorf("unauthorized")
	}

	event, err := s.eventRepo.Find(ctx, id)
	if err != nil {
		return false, err
	}
	if event == nil {
		return false, fmt.Errorf("event not found")
	}
	if err := s.requireEventManager(ctx, userId, event); err != nil {
		return false, err
	}

	event.Status = events.StatusCancelled
	if _, err := s.eventRepo.Save(ctx, event); err != nil {
		return false, err
	}
	return true, nil
}

func (s *eventService) DeleteEvent(ctx context.Context, id string) (bool, error) {
	userId, ok := ctx.Value("user_id").(string)
	if !ok || userId == "" {
		return false, fmt.Errorf("unauthorized")
	}

	event, err := s.eventRepo.Find(ctx, id)
	if err != nil {
		return false, err
	}
	if event == nil {
		return false, fmt.Errorf("event not found")
	}
	if err := s.requireEventManager(ctx, userId, event); err != nil {
		return false, err
	}

	return s.eventRepo.Delete(ctx, id)
}

func (s *eventService) OpenEvent(ctx context.Context, id string) (bool, error) {
	userId, ok := ctx.Value("user_id").(string)
	if !ok || userId == "" {
		return false, fmt.Errorf("unauthorized")
	}

	event, err := s.eventRepo.Find(ctx, id)
	if err != nil {
		return false, err
	}
	if event == nil {
		return false, fmt.Errorf("event not found")
	}
	if err := s.requireEventManager(ctx, userId, event); err != nil {
		return false, err
	}

	// For competitions, every category must have a weigh-in time set.
	if event.Type == events.TypeCompetition {
		cats, err := s.categoryRepo.FindByEvent(ctx, id)
		if err != nil {
			return false, err
		}
		if len(cats) == 0 {
			return false, fmt.Errorf("%w: une compétition doit avoir au moins une catégorie", ErrValidation)
		}
		judoCats, err := s.judoCatRepo.FindAll(ctx)
		if err != nil {
			return false, err
		}
		judoCatMap := make(map[string]*events.JudoCategory, len(judoCats))
		for _, jc := range judoCats {
			judoCatMap[jc.Id.String()] = jc
		}
		for _, cat := range cats {
			if cat.WeighInAt == "" {
				jc := judoCatMap[cat.JudoCategoryId.String()]
				label := cat.JudoCategoryId.String()
				if jc != nil {
					label = jc.AgeGroup
					if jc.Gender == "man" {
						label += " Masculin"
					} else if jc.Gender == "woman" {
						label += " Féminin"
					}
				}
				return false, fmt.Errorf("%w: heure de pesée manquante pour la catégorie %q", ErrValidation, label)
			}
		}
	}

	event.Status = events.StatusOpen
	if _, err := s.eventRepo.Save(ctx, event); err != nil {
		return false, err
	}
	return true, nil
}

func (s *eventService) ReopenEvent(ctx context.Context, id string) (bool, error) {
	userId, ok := ctx.Value("user_id").(string)
	if !ok || userId == "" {
		return false, fmt.Errorf("unauthorized")
	}

	event, err := s.eventRepo.Find(ctx, id)
	if err != nil {
		return false, err
	}
	if event == nil {
		return false, fmt.Errorf("event not found")
	}
	if err := s.requireEventManager(ctx, userId, event); err != nil {
		return false, err
	}

	event.Status = events.StatusDraft
	if _, err := s.eventRepo.Save(ctx, event); err != nil {
		return false, err
	}
	return true, nil
}

// requireEventManager checks that userId is president, staff, or the event creator.
func (s *eventService) requireEventManager(ctx context.Context, userId string, event *events.Event) error {
	if event.CreatedBy.String() == userId {
		return nil
	}
	isPresident, err := s.roleChecker.HasRole(ctx, userId, event.ClubId.String(), roles.RolePresident)
	if err != nil {
		return err
	}
	if isPresident {
		return nil
	}
	isStaff, err := s.roleChecker.HasRole(ctx, userId, event.ClubId.String(), roles.RoleStaff)
	if err != nil {
		return err
	}
	if isStaff {
		return nil
	}
	isSuperAdmin, err := s.roleChecker.IsSuperAdmin(ctx, userId)
	if err != nil {
		return err
	}
	if isSuperAdmin {
		return nil
	}
	return fmt.Errorf("unauthorized: only the event creator, president or staff can manage this event")
}

// ── Participants ─────────────────────────────────────────────────────────────

func (s *eventService) RegisterParticipant(ctx context.Context, data *dto.RegisterParticipantRequest) (*dto.RegisterParticipantResponse, error) {
	event, err := s.eventRepo.Find(ctx, data.EventId)
	if err != nil {
		return nil, err
	}
	if event == nil {
		return &dto.RegisterParticipantResponse{Errors: map[string]string{"event": "Event not found."}}, nil
	}

	// Event must be open
	if event.Status != events.StatusOpen {
		return &dto.RegisterParticipantResponse{
			Errors: map[string]string{"event": "Event is not open for registration."},
		}, nil
	}

	// Registration period must be active
	now := time.Now()
	tOpen, err := parseTimestamp(event.RegistrationOpenAt)
	if err != nil {
		return nil, fmt.Errorf("invalid event registration_open_at format: %w", err)
	}
	tClose, err := parseTimestamp(event.RegistrationCloseAt)
	if err != nil {
		return nil, fmt.Errorf("invalid event registration_close_at format: %w", err)
	}
	if now.Before(tOpen) || now.After(tClose) {
		return &dto.RegisterParticipantResponse{
			Errors: map[string]string{"event": "Registration period is not active."},
		}, nil
	}

	// Role must be valid for event type
	if !events.IsRoleAllowed(event.Type, data.Role) {
		return &dto.RegisterParticipantResponse{
			Errors: map[string]string{"role": fmt.Sprintf("Role '%s' is not allowed for event type '%s'.", data.Role, event.Type)},
		}, nil
	}

	// Get member to access user_id and club_id
	member, err := s.memberRepo.Find(ctx, data.MemberId)
	if err != nil {
		return nil, err
	}
	if member == nil {
		return &dto.RegisterParticipantResponse{Errors: map[string]string{"member": "Member not found."}}, nil
	}

	// Licence check for competitor and referee in competition/stage
	if events.RequiresLicence(data.Role) {
		if err := s.checkActiveLicence(ctx, data.MemberId, event.Date); err != nil {
			return &dto.RegisterParticipantResponse{
				Errors: map[string]string{"licence": err.Error()},
			}, nil
		}
	}

	// Coach role: must have coach role in the club that organises the event.
	if data.Role == events.RoleCoach {
		ok, err := s.roleChecker.HasRole(ctx, member.UserId.String(), event.ClubId.String(), roles.RoleCoach)
		if err != nil {
			return nil, err
		}
		if !ok {
			return &dto.RegisterParticipantResponse{
				Errors: map[string]string{"role": "Member does not have the coach role in this club."},
			}, nil
		}
	}

	// Competitor in competition must have a category
	var catId *uuid.UUID
	if data.Role == events.RoleCompetitor {
		if event.Type == events.TypeCompetition {
			if data.EventCategoryId == "" {
				return &dto.RegisterParticipantResponse{
					Errors: map[string]string{"event_category_id": "Competitors must select a category."},
				}, nil
			}
			id, err := uuid.Parse(data.EventCategoryId)
			if err != nil {
				return &dto.RegisterParticipantResponse{
					Errors: map[string]string{"event_category_id": "Invalid category ID."},
				}, nil
			}
			catId = &id
		}
	}

	memberId, _ := uuid.Parse(data.MemberId)
	eventId, _ := uuid.Parse(data.EventId)

	p := &events.EventParticipant{
		EventId:         eventId,
		MemberId:        memberId,
		Role:            data.Role,
		EventCategoryId: catId,
	}

	// Register atomically: blocks on duplicate or capacity
	registered, err := s.participantRepo.Register(ctx, p, event.MaxParticipants)
	if err != nil {
		return nil, err
	}
	if registered == nil {
		// Distinguish cause: check if already registered or at capacity
		existing, err := s.participantRepo.FindByMember(ctx, data.EventId, data.MemberId)
		if err != nil {
			return nil, err
		}
		if existing != nil {
			return &dto.RegisterParticipantResponse{
				Errors: map[string]string{"member": "Member is already registered for this event."},
			}, nil
		}
		return &dto.RegisterParticipantResponse{
			Errors: map[string]string{"event": "Event has reached maximum capacity."},
		}, nil
	}
	return &dto.RegisterParticipantResponse{Participant: registered, Errors: make(map[string]string)}, nil
}

func (s *eventService) checkActiveLicence(ctx context.Context, memberId, eventDate string) error {
	list, err := s.licenceRepo.Search(ctx, &domain.SearchParams{
		Fields:    map[string]any{"member_id": memberId, "status": "active"},
		Keys:      map[string]bool{"member_id": true, "status": true},
		Connector: "AND",
	})
	if err != nil {
		return err
	}

	tEventFull, err := time.Parse(time.RFC3339, eventDate)
	if err != nil {
		return fmt.Errorf("invalid event date format")
	}
	// Truncate to date-only so a licence expiring on the event day is still valid.
	tEvent := time.Date(tEventFull.Year(), tEventFull.Month(), tEventFull.Day(), 0, 0, 0, 0, time.UTC)

	for _, l := range list {
		tUntil, err := time.Parse("2006-01-02", l.ValidUntil)
		if err != nil {
			continue
		}
		if !tUntil.Before(tEvent) {
			return nil // valid licence found
		}
	}
	return fmt.Errorf("an active licence valid for the event date is required")
}

func (s *eventService) UnregisterParticipant(ctx context.Context, data *dto.UnregisterParticipantRequest) (bool, error) {
	userId, ok := ctx.Value("user_id").(string)
	if !ok || userId == "" {
		return false, fmt.Errorf("unauthorized")
	}

	member, err := s.memberRepo.Find(ctx, data.MemberId)
	if err != nil {
		return false, err
	}
	if member == nil || member.UserId.String() != userId {
		return false, fmt.Errorf("unauthorized: you can only unregister yourself")
	}

	event, err := s.eventRepo.Find(ctx, data.EventId)
	if err != nil {
		return false, err
	}
	if event == nil {
		return false, fmt.Errorf("event not found")
	}

	if !data.Force {
		now := time.Now()
		tClose, err := parseTimestamp(event.RegistrationCloseAt)
		if err != nil {
			return false, fmt.Errorf("invalid event registration_close_at format: %w", err)
		}
		if now.After(tClose) {
			return false, fmt.Errorf("registration period is closed, cannot unregister")
		}
	}

	return s.participantRepo.Unregister(ctx, data.EventId, data.MemberId)
}

func (s *eventService) GetParticipants(ctx context.Context, eventId string) (*dto.GetParticipantsResponse, error) {
	list, err := s.participantRepo.FindByEvent(ctx, eventId)
	if err != nil {
		return nil, err
	}
	return &dto.GetParticipantsResponse{Participants: list, Errors: make(map[string]string)}, nil
}

func (s *eventService) ExportParticipants(ctx context.Context, eventId string) ([]byte, error) {
	participants, err := s.participantRepo.FindByEvent(ctx, eventId)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	w := csv.NewWriter(&buf)

	if err := w.Write([]string{"Nom", "Prénom", "Date de naissance", "Rôle", "Catégorie"}); err != nil {
		return nil, fmt.Errorf("failed to write CSV header: %w", err)
	}

	for _, p := range participants {
		member, err := s.memberRepo.Find(ctx, p.MemberId.String())
		if err != nil || member == nil {
			continue
		}

		category := ""
		if p.EventCategoryId != nil {
			cat, err := s.categoryRepo.FindById(ctx, p.EventCategoryId.String())
			if err == nil && cat != nil {
				jc, err := s.judoCatRepo.FindById(ctx, cat.JudoCategoryId.String())
				if err == nil && jc != nil {
					category = fmt.Sprintf("%s %s", jc.AgeGroup, jc.WeightLabel)
				}
			}
		}

		if err := w.Write([]string{
			member.Lastname,
			member.Firstname,
			member.Birthdate,
			p.Role,
			category,
		}); err != nil {
			return nil, fmt.Errorf("failed to write CSV row: %w", err)
		}
	}

	w.Flush()
	return buf.Bytes(), w.Error()
}

func (s *eventService) UpdateCategoryStatus(ctx context.Context, data *dto.UpdateCategoryStatusRequest) (*dto.UpdateCategoryStatusResponse, error) {
	validStatuses := map[string]bool{
		events.CategoryStatusPending:    true,
		events.CategoryStatusWeighIn:    true,
		events.CategoryStatusInProgress: true,
		events.CategoryStatusCompleted:  true,
	}
	if !validStatuses[data.Status] {
		return &dto.UpdateCategoryStatusResponse{
			Errors: map[string]string{"status": "Invalid category status."},
		}, nil
	}

	cat, err := s.categoryRepo.UpdateStatus(ctx, data.CategoryId, data.Status)
	if err != nil {
		return nil, err
	}
	return &dto.UpdateCategoryStatusResponse{Category: cat, Errors: make(map[string]string)}, nil
}

// ── Carpool ──────────────────────────────────────────────────────────────────

func (s *eventService) CreateCarpoolOffer(ctx context.Context, data *dto.CreateCarpoolOfferRequest) (*dto.CreateCarpoolOfferResponse, error) {
	// Must be registered for the event
	existing, err := s.participantRepo.FindByMember(ctx, data.EventId, data.MemberId)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return &dto.CreateCarpoolOfferResponse{
			Errors: map[string]string{"member": "Must be registered for the event to offer a carpool."},
		}, nil
	}

	// One offer per member per event (enforced by UNIQUE constraint, but check first for a clean error)
	existingOffer, err := s.carpoolRepo.FindOfferByMember(ctx, data.EventId, data.MemberId)
	if err != nil {
		return nil, err
	}
	if existingOffer != nil {
		return &dto.CreateCarpoolOfferResponse{
			Errors: map[string]string{"offer": "You already have a carpool offer for this event."},
		}, nil
	}

	if data.AvailableSeats <= 0 {
		return &dto.CreateCarpoolOfferResponse{
			Errors: map[string]string{"available_seats": "Available seats must be greater than 0."},
		}, nil
	}

	eventId, _ := uuid.Parse(data.EventId)
	memberId, _ := uuid.Parse(data.MemberId)

	offer := &events.CarpoolOffer{
		EventId:          eventId,
		MemberId:         memberId,
		DepartureAddress: data.DepartureAddress,
		AvailableSeats:   data.AvailableSeats,
		DepartureAt:      data.DepartureAt,
	}

	created, err := s.carpoolRepo.CreateOffer(ctx, offer)
	if err != nil {
		return nil, err
	}
	return &dto.CreateCarpoolOfferResponse{Offer: created, Errors: make(map[string]string)}, nil
}

func (s *eventService) JoinCarpoolOffer(ctx context.Context, offerId, memberId string) (bool, error) {
	offer, err := s.carpoolRepo.FindOfferById(ctx, offerId)
	if err != nil {
		return false, err
	}
	if offer == nil {
		return false, fmt.Errorf("carpool offer not found")
	}

	// Member must be registered for the event
	registered, err := s.participantRepo.FindByMember(ctx, offer.EventId.String(), memberId)
	if err != nil {
		return false, err
	}
	if registered == nil {
		return false, fmt.Errorf("must be registered for the event to join a carpool")
	}

	// Member must not be offering a carpool themselves
	existingOffer, err := s.carpoolRepo.FindOfferByMember(ctx, offer.EventId.String(), memberId)
	if err != nil {
		return false, err
	}
	if existingOffer != nil {
		return false, fmt.Errorf("cannot join a carpool when you have an active offer")
	}

	// AddPassenger now handles duplicate passenger and capacity atomically
	added, err := s.carpoolRepo.AddPassenger(ctx, offerId, memberId)
	if err != nil {
		return false, err
	}
	if !added {
		// Distinguish: already passenger vs no seats
		isPassenger, err := s.carpoolRepo.IsPassenger(ctx, offer.EventId.String(), memberId)
		if err != nil {
			return false, err
		}
		if isPassenger {
			return false, fmt.Errorf("already a passenger for this event")
		}
		return false, fmt.Errorf("no seats available in this carpool")
	}
	return true, nil
}

func (s *eventService) LeaveCarpoolOffer(ctx context.Context, offerId, memberId string) (bool, error) {
	userId, ok := ctx.Value("user_id").(string)
	if !ok || userId == "" {
		return false, fmt.Errorf("unauthorized")
	}

	member, err := s.memberRepo.Find(ctx, memberId)
	if err != nil {
		return false, err
	}
	if member == nil || member.UserId.String() != userId {
		return false, fmt.Errorf("unauthorized: you can only leave your own carpool")
	}

	return s.carpoolRepo.RemovePassenger(ctx, offerId, memberId)
}

func (s *eventService) GetEventCarpools(ctx context.Context, eventId string) (*dto.GetEventCarpoolsResponse, error) {
	list, err := s.carpoolRepo.FindOffersByEvent(ctx, eventId)
	if err != nil {
		return nil, err
	}
	return &dto.GetEventCarpoolsResponse{Offers: list, Errors: make(map[string]string)}, nil
}

func (s *eventService) GetEventCarpoolPassengers(ctx context.Context, eventId string) ([]*events.CarpoolPassenger, error) {
	return s.carpoolRepo.FindPassengersByEvent(ctx, eventId)
}

func (s *eventService) CancelCarpoolOffer(ctx context.Context, offerId string) (bool, error) {
	userId, ok := ctx.Value("user_id").(string)
	if !ok || userId == "" {
		return false, fmt.Errorf("unauthorized")
	}

	offer, err := s.carpoolRepo.FindOfferById(ctx, offerId)
	if err != nil {
		return false, err
	}
	if offer == nil {
		return false, fmt.Errorf("carpool offer not found")
	}

	// Verify ownership: offer.MemberId → member.UserId must match context userId
	member, err := s.memberRepo.Find(ctx, offer.MemberId.String())
	if err != nil {
		return false, err
	}
	if member == nil || member.UserId.String() != userId {
		return false, fmt.Errorf("unauthorized: you can only cancel your own carpool offers")
	}

	return s.carpoolRepo.DeleteOffer(ctx, offerId)
}

// ── Judo categories ──────────────────────────────────────────────────────────

func (s *eventService) ListJudoCategories(ctx context.Context) (*dto.ListJudoCategoriesResponse, error) {
	list, err := s.judoCatRepo.FindAll(ctx)
	if err != nil {
		return nil, err
	}
	return &dto.ListJudoCategoriesResponse{Categories: list}, nil
}

