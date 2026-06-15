package services

import (
	"clubmanager/internal/adapters/api/grpc/dto"
	"clubmanager/internal/domain/clubs"
	"clubmanager/internal/domain/posts"
	"clubmanager/internal/domain/roles"
	"context"
	"fmt"
)

// MembershipChecker is the port that the post service uses to determine
// whether a user has a valid club membership (for visibility enforcement).
type MembershipChecker interface {
	HasMembership(ctx context.Context, userId, clubId string) (bool, error)
}

type PostService interface {
	CreatePost(context.Context, *dto.CreatePostRequest) (*dto.CreatePostResponse, error)
	PublishPost(context.Context, *dto.PublishPostRequest) (*dto.PublishPostResponse, error)
	UnpublishPost(context.Context, *dto.PublishPostRequest) (*dto.PublishPostResponse, error)
	UpdatePost(context.Context, *dto.UpdatePostRequest) (*dto.UpdatePostResponse, error)
	DeletePost(context.Context, string) (bool, error)
	GetClubPosts(context.Context, *dto.GetClubPostsRequest) (*dto.GetClubPostsResponse, error)
	GetPost(context.Context, string) (*posts.Post, error)
	HasPostForEvent(ctx context.Context, eventId string) (bool, error)
}

type PostServiceConfig struct {
	Repository        posts.PostRepository
	RoleChecker       roles.RoleChecker
	ClubStatusChecker clubs.ClubStatusChecker
	MembershipChecker MembershipChecker
}

type postService struct {
	repo              posts.PostRepository
	roleChecker       roles.RoleChecker
	clubStatusChecker clubs.ClubStatusChecker
	memberChecker     MembershipChecker
}

func NewPostService(cfg PostServiceConfig) *postService {
	return &postService{
		repo:              cfg.Repository,
		roleChecker:       cfg.RoleChecker,
		clubStatusChecker: cfg.ClubStatusChecker,
		memberChecker:     cfg.MembershipChecker,
	}
}

// ── Helpers ──────────────────────────────────────────────────────────────────

func (s *postService) callerUserId(ctx context.Context) (string, bool) {
	id, ok := ctx.Value("user_id").(string)
	return id, ok && id != ""
}

// isClubManager returns true when userId is president, staff or coach in clubId, or is superadmin.
func (s *postService) isClubManager(ctx context.Context, userId, clubId string) (bool, error) {
	isSuperAdmin, err := s.roleChecker.IsSuperAdmin(ctx, userId)
	if err != nil {
		return false, err
	}
	if isSuperAdmin {
		return true, nil
	}
	return s.roleChecker.HasRole(ctx, userId, clubId, roles.RolePresident, roles.RoleStaff, roles.RoleCoach)
}

// ── Operations ────────────────────────────────────────────────────────────────

func (s *postService) CreatePost(ctx context.Context, data *dto.CreatePostRequest) (*dto.CreatePostResponse, error) {
	userId, ok := s.callerUserId(ctx)
	if !ok {
		return &dto.CreatePostResponse{Errors: map[string]string{"auth": "Missing user context."}}, nil
	}

	ok, err := s.isClubManager(ctx, userId, data.ClubId)
	if err != nil {
		return nil, err
	}
	if !ok {
		return &dto.CreatePostResponse{
			Errors: map[string]string{"auth": "Only club managers can create posts."},
		}, nil
	}

	if s.clubStatusChecker != nil {
		active, err := s.clubStatusChecker.IsActive(ctx, data.ClubId)
		if err != nil {
			return nil, fmt.Errorf("check club status: %w", err)
		}
		if !active {
			return &dto.CreatePostResponse{
				Errors: map[string]string{"club": "Club is not active."},
			}, nil
		}
	}

	post, errs := posts.NewPost(data.ClubId, userId, data.Map())
	if len(errs) > 0 {
		return &dto.CreatePostResponse{Errors: errs}, nil
	}

	created, err := s.repo.Save(ctx, post)
	if err != nil {
		return nil, err
	}
	return &dto.CreatePostResponse{Post: created, Errors: make(map[string]string)}, nil
}

func (s *postService) PublishPost(ctx context.Context, data *dto.PublishPostRequest) (*dto.PublishPostResponse, error) {
	userId, ok := s.callerUserId(ctx)
	if !ok {
		return &dto.PublishPostResponse{Errors: map[string]string{"auth": "Missing user context."}}, nil
	}

	post, err := s.repo.Find(ctx, data.Id)
	if err != nil {
		return nil, err
	}
	if post == nil {
		return &dto.PublishPostResponse{Errors: map[string]string{"post": "Post not found."}}, nil
	}
	if post.Status == posts.StatusPublished {
		return &dto.PublishPostResponse{Errors: map[string]string{"post": "Post is already published."}}, nil
	}

	isAuthor := post.AuthorId.String() == userId
	if !isAuthor {
		ok, err := s.isClubManager(ctx, userId, post.ClubId.String())
		if err != nil {
			return nil, err
		}
		if !ok {
			return &dto.PublishPostResponse{Errors: map[string]string{"auth": "Insufficient permissions."}}, nil
		}
	}

	post.Status = posts.StatusPublished
	updated, err := s.repo.Save(ctx, post)
	if err != nil {
		return nil, err
	}
	return &dto.PublishPostResponse{Post: updated, Errors: make(map[string]string)}, nil
}

