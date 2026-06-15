package handlers

import (
	"clubmanager/internal/adapters/api/grpc/dto"
	"clubmanager/internal/adapters/api/http/views/pages"
	"clubmanager/internal/app/services"
	eventDomain "clubmanager/internal/domain/events"
	memberDomain "clubmanager/internal/domain/members"
	"clubmanager/internal/domain/roles"
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v5"
)

type EventHandler struct {
	eventSvc    services.EventService
	memberSvc   services.MemberService
	clubSvc     services.ClubService
	roleChecker roles.RoleChecker
	postSvc     services.PostService
}

func NewEventHandler(eventSvc services.EventService, memberSvc services.MemberService, clubSvc services.ClubService, roleChecker roles.RoleChecker, postSvc services.PostService) *EventHandler {
	return &EventHandler{eventSvc: eventSvc, memberSvc: memberSvc, clubSvc: clubSvc, roleChecker: roleChecker, postSvc: postSvc}
}

func parseHTMLDatetime(s string) string {
	if s == "" {
		return ""
	}
	t, err := time.ParseInLocation("2006-01-02T15:04", s, time.UTC)
	if err != nil {
		return s
	}
	return t.Format(time.RFC3339)
}

func parseDateWithTime(date, timeStr string) string {
	if date == "" || timeStr == "" {
		return ""
	}
	t, err := time.ParseInLocation("2006-01-02 15:04", date+" "+timeStr, time.UTC)
	if err != nil {
		return ""
	}
	return t.Format(time.RFC3339)
}

func parseDateOnly(s string) string {
	if s == "" {
		return ""
	}
	t, err := time.ParseInLocation("2006-01-02", s, time.UTC)
	if err != nil {
		return s
	}
	return t.Format(time.RFC3339)
}

func parseAnyTimestamp(s string) (time.Time, bool) {
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t, true
	}
	if t, err := time.Parse("2006-01-02 15:04:05-07", s); err == nil {
		return t, true
	}
	if t, err := time.Parse("2006-01-02", s); err == nil {
		return t, true
	}
	return time.Time{}, false
}

func timestampToDate(s string) string {
	if t, ok := parseAnyTimestamp(s); ok {
		return t.UTC().Format("2006-01-02")
	}
	if len(s) >= 10 {
		return s[:10]
	}
	return s
}

func ageGroupDisplayLabel(ageGroup, gender string) string {
	key := ageGroup
	if gender == "man" {
		key += " Masculin"
	} else if gender == "woman" {
		key += " Féminin"
	}
	labels := map[string]string{
		"Eveil":             "Eveil",
		"Poussinet":         "Poussinet",
		"Poussin":           "Poussin",
		"Benjamin Masculin": "Benjamin",
		"Benjamin Féminin":  "Benjamine",
		"Minime Masculin":   "Minime",
		"Minime Féminin":    "Minimette",
		"Cadet Masculin":    "Cadet",
		"Cadet Féminin":     "Cadette",
		"Junior Masculin":   "Junior",
		"Junior Féminin":    "Juniorette",
		"Senior Masculin":   "Senior",
		"Senior Féminin":    "Sénior Féminine",
		"Vétéran Masculin":  "Vétéran",
		"Vétéran Féminin":   "Vétérane",
	}
	if l, ok := labels[key]; ok {
		return l
	}
	return key
}

func memberCategoryAge(birthdate, eventDate string) int {
	birth, err := time.Parse("2006-01-02", birthdate)
	if err != nil {
		return 0
	}
	ev, ok := parseAnyTimestamp(eventDate)
	if !ok {
		return 0
	}
	seasonEndYear := ev.Year()
	if ev.Month() >= time.September {
		seasonEndYear++
	}
	return seasonEndYear - birth.Year()
}

func (h *EventHandler) HandleEventList(c *echo.Context) error {
	clubId := c.Param("id")
	userId, _ := c.Get("userId").(string)

	clubResp, err := h.clubSvc.GetClub(c.Request().Context(), clubId)
	if err != nil {
		return err
	}
	if clubResp.Club == nil {
		return echo.NewHTTPError(http.StatusNotFound, "Club introuvable.")
	}

	isMember, err := h.roleChecker.HasRole(c.Request().Context(), userId, clubId, "president", "staff", "coach", "associate")
	if err != nil {
		return err
	}
	if !isMember {
		return echo.NewHTTPError(http.StatusForbidden, "Accès refusé.")
	}

	canManage, err := h.roleChecker.HasRole(c.Request().Context(), userId, clubId, "president", "staff")
	if err != nil {
		return err
	}

	pageStr := c.QueryParam("page")
	page, _ := strconv.Atoi(pageStr)
	if page <= 0 {
		page = 1
	}
	const eventsPageSize = 10

	eventsResp, err := h.eventSvc.ListClubEvents(c.Request().Context(), &dto.ListClubEventsRequest{
		ClubId: clubId, Page: page, PageSize: eventsPageSize,
	})
	if err != nil {
		return err
	}

	return render(c, pages.EventsList(clubResp.Club, eventsResp.Events, eventsResp.Total, page, eventsPageSize, canManage))
}

func (h *EventHandler) HandleCreateEventPage(c *echo.Context) error {
	clubId := c.Param("id")
	userId, _ := c.Get("userId").(string)

	clubResp, err := h.clubSvc.GetClub(c.Request().Context(), clubId)
	if err != nil {
		return err
	}
	if clubResp.Club == nil {
		return echo.NewHTTPError(http.StatusNotFound, "Club introuvable.")
	}

	canManage, err := h.roleChecker.HasRole(c.Request().Context(), userId, clubId, "president", "staff")
	if err != nil {
		return err
	}
	if !canManage {
		return echo.NewHTTPError(http.StatusForbidden, "Accès refusé.")
	}

	judoCatsResp, err := h.eventSvc.ListJudoCategories(c.Request().Context())
	if err != nil {
		return err
	}

	return render(c, pages.EventCreate(clubResp.Club, judoCatsResp.Categories, nil, csrf(c)))
}

