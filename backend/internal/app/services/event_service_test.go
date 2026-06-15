package services

import (
	"clubmanager/internal/adapters/api/grpc/dto"
	"clubmanager/internal/domain"
	"clubmanager/internal/domain/events"
	"clubmanager/internal/domain/licences"
	"clubmanager/internal/domain/members"
	"clubmanager/internal/domain/roles"
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// ── Mock repositories ────────────────────────────────────────────────────────

type mockEventRepo struct {
	saveFunc   func(context.Context, *events.Event) (*events.Event, error)
	findFunc   func(context.Context, string) (*events.Event, error)
	searchFunc func(context.Context, *domain.SearchParams) ([]*events.Event, error)
	deleteFunc func(context.Context, string) (bool, error)
}

func (m mockEventRepo) Save(ctx context.Context, e *events.Event) (*events.Event, error) {
	return m.saveFunc(ctx, e)
}
func (m mockEventRepo) Find(ctx context.Context, id string) (*events.Event, error) {
	return m.findFunc(ctx, id)
}
func (m mockEventRepo) Search(ctx context.Context, p *domain.SearchParams) ([]*events.Event, error) {
	return m.searchFunc(ctx, p)
}
func (m mockEventRepo) Delete(ctx context.Context, id string) (bool, error) {
	return m.deleteFunc(ctx, id)
}

type mockCategoryRepo struct {
	saveFunc         func(context.Context, *events.EventCategory) (*events.EventCategory, error)
	findByEventFunc  func(context.Context, string) ([]*events.EventCategory, error)
	findByIdFunc     func(context.Context, string) (*events.EventCategory, error)
	updateStatusFunc func(context.Context, string, string) (*events.EventCategory, error)
	deleteByEventFunc func(context.Context, string) error
}

func (m mockCategoryRepo) Save(ctx context.Context, c *events.EventCategory) (*events.EventCategory, error) {
	return m.saveFunc(ctx, c)
}
func (m mockCategoryRepo) FindByEvent(ctx context.Context, id string) ([]*events.EventCategory, error) {
	return m.findByEventFunc(ctx, id)
}
func (m mockCategoryRepo) FindById(ctx context.Context, id string) (*events.EventCategory, error) {
	return m.findByIdFunc(ctx, id)
}
func (m mockCategoryRepo) UpdateStatus(ctx context.Context, id, status string) (*events.EventCategory, error) {
	return m.updateStatusFunc(ctx, id, status)
}
func (m mockCategoryRepo) DeleteByEvent(ctx context.Context, id string) error {
	return m.deleteByEventFunc(ctx, id)
}

type mockParticipantRepo struct {
	registerFunc    func(context.Context, *events.EventParticipant, int) (*events.EventParticipant, error)
	unregisterFunc  func(context.Context, string, string) (bool, error)
	findByEventFunc func(context.Context, string) ([]*events.EventParticipant, error)
	findByMemberFunc func(context.Context, string, string) (*events.EventParticipant, error)
	countByEventFunc func(context.Context, string) (int, error)
}

func (m mockParticipantRepo) Register(ctx context.Context, p *events.EventParticipant, max int) (*events.EventParticipant, error) {
	return m.registerFunc(ctx, p, max)
}
func (m mockParticipantRepo) Unregister(ctx context.Context, eventId, memberId string) (bool, error) {
	return m.unregisterFunc(ctx, eventId, memberId)
}
func (m mockParticipantRepo) FindByEvent(ctx context.Context, eventId string) ([]*events.EventParticipant, error) {
	return m.findByEventFunc(ctx, eventId)
}
func (m mockParticipantRepo) FindByMember(ctx context.Context, eventId, memberId string) (*events.EventParticipant, error) {
	return m.findByMemberFunc(ctx, eventId, memberId)
}
func (m mockParticipantRepo) CountByEvent(ctx context.Context, eventId string) (int, error) {
	return m.countByEventFunc(ctx, eventId)
}

type mockCarpoolRepo struct {
	createOfferFunc          func(context.Context, *events.CarpoolOffer) (*events.CarpoolOffer, error)
	findOfferByIdFunc        func(context.Context, string) (*events.CarpoolOffer, error)
	findOffersByEventFunc    func(context.Context, string) ([]*events.CarpoolOffer, error)
	findOfferByMemberFunc    func(context.Context, string, string) (*events.CarpoolOffer, error)
	deleteOfferFunc          func(context.Context, string) (bool, error)
	addPassengerFunc         func(context.Context, string, string) (bool, error)
	removePassengerFunc      func(context.Context, string, string) (bool, error)
	countPassengersFunc      func(context.Context, string) (int, error)
	isPassengerFunc          func(context.Context, string, string) (bool, error)
	findPassengersByEventFunc func(context.Context, string) ([]*events.CarpoolPassenger, error)
}

func (m mockCarpoolRepo) CreateOffer(ctx context.Context, o *events.CarpoolOffer) (*events.CarpoolOffer, error) {
	return m.createOfferFunc(ctx, o)
}
func (m mockCarpoolRepo) FindOfferById(ctx context.Context, id string) (*events.CarpoolOffer, error) {
	return m.findOfferByIdFunc(ctx, id)
}
func (m mockCarpoolRepo) FindOffersByEvent(ctx context.Context, id string) ([]*events.CarpoolOffer, error) {
	return m.findOffersByEventFunc(ctx, id)
}
func (m mockCarpoolRepo) FindOfferByMember(ctx context.Context, eventId, memberId string) (*events.CarpoolOffer, error) {
	return m.findOfferByMemberFunc(ctx, eventId, memberId)
}
func (m mockCarpoolRepo) DeleteOffer(ctx context.Context, id string) (bool, error) {
	return m.deleteOfferFunc(ctx, id)
}
func (m mockCarpoolRepo) AddPassenger(ctx context.Context, offerId, memberId string) (bool, error) {
	return m.addPassengerFunc(ctx, offerId, memberId)
}
func (m mockCarpoolRepo) RemovePassenger(ctx context.Context, offerId, memberId string) (bool, error) {
	return m.removePassengerFunc(ctx, offerId, memberId)
}
func (m mockCarpoolRepo) CountPassengers(ctx context.Context, offerId string) (int, error) {
	return m.countPassengersFunc(ctx, offerId)
}
func (m mockCarpoolRepo) IsPassenger(ctx context.Context, eventId, memberId string) (bool, error) {
	return m.isPassengerFunc(ctx, eventId, memberId)
}
func (m mockCarpoolRepo) FindPassengersByEvent(ctx context.Context, eventId string) ([]*events.CarpoolPassenger, error) {
	if m.findPassengersByEventFunc != nil {
		return m.findPassengersByEventFunc(ctx, eventId)
	}
	return nil, nil
}

type mockJudoCatRepo struct {
	findAllFunc  func(context.Context) ([]*events.JudoCategory, error)
	findByIdFunc func(context.Context, string) (*events.JudoCategory, error)
}

func (m mockJudoCatRepo) FindAll(ctx context.Context) ([]*events.JudoCategory, error) {
	return m.findAllFunc(ctx)
}
func (m mockJudoCatRepo) FindById(ctx context.Context, id string) (*events.JudoCategory, error) {
	return m.findByIdFunc(ctx, id)
}

// ── Test fixtures ────────────────────────────────────────────────────────────

var (
	eventUUID    = uuid.MustParse("00000000-0000-0000-0000-000000000010")
	eventCreator = uuid.MustParse("00000000-0000-0000-0000-000000000001") // same as memberUser
	eventClubId  = uuid.MustParse("00000000-0000-0000-0000-000000000002") // same as memberClub
	offerUUID    = uuid.MustParse("00000000-0000-0000-0000-000000000050")

	// openEvent: registration period is active relative to 2026-06-08
	openEvent = &events.Event{
		Id:                  eventUUID,
		ClubId:              eventClubId,
		CreatedBy:           eventCreator,
		Title:               "Championnat",
		Type:                events.TypeCompetition,
		Status:              events.StatusOpen,
		Location:            "Dojo Paris",
		RegistrationOpenAt:  "2026-05-01T00:00:00Z",
		RegistrationCloseAt: "2026-12-31T00:00:00Z",
		Date:                "2027-01-15T09:00:00Z",
		MaxParticipants:     0,
	}

	// eventMember: a member whose UserId matches eventCreator
	eventMember = &members.Member{
		Id:        memberUUID, // reuse from member_service_test.go
		UserId:    eventCreator,
		Firstname: "Jean",
		Lastname:  "Dupont",
		Birthdate: "2000-06-15",
		Gender:    "man",
		IsPrimary: true,
	}

	activeLicence = &licences.Licence{
		Id:            licenceUUID, // reuse from licence_service_test.go
		MemberId:      memberUUID,
		LicenceNumber: "LIC-2026-001",
		ValidFrom:     "2026-01-01",
		ValidUntil:    "2027-12-31",
		Status:        "active",
	}

	testParticipant = &events.EventParticipant{
		Id:       uuid.MustParse("00000000-0000-0000-0000-000000000060"),
		EventId:  eventUUID,
		MemberId: memberUUID,
		Role:     events.RoleAccompanist,
	}

	testCarpoolOffer = &events.CarpoolOffer{
		Id:               offerUUID,
		EventId:          eventUUID,
		MemberId:         memberUUID,
		DepartureAddress: "10 rue de la Paix",
		AvailableSeats:   3,
		DepartureAt:      "2027-01-15T07:00:00Z",
	}
)

// ── Helpers ───────────────────────────────────────────────────────────────────

func defaultEventSvc(
	eRepo domain.Repository[events.Event, string],
	cRepo events.EventCategoryRepository,
	pRepo events.ParticipantRepository,
	carpRepo events.CarpoolRepository,
	jRepo events.JudoCategoryRepository,
	mRepo domain.Repository[members.Member, string],
	lRepo domain.Repository[licences.Licence, string],
	rc roles.RoleChecker,
) *eventService {
	return NewEventService(EventServiceConfig{
		EventRepo:       eRepo,
		CategoryRepo:    cRepo,
		ParticipantRepo: pRepo,
		CarpoolRepo:     carpRepo,
		JudoCatRepo:     jRepo,
		MemberRepo:      mRepo,
		LicenceRepo:     lRepo,
		RoleChecker:     rc,
	})
}

// noopCategory returns a categoryRepo whose methods are never expected to be called.
func noopCategory() mockCategoryRepo {
	return mockCategoryRepo{
		findByEventFunc: func(_ context.Context, _ string) ([]*events.EventCategory, error) { return nil, nil },
	}
}

// noopCarpool returns a carpoolRepo whose methods are never expected to be called.
func noopCarpool() mockCarpoolRepo { return mockCarpoolRepo{} }

// noopJudo returns a judoCatRepo whose methods are never expected to be called.
func noopJudo() mockJudoCatRepo { return mockJudoCatRepo{} }

// noopRoleChecker returns a RoleChecker that always denies.
func noopRoleChecker() mockRoleChecker {
	return mockRoleChecker{
		hasRoleFunc:      func(_ context.Context, _, _ string, _ ...string) (bool, error) { return false, nil },
		isSuperAdminFunc: func(_ context.Context, _ string) (bool, error) { return false, nil },
	}
}

// creatorRoleChecker returns a RoleChecker that grants president to the given userId.
func presidentRoleChecker(presidentId string) mockRoleChecker {
	return mockRoleChecker{
		hasRoleFunc: func(_ context.Context, userId, _ string, r ...string) (bool, error) {
			if userId == presidentId {
				for _, role := range r {
					if role == roles.RolePresident {
						return true, nil
					}
				}
			}
			return false, nil
		},
		isSuperAdminFunc: func(_ context.Context, _ string) (bool, error) { return false, nil },
	}
}

// ── RegisterParticipant ──────────────────────────────────────────────────────

func TestRegisterParticipant(t *testing.T) {
	otherUserId := "00000000-0000-0000-0000-000000000099"

	tests := []struct {
		name           string
		req            *dto.RegisterParticipantRequest
		event          *events.Event
		eventErr       error
		member         *members.Member
		memberErr      error
		licences       []*licences.Licence
		licenceErr     error
		hasRole        bool
		hasRoleErr     error
		registered     *events.EventParticipant // result of Register()
		registerErr    error
		existingPart   *events.EventParticipant // FindByMember when register returns nil
		wantErrors     bool
		wantErr        bool
		wantErrKey     string
	}{
		{
			name:       "Event not found",
			req:        &dto.RegisterParticipantRequest{EventId: eventUUID.String(), MemberId: memberUUID.String(), Role: events.RoleAccompanist},
			event:      nil,
			wantErrors: true,
			wantErrKey: "event",
		},
		{
			name:       "Event not open (draft)",
			req:        &dto.RegisterParticipantRequest{EventId: eventUUID.String(), MemberId: memberUUID.String(), Role: events.RoleAccompanist},
			event:      &events.Event{Id: eventUUID, Status: events.StatusDraft, Type: events.TypeCompetition, RegistrationOpenAt: "2026-05-01T00:00:00Z", RegistrationCloseAt: "2026-12-31T00:00:00Z", Date: "2027-01-15T09:00:00Z"},
			wantErrors: true,
			wantErrKey: "event",
		},
		{
			name: "Registration period not active (closed)",
			req:  &dto.RegisterParticipantRequest{EventId: eventUUID.String(), MemberId: memberUUID.String(), Role: events.RoleAccompanist},
			event: &events.Event{
				Id: eventUUID, Status: events.StatusOpen, Type: events.TypeCompetition,
				RegistrationOpenAt:  "2025-01-01T00:00:00Z",
				RegistrationCloseAt: "2025-06-01T00:00:00Z", // past
				Date:                "2027-01-15T09:00:00Z",
			},
			wantErrors: true,
			wantErrKey: "event",
		},
		{
			name:       "Role not allowed for event type (autre + competitor)",
			req:        &dto.RegisterParticipantRequest{EventId: eventUUID.String(), MemberId: memberUUID.String(), Role: events.RoleCompetitor},
			event:      &events.Event{Id: eventUUID, Status: events.StatusOpen, Type: events.TypeAutre, RegistrationOpenAt: "2026-05-01T00:00:00Z", RegistrationCloseAt: "2026-12-31T00:00:00Z", Date: "2027-01-15T09:00:00Z"},
			wantErrors: true,
			wantErrKey: "role",
		},
		{
			name:       "Member not found",
			req:        &dto.RegisterParticipantRequest{EventId: eventUUID.String(), MemberId: memberUUID.String(), Role: events.RoleAccompanist},
			event:      openEvent,
			member:     nil,
			wantErrors: true,
			wantErrKey: "member",
		},
		{
			name:       "Licence required but none active",
			req:        &dto.RegisterParticipantRequest{EventId: eventUUID.String(), MemberId: memberUUID.String(), Role: events.RoleCompetitor},
			event:      openEvent,
			member:     eventMember,
			licences:   []*licences.Licence{},
			wantErrors: true,
			wantErrKey: "licence",
		},
		{
			name:       "Coach without coach role in club",
			req:        &dto.RegisterParticipantRequest{EventId: eventUUID.String(), MemberId: memberUUID.String(), Role: events.RoleCoach},
			event:      openEvent,
			member:     eventMember,
			hasRole:    false,
			wantErrors: true,
			wantErrKey: "role",
		},
		{
			name:         "Already registered (atomic block + FindByMember returns existing)",
			req:          &dto.RegisterParticipantRequest{EventId: eventUUID.String(), MemberId: memberUUID.String(), Role: events.RoleAccompanist},
			event:        openEvent,
			member:       eventMember,
			registered:   nil,
			existingPart: testParticipant,
			wantErrors:   true,
			wantErrKey:   "member",
		},
		{
			name:         "Capacity reached (atomic block + FindByMember returns nil)",
			req:          &dto.RegisterParticipantRequest{EventId: eventUUID.String(), MemberId: memberUUID.String(), Role: events.RoleAccompanist},
			event:        openEvent,
			member:       eventMember,
			registered:   nil,
			existingPart: nil,
			wantErrors:   true,
			wantErrKey:   "event",
		},
		{
			name:       "Valid accompanist registration",
			req:        &dto.RegisterParticipantRequest{EventId: eventUUID.String(), MemberId: memberUUID.String(), Role: events.RoleAccompanist},
			event:      openEvent,
			member:     eventMember,
			registered: testParticipant,
			wantErrors: false,
		},
		{
			name:       "Valid competitor registration with active licence",
			req:        &dto.RegisterParticipantRequest{EventId: eventUUID.String(), MemberId: memberUUID.String(), Role: events.RoleCompetitor, EventCategoryId: "00000000-0000-0000-0000-000000000070"},
			event:      openEvent,
			member:     eventMember,
			licences:   []*licences.Licence{activeLicence},
			registered: testParticipant,
			wantErrors: false,
		},
		{
			name:       "Repository error on Find event",
			req:        &dto.RegisterParticipantRequest{EventId: eventUUID.String(), MemberId: memberUUID.String(), Role: events.RoleAccompanist},
			eventErr:   errors.New("db error"),
			wantErr:    true,
		},
		{
			name:      "Repository error on Find member",
			req:       &dto.RegisterParticipantRequest{EventId: eventUUID.String(), MemberId: memberUUID.String(), Role: events.RoleAccompanist},
			event:     openEvent,
			memberErr: errors.New("db error"),
			wantErr:   true,
		},
	}

	_ = otherUserId // used implicitly through eventMember.UserId != ctx user

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			eRepo := mockEventRepo{
				findFunc: func(_ context.Context, _ string) (*events.Event, error) {
					return tt.event, tt.eventErr
				},
				saveFunc: func(_ context.Context, e *events.Event) (*events.Event, error) { return e, nil },
			}

			mRepo := mockMemberRepository{
				findFunc: func(_ context.Context, _ string) (*members.Member, error) {
					return tt.member, tt.memberErr
				},
			}

			lRepo := mockLicenceRepository{
				searchFunc: func(_ context.Context, _ *domain.SearchParams) ([]*licences.Licence, error) {
					return tt.licences, tt.licenceErr
				},
			}

			pRepo := mockParticipantRepo{
				registerFunc: func(_ context.Context, p *events.EventParticipant, max int) (*events.EventParticipant, error) {
					if tt.registerErr != nil {
						return nil, tt.registerErr
					}
					return tt.registered, nil
				},
				findByMemberFunc: func(_ context.Context, _, _ string) (*events.EventParticipant, error) {
					return tt.existingPart, nil
				},
			}

			rc := mockRoleChecker{
				hasRoleFunc: func(_ context.Context, _, _ string, _ ...string) (bool, error) {
					return tt.hasRole, tt.hasRoleErr
				},
				isSuperAdminFunc: func(_ context.Context, _ string) (bool, error) { return false, nil },
			}

			svc := defaultEventSvc(eRepo, noopCategory(), pRepo, noopCarpool(), noopJudo(), mRepo, lRepo, rc)
			res, err := svc.RegisterParticipant(context.Background(), tt.req)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			if tt.wantErrors {
				assert.True(t, len(res.Errors) > 0)
				if tt.wantErrKey != "" {
					assert.Contains(t, res.Errors, tt.wantErrKey)
				}
				assert.Nil(t, res.Participant)
			} else {
				assert.Empty(t, res.Errors)
				assert.NotNil(t, res.Participant)
			}
		})
	}
}

