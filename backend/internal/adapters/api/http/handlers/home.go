package handlers

import (
	"clubmanager/internal/adapters/api/grpc/dto"
	"clubmanager/internal/adapters/api/http/views/pages"
	"clubmanager/internal/app/services"
	"context"
	"net/http"

	"github.com/labstack/echo/v5"
)

type HomeHandler struct {
	memberSvc services.MemberService
	postSvc   services.PostService
}

func NewHomeHandler(memberSvc services.MemberService, postSvc services.PostService) *HomeHandler {
	return &HomeHandler{memberSvc: memberSvc, postSvc: postSvc}
}

func (h *HomeHandler) HandleLandingPage(c *echo.Context) error {
	return render(c, pages.Landing())
}

func (h *HomeHandler) HandleHomePage(c *echo.Context) error {
	userId, _ := c.Get("userId").(string)

	clubsResp, err := h.memberSvc.GetUserClubs(c.Request().Context(), &dto.GetUserClubsRequest{UserId: userId})
	if err != nil {
		return err
	}

	ctx := context.WithValue(c.Request().Context(), "user_id", userId)
	var recentPosts []pages.RecentPost
	for _, cl := range clubsResp.Clubs {
		postResp, err := h.postSvc.GetClubPosts(ctx, &dto.GetClubPostsRequest{
			ClubId:   cl.Id.String(),
			Page:     1,
			PageSize: 5,
		})
		if err != nil {
			return err
		}
		for _, p := range postResp.Posts {
			recentPosts = append(recentPosts, pages.RecentPost{
				Id:        p.Id.String(),
				EventId:   p.EventId.String(),
				EventDate: p.EventDate,
				Title:     p.Title,
				ClubName:  cl.Name,
				CreatedAt: p.CreatedAt,
			})
		}
	}

	return render(c, pages.Home(clubsResp.Clubs, recentPosts))
}

func (h *HomeHandler) HandleNotFound(c *echo.Context) error {
	return c.String(http.StatusNotFound, "Page introuvable")
}