func (h *EventHandler) HandleCreateEvent(c *echo.Context) error {
	clubId := c.Param("id")
	userId, _ := c.Get("userId").(string)

	clubResp, err := h.clubSvc.GetClub(c.Request().Context(), clubId)
	if err != nil {
		return err
	}
	if clubResp.Club == nil {
		return echo.NewHTTPError(http.StatusNotFound, "Club introuvable.")
	}

	canManage, err := h.roleChecker.HasRole(c.Request().Context(), userId, clubId, "president", "staff")
	if err != nil {
		return err
	}
	if !canManage {
		return echo.NewHTTPError(http.StatusForbidden, "Accès refusé.")
	}

	if err := c.Request().ParseForm(); err != nil {
		return err
	}
	eventDate := c.FormValue("date") // YYYY-MM-DD
	judoCatIds := c.Request().Form["category_judo_id"]
	weighIns := c.Request().Form["category_weigh_in"]

	var categories []*dto.CreateEventCategoryRequest
	for i, jcId := range judoCatIds {
		if jcId == "" {
			continue
		}
		cr := &dto.CreateEventCategoryRequest{JudoCategoryId: jcId}
		if i < len(weighIns) && weighIns[i] != "" {
			cr.WeighInAt = parseDateWithTime(eventDate, weighIns[i])
		}
		categories = append(categories, cr)
	}

	maxPart, _ := strconv.Atoi(c.FormValue("max_participants"))

	ctx := context.WithValue(c.Request().Context(), "user_id", userId)
	resp, err := h.eventSvc.CreateEvent(ctx, &dto.CreateEventRequest{
		ClubId:              clubId,
		Title:               c.FormValue("title"),
		Description:         c.FormValue("description"),
		Type:                c.FormValue("type"),
		Location:            c.FormValue("location"),
		RegistrationOpenAt:  parseHTMLDatetime(c.FormValue("registration_open_at")),
		RegistrationCloseAt: parseHTMLDatetime(c.FormValue("registration_close_at")),
		Date:                parseDateOnly(c.FormValue("date")),
		MaxParticipants:     maxPart,
		Categories:          categories,
	})
	if err != nil {
		return err
	}

	if len(resp.Errors) > 0 {
		var judoCats []*eventDomain.JudoCategory
		if judoCatsResp, err2 := h.eventSvc.ListJudoCategories(c.Request().Context()); err2 == nil {
			judoCats = judoCatsResp.Categories
		}
		return render(c, pages.EventCreate(clubResp.Club, judoCats, resp.Errors, csrf(c)))
	}

	return c.Redirect(http.StatusSeeOther, "/clubs/"+clubId+"/events")
}