// ── CancelEvent ──────────────────────────────────────────────────────────────

func TestCancelEvent(t *testing.T) {
	creatorCtx := ctxWithUser(eventCreator.String())
	outsiderCtx := ctxWithUser("00000000-0000-0000-0000-000000000099")
	noUserCtx := context.Background()

	tests := []struct {
		name     string
		ctx      context.Context
		event    *events.Event
		eventErr error
		rc       mockRoleChecker
		saveErr  error
		wantOk   bool
		wantErr  bool
	}{
		{
			name:    "No user in context",
			ctx:     noUserCtx,
			wantErr: true,
		},
		{
			name:     "Event not found",
			ctx:      creatorCtx,
			event:    nil,
			wantErr:  true,
		},
		{
			name:    "Unauthorized outsider",
			ctx:     outsiderCtx,
			event:   openEvent,
			rc:      noopRoleChecker(),
			wantErr: true,
		},
		{
			name:   "Creator can cancel",
			ctx:    creatorCtx,
			event:  openEvent,
			rc:     noopRoleChecker(),
			wantOk: true,
		},
		{
			name:   "President can cancel",
			ctx:    outsiderCtx,
			event:  openEvent,
			rc:     presidentRoleChecker("00000000-0000-0000-0000-000000000099"),
			wantOk: true,
		},
		{
			name:     "Repository error on Save",
			ctx:      creatorCtx,
			event:    openEvent,
			rc:       noopRoleChecker(),
			saveErr:  errors.New("db error"),
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			eRepo := mockEventRepo{
				findFunc: func(_ context.Context, _ string) (*events.Event, error) {
					return tt.event, tt.eventErr
				},
				saveFunc: func(_ context.Context, e *events.Event) (*events.Event, error) {
					if tt.saveErr != nil {
						return nil, tt.saveErr
					}
					return e, nil
				},
			}

			svc := defaultEventSvc(eRepo, noopCategory(), mockParticipantRepo{}, noopCarpool(), noopJudo(),
				mockMemberRepository{}, mockLicenceRepository{}, tt.rc)
			ok, err := svc.CancelEvent(tt.ctx, eventUUID.String())

			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.wantOk, ok)
		})
	}
}