func (s *postService) UnpublishPost(ctx context.Context, data *dto.PublishPostRequest) (*dto.PublishPostResponse, error) {
	userId, ok := s.callerUserId(ctx)
	if !ok {
		return &dto.PublishPostResponse{Errors: map[string]string{"auth": "Missing user context."}}, nil
	}

	post, err := s.repo.Find(ctx, data.Id)
	if err != nil {
		return nil, err
	}
	if post == nil {
		return &dto.PublishPostResponse{Errors: map[string]string{"post": "Post not found."}}, nil
	}
	if post.Status == posts.StatusDraft {
		return &dto.PublishPostResponse{Errors: map[string]string{"post": "Post is already a draft."}}, nil
	}

	isAuthor := post.AuthorId.String() == userId
	if !isAuthor {
		ok, err := s.isClubManager(ctx, userId, post.ClubId.String())
		if err != nil {
			return nil, err
		}
		if !ok {
			return &dto.PublishPostResponse{Errors: map[string]string{"auth": "Insufficient permissions."}}, nil
		}
	}

	post.Status = posts.StatusDraft
	updated, err := s.repo.Save(ctx, post)
	if err != nil {
		return nil, err
	}
	return &dto.PublishPostResponse{Post: updated, Errors: make(map[string]string)}, nil
}

func (s *postService) UpdatePost(ctx context.Context, data *dto.UpdatePostRequest) (*dto.UpdatePostResponse, error) {
	userId, ok := s.callerUserId(ctx)
	if !ok {
		return &dto.UpdatePostResponse{Errors: map[string]string{"auth": "Missing user context."}}, nil
	}

	post, err := s.repo.Find(ctx, data.Id)
	if err != nil {
		return nil, err
	}
	if post == nil {
		return &dto.UpdatePostResponse{Errors: map[string]string{"post": "Post not found."}}, nil
	}

	if post.Status != posts.StatusDraft {
		return &dto.UpdatePostResponse{Errors: map[string]string{"post": "Only draft posts can be edited."}}, nil
	}

	isAuthor := post.AuthorId.String() == userId
	if !isAuthor {
		ok, err := s.isClubManager(ctx, userId, post.ClubId.String())
		if err != nil {
			return nil, err
		}
		if !ok {
			return &dto.UpdatePostResponse{Errors: map[string]string{"auth": "Insufficient permissions."}}, nil
		}
	}

	updated, errs := post.Update(data.Map())
	if len(errs) > 0 {
		return &dto.UpdatePostResponse{Errors: errs}, nil
	}

	saved, err := s.repo.Save(ctx, updated)
	if err != nil {
		return nil, err
	}
	return &dto.UpdatePostResponse{Post: saved, Errors: make(map[string]string)}, nil
}

func (s *postService) DeletePost(ctx context.Context, id string) (bool, error) {
	userId, ok := s.callerUserId(ctx)
	if !ok {
		return false, fmt.Errorf("missing user context")
	}

	post, err := s.repo.Find(ctx, id)
	if err != nil {
		return false, err
	}
	if post == nil {
		return false, nil
	}

	if post.Status != posts.StatusDraft {
		return false, fmt.Errorf("only draft posts can be deleted")
	}

	isAuthor := post.AuthorId.String() == userId
	if !isAuthor {
		ok, err := s.isClubManager(ctx, userId, post.ClubId.String())
		if err != nil {
			return false, err
		}
		if !ok {
			return false, fmt.Errorf("insufficient permissions to delete this post")
		}
	}

	return s.repo.Delete(ctx, id)
}

func (s *postService) GetPost(ctx context.Context, id string) (*posts.Post, error) {
	return s.repo.Find(ctx, id)
}

func (s *postService) HasPostForEvent(ctx context.Context, eventId string) (bool, error) {
	p, err := s.repo.FindByEventId(ctx, eventId)
	if err != nil {
		return false, err
	}
	return p != nil, nil
}

func (s *postService) GetClubPosts(ctx context.Context, data *dto.GetClubPostsRequest) (*dto.GetClubPostsResponse, error) {
	userId, _ := s.callerUserId(ctx)

	// Determine what the caller is allowed to see.
	canSeeDrafts := false
	canSeeAdherent := false

	if userId != "" && s.roleChecker != nil {
		manager, err := s.isClubManager(ctx, userId, data.ClubId)
		if err != nil {
			return nil, err
		}
		if manager {
			canSeeDrafts = true
			canSeeAdherent = true
		}
	}

	if !canSeeAdherent && userId != "" && s.memberChecker != nil {
		member, err := s.memberChecker.HasMembership(ctx, userId, data.ClubId)
		if err != nil {
			return nil, err
		}
		if member {
			canSeeAdherent = true
		}
	}

	params := posts.PostSearchParams{
		Page:     data.Page,
		PageSize: data.PageSize,
	}

	// Non-managers always see only published posts.
	if !canSeeDrafts {
		params.Status = posts.StatusPublished
	}
	// Non-members see only public posts.
	if !canSeeAdherent {
		params.Visibility = posts.VisibilityPublic
	}

	list, total, err := s.repo.FindByClub(ctx, data.ClubId, params)
	if err != nil {
		return nil, err
	}

	return &dto.GetClubPostsResponse{
		Posts:    list,
		Total:    total,
		Page:     params.Page,
		PageSize: params.PageSize,
		Errors:   make(map[string]string),
	}, nil
}