func (h *EventHandler) HandleEventDetail(c *echo.Context) error {
	eventId := c.Param("id")
	userId, _ := c.Get("userId").(string)

	eventResp, err := h.eventSvc.GetEvent(c.Request().Context(), eventId)
	if err != nil {
		return err
	}
	if eventResp.Event == nil {
		return echo.NewHTTPError(http.StatusNotFound, "Événement introuvable.")
	}

	clubId := eventResp.Event.ClubId.String()
	hasRole, err := h.roleChecker.HasRole(c.Request().Context(), userId, clubId, "president", "staff", "coach")
	if err != nil {
		return err
	}

	// Load participants once — used for access check and main logic below
	participantsResp, err := h.eventSvc.GetParticipants(c.Request().Context(), eventId)
	if err != nil {
		return err
	}

	// Build current user's member map and their participations (used for access check, carpool and adherent view)
	type myParticipant struct {
		member      *memberDomain.Member
		participant *eventDomain.EventParticipant
	}
	var myParticipants []myParticipant
	myMemberIds := make(map[string]bool)
	for _, p := range participantsResp.Participants {
		m, err2 := h.memberSvc.GetMember(c.Request().Context(), p.MemberId.String())
		if err2 == nil && m != nil && m.UserId.String() == userId {
			myMemberIds[p.MemberId.String()] = true
			myParticipants = append(myParticipants, myParticipant{member: m, participant: p})
		}
	}

	isRegistered := len(myParticipants) > 0
	if !hasRole && !isRegistered {
		return c.Redirect(http.StatusSeeOther, "/events/"+eventId+"/join")
	}

	isManager, err := h.roleChecker.HasRole(c.Request().Context(), userId, clubId, "president", "staff")
	if err != nil {
		return err
	}

	// Build judoCat lookup
	judoCatsResp, err := h.eventSvc.ListJudoCategories(c.Request().Context())
	if err != nil {
		return err
	}
	judoCatMap := make(map[string]*eventDomain.JudoCategory)
	for _, jc := range judoCatsResp.Categories {
		judoCatMap[jc.Id.String()] = jc
	}

	// Build event category → judo category map and eventCat lookup (id → EventCategory)
	eventCatToJudoCat := make(map[string]*eventDomain.JudoCategory)
	eventCatById := make(map[string]*eventDomain.EventCategory)
	for _, ec := range eventResp.Categories {
		eventCatById[ec.Id.String()] = ec
		if jc := judoCatMap[ec.JudoCategoryId.String()]; jc != nil {
			eventCatToJudoCat[ec.Id.String()] = jc
		}
	}

	// Build adherent-specific participation rows (weigh-in times per registered member)
	var myParticipationRows []dto.MyParticipationRow
	for _, mp := range myParticipants {
		row := dto.MyParticipationRow{
			MemberName: mp.member.Firstname + " " + mp.member.Lastname,
			Role:       mp.participant.Role,
		}
		if mp.participant.EventCategoryId != nil {
			ecId := mp.participant.EventCategoryId.String()
			if jc, ok := eventCatToJudoCat[ecId]; ok {
				row.CategoryName = ageGroupDisplayLabel(jc.AgeGroup, jc.Gender)
			}
			if ec, ok := eventCatById[ecId]; ok {
				row.WeighInAt = ec.WeighInAt
				row.StartsAt = ec.StartsAt
			}
		}
		myParticipationRows = append(myParticipationRows, row)
	}

	// Load carpool data for open events
	var carpoolOffers []dto.CarpoolOfferRow
	var canCreateOffer bool
	if eventResp.Event.Status == "open" {
		carpoolResp, err2 := h.eventSvc.GetEventCarpools(c.Request().Context(), eventId)
		if err2 == nil && carpoolResp != nil {
			passengers, _ := h.eventSvc.GetEventCarpoolPassengers(c.Request().Context(), eventId)

			passengerCountByOffer := make(map[string]int)
			passengersByOffer := make(map[string]map[string]bool)
			for _, p := range passengers {
				oid := p.OfferId.String()
				passengerCountByOffer[oid]++
				if passengersByOffer[oid] == nil {
					passengersByOffer[oid] = make(map[string]bool)
				}
				passengersByOffer[oid][p.MemberId.String()] = true
			}

			isAlreadyDriver := false
			isAlreadyPassenger := false

			for _, offer := range carpoolResp.Offers {
				isOwnOffer := myMemberIds[offer.MemberId.String()]
				if isOwnOffer {
					isAlreadyDriver = true
				}

				isPassenger := false
				for mid := range myMemberIds {
					if passengersByOffer[offer.Id.String()][mid] {
						isPassenger = true
						isAlreadyPassenger = true
						break
					}
				}

				driverName := "Conducteur inconnu"
				driver, _ := h.memberSvc.GetMember(c.Request().Context(), offer.MemberId.String())
				if driver != nil {
					driverName = driver.Firstname + " " + driver.Lastname
				}

				carpoolOffers = append(carpoolOffers, dto.CarpoolOfferRow{
					Offer:          offer,
					DriverName:     driverName,
					PassengerCount: passengerCountByOffer[offer.Id.String()],
					IsOwnOffer:     isOwnOffer,
					IsPassenger:    isPassenger,
				})
			}

			canCreateOffer = len(myMemberIds) > 0 && !isAlreadyDriver && !isAlreadyPassenger
		}
	}

	// For competition: build age group rows with participant counts
	var ageGroups []dto.AgeGroupRow
	if eventResp.Event.Type == eventDomain.TypeCompetition {
		// Count participants per age group + gender
		type agKey struct{ AgeGroup, Gender string }
		ageGroupCount := make(map[agKey]int)
		ageGroupOrder := make([]agKey, 0)
		seen := make(map[agKey]bool)
		for _, ec := range eventResp.Categories {
			if jc := judoCatMap[ec.JudoCategoryId.String()]; jc != nil {
				k := agKey{jc.AgeGroup, jc.Gender}
				if !seen[k] {
					seen[k] = true
					ageGroupOrder = append(ageGroupOrder, k)
				}
			}
		}
		for _, p := range participantsResp.Participants {
			if p.EventCategoryId != nil {
				if jc, ok := eventCatToJudoCat[p.EventCategoryId.String()]; ok {
					ageGroupCount[agKey{jc.AgeGroup, jc.Gender}]++
				}
			}
		}
		for _, k := range ageGroupOrder {
			ageGroups = append(ageGroups, dto.AgeGroupRow{
				AgeGroup:         k.AgeGroup,
				Gender:           k.Gender,
				ParticipantCount: ageGroupCount[k],
			})
		}

		// Count supervisors (coach/referee) and accompanists for modal buttons
		supervisorCount := 0
		accompanistCount := 0
		for _, p := range participantsResp.Participants {
			if p.EventCategoryId != nil {
				continue
			}
			switch p.Role {
			case "coach", "referee":
				supervisorCount++
			case "accompanist":
				accompanistCount++
			}
		}
		return render(c, pages.EventDetail(eventResp.Event, ageGroups, supervisorCount, accompanistCount, nil, isManager, csrf(c), c.QueryParam("error"), carpoolOffers, canCreateOffer, myParticipationRows))
	}

	// Non-competition: show participants directly
	var rows []dto.ParticipantRow
	for _, p := range participantsResp.Participants {
		memberId := p.MemberId.String()
		m, err2 := h.memberSvc.GetMember(c.Request().Context(), memberId)
		name := "Membre inconnu"
		if err2 == nil && m != nil {
			name = m.Firstname + " " + m.Lastname
		}
		rows = append(rows, dto.ParticipantRow{
			Participant: p,
			MemberName:  name,
		})
	}
	return render(c, pages.EventDetail(eventResp.Event, nil, 0, 0, rows, isManager, csrf(c), c.QueryParam("error"), carpoolOffers, canCreateOffer, myParticipationRows))
}