// ── OpenEvent ────────────────────────────────────────────────────────────────

func TestOpenEvent(t *testing.T) {
	creatorCtx := ctxWithUser(eventCreator.String())
	outsiderCtx := ctxWithUser("00000000-0000-0000-0000-000000000099")

	tests := []struct {
		name    string
		ctx     context.Context
		event   *events.Event
		rc      mockRoleChecker
		wantOk  bool
		wantErr bool
	}{
		{
			name:    "No user in context",
			ctx:     context.Background(),
			wantErr: true,
		},
		{
			name:    "Event not found",
			ctx:     creatorCtx,
			event:   nil,
			wantErr: true,
		},
		{
			name:    "Unauthorized",
			ctx:     outsiderCtx,
			event:   openEvent,
			rc:      noopRoleChecker(),
			wantErr: true,
		},
		{
			name:   "Creator opens event",
			ctx:    creatorCtx,
			event:  openEvent,
			rc:     noopRoleChecker(),
			wantOk: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			eRepo := mockEventRepo{
				findFunc: func(_ context.Context, _ string) (*events.Event, error) { return tt.event, nil },
				saveFunc: func(_ context.Context, e *events.Event) (*events.Event, error) { return e, nil },
			}

			svc := defaultEventSvc(eRepo, noopCategory(), mockParticipantRepo{}, noopCarpool(), noopJudo(),
				mockMemberRepository{}, mockLicenceRepository{}, tt.rc)
			ok, err := svc.OpenEvent(tt.ctx, eventUUID.String())

			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.wantOk, ok)
		})
	}
}

