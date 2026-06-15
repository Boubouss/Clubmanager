package handlers

import (
	"clubmanager/internal/adapters/api/grpc/dto"
	"clubmanager/internal/adapters/api/http/views/pages"
	"clubmanager/internal/app/services"
	memberDomain "clubmanager/internal/domain/members"
	"clubmanager/internal/domain/roles"
	"context"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v5"
)

type MemberHandler struct {
	memberSvc   services.MemberService
	clubSvc     services.ClubService
	roleSvc     services.RoleService
	roleChecker roles.RoleChecker
}

func NewMemberHandler(memberSvc services.MemberService, clubSvc services.ClubService, roleSvc services.RoleService, roleChecker roles.RoleChecker) *MemberHandler {
	return &MemberHandler{memberSvc: memberSvc, clubSvc: clubSvc, roleSvc: roleSvc, roleChecker: roleChecker}
}

func (h *MemberHandler) HandleMemberList(c *echo.Context) error {
	clubId := c.Param("id")
	userId, _ := c.Get("userId").(string)

	clubResp, err := h.clubSvc.GetClub(c.Request().Context(), clubId)
	if err != nil {
		return err
	}
	if clubResp.Club == nil {
		return echo.NewHTTPError(http.StatusNotFound, "Club introuvable.")
	}

	canManage, err := h.roleChecker.HasRole(c.Request().Context(), userId, clubId, "president", "staff", "coach")
	if err != nil {
		return err
	}
	isPresident, err := h.roleChecker.HasRole(c.Request().Context(), userId, clubId, "president")
	if err != nil {
		return err
	}

	pageStr := c.QueryParam("page")
	page, _ := strconv.Atoi(pageStr)
	if page <= 0 {
		page = 1
	}
	const membersPageSize = 10

	membersResp, err := h.memberSvc.GetMembersByClub(c.Request().Context(), &dto.GetMembersByClubRequest{
		ClubId: clubId, Page: page, PageSize: membersPageSize,
	})
	if err != nil {
		return err
	}

	rolesResp, err := h.roleSvc.GetClubRoles(c.Request().Context(), &dto.GetClubRolesRequest{ClubId: clubId})
	if err != nil {
		return err
	}

	// Build userId → role map
	roleByUserId := make(map[string]string)
	roleIdByUserId := make(map[string]string)
	for _, r := range rolesResp.Roles {
		roleByUserId[r.UserId.String()] = r.Role
		roleIdByUserId[r.UserId.String()] = r.Id.String()
	}

	// Non-managers only see their own members
	members := membersResp.Members
	total := membersResp.Total
	var staffContacts []dto.StaffContact
	if !canManage {
		var own []dto.MemberWithMembership
		for _, mwm := range members {
			if mwm.Member.UserId.String() == userId {
				own = append(own, mwm)
			}
		}
		members = own
		total = len(own)

		// Fetch president/coach contacts to display to associates
		staffResp, err := h.memberSvc.GetStaffContacts(c.Request().Context(), clubId)
		if err != nil {
			return err
		}
		staffContacts = staffResp.Contacts
	}

	return render(c, pages.MembersList(clubResp.Club, members, total, page, membersPageSize, roleByUserId, roleIdByUserId, canManage, isPresident, staffContacts, csrf(c)))
}

func (h *MemberHandler) HandleMyMembersPage(c *echo.Context) error {
	userId, _ := c.Get("userId").(string)

	resp, err := h.memberSvc.GetMembersByUser(c.Request().Context(), &dto.GetMembersByUserRequest{UserId: userId})
	if err != nil {
		return err
	}

	return render(c, pages.MemberList(resp.Members))
}

func (h *MemberHandler) HandleCreateMemberPage(c *echo.Context) error {
	return render(c, pages.MemberCreate(nil, csrf(c)))
}