func (h *EventHandler) HandleAgeGroupParticipants(c *echo.Context) error {
	eventId := c.Param("id")
	ageGroup := c.QueryParam("age_group")
	roleParam := c.QueryParam("role")   // e.g. "coach,referee" or "accompanist"
	genderParam := c.QueryParam("gender") // e.g. "man", "woman", "mixed"
	userId, _ := c.Get("userId").(string)

	eventResp, err := h.eventSvc.GetEvent(c.Request().Context(), eventId)
	if err != nil {
		return err
	}
	if eventResp.Event == nil {
		return echo.NewHTTPError(http.StatusNotFound, "Événement introuvable.")
	}

	organizingClubId := eventResp.Event.ClubId.String()
	isOrganizer, err := h.roleChecker.HasRole(c.Request().Context(), userId, organizingClubId, "president", "staff")
	if err != nil {
		return err
	}

	// Determine which club's members the user can see
	var viewingClubMemberIds map[string]bool
	if !isOrganizer {
		userClubsResp, err := h.memberSvc.GetUserClubs(c.Request().Context(), &dto.GetUserClubsRequest{UserId: userId})
		if err != nil {
			return err
		}
		var viewingClubId string
		for _, club := range userClubsResp.Clubs {
			hasRole, err := h.roleChecker.HasRole(c.Request().Context(), userId, club.Id.String(), "president", "staff")
			if err != nil {
				continue
			}
			if hasRole {
				viewingClubId = club.Id.String()
				break
			}
		}
		if viewingClubId == "" {
			return echo.NewHTTPError(http.StatusForbidden, "Accès refusé.")
		}
		clubMembersResp, err := h.memberSvc.GetMembersByClub(c.Request().Context(), &dto.GetMembersByClubRequest{ClubId: viewingClubId})
		if err != nil {
			return err
		}
		viewingClubMemberIds = make(map[string]bool)
		for _, mwm := range clubMembersResp.Members {
			viewingClubMemberIds[mwm.Member.Id.String()] = true
		}
	}

	participantsResp, err := h.eventSvc.GetParticipants(c.Request().Context(), eventId)
	if err != nil {
		return err
	}

	// Role-based filter (coach/referee or accompanist)
	if roleParam != "" {
		rolesSet := make(map[string]bool)
		for _, r := range strings.Split(roleParam, ",") {
			rolesSet[strings.TrimSpace(r)] = true
		}
		var rows []dto.ParticipantRow
		for _, p := range participantsResp.Participants {
			if !rolesSet[p.Role] {
				continue
			}
			memberId := p.MemberId.String()
			if viewingClubMemberIds != nil && !viewingClubMemberIds[memberId] {
				continue
			}
			m, err2 := h.memberSvc.GetMember(c.Request().Context(), memberId)
			name := "Inconnu"
			if err2 == nil && m != nil {
				name = m.Firstname + " " + m.Lastname
			}
			rows = append(rows, dto.ParticipantRow{Participant: p, MemberName: name})
		}
		title := "Encadrants"
		if roleParam == "accompanist" {
			title = "Accompagnateurs"
		}
		return render(c, pages.AgeGroupParticipantsFragment(title, rows, isOrganizer, eventId, csrf(c)))
	}

	// Age-group filter (competitors)
	judoCatsResp, err := h.eventSvc.ListJudoCategories(c.Request().Context())
	if err != nil {
		return err
	}
	judoCatMap := make(map[string]*eventDomain.JudoCategory)
	for _, jc := range judoCatsResp.Categories {
		judoCatMap[jc.Id.String()] = jc
	}
	eventCatToJudoCat := make(map[string]*eventDomain.JudoCategory)
	for _, ec := range eventResp.Categories {
		jc := judoCatMap[ec.JudoCategoryId.String()]
		if jc == nil || jc.AgeGroup != ageGroup {
			continue
		}
		if genderParam != "" && jc.Gender != genderParam {
			continue
		}
		eventCatToJudoCat[ec.Id.String()] = jc
	}

	var rows []dto.ParticipantRow
	for _, p := range participantsResp.Participants {
		if p.EventCategoryId == nil {
			continue
		}
		jc, ok := eventCatToJudoCat[p.EventCategoryId.String()]
		if !ok {
			continue
		}
		memberId := p.MemberId.String()
		if viewingClubMemberIds != nil && !viewingClubMemberIds[memberId] {
			continue
		}
		m, err2 := h.memberSvc.GetMember(c.Request().Context(), memberId)
		name := "Inconnu"
		if err2 == nil && m != nil {
			name = m.Firstname + " " + m.Lastname
		}
		rows = append(rows, dto.ParticipantRow{
			Participant:   p,
			MemberName:    name,
			CategoryLabel: jc.WeightLabel,
		})
	}

	title := ageGroupDisplayLabel(ageGroup, genderParam) + " — Compétiteurs"
	return render(c, pages.AgeGroupParticipantsFragment(title, rows, isOrganizer, eventId, csrf(c)))
}

func (h *EventHandler) HandleOpenEvent(c *echo.Context) error {
	eventId := c.Param("id")
	userId, _ := c.Get("userId").(string)

	eventResp, err := h.eventSvc.GetEvent(c.Request().Context(), eventId)
	if err != nil {
		return err
	}
	if eventResp.Event == nil {
		return echo.NewHTTPError(http.StatusNotFound, "Événement introuvable.")
	}
	ev := eventResp.Event

	canManage, err := h.roleChecker.HasRole(c.Request().Context(), userId, ev.ClubId.String(), "president", "staff")
	if err != nil {
		return err
	}
	if !canManage {
		return echo.NewHTTPError(http.StatusForbidden, "Accès refusé.")
	}

	ctx := context.WithValue(c.Request().Context(), "user_id", userId)
	if _, err := h.eventSvc.OpenEvent(ctx, eventId); err != nil {
		if errors.Is(err, services.ErrValidation) {
			return c.Redirect(http.StatusSeeOther, "/events/"+eventId+"?error="+url.QueryEscape(err.Error()))
		}
		return err
	}

	// Auto-create an announcement post linked to this event, only if none exists yet.
	hasPost, _ := h.postSvc.HasPostForEvent(ctx, eventId)
	if !hasPost {
		visibility := c.FormValue("visibility")
		if visibility != "public" {
			visibility = "adherent"
		}
		content := buildEventOpenContent(ev)
		if _, postErr := h.postSvc.CreatePost(ctx, &dto.CreatePostRequest{
			ClubId:     ev.ClubId.String(),
			EventId:    eventId,
			Title:      ev.Title + " — Inscriptions ouvertes",
			Content:    content,
			Visibility: visibility,
			Status:     "published",
		}); postErr != nil {
			fmt.Printf("[WARN] HandleOpenEvent: failed to create announcement post for event %s: %v\n", eventId, postErr)
		}
	}

	return c.Redirect(http.StatusSeeOther, "/events/"+eventId)
}