// ── UpdateEvent ──────────────────────────────────────────────────────────────

func TestUpdateEvent(t *testing.T) {
	creatorCtx := ctxWithUser(eventCreator.String())

	tests := []struct {
		name       string
		ctx        context.Context
		req        *dto.UpdateEventRequest
		event      *events.Event
		rc         mockRoleChecker
		wantErrors bool
		wantErr    bool
	}{
		{
			name:    "No user in context",
			ctx:     context.Background(),
			req:     &dto.UpdateEventRequest{Id: eventUUID.String()},
			wantErr: true,
		},
		{
			name:       "Event not found",
			ctx:        creatorCtx,
			req:        &dto.UpdateEventRequest{Id: eventUUID.String()},
			event:      nil,
			rc:         noopRoleChecker(),
			wantErrors: true,
		},
		{
			name:    "Unauthorized user",
			ctx:     ctxWithUser("00000000-0000-0000-0000-000000000099"),
			req:     &dto.UpdateEventRequest{Id: eventUUID.String(), Title: "New"},
			event:   openEvent,
			rc:      noopRoleChecker(),
			wantErr: true,
		},
		{
			name:       "Invalid update data (close before open)",
			ctx:        creatorCtx,
			req:        &dto.UpdateEventRequest{Id: eventUUID.String(), RegistrationCloseAt: "2026-01-01T00:00:00Z"},
			event:      openEvent,
			rc:         noopRoleChecker(),
			wantErrors: true,
		},
		{
			name:   "Valid update by creator",
			ctx:    creatorCtx,
			req:    &dto.UpdateEventRequest{Id: eventUUID.String(), Title: "Updated Title"},
			event:  openEvent,
			rc:     noopRoleChecker(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			eRepo := mockEventRepo{
				findFunc: func(_ context.Context, _ string) (*events.Event, error) { return tt.event, nil },
				saveFunc: func(_ context.Context, e *events.Event) (*events.Event, error) { return e, nil },
			}

			svc := defaultEventSvc(eRepo, noopCategory(), mockParticipantRepo{}, noopCarpool(), noopJudo(),
				mockMemberRepository{}, mockLicenceRepository{}, tt.rc)
			res, err := svc.UpdateEvent(tt.ctx, tt.req)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			if tt.wantErrors {
				assert.True(t, len(res.Errors) > 0)
			} else {
				assert.Empty(t, res.Errors)
				assert.NotNil(t, res.Event)
			}
		})
	}
}