func (h *MemberHandler) HandleCreateMember(c *echo.Context) error {
	userId, _ := c.Get("userId").(string)

	resp, err := h.memberSvc.CreateMember(c.Request().Context(), &dto.CreateMemberRequest{
		UserId:    userId,
		Firstname: c.FormValue("firstname"),
		Lastname:  c.FormValue("lastname"),
		Birthdate: c.FormValue("birthdate"),
		Gender:    c.FormValue("gender"),
		IsPrimary: c.FormValue("is_primary") == "true",
	})
	if err != nil {
		return err
	}

	if len(resp.Errors) > 0 {
		return render(c, pages.MemberCreate(resp.Errors, csrf(c)))
	}

	return c.Redirect(http.StatusSeeOther, "/members")
}

func (h *MemberHandler) HandleEditMemberPage(c *echo.Context) error {
	memberId := c.Param("id")
	userId, _ := c.Get("userId").(string)

	resp, err := h.memberSvc.GetMembersByUser(c.Request().Context(), &dto.GetMembersByUserRequest{UserId: userId})
	if err != nil {
		return err
	}
	var target *memberDomain.Member
	for _, m := range resp.Members {
		if m.Id.String() == memberId {
			target = m
			break
		}
	}
	if target == nil {
		return echo.NewHTTPError(http.StatusNotFound, "Profil introuvable.")
	}

	return render(c, pages.MemberEdit(target, nil, csrf(c)))
}

func (h *MemberHandler) HandleEditMember(c *echo.Context) error {
	memberId := c.Param("id")
	userId, _ := c.Get("userId").(string)

	// Verify ownership
	resp, err := h.memberSvc.GetMembersByUser(c.Request().Context(), &dto.GetMembersByUserRequest{UserId: userId})
	if err != nil {
		return err
	}
	var target *memberDomain.Member
	for _, m := range resp.Members {
		if m.Id.String() == memberId {
			target = m
			break
		}
	}
	if target == nil {
		return echo.NewHTTPError(http.StatusForbidden, "Accès refusé.")
	}

	updateResp, err := h.memberSvc.UpdateMember(c.Request().Context(), &dto.UpdateMemberRequest{
		Id:        memberId,
		Firstname: c.FormValue("firstname"),
		Lastname:  c.FormValue("lastname"),
		Birthdate: c.FormValue("birthdate"),
		Gender:    c.FormValue("gender"),
	})
	if err != nil {
		return err
	}
	if len(updateResp.Errors) > 0 {
		return render(c, pages.MemberEdit(target, updateResp.Errors, csrf(c)))
	}

	return c.Redirect(http.StatusSeeOther, "/members")
}

func (h *MemberHandler) HandleDeleteMember(c *echo.Context) error {
	memberId := c.Param("id")
	userId, _ := c.Get("userId").(string)

	// Verify ownership
	resp, err := h.memberSvc.GetMembersByUser(c.Request().Context(), &dto.GetMembersByUserRequest{UserId: userId})
	if err != nil {
		return err
	}
	var found bool
	for _, m := range resp.Members {
		if m.Id.String() == memberId {
			found = true
			break
		}
	}
	if !found {
		return echo.NewHTTPError(http.StatusForbidden, "Accès refusé.")
	}

	if _, err := h.memberSvc.RemoveMember(c.Request().Context(), memberId); err != nil {
		return err
	}

	return c.Redirect(http.StatusSeeOther, "/members")
}

func (h *MemberHandler) HandleRequestMembershipPage(c *echo.Context) error {
	clubId := c.Param("id")
	userId, _ := c.Get("userId").(string)

	clubResp, err := h.clubSvc.GetClub(c.Request().Context(), clubId)
	if err != nil {
		return err
	}
	if clubResp.Club == nil {
		return echo.NewHTTPError(http.StatusNotFound, "Club introuvable.")
	}

	membersResp, err := h.memberSvc.GetMembersByUser(c.Request().Context(), &dto.GetMembersByUserRequest{UserId: userId})
	if err != nil {
		return err
	}

	return render(c, pages.ClubJoin(clubResp.Club, membersResp.Members, nil, csrf(c)))
}

