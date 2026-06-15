package handlers

import (
	"clubmanager/internal/adapters/api/grpc/dto"
	"clubmanager/internal/adapters/api/http/views/components"
	"clubmanager/internal/adapters/api/http/views/pages"
	"clubmanager/internal/app/services"
	membersDomain "clubmanager/internal/domain/members"
	"context"
	"net/http"

	"github.com/labstack/echo/v5"
)

type ClubHandler struct {
	clubSvc   services.ClubService
	memberSvc services.MemberService
	roleSvc   services.RoleService
	eventSvc  services.EventService
	postSvc   services.PostService
}

func NewClubHandler(clubSvc services.ClubService, memberSvc services.MemberService, roleSvc services.RoleService, eventSvc services.EventService, postSvc services.PostService) *ClubHandler {
	return &ClubHandler{clubSvc: clubSvc, memberSvc: memberSvc, roleSvc: roleSvc, eventSvc: eventSvc, postSvc: postSvc}
}

func (h *ClubHandler) HandleClubList(c *echo.Context) error {
	query := c.QueryParam("q")
	userId, _ := c.Get("userId").(string)

	if query == "" {
		// No search — load the user's own clubs
		resp, err := h.memberSvc.GetUserClubs(c.Request().Context(), &dto.GetUserClubsRequest{UserId: userId})
		if err != nil {
			return err
		}
		return render(c, pages.ClubsList(resp.Clubs, ""))
	}

	resp, err := h.clubSvc.ReadClub(c.Request().Context(), &dto.ReadClubRequest{Params: map[string]any{"name": query}})
	if err != nil {
		return err
	}

	return render(c, pages.ClubsList(resp.Clubs, query))
}

func (h *ClubHandler) HandleCreateClubPage(c *echo.Context) error {
	userId, _ := c.Get("userId").(string)
	membersResp, err := h.memberSvc.GetMembersByUser(c.Request().Context(), &dto.GetMembersByUserRequest{UserId: userId})
	if err != nil {
		return err
	}
	return render(c, pages.ClubCreate(nil, csrf(c), membersResp.Members))
}

func (h *ClubHandler) HandleMemberFieldsFragment(c *echo.Context) error {
	memberId := c.QueryParam("president_member_id")
	if memberId == "" || memberId == "new" {
		return render(c, components.MemberFieldsFragment(nil))
	}

	userId, _ := c.Get("userId").(string)
	membersResp, err := h.memberSvc.GetMembersByUser(c.Request().Context(), &dto.GetMembersByUserRequest{UserId: userId})
	if err != nil {
		return err
	}
	for _, m := range membersResp.Members {
		if m.Id.String() == memberId {
			return render(c, components.MemberFieldsFragment(m))
		}
	}
	return render(c, components.MemberFieldsFragment(nil))
}

func (h *ClubHandler) HandleCreateClub(c *echo.Context) error {
	userId, _ := c.Get("userId").(string)
	req := &dto.CreateClubRequest{
		UserId:      userId,
		Siren:       c.FormValue("siren"),
		Name:        c.FormValue("name"),
		Address:     c.FormValue("address"),
		City:        c.FormValue("city"),
		PostalCode:  c.FormValue("postal_code"),
		Country:     c.FormValue("country"),
		Phonenumber: c.FormValue("phonenumber"),
	}

	clubResp, err := h.clubSvc.CreateClub(c.Request().Context(), req)
	if err != nil {
		return err
	}

	if len(clubResp.Errors) > 0 {
		membersResp, _ := h.memberSvc.GetMembersByUser(c.Request().Context(), &dto.GetMembersByUserRequest{UserId: userId})
		var userMembers []*membersDomain.Member
		if membersResp != nil {
			userMembers = membersResp.Members
		}
		return render(c, pages.ClubCreate(clubResp.Errors, csrf(c), userMembers))
	}

	clubId := clubResp.Club.Id.String()
	presidentMemberId := c.FormValue("president_member_id")

	var memberId string
	if presidentMemberId == "" || presidentMemberId == "new" {
		// Create a new member profile
		createResp, err := h.memberSvc.CreateMember(c.Request().Context(), &dto.CreateMemberRequest{
			UserId:    userId,
			Firstname: c.FormValue("president_firstname"),
			Lastname:  c.FormValue("president_lastname"),
			Birthdate: c.FormValue("president_birthdate"),
			Gender:    c.FormValue("president_gender"),
			IsPrimary: true,
		})
		if err != nil {
			return err
		}
		if len(createResp.Errors) > 0 {
			membersResp, _ := h.memberSvc.GetMembersByUser(c.Request().Context(), &dto.GetMembersByUserRequest{UserId: userId})
			var userMembers []*membersDomain.Member
			if membersResp != nil {
				userMembers = membersResp.Members
			}
			return render(c, pages.ClubCreate(createResp.Errors, csrf(c), userMembers))
		}
		memberId = createResp.Member.Id.String()
	} else {
		memberId = presidentMemberId
	}

	// Create validated membership for the president
	if _, err := h.memberSvc.AddMembership(c.Request().Context(), memberId, clubId); err != nil {
		return err
	}

	// Assign president role
	ctx := context.WithValue(c.Request().Context(), "user_id", userId)
	if _, err := h.roleSvc.InitPresidentRole(ctx, userId, clubId); err != nil {
		return err
	}

	return c.Redirect(http.StatusSeeOther, "/clubs/"+clubId)
}

func (h *ClubHandler) HandleClubDetail(c *echo.Context) error {
	id := c.Param("id")
	userId, _ := c.Get("userId").(string)

	clubResp, err := h.clubSvc.GetClub(c.Request().Context(), id)
	if err != nil {
		return err
	}
	if clubResp.Club == nil {
		return echo.NewHTTPError(http.StatusNotFound, "Club introuvable.")
	}

	ctx := context.WithValue(c.Request().Context(), "user_id", userId)
	postResp, err := h.postSvc.GetClubPosts(ctx, &dto.GetClubPostsRequest{
		ClubId:   id,
		Page:     1,
		PageSize: 10,
	})
	if err != nil {
		return err
	}

	eventsResp, err := h.eventSvc.ListClubEvents(c.Request().Context(), &dto.ListClubEventsRequest{
		ClubId: id, Page: 1, PageSize: 5,
	})
	if err != nil {
		return err
	}

	return render(c, pages.ClubDetail(clubResp.Club, eventsResp.Events, postResp.Posts))
}