// ── UnregisterParticipant ────────────────────────────────────────────────────

func TestUnregisterParticipant(t *testing.T) {
	// An event whose registration period is closed
	closedEvent := &events.Event{
		Id:                  eventUUID,
		ClubId:              eventClubId,
		CreatedBy:           eventCreator,
		Status:              events.StatusOpen,
		Type:                events.TypeCompetition,
		RegistrationOpenAt:  "2025-01-01T00:00:00Z",
		RegistrationCloseAt: "2025-06-01T00:00:00Z",
		Date:                "2025-07-01T09:00:00Z",
	}

	tests := []struct {
		name        string
		ctx         context.Context
		member      *members.Member
		memberErr   error
		event       *events.Event
		eventErr    error
		unregResult bool
		unregErr    error
		wantOk      bool
		wantErr     bool
	}{
		{
			name:    "No user in context",
			ctx:     context.Background(),
			wantErr: true,
		},
		{
			name:      "Member not found (unauthorized)",
			ctx:       ctxWithUser(eventCreator.String()),
			member:    nil,
			wantErr:   true,
		},
		{
			name: "Member belongs to different user (unauthorized)",
			ctx:  ctxWithUser("00000000-0000-0000-0000-000000000099"),
			member: &members.Member{
				Id:     memberUUID,
				UserId: eventCreator, // different from ctx user
			},
			wantErr: true,
		},
		{
			name:    "Event not found",
			ctx:     ctxWithUser(eventCreator.String()),
			member:  eventMember,
			event:   nil,
			wantErr: true,
		},
		{
			name:    "Registration period closed",
			ctx:     ctxWithUser(eventCreator.String()),
			member:  eventMember,
			event:   closedEvent,
			wantErr: true,
		},
		{
			name:        "Valid unregister",
			ctx:         ctxWithUser(eventCreator.String()),
			member:      eventMember,
			event:       openEvent,
			unregResult: true,
			wantOk:      true,
		},
		{
			name:      "Repository error on Find member",
			ctx:       ctxWithUser(eventCreator.String()),
			memberErr: errors.New("db error"),
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mRepo := mockMemberRepository{
				findFunc: func(_ context.Context, _ string) (*members.Member, error) {
					return tt.member, tt.memberErr
				},
			}
			eRepo := mockEventRepo{
				findFunc: func(_ context.Context, _ string) (*events.Event, error) {
					return tt.event, tt.eventErr
				},
			}
			pRepo := mockParticipantRepo{
				unregisterFunc: func(_ context.Context, _, _ string) (bool, error) {
					return tt.unregResult, tt.unregErr
				},
			}

			svc := defaultEventSvc(eRepo, noopCategory(), pRepo, noopCarpool(), noopJudo(),
				mRepo, mockLicenceRepository{}, noopRoleChecker())
			ok, err := svc.UnregisterParticipant(tt.ctx, &dto.UnregisterParticipantRequest{
				EventId: eventUUID.String(), MemberId: memberUUID.String(),
			})

			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.wantOk, ok)
		})
	}
}