func (h *MemberHandler) HandleRequestMembership(c *echo.Context) error {
	clubId := c.Param("id")
	userId, _ := c.Get("userId").(string)

	clubResp, err := h.clubSvc.GetClub(c.Request().Context(), clubId)
	if err != nil {
		return err
	}
	if clubResp.Club == nil {
		return echo.NewHTTPError(http.StatusNotFound, "Club introuvable.")
	}

	resp, err := h.memberSvc.RequestMembership(c.Request().Context(), &dto.RequestMembershipRequest{
		MemberId: c.FormValue("member_id"),
		ClubId:   clubId,
		UserId:   userId,
	})
	if err != nil {
		return err
	}

	if len(resp.Errors) > 0 {
		membersResp, err := h.memberSvc.GetMembersByUser(c.Request().Context(), &dto.GetMembersByUserRequest{UserId: userId})
		if err != nil {
			return err
		}
		return render(c, pages.ClubJoin(clubResp.Club, membersResp.Members, resp.Errors, csrf(c)))
	}

	return c.Redirect(http.StatusSeeOther, "/clubs/"+clubId+"/members")
}

func (h *MemberHandler) HandleMemberAddPage(c *echo.Context) error {
	clubId := c.Param("id")
	userId, _ := c.Get("userId").(string)

	clubResp, err := h.clubSvc.GetClub(c.Request().Context(), clubId)
	if err != nil {
		return err
	}
	if clubResp.Club == nil {
		return echo.NewHTTPError(http.StatusNotFound, "Club introuvable.")
	}

	// Load the user's own member profiles
	userMembersResp, err := h.memberSvc.GetMembersByUser(c.Request().Context(), &dto.GetMembersByUserRequest{UserId: userId})
	if err != nil {
		return err
	}

	// Build alreadyInClub map
	clubMembersResp, err := h.memberSvc.GetMembersByClub(c.Request().Context(), &dto.GetMembersByClubRequest{ClubId: clubId, Page: 1, PageSize: 100})
	if err != nil {
		return err
	}
	alreadyInClub := make(map[string]bool)
	for _, mwm := range clubMembersResp.Members {
		alreadyInClub[mwm.Member.Id.String()] = true
	}

	// Only pass members not already in the club
	var linkableMembers []*memberDomain.Member
	for _, m := range userMembersResp.Members {
		if !alreadyInClub[m.Id.String()] {
			linkableMembers = append(linkableMembers, m)
		}
	}

	return render(c, pages.MemberAdd(clubResp.Club, linkableMembers, nil, csrf(c)))
}