func buildEventOpenContent(ev *eventDomain.Event) string {
	date := ""
	if t, ok := parseAnyTimestamp(ev.Date); ok {
		months := [...]string{"janvier", "février", "mars", "avril", "mai", "juin", "juillet", "août", "septembre", "octobre", "novembre", "décembre"}
		date = fmt.Sprintf("%d %s %d", t.Day(), months[t.Month()-1], t.Year())
	}
	content := "Les inscriptions sont maintenant ouvertes !"
	if date != "" {
		content += "\n\nDate : " + date
	}
	if ev.Location != "" {
		content += "\nLieu : " + ev.Location
	}
	if ev.Description != "" {
		content += "\n\n" + ev.Description
	}
	return content
}

func (h *EventHandler) HandleCancelEvent(c *echo.Context) error {
	eventId := c.Param("id")
	userId, _ := c.Get("userId").(string)

	eventResp, err := h.eventSvc.GetEvent(c.Request().Context(), eventId)
	if err != nil {
		return err
	}
	if eventResp.Event == nil {
		return echo.NewHTTPError(http.StatusNotFound, "Événement introuvable.")
	}

	canManage, err := h.roleChecker.HasRole(c.Request().Context(), userId, eventResp.Event.ClubId.String(), "president", "staff")
	if err != nil {
		return err
	}
	if !canManage {
		return echo.NewHTTPError(http.StatusForbidden, "Accès refusé.")
	}

	ctx := context.WithValue(c.Request().Context(), "user_id", userId)
	if _, err := h.eventSvc.CancelEvent(ctx, eventId); err != nil {
		return err
	}
	return c.Redirect(http.StatusSeeOther, "/events/"+eventId)
}

func (h *EventHandler) HandleReopenEvent(c *echo.Context) error {
	eventId := c.Param("id")
	userId, _ := c.Get("userId").(string)

	eventResp, err := h.eventSvc.GetEvent(c.Request().Context(), eventId)
	if err != nil {
		return err
	}
	if eventResp.Event == nil {
		return echo.NewHTTPError(http.StatusNotFound, "Événement introuvable.")
	}

	canManage, err := h.roleChecker.HasRole(c.Request().Context(), userId, eventResp.Event.ClubId.String(), "president", "staff")
	if err != nil {
		return err
	}
	if !canManage {
		return echo.NewHTTPError(http.StatusForbidden, "Accès refusé.")
	}

	ctx := context.WithValue(c.Request().Context(), "user_id", userId)
	if _, err := h.eventSvc.ReopenEvent(ctx, eventId); err != nil {
		return err
	}
	return c.Redirect(http.StatusSeeOther, "/events/"+eventId)
}

func (h *EventHandler) HandleDeleteEvent(c *echo.Context) error {
	eventId := c.Param("id")
	userId, _ := c.Get("userId").(string)

	eventResp, err := h.eventSvc.GetEvent(c.Request().Context(), eventId)
	if err != nil {
		return err
	}
	if eventResp.Event == nil {
		return echo.NewHTTPError(http.StatusNotFound, "Événement introuvable.")
	}

	canManage, err := h.roleChecker.HasRole(c.Request().Context(), userId, eventResp.Event.ClubId.String(), "president", "staff")
	if err != nil {
		return err
	}
	if !canManage {
		return echo.NewHTTPError(http.StatusForbidden, "Accès refusé.")
	}

	clubId := eventResp.Event.ClubId.String()
	ctx := context.WithValue(c.Request().Context(), "user_id", userId)
	if _, err := h.eventSvc.DeleteEvent(ctx, eventId); err != nil {
		return err
	}
	return c.Redirect(http.StatusSeeOther, "/clubs/"+clubId+"/events")
}