// ── UnregisterParticipant (Force / Manager) ───────────────────────────────────

func TestUnregisterParticipantForce(t *testing.T) {
	// An event whose registration period is closed
	closedEvent := &events.Event{
		Id:                  eventUUID,
		ClubId:              eventClubId,
		CreatedBy:           eventCreator,
		Status:              events.StatusOpen,
		Type:                events.TypeCompetition,
		RegistrationOpenAt:  "2025-01-01T00:00:00Z",
		RegistrationCloseAt: "2025-06-01T00:00:00Z",
		Date:                "2025-07-01T09:00:00Z",
	}

	t.Run("Manager removes participant after period close (Force=true)", func(t *testing.T) {
		mRepo := mockMemberRepository{
			findFunc: func(_ context.Context, _ string) (*members.Member, error) {
				return eventMember, nil
			},
		}
		eRepo := mockEventRepo{
			findFunc: func(_ context.Context, _ string) (*events.Event, error) {
				return closedEvent, nil
			},
		}
		pRepo := mockParticipantRepo{
			unregisterFunc: func(_ context.Context, _, _ string) (bool, error) {
				return true, nil
			},
		}

		svc := defaultEventSvc(eRepo, noopCategory(), pRepo, noopCarpool(), noopJudo(),
			mRepo, mockLicenceRepository{}, noopRoleChecker())
		ok, err := svc.UnregisterParticipant(ctxWithUser(eventCreator.String()), &dto.UnregisterParticipantRequest{
			EventId: eventUUID.String(), MemberId: memberUUID.String(), Force: true,
		})

		assert.NoError(t, err)
		assert.True(t, ok)
	})

	t.Run("Member cannot unregister after period close (Force=false)", func(t *testing.T) {
		mRepo := mockMemberRepository{
			findFunc: func(_ context.Context, _ string) (*members.Member, error) {
				return eventMember, nil
			},
		}
		eRepo := mockEventRepo{
			findFunc: func(_ context.Context, _ string) (*events.Event, error) {
				return closedEvent, nil
			},
		}
		pRepo := mockParticipantRepo{
			unregisterFunc: func(_ context.Context, _, _ string) (bool, error) {
				return false, nil
			},
		}

		svc := defaultEventSvc(eRepo, noopCategory(), pRepo, noopCarpool(), noopJudo(),
			mRepo, mockLicenceRepository{}, noopRoleChecker())
		_, err := svc.UnregisterParticipant(ctxWithUser(eventCreator.String()), &dto.UnregisterParticipantRequest{
			EventId: eventUUID.String(), MemberId: memberUUID.String(), Force: false,
		})

		assert.Error(t, err)
	})
}

// ── CreateCarpoolOffer ───────────────────────────────────────────────────────