func (h *MemberHandler) HandleMemberAdd(c *echo.Context) error {
	clubId := c.Param("id")
	userId, _ := c.Get("userId").(string)

	clubResp, err := h.clubSvc.GetClub(c.Request().Context(), clubId)
	if err != nil {
		return err
	}
	if clubResp.Club == nil {
		return echo.NewHTTPError(http.StatusNotFound, "Club introuvable.")
	}

	// Link existing member profile
	if existingMemberId := c.FormValue("existing_member_id"); existingMemberId != "" {
		// Fetch the member to get their userId before creating membership
		clubMembersResp, err := h.memberSvc.GetMembersByClub(c.Request().Context(), &dto.GetMembersByClubRequest{ClubId: clubId, Page: 1, PageSize: 100})
		if err != nil {
			return err
		}
		// Check not already in club
		for _, mwm := range clubMembersResp.Members {
			if mwm.Member.Id.String() == existingMemberId {
				return c.Redirect(http.StatusSeeOther, "/clubs/"+clubId+"/members")
			}
		}

		if _, err := h.memberSvc.AddMembership(c.Request().Context(), existingMemberId, clubId); err != nil {
			return err
		}

		// Assign associate role if not already set — need userId from member search
		// We search clubs members again after add is too late; fetch user members instead
		// We'll look up via GetMembersByClub post-add to find the just-added member
		updatedResp, err := h.memberSvc.GetMembersByClub(c.Request().Context(), &dto.GetMembersByClubRequest{ClubId: clubId, Page: 1, PageSize: 100})
		if err == nil {
			for _, mwm := range updatedResp.Members {
				if mwm.Member.Id.String() == existingMemberId {
					targetUserId := mwm.Member.UserId.String()
					hasRole, _ := h.roleChecker.HasRole(c.Request().Context(), targetUserId, clubId, "president", "staff", "coach", "associate")
					if !hasRole {
						ctx := context.WithValue(c.Request().Context(), "user_id", userId)
						h.roleSvc.AssignRole(ctx, &dto.AssignRoleRequest{
							UserId: targetUserId,
							ClubId: clubId,
							Role:   "associate",
						})
					}
					break
				}
			}
		}

		return c.Redirect(http.StatusSeeOther, "/clubs/"+clubId+"/members")
	}

	resp, err := h.memberSvc.AddMember(c.Request().Context(), &dto.AddMemberRequest{
		UserId:    userId,
		ClubId:    clubId,
		Firstname: c.FormValue("firstname"),
		Lastname:  c.FormValue("lastname"),
		Birthdate: c.FormValue("birthdate"),
		Gender:    c.FormValue("gender"),
		IsPrimary: c.FormValue("is_primary") == "true",
	})
	if err != nil {
		return err
	}

	if len(resp.Errors) > 0 {
		var linkableMembers []*memberDomain.Member
		if userMembersResp, err2 := h.memberSvc.GetMembersByUser(c.Request().Context(), &dto.GetMembersByUserRequest{UserId: userId}); err2 == nil {
			if clubMembersResp, err3 := h.memberSvc.GetMembersByClub(c.Request().Context(), &dto.GetMembersByClubRequest{ClubId: clubId, Page: 1, PageSize: 100}); err3 == nil {
				inClub := make(map[string]bool)
				for _, mwm := range clubMembersResp.Members {
					inClub[mwm.Member.Id.String()] = true
				}
				for _, m := range userMembersResp.Members {
					if !inClub[m.Id.String()] {
						linkableMembers = append(linkableMembers, m)
					}
				}
			}
		}
		return render(c, pages.MemberAdd(clubResp.Club, linkableMembers, resp.Errors, csrf(c)))
	}

	// Only assign associate if the user has no role yet in this club
	targetUserId := resp.Member.UserId.String()
	hasRole, err := h.roleChecker.HasRole(c.Request().Context(), targetUserId, clubId, "president", "staff", "coach", "associate")
	if err != nil {
		return err
	}
	if !hasRole {
		ctx := context.WithValue(c.Request().Context(), "user_id", userId)
		h.roleSvc.AssignRole(ctx, &dto.AssignRoleRequest{
			UserId: targetUserId,
			ClubId: clubId,
			Role:   "associate",
		})
	}

	return c.Redirect(http.StatusSeeOther, "/clubs/"+clubId+"/members")
}

func (h *MemberHandler) HandleValidateMember(c *echo.Context) error {
	clubId := c.Param("id")
	membershipId := c.Param("membershipId")
	userId, _ := c.Get("userId").(string)

	canManage, err := h.roleChecker.HasRole(c.Request().Context(), userId, clubId, "president", "staff")
	if err != nil {
		return err
	}
	if !canManage {
		return echo.NewHTTPError(http.StatusForbidden, "Accès refusé.")
	}

	resp, err := h.memberSvc.ValidateMember(c.Request().Context(), &dto.ValidateMemberRequest{Id: membershipId})
	if err != nil {
		return err
	}
	if len(resp.Errors) > 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "Validation impossible.")
	}

	// Only assign associate if the user has no role yet in this club
	targetUserId := resp.Member.UserId.String()
	hasRole, err := h.roleChecker.HasRole(c.Request().Context(), targetUserId, clubId, "president", "staff", "coach", "associate")
	if err != nil {
		return err
	}
	if !hasRole {
		ctx := context.WithValue(c.Request().Context(), "user_id", userId)
		h.roleSvc.AssignRole(ctx, &dto.AssignRoleRequest{
			UserId: targetUserId,
			ClubId: clubId,
			Role:   "associate",
		})
	}

	return c.Redirect(http.StatusSeeOther, "/clubs/"+clubId+"/members")
}

func (h *MemberHandler) HandleRefuseMember(c *echo.Context) error {
	clubId := c.Param("id")
	membershipId := c.Param("membershipId")
	userId, _ := c.Get("userId").(string)

	isPresident, err := h.roleChecker.HasRole(c.Request().Context(), userId, clubId, "president")
	if err != nil {
		return err
	}
	if !isPresident {
		return echo.NewHTTPError(http.StatusForbidden, "Accès refusé.")
	}

	if _, err := h.memberSvc.RemoveMembership(c.Request().Context(), membershipId); err != nil {
		return err
	}

	return c.Redirect(http.StatusSeeOther, "/clubs/"+clubId+"/members")
}