func (h *EventHandler) HandleJoinEventPage(c *echo.Context) error {
	eventId := c.Param("id")
	userId, _ := c.Get("userId").(string)

	eventResp, err := h.eventSvc.GetEvent(c.Request().Context(), eventId)
	if err != nil {
		return err
	}
	if eventResp.Event == nil {
		return echo.NewHTTPError(http.StatusNotFound, "Événement introuvable.")
	}

	clubId := eventResp.Event.ClubId.String()

	// Get user's own members with valid membership in this club
	clubMembersResp, err := h.memberSvc.GetMembersByClub(c.Request().Context(), &dto.GetMembersByClubRequest{ClubId: clubId})
	if err != nil {
		return err
	}
	var myMembers []*memberDomain.Member
	for _, mwm := range clubMembersResp.Members {
		if mwm.Member.UserId.String() == userId && mwm.Membership.IsValid {
			myMembers = append(myMembers, mwm.Member)
		}
	}
	if len(myMembers) == 0 {
		return echo.NewHTTPError(http.StatusForbidden, "Vous n'êtes pas membre de ce club.")
	}

	// Check existing participations per member
	participantsResp, err := h.eventSvc.GetParticipants(c.Request().Context(), eventId)
	if err != nil {
		return err
	}
	myMemberIds := make(map[string]bool)
	for _, mm := range myMembers {
		myMemberIds[mm.Id.String()] = true
	}
	myParticipations := make(map[string]*eventDomain.EventParticipant)
	for _, p := range participantsResp.Participants {
		if myMemberIds[p.MemberId.String()] {
			myParticipations[p.MemberId.String()] = p
		}
	}

	// For competition: compute matching event categories per member, then sort members
	// so those with matching categories appear first.
	memberMatchingCats := make(map[string][]dto.EnrichedEventCategory)
	if eventResp.Event.Type == eventDomain.TypeCompetition {
		judoCatsResp, err := h.eventSvc.ListJudoCategories(c.Request().Context())
		if err != nil {
			return err
		}
		judoCatMap := make(map[string]*eventDomain.JudoCategory)
		for _, jc := range judoCatsResp.Categories {
			judoCatMap[jc.Id.String()] = jc
		}
		for _, mm := range myMembers {
			age := memberCategoryAge(mm.Birthdate, eventResp.Event.Date)
			var matching []dto.EnrichedEventCategory
			for _, ec := range eventResp.Categories {
				jc := judoCatMap[ec.JudoCategoryId.String()]
				if jc == nil {
					continue
				}
				// mixed categories are open to all genders
				if jc.Gender != "mixed" && jc.Gender != mm.Gender {
					continue
				}
				if age < jc.AgeMin {
					continue
				}
				if jc.AgeMax != nil && age > *jc.AgeMax {
					continue
				}
				matching = append(matching, dto.EnrichedEventCategory{EventCat: ec, JudoCat: jc})
			}
			memberMatchingCats[mm.Id.String()] = matching
		}

		// Sort: members with categories first
		sort.Slice(myMembers, func(i, j int) bool {
			iHas := len(memberMatchingCats[myMembers[i].Id.String()]) > 0
			jHas := len(memberMatchingCats[myMembers[j].Id.String()]) > 0
			return iHas && !jHas
		})
	}

	isCoach, err := h.roleChecker.HasRole(c.Request().Context(), userId, clubId, "coach", "president", "staff")
	if err != nil {
		return err
	}

	return render(c, pages.EventJoin(eventResp.Event, myMembers, memberMatchingCats, myParticipations, isCoach, csrf(c)))
}

func (h *EventHandler) HandleJoinEvent(c *echo.Context) error {
	eventId := c.Param("id")
	userId, _ := c.Get("userId").(string)

	memberId := c.FormValue("member_id")

	// Get event to verify it exists
	eventResp, err := h.eventSvc.GetEvent(c.Request().Context(), eventId)
	if err != nil {
		return err
	}
	if eventResp.Event == nil {
		return echo.NewHTTPError(http.StatusNotFound, "Événement introuvable.")
	}

	// Verify member belongs to the user
	userMembersResp, err := h.memberSvc.GetMembersByUser(c.Request().Context(), &dto.GetMembersByUserRequest{UserId: userId})
	if err != nil {
		return err
	}
	var found bool
	for _, m := range userMembersResp.Members {
		if m.Id.String() == memberId {
			found = true
			break
		}
	}
	if !found {
		return echo.NewHTTPError(http.StatusForbidden, "Accès refusé.")
	}

	resp, err := h.eventSvc.RegisterParticipant(c.Request().Context(), &dto.RegisterParticipantRequest{
		EventId:         eventId,
		MemberId:        memberId,
		Role:            c.FormValue("role"),
		EventCategoryId: c.FormValue("event_category_id"),
	})
	if err != nil {
		return err
	}
	if len(resp.Errors) > 0 {
		// Concatenate all error messages and return as a user-facing error
		var msgs []string
		for _, v := range resp.Errors {
			msgs = append(msgs, v)
		}
		return echo.NewHTTPError(http.StatusBadRequest, strings.Join(msgs, " "))
	}

	return c.Redirect(http.StatusSeeOther, "/events/"+eventId+"/join")
}

// HandleRemoveParticipant allows a club manager to remove any participant from an event.
func (h *EventHandler) HandleRemoveParticipant(c *echo.Context) error {
	eventId := c.Param("id")
	userId, _ := c.Get("userId").(string)
	memberId := c.FormValue("member_id")

	eventResp, err := h.eventSvc.GetEvent(c.Request().Context(), eventId)
	if err != nil {
		return err
	}
	if eventResp.Event == nil {
		return echo.NewHTTPError(http.StatusNotFound, "Événement introuvable.")
	}

	isManager, err := h.roleChecker.HasRole(c.Request().Context(), userId, eventResp.Event.ClubId.String(), "president", "staff")
	if err != nil {
		return err
	}
	if !isManager {
		return echo.NewHTTPError(http.StatusForbidden, "Accès refusé.")
	}

	// Fetch the member to obtain their userId — required by the service's ownership check.
	member, err := h.memberSvc.GetMember(c.Request().Context(), memberId)
	if err != nil {
		return err
	}
	if member == nil {
		return echo.NewHTTPError(http.StatusNotFound, "Membre introuvable.")
	}

	ctx := context.WithValue(c.Request().Context(), "user_id", member.UserId.String())
	if _, err := h.eventSvc.UnregisterParticipant(ctx, &dto.UnregisterParticipantRequest{
		EventId:  eventId,
		MemberId: memberId,
		Force:    true,
	}); err != nil {
		return err
	}

	return c.Redirect(http.StatusSeeOther, "/events/"+eventId)
}