func TestCreateCarpoolOffer(t *testing.T) {
	tests := []struct {
		name         string
		req          *dto.CreateCarpoolOfferRequest
		participant  *events.EventParticipant
		partErr      error
		existOffer   *events.CarpoolOffer
		offerErr     error
		saveResult   *events.CarpoolOffer
		saveErr      error
		wantErrors   bool
		wantErr      bool
		wantErrKey   string
	}{
		{
			name:        "Not registered for event",
			req:         &dto.CreateCarpoolOfferRequest{EventId: eventUUID.String(), MemberId: memberUUID.String(), AvailableSeats: 2},
			participant: nil,
			wantErrors:  true,
			wantErrKey:  "member",
		},
		{
			name:       "Already has an offer",
			req:        &dto.CreateCarpoolOfferRequest{EventId: eventUUID.String(), MemberId: memberUUID.String(), AvailableSeats: 2},
			participant: testParticipant,
			existOffer: testCarpoolOffer,
			wantErrors: true,
			wantErrKey: "offer",
		},
		{
			name:        "Zero available seats",
			req:         &dto.CreateCarpoolOfferRequest{EventId: eventUUID.String(), MemberId: memberUUID.String(), AvailableSeats: 0},
			participant: testParticipant,
			existOffer:  nil,
			wantErrors:  true,
			wantErrKey:  "available_seats",
		},
		{
			name:        "Valid offer created",
			req:         &dto.CreateCarpoolOfferRequest{EventId: eventUUID.String(), MemberId: memberUUID.String(), AvailableSeats: 3, DepartureAddress: "10 rue de la Paix"},
			participant: testParticipant,
			existOffer:  nil,
			saveResult:  testCarpoolOffer,
			wantErrors:  false,
		},
		{
			name:      "Repository error on FindByMember (participant check)",
			req:       &dto.CreateCarpoolOfferRequest{EventId: eventUUID.String(), MemberId: memberUUID.String(), AvailableSeats: 2},
			partErr:   errors.New("db error"),
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pRepo := mockParticipantRepo{
				findByMemberFunc: func(_ context.Context, _, _ string) (*events.EventParticipant, error) {
					return tt.participant, tt.partErr
				},
			}
			carpRepo := mockCarpoolRepo{
				findOfferByMemberFunc: func(_ context.Context, _, _ string) (*events.CarpoolOffer, error) {
					return tt.existOffer, tt.offerErr
				},
				createOfferFunc: func(_ context.Context, o *events.CarpoolOffer) (*events.CarpoolOffer, error) {
					if tt.saveErr != nil {
						return nil, tt.saveErr
					}
					return tt.saveResult, nil
				},
			}

			svc := defaultEventSvc(mockEventRepo{}, noopCategory(), pRepo, carpRepo, noopJudo(),
				mockMemberRepository{}, mockLicenceRepository{}, noopRoleChecker())
			res, err := svc.CreateCarpoolOffer(context.Background(), tt.req)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			if tt.wantErrors {
				assert.True(t, len(res.Errors) > 0)
				if tt.wantErrKey != "" {
					assert.Contains(t, res.Errors, tt.wantErrKey)
				}
				assert.Nil(t, res.Offer)
			} else {
				assert.Empty(t, res.Errors)
				assert.NotNil(t, res.Offer)
			}
		})
	}
}

// ── JoinCarpoolOffer ─────────────────────────────────────────────────────────

func TestJoinCarpoolOffer(t *testing.T) {
	tests := []struct {
		name          string
		offer         *events.CarpoolOffer
		offerErr      error
		registered    *events.EventParticipant
		regErr        error
		memberOffer   *events.CarpoolOffer
		memberOfferErr error
		added         bool
		addErr        error
		isPassenger   bool
		isPassErr     error
		wantOk        bool
		wantErr       bool
		wantErrMsg    string
	}{
		{
			name:       "Offer not found",
			offer:      nil,
			wantErr:    true,
			wantErrMsg: "carpool offer not found",
		},
		{
			name:       "Not registered for event",
			offer:      testCarpoolOffer,
			registered: nil,
			wantErr:    true,
			wantErrMsg: "must be registered",
		},
		{
			name:        "Joiner has active offer (cannot join)",
			offer:       testCarpoolOffer,
			registered:  testParticipant,
			memberOffer: testCarpoolOffer,
			wantErr:     true,
			wantErrMsg:  "cannot join",
		},
		{
			name:        "Already a passenger (atomic block)",
			offer:       testCarpoolOffer,
			registered:  testParticipant,
			memberOffer: nil,
			added:       false,
			isPassenger: true,
			wantErr:     true,
			wantErrMsg:  "already a passenger",
		},
		{
			name:        "No seats available (atomic block)",
			offer:       testCarpoolOffer,
			registered:  testParticipant,
			memberOffer: nil,
			added:       false,
			isPassenger: false,
			wantErr:     true,
			wantErrMsg:  "no seats",
		},
		{
			name:        "Successfully joined",
			offer:       testCarpoolOffer,
			registered:  testParticipant,
			memberOffer: nil,
			added:       true,
			wantOk:      true,
		},
		{
			name:     "Repository error on FindOfferById",
			offerErr: errors.New("db error"),
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			carpRepo := mockCarpoolRepo{
				findOfferByIdFunc: func(_ context.Context, _ string) (*events.CarpoolOffer, error) {
					return tt.offer, tt.offerErr
				},
				findOfferByMemberFunc: func(_ context.Context, _, _ string) (*events.CarpoolOffer, error) {
					return tt.memberOffer, tt.memberOfferErr
				},
				addPassengerFunc: func(_ context.Context, _, _ string) (bool, error) {
					return tt.added, tt.addErr
				},
				isPassengerFunc: func(_ context.Context, _, _ string) (bool, error) {
					return tt.isPassenger, tt.isPassErr
				},
			}
			pRepo := mockParticipantRepo{
				findByMemberFunc: func(_ context.Context, _, _ string) (*events.EventParticipant, error) {
					return tt.registered, tt.regErr
				},
			}

			svc := defaultEventSvc(mockEventRepo{}, noopCategory(), pRepo, carpRepo, noopJudo(),
				mockMemberRepository{}, mockLicenceRepository{}, noopRoleChecker())
			ok, err := svc.JoinCarpoolOffer(context.Background(), offerUUID.String(), memberUUID.String())

			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantErrMsg != "" {
					assert.Contains(t, err.Error(), tt.wantErrMsg)
				}
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.wantOk, ok)
		})
	}
}

