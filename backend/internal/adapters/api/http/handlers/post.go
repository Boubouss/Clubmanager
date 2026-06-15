package handlers

import (
	"clubmanager/internal/adapters/api/grpc/dto"
	"clubmanager/internal/adapters/api/http/views/pages"
	"clubmanager/internal/app/services"
	"clubmanager/internal/domain/roles"
	"context"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v5"
)

type PostHandler struct {
	postSvc     services.PostService
	clubSvc     services.ClubService
	roleChecker roles.RoleChecker
}

func NewPostHandler(postSvc services.PostService, clubSvc services.ClubService, roleChecker roles.RoleChecker) *PostHandler {
	return &PostHandler{postSvc: postSvc, clubSvc: clubSvc, roleChecker: roleChecker}
}

func (h *PostHandler) HandlePostList(c *echo.Context) error {
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

	page := 1
	if p := c.QueryParam("page"); p != "" {
		if n, err := strconv.Atoi(p); err == nil && n > 0 {
			page = n
		}
	}

	ctx := context.WithValue(c.Request().Context(), "user_id", userId)
	resp, err := h.postSvc.GetClubPosts(ctx, &dto.GetClubPostsRequest{
		ClubId:   clubId,
		Page:     page,
		PageSize: 10,
	})
	if err != nil {
		return err
	}

	return render(c, pages.PostsList(clubResp.Club, resp.Posts, resp.Total, resp.Page, resp.PageSize, canManage, csrf(c)))
}

func (h *PostHandler) HandleNewPostPage(c *echo.Context) error {
	clubId := c.Param("id")
	userId, _ := c.Get("userId").(string)

	canManage, err := h.roleChecker.HasRole(c.Request().Context(), userId, clubId, "president", "staff", "coach")
	if err != nil {
		return err
	}
	if !canManage {
		return echo.NewHTTPError(http.StatusForbidden, "Accès refusé.")
	}

	return render(c, pages.PostForm(clubId, nil, nil, csrf(c)))
}

func (h *PostHandler) HandleCreatePost(c *echo.Context) error {
	clubId := c.Param("id")
	userId, _ := c.Get("userId").(string)

	ctx := context.WithValue(c.Request().Context(), "user_id", userId)
	resp, err := h.postSvc.CreatePost(ctx, &dto.CreatePostRequest{
		ClubId:     clubId,
		Title:      c.FormValue("title"),
		Content:    c.FormValue("content"),
		Visibility: c.FormValue("visibility"),
		Status:     c.FormValue("status"),
	})
	if err != nil {
		return err
	}
	if len(resp.Errors) > 0 {
		return render(c, pages.PostForm(clubId, nil, resp.Errors, csrf(c)))
	}

	return c.Redirect(http.StatusSeeOther, "/posts/"+resp.Post.Id.String())
}

func (h *PostHandler) HandlePostDetail(c *echo.Context) error {
	postId := c.Param("id")
	userId, _ := c.Get("userId").(string)

	ctx := context.WithValue(c.Request().Context(), "user_id", userId)
	post, err := h.postSvc.GetPost(ctx, postId)
	if err != nil {
		return err
	}
	if post == nil {
		return echo.NewHTTPError(http.StatusNotFound, "Post introuvable.")
	}

	clubResp, err := h.clubSvc.GetClub(c.Request().Context(), post.ClubId.String())
	if err != nil {
		return err
	}
	clubName := ""
	if clubResp.Club != nil {
		clubName = clubResp.Club.Name
	}

	canManage, err := h.roleChecker.HasRole(c.Request().Context(), userId, post.ClubId.String(), "president", "staff", "coach")
	if err != nil {
		return err
	}

	isMember, err := h.roleChecker.HasRole(c.Request().Context(), userId, post.ClubId.String(), "associate")
	if err != nil {
		return err
	}

	isAuthor := post.AuthorId.String() == userId

	return render(c, pages.PostDetail(post, clubName, isAuthor, canManage, isMember, csrf(c)))
}

func (h *PostHandler) HandleEditPostPage(c *echo.Context) error {
	postId := c.Param("id")
	userId, _ := c.Get("userId").(string)

	ctx := context.WithValue(c.Request().Context(), "user_id", userId)
	post, err := h.postSvc.GetPost(ctx, postId)
	if err != nil {
		return err
	}
	if post == nil {
		return echo.NewHTTPError(http.StatusNotFound, "Post introuvable.")
	}

	if post.Status != "draft" {
		return echo.NewHTTPError(http.StatusForbidden, "Seuls les brouillons peuvent être modifiés.")
	}

	isAuthor := post.AuthorId.String() == userId
	canManage, err := h.roleChecker.HasRole(c.Request().Context(), userId, post.ClubId.String(), "president", "staff", "coach")
	if err != nil {
		return err
	}
	if !isAuthor && !canManage {
		return echo.NewHTTPError(http.StatusForbidden, "Accès refusé.")
	}

	return render(c, pages.PostForm(post.ClubId.String(), post, nil, csrf(c)))
}

func (h *PostHandler) HandleEditPost(c *echo.Context) error {
	postId := c.Param("id")
	userId, _ := c.Get("userId").(string)

	ctx := context.WithValue(c.Request().Context(), "user_id", userId)

	// Need club_id for the form on error — fetch the post first
	post, err := h.postSvc.GetPost(ctx, postId)
	if err != nil {
		return err
	}
	if post == nil {
		return echo.NewHTTPError(http.StatusNotFound, "Post introuvable.")
	}

	resp, err := h.postSvc.UpdatePost(ctx, &dto.UpdatePostRequest{
		Id:         postId,
		Title:      c.FormValue("title"),
		Content:    c.FormValue("content"),
		Visibility: c.FormValue("visibility"),
	})
	if err != nil {
		return err
	}
	if len(resp.Errors) > 0 {
		return render(c, pages.PostForm(post.ClubId.String(), post, resp.Errors, csrf(c)))
	}

	return c.Redirect(http.StatusSeeOther, "/posts/"+postId)
}

func (h *PostHandler) HandlePublishPost(c *echo.Context) error {
	postId := c.Param("id")
	userId, _ := c.Get("userId").(string)

	ctx := context.WithValue(c.Request().Context(), "user_id", userId)
	if _, err := h.postSvc.PublishPost(ctx, &dto.PublishPostRequest{Id: postId}); err != nil {
		return err
	}

	return c.Redirect(http.StatusSeeOther, "/posts/"+postId)
}

func (h *PostHandler) HandleUnpublishPost(c *echo.Context) error {
	postId := c.Param("id")
	userId, _ := c.Get("userId").(string)

	ctx := context.WithValue(c.Request().Context(), "user_id", userId)
	if _, err := h.postSvc.UnpublishPost(ctx, &dto.PublishPostRequest{Id: postId}); err != nil {
		return err
	}

	return c.Redirect(http.StatusSeeOther, "/posts/"+postId)
}

func (h *PostHandler) HandleDeletePost(c *echo.Context) error {
	postId := c.Param("id")
	userId, _ := c.Get("userId").(string)

	ctx := context.WithValue(c.Request().Context(), "user_id", userId)

	// Fetch post to get club_id for redirect
	post, err := h.postSvc.GetPost(ctx, postId)
	if err != nil {
		return err
	}
	if post == nil {
		return c.Redirect(http.StatusSeeOther, "/home")
	}
	clubId := post.ClubId.String()

	if _, err := h.postSvc.DeletePost(ctx, postId); err != nil {
		return err
	}

	return c.Redirect(http.StatusSeeOther, "/clubs/"+clubId+"/posts")
}