func (h *MemberHandler) HandleRemoveMember(c *echo.Context) error {
	clubId := c.Param("id")
	membershipId := c.Param("membershipId")
	userId, _ := c.Get("userId").(string)

	isPresident, err := h.roleChecker.HasRole(c.Request().Context(), userId, clubId, "president")
	if err != nil {
		return err
	}
	if !isPresident {
		return echo.NewHTTPError(http.StatusForbidden, "Accès refusé.")
	}

	if _, err := h.memberSvc.RemoveMembership(c.Request().Context(), membershipId); err != nil {
		return err
	}

	return h.renderMembersFragment(c, clubId, userId)
}

func (h *MemberHandler) HandleAssignRole(c *echo.Context) error {
	clubId := c.Param("id")
	userId, _ := c.Get("userId").(string)
	targetUserId := c.FormValue("target_user_id")

	ctx := context.WithValue(c.Request().Context(), "user_id", userId)

	// Remove existing role for this user in this club before assigning a new one
	existingRoles, err := h.roleSvc.GetClubRoles(ctx, &dto.GetClubRolesRequest{ClubId: clubId})
	if err != nil {
		return err
	}
	for _, r := range existingRoles.Roles {
		if r.UserId.String() == targetUserId {
			h.roleSvc.RemoveRole(ctx, &dto.RemoveRoleRequest{Id: r.Id.String()})
			break
		}
	}

	resp, err := h.roleSvc.AssignRole(ctx, &dto.AssignRoleRequest{
		UserId: targetUserId,
		ClubId: clubId,
		Role:   c.FormValue("role"),
	})
	if err != nil {
		return err
	}
	if len(resp.Errors) > 0 {
		return echo.NewHTTPError(http.StatusForbidden, "Accès refusé.")
	}

	return h.renderMembersFragment(c, clubId, userId)
}

func (h *MemberHandler) HandleRemoveRole(c *echo.Context) error {
	clubId := c.Param("id")
	roleId := c.Param("roleId")
	userId, _ := c.Get("userId").(string)

	ctx := context.WithValue(c.Request().Context(), "user_id", userId)

	if _, err := h.roleSvc.RemoveRole(ctx, &dto.RemoveRoleRequest{Id: roleId}); err != nil {
		return echo.NewHTTPError(http.StatusForbidden, "Accès refusé.")
	}

	return h.renderMembersFragment(c, clubId, userId)
}

func (h *MemberHandler) renderMembersFragment(c *echo.Context, clubId, userId string) error {
	canManage, err := h.roleChecker.HasRole(c.Request().Context(), userId, clubId, "president", "staff", "coach")
	if err != nil {
		return err
	}
	isPresident, err := h.roleChecker.HasRole(c.Request().Context(), userId, clubId, "president")
	if err != nil {
		return err
	}

	clubResp, err := h.clubSvc.GetClub(c.Request().Context(), clubId)
	if err != nil {
		return err
	}

	membersResp, err := h.memberSvc.GetMembersByClub(c.Request().Context(), &dto.GetMembersByClubRequest{
		ClubId: clubId, Page: 1, PageSize: 100,
	})
	if err != nil {
		return err
	}

	rolesResp, err := h.roleSvc.GetClubRoles(c.Request().Context(), &dto.GetClubRolesRequest{ClubId: clubId})
	if err != nil {
		return err
	}

	roleByUserId := make(map[string]string)
	roleIdByUserId := make(map[string]string)
	for _, r := range rolesResp.Roles {
		roleByUserId[r.UserId.String()] = r.Role
		roleIdByUserId[r.UserId.String()] = r.Id.String()
	}

	members := membersResp.Members
	if !canManage {
		var own []dto.MemberWithMembership
		for _, mwm := range members {
			if mwm.Member.UserId.String() == userId {
				own = append(own, mwm)
			}
		}
		members = own
	}

	return render(c, pages.MembersTable(clubResp.Club, members, roleByUserId, roleIdByUserId, canManage, isPresident, csrf(c)))
}