func (h *EventHandler) HandleEditEventPage(c *echo.Context) error {
	eventId := c.Param("id")
	userId, _ := c.Get("userId").(string)

	eventResp, err := h.eventSvc.GetEvent(c.Request().Context(), eventId)
	if err != nil {
		return err
	}
	if eventResp.Event == nil {
		return echo.NewHTTPError(http.StatusNotFound, "Événement introuvable.")
	}
	if eventResp.Event.Status != "draft" {
		return echo.NewHTTPError(http.StatusForbidden, "Seuls les événements en brouillon peuvent être modifiés.")
	}

	canManage, err := h.roleChecker.HasRole(c.Request().Context(), userId, eventResp.Event.ClubId.String(), "president", "staff")
	if err != nil {
		return err
	}
	if !canManage {
		return echo.NewHTTPError(http.StatusForbidden, "Accès refusé.")
	}

	var judoCats []*eventDomain.JudoCategory
	if eventResp.Event.Type == eventDomain.TypeCompetition {
		judoCatsResp, err := h.eventSvc.ListJudoCategories(c.Request().Context())
		if err != nil {
			return err
		}
		judoCats = judoCatsResp.Categories
	}

	return render(c, pages.EventEdit(eventResp.Event, eventResp.Categories, judoCats, nil, csrf(c)))
}

func (h *EventHandler) HandleEditEvent(c *echo.Context) error {
	eventId := c.Param("id")
	userId, _ := c.Get("userId").(string)

	eventResp, err := h.eventSvc.GetEvent(c.Request().Context(), eventId)
	if err != nil {
		return err
	}
	if eventResp.Event == nil {
		return echo.NewHTTPError(http.StatusNotFound, "Événement introuvable.")
	}
	if eventResp.Event.Status != "draft" {
		return echo.NewHTTPError(http.StatusForbidden, "Seuls les événements en brouillon peuvent être modifiés.")
	}

	canManage, err := h.roleChecker.HasRole(c.Request().Context(), userId, eventResp.Event.ClubId.String(), "president", "staff")
	if err != nil {
		return err
	}
	if !canManage {
		return echo.NewHTTPError(http.StatusForbidden, "Accès refusé.")
	}

	if err := c.Request().ParseForm(); err != nil {
		return err
	}

	maxPart, _ := strconv.Atoi(c.FormValue("max_participants"))

	// Build category update when the form includes the categories sentinel.
	var updateCats []*dto.UpdateEventCategoryRequest
	if c.FormValue("categories_submitted") == "1" {
		eventDate := timestampToDate(eventResp.Event.Date)
		judoCatIds := c.Request().Form["category_judo_id"]
		weighIns := c.Request().Form["category_weigh_in"]
		for i, jcId := range judoCatIds {
			if jcId == "" {
				continue
			}
			cr := &dto.UpdateEventCategoryRequest{JudoCategoryId: jcId}
			if i < len(weighIns) && weighIns[i] != "" {
				cr.WeighInAt = parseDateWithTime(eventDate, weighIns[i])
			}
			updateCats = append(updateCats, cr)
		}
		if updateCats == nil {
			updateCats = []*dto.UpdateEventCategoryRequest{}
		}
	}

	ctx := context.WithValue(c.Request().Context(), "user_id", userId)
	resp, err := h.eventSvc.UpdateEvent(ctx, &dto.UpdateEventRequest{
		Id:                  eventId,
		Title:               c.FormValue("title"),
		Description:         c.FormValue("description"),
		Location:            c.FormValue("location"),
		RegistrationOpenAt:  parseHTMLDatetime(c.FormValue("registration_open_at")),
		RegistrationCloseAt: parseHTMLDatetime(c.FormValue("registration_close_at")),
		Date:                parseDateOnly(c.FormValue("date")),
		MaxParticipants:     maxPart,
		Categories:          updateCats,
	})
	if err != nil {
		return err
	}
	if len(resp.Errors) > 0 {
		var judoCats []*eventDomain.JudoCategory
		if eventResp.Event.Type == eventDomain.TypeCompetition {
			if judoCatsResp, err2 := h.eventSvc.ListJudoCategories(c.Request().Context()); err2 == nil {
				judoCats = judoCatsResp.Categories
			}
		}
		return render(c, pages.EventEdit(eventResp.Event, eventResp.Categories, judoCats, resp.Errors, csrf(c)))
	}

	return c.Redirect(http.StatusSeeOther, "/events/"+eventId)
}

func (h *EventHandler) HandleLeaveEvent(c *echo.Context) error {
	eventId := c.Param("id")
	userId, _ := c.Get("userId").(string)
	memberId := c.FormValue("member_id")

	ctx := context.WithValue(c.Request().Context(), "user_id", userId)
	if _, err := h.eventSvc.UnregisterParticipant(ctx, &dto.UnregisterParticipantRequest{
		EventId:  eventId,
		MemberId: memberId,
	}); err != nil {
		return err
	}

	return c.Redirect(http.StatusSeeOther, "/events/"+eventId+"/join")
}

// ── Carpool handlers ──────────────────────────────────────────────────────────

func (h *EventHandler) HandleCarpoolOfferPage(c *echo.Context) error {
	eventId := c.Param("id")
	userId, _ := c.Get("userId").(string)

	eventResp, err := h.eventSvc.GetEvent(c.Request().Context(), eventId)
	if err != nil {
		return err
	}
	if eventResp.Event == nil {
		return echo.NewHTTPError(http.StatusNotFound, "Événement introuvable.")
	}
	if eventResp.Event.Status != "open" {
		return echo.NewHTTPError(http.StatusForbidden, "Les inscriptions ne sont pas ouvertes.")
	}

	// Find the user's registered members for this event
	participantsResp, err := h.eventSvc.GetParticipants(c.Request().Context(), eventId)
	if err != nil {
		return err
	}
	var myMembers []*memberDomain.Member
	for _, p := range participantsResp.Participants {
		m, err2 := h.memberSvc.GetMember(c.Request().Context(), p.MemberId.String())
		if err2 == nil && m != nil && m.UserId.String() == userId {
			myMembers = append(myMembers, m)
		}
	}
	if len(myMembers) == 0 {
		return echo.NewHTTPError(http.StatusForbidden, "Vous devez être inscrit à l'événement pour proposer un covoiturage.")
	}

	return render(c, pages.CarpoolForm(eventResp.Event, myMembers, nil, csrf(c)))
}