// ── CancelCarpoolOffer ───────────────────────────────────────────────────────

func TestCancelCarpoolOffer(t *testing.T) {
	ownerCtx := ctxWithUser(eventCreator.String())
	otherCtx := ctxWithUser("00000000-0000-0000-0000-000000000099")

	tests := []struct {
		name      string
		ctx       context.Context
		offer     *events.CarpoolOffer
		offerErr  error
		member    *members.Member
		memberErr error
		deleteOk  bool
		deleteErr error
		wantOk    bool
		wantErr   bool
		wantErrMsg string
	}{
		{
			name:       "No user in context",
			ctx:        context.Background(),
			wantErr:    true,
		},
		{
			name:       "Offer not found",
			ctx:        ownerCtx,
			offer:      nil,
			wantErr:    true,
			wantErrMsg: "carpool offer not found",
		},
		{
			name:       "Member not found (unauthorized)",
			ctx:        ownerCtx,
			offer:      testCarpoolOffer,
			member:     nil,
			wantErr:    true,
			wantErrMsg: "unauthorized",
		},
		{
			name:   "User is not the offer owner",
			ctx:    otherCtx,
			offer:  testCarpoolOffer,
			member: eventMember, // UserId = eventCreator, not otherCtx user
			wantErr:    true,
			wantErrMsg: "unauthorized",
		},
		{
			name:     "Valid cancel",
			ctx:      ownerCtx,
			offer:    testCarpoolOffer,
			member:   eventMember,
			deleteOk: true,
			wantOk:   true,
		},
		{
			name:      "Repository error on FindOfferById",
			ctx:       ownerCtx,
			offerErr:  errors.New("db error"),
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			carpRepo := mockCarpoolRepo{
				findOfferByIdFunc: func(_ context.Context, _ string) (*events.CarpoolOffer, error) {
					return tt.offer, tt.offerErr
				},
				deleteOfferFunc: func(_ context.Context, _ string) (bool, error) {
					return tt.deleteOk, tt.deleteErr
				},
			}
			mRepo := mockMemberRepository{
				findFunc: func(_ context.Context, _ string) (*members.Member, error) {
					return tt.member, tt.memberErr
				},
			}

			svc := defaultEventSvc(mockEventRepo{}, noopCategory(), mockParticipantRepo{}, carpRepo, noopJudo(),
				mRepo, mockLicenceRepository{}, noopRoleChecker())
			ok, err := svc.CancelCarpoolOffer(tt.ctx, offerUUID.String())

			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantErrMsg != "" {
					assert.Contains(t, err.Error(), tt.wantErrMsg)
				}
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.wantOk, ok)
		})
	}
}

// ── LeaveCarpoolOffer ────────────────────────────────────────────────────────

func TestLeaveCarpoolOffer(t *testing.T) {
	ownerCtx := ctxWithUser(eventCreator.String())

	tests := []struct {
		name       string
		ctx        context.Context
		member     *members.Member
		memberErr  error
		removeOk   bool
		removeErr  error
		wantOk     bool
		wantErr    bool
		wantErrMsg string
	}{
		{
			name:    "No user in context",
			ctx:     context.Background(),
			wantErr: true,
		},
		{
			name:       "Member not found (unauthorized)",
			ctx:        ownerCtx,
			member:     nil,
			wantErr:    true,
			wantErrMsg: "unauthorized",
		},
		{
			name: "Member belongs to different user",
			ctx:  ctxWithUser("00000000-0000-0000-0000-000000000099"),
			member: &members.Member{
				Id:     memberUUID,
				UserId: eventCreator, // context user is 099, member.UserId is creator
			},
			wantErr:    true,
			wantErrMsg: "unauthorized",
		},
		{
			name:      "Valid leave",
			ctx:       ownerCtx,
			member:    eventMember,
			removeOk:  true,
			wantOk:    true,
		},
		{
			name:      "Repository error on Find member",
			ctx:       ownerCtx,
			memberErr: errors.New("db error"),
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mRepo := mockMemberRepository{
				findFunc: func(_ context.Context, _ string) (*members.Member, error) {
					return tt.member, tt.memberErr
				},
			}
			carpRepo := mockCarpoolRepo{
				removePassengerFunc: func(_ context.Context, _, _ string) (bool, error) {
					return tt.removeOk, tt.removeErr
				},
			}

			svc := defaultEventSvc(mockEventRepo{}, noopCategory(), mockParticipantRepo{}, carpRepo, noopJudo(),
				mRepo, mockLicenceRepository{}, noopRoleChecker())
			ok, err := svc.LeaveCarpoolOffer(tt.ctx, offerUUID.String(), memberUUID.String())

			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantErrMsg != "" {
					assert.Contains(t, err.Error(), tt.wantErrMsg)
				}
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.wantOk, ok)
		})
	}
}