func (h *EventHandler) HandleCreateCarpoolOffer(c *echo.Context) error {
	eventId := c.Param("id")
	userId, _ := c.Get("userId").(string)
	memberId := c.FormValue("member_id")

	// Verify the member belongs to the current user
	m, err := h.memberSvc.GetMember(c.Request().Context(), memberId)
	if err != nil {
		return err
	}
	if m == nil || m.UserId.String() != userId {
		return echo.NewHTTPError(http.StatusForbidden, "Accès refusé.")
	}

	seats, _ := strconv.Atoi(c.FormValue("available_seats"))
	ctx := context.WithValue(c.Request().Context(), "user_id", userId)
	resp, err := h.eventSvc.CreateCarpoolOffer(ctx, &dto.CreateCarpoolOfferRequest{
		EventId:          eventId,
		MemberId:         memberId,
		DepartureAddress: c.FormValue("departure_address"),
		AvailableSeats:   seats,
		DepartureAt:      parseHTMLDatetime(c.FormValue("departure_at")),
	})
	if err != nil {
		return err
	}
	if len(resp.Errors) > 0 {
		eventResp, _ := h.eventSvc.GetEvent(c.Request().Context(), eventId)
		var myMembers []*memberDomain.Member
		if eventResp != nil {
			participantsResp, _ := h.eventSvc.GetParticipants(c.Request().Context(), eventId)
			for _, p := range participantsResp.Participants {
				mm, err2 := h.memberSvc.GetMember(c.Request().Context(), p.MemberId.String())
				if err2 == nil && mm != nil && mm.UserId.String() == userId {
					myMembers = append(myMembers, mm)
				}
			}
			return render(c, pages.CarpoolForm(eventResp.Event, myMembers, resp.Errors, csrf(c)))
		}
		return echo.NewHTTPError(http.StatusBadRequest, "Erreur lors de la création de l'offre.")
	}

	return c.Redirect(http.StatusSeeOther, "/events/"+eventId)
}

func (h *EventHandler) HandleJoinCarpoolOffer(c *echo.Context) error {
	eventId := c.Param("id")
	offerId := c.Param("offerId")
	userId, _ := c.Get("userId").(string)

	// Auto-detect user's registered member for this event
	participantsResp, err := h.eventSvc.GetParticipants(c.Request().Context(), eventId)
	if err != nil {
		return err
	}
	memberId := ""
	for _, p := range participantsResp.Participants {
		m, err2 := h.memberSvc.GetMember(c.Request().Context(), p.MemberId.String())
		if err2 == nil && m != nil && m.UserId.String() == userId {
			memberId = p.MemberId.String()
			break
		}
	}
	if memberId == "" {
		return echo.NewHTTPError(http.StatusForbidden, "Vous devez être inscrit à l'événement.")
	}

	ctx := context.WithValue(c.Request().Context(), "user_id", userId)
	if _, err := h.eventSvc.JoinCarpoolOffer(ctx, offerId, memberId); err != nil {
		return c.Redirect(http.StatusSeeOther, "/events/"+eventId+"?error="+url.QueryEscape(err.Error()))
	}

	return c.Redirect(http.StatusSeeOther, "/events/"+eventId)
}

func (h *EventHandler) HandleLeaveCarpoolOffer(c *echo.Context) error {
	offerId := c.Param("offerId")
	userId, _ := c.Get("userId").(string)
	eventId := c.FormValue("event_id")

	if eventId == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "event_id manquant.")
	}

	// Find user's registered member for this event
	participantsResp, err := h.eventSvc.GetParticipants(c.Request().Context(), eventId)
	if err != nil {
		return err
	}
	memberId := ""
	for _, p := range participantsResp.Participants {
		m, err2 := h.memberSvc.GetMember(c.Request().Context(), p.MemberId.String())
		if err2 == nil && m != nil && m.UserId.String() == userId {
			memberId = p.MemberId.String()
			break
		}
	}
	if memberId == "" {
		return c.Redirect(http.StatusSeeOther, "/home")
	}

	ctx := context.WithValue(c.Request().Context(), "user_id", userId)
	if _, err := h.eventSvc.LeaveCarpoolOffer(ctx, offerId, memberId); err != nil {
		return c.Redirect(http.StatusSeeOther, "/events/"+eventId+"?error="+url.QueryEscape(err.Error()))
	}

	return c.Redirect(http.StatusSeeOther, "/events/"+eventId)
}

func (h *EventHandler) HandleCancelCarpoolOffer(c *echo.Context) error {
	offerId := c.Param("offerId")
	userId, _ := c.Get("userId").(string)
	eventId := c.FormValue("event_id")

	if eventId == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "event_id manquant.")
	}

	ctx := context.WithValue(c.Request().Context(), "user_id", userId)
	if _, err := h.eventSvc.CancelCarpoolOffer(ctx, offerId); err != nil {
		if eventId != "" {
			return c.Redirect(http.StatusSeeOther, "/events/"+eventId+"?error="+url.QueryEscape(err.Error()))
		}
		return err
	}

	if eventId != "" {
		return c.Redirect(http.StatusSeeOther, "/events/"+eventId)
	}
	return c.Redirect(http.StatusSeeOther, "/home")
}
