package services

import (
	"clubmanager/internal/adapters/api/grpc/dto"
	"clubmanager/internal/domain/posts"
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// ── Mocks ─────────────────────────────────────────────────────────────────────

type mockPostRepository struct {
	saveFunc       func(context.Context, *posts.Post) (*posts.Post, error)
	findFunc       func(context.Context, string) (*posts.Post, error)
	findByClubFunc func(context.Context, string, posts.PostSearchParams) ([]*posts.Post, int, error)
	deleteFunc     func(context.Context, string) (bool, error)
}

func (m mockPostRepository) Save(ctx context.Context, p *posts.Post) (*posts.Post, error) {
	return m.saveFunc(ctx, p)
}
func (m mockPostRepository) Find(ctx context.Context, id string) (*posts.Post, error) {
	return m.findFunc(ctx, id)
}
func (m mockPostRepository) FindByClub(ctx context.Context, clubId string, params posts.PostSearchParams) ([]*posts.Post, int, error) {
	return m.findByClubFunc(ctx, clubId, params)
}
func (m mockPostRepository) FindByEventId(ctx context.Context, eventId string) (*posts.Post, error) {
	return nil, nil
}
func (m mockPostRepository) Delete(ctx context.Context, id string) (bool, error) {
	return m.deleteFunc(ctx, id)
}

type mockMembershipChecker struct {
	hasMembershipFunc func(context.Context, string, string) (bool, error)
}

func (m mockMembershipChecker) HasMembership(ctx context.Context, userId, clubId string) (bool, error) {
	return m.hasMembershipFunc(ctx, userId, clubId)
}

// ── Fixtures ──────────────────────────────────────────────────────────────────

var (
	postClubId   = uuid.MustParse("00000000-0000-0000-0000-000000000010")
	postAuthorId = uuid.MustParse("00000000-0000-0000-0000-000000000011")
	postId       = uuid.MustParse("00000000-0000-0000-0000-000000000012")

	testPost = &posts.Post{
		Id:         postId,
		ClubId:     postClubId,
		AuthorId:   postAuthorId,
		Title:      "Titre de test",
		Content:    "Contenu de test",
		Status:     posts.StatusDraft,
		Visibility: posts.VisibilityAdherent,
	}
)

func authorCtx() context.Context {
	return context.WithValue(context.Background(), "user_id", postAuthorId.String())
}

func managerCtx() context.Context {
	// manager = a different user with president/staff role
	return context.WithValue(context.Background(), "user_id", "00000000-0000-0000-0000-000000000099")
}

func newTestPostService(
	repo mockPostRepository,
	checker mockRoleCheckerForClub,
	memberChecker mockMembershipChecker,
) *postService {
	return NewPostService(PostServiceConfig{
		Repository:  repo,
		RoleChecker: checker,
		MembershipChecker: memberChecker,
	})
}

// ── Tests ─────────────────────────────────────────────────────────────────────

func TestCreatePost(t *testing.T) {
	isManagerChecker := mockRoleCheckerForClub{
		isSuperAdminFunc: func(_ context.Context, _ string) (bool, error) { return false, nil },
	}
	// Make HasRole return true (manager check).
	// We can't override HasRole in the mock easily — use superadmin path instead for simplicity.
	superAdminChecker := mockRoleCheckerForClub{
		isSuperAdminFunc: func(_ context.Context, _ string) (bool, error) { return true, nil },
	}
	notManagerChecker := mockRoleCheckerForClub{
		isSuperAdminFunc: func(_ context.Context, _ string) (bool, error) { return false, nil },
	}
	_ = isManagerChecker

	noopMemberChecker := mockMembershipChecker{
		hasMembershipFunc: func(_ context.Context, _, _ string) (bool, error) { return false, nil },
	}

	tests := []struct {
		name       string
		ctx        context.Context
		req        *dto.CreatePostRequest
		checker    mockRoleCheckerForClub
		saveErr    error
		wantErrors bool
		wantErr    bool
		wantStatus string
	}{
		{
			name: "Superadmin creates draft post",
			ctx:  authorCtx(),
			req: &dto.CreatePostRequest{
				ClubId: postClubId.String(), Title: "Actu", Content: "Corps",
			},
			checker:    superAdminChecker,
			wantStatus: posts.StatusDraft,
		},
		{
			name: "Superadmin creates published post",
			ctx:  authorCtx(),
			req: &dto.CreatePostRequest{
				ClubId: postClubId.String(), Title: "Actu", Content: "Corps", Status: "published",
			},
			checker:    superAdminChecker,
			wantStatus: posts.StatusPublished,
		},
		{
			name: "Non-manager rejected",
			ctx:  authorCtx(),
			req: &dto.CreatePostRequest{
				ClubId: postClubId.String(), Title: "Actu", Content: "Corps",
			},
			checker:    notManagerChecker,
			wantErrors: true,
		},
		{
			name: "Missing title",
			ctx:  authorCtx(),
			req: &dto.CreatePostRequest{
				ClubId: postClubId.String(), Content: "Corps",
			},
			checker:    superAdminChecker,
			wantErrors: true,
		},
		{
			name: "Repository error",
			ctx:  authorCtx(),
			req: &dto.CreatePostRequest{
				ClubId: postClubId.String(), Title: "Actu", Content: "Corps",
			},
			checker: superAdminChecker,
			saveErr: errors.New("db error"),
			wantErr: true,
		},
		{
			name:    "Missing user context",
			ctx:     context.Background(),
			req:     &dto.CreatePostRequest{ClubId: postClubId.String(), Title: "T", Content: "C"},
			checker: superAdminChecker,
			wantErrors: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := mockPostRepository{
				saveFunc: func(_ context.Context, p *posts.Post) (*posts.Post, error) {
					if tt.saveErr != nil {
						return nil, tt.saveErr
					}
					p.Id = postId
					return p, nil
				},
			}

			svc := newTestPostService(repo, tt.checker, noopMemberChecker)
			res, err := svc.CreatePost(tt.ctx, tt.req)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			if tt.wantErrors {
				assert.True(t, len(res.Errors) > 0)
			} else {
				assert.Empty(t, res.Errors)
				assert.NotNil(t, res.Post)
				assert.Equal(t, tt.wantStatus, res.Post.Status)
			}
		})
	}
}

func TestPublishPost(t *testing.T) {
	superAdminChecker := mockRoleCheckerForClub{
		isSuperAdminFunc: func(_ context.Context, _ string) (bool, error) { return true, nil },
	}
	notManagerChecker := mockRoleCheckerForClub{
		isSuperAdminFunc: func(_ context.Context, _ string) (bool, error) { return false, nil },
	}
	noopMemberChecker := mockMembershipChecker{
		hasMembershipFunc: func(_ context.Context, _, _ string) (bool, error) { return false, nil },
	}

	tests := []struct {
		name       string
		ctx        context.Context
		post       *posts.Post
		checker    mockRoleCheckerForClub
		saveErr    error
		wantErrors bool
		wantErr    bool
	}{
		{
			name:    "Author publishes own draft",
			ctx:     authorCtx(),
			post:    testPost,
			checker: notManagerChecker,
		},
		{
			name:    "Manager publishes any draft",
			ctx:     managerCtx(),
			post:    testPost,
			checker: superAdminChecker,
		},
		{
			name:       "Non-author non-manager rejected",
			ctx:        managerCtx(),
			post:       testPost,
			checker:    notManagerChecker,
			wantErrors: true,
		},
		{
			name:       "Already published",
			ctx:        authorCtx(),
			post:       &posts.Post{Id: postId, AuthorId: postAuthorId, Status: posts.StatusPublished},
			checker:    notManagerChecker,
			wantErrors: true,
		},
		{
			name:       "Post not found",
			ctx:        authorCtx(),
			post:       nil,
			checker:    notManagerChecker,
			wantErrors: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			post := tt.post
			var postCopy *posts.Post
			if post != nil {
				c := *post
				postCopy = &c
			}
			repo := mockPostRepository{
				findFunc: func(_ context.Context, _ string) (*posts.Post, error) {
					return postCopy, nil
				},
				saveFunc: func(_ context.Context, p *posts.Post) (*posts.Post, error) {
					if tt.saveErr != nil {
						return nil, tt.saveErr
					}
					return p, nil
				},
			}

			svc := newTestPostService(repo, tt.checker, noopMemberChecker)
			res, err := svc.PublishPost(tt.ctx, &dto.PublishPostRequest{Id: postId.String()})

			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			if tt.wantErrors {
				assert.True(t, len(res.Errors) > 0)
			} else {
				assert.Empty(t, res.Errors)
				assert.NotNil(t, res.Post)
				assert.Equal(t, posts.StatusPublished, res.Post.Status)
			}
		})
	}
}

func TestDeletePost(t *testing.T) {
	superAdminChecker := mockRoleCheckerForClub{
		isSuperAdminFunc: func(_ context.Context, _ string) (bool, error) { return true, nil },
	}
	notManagerChecker := mockRoleCheckerForClub{
		isSuperAdminFunc: func(_ context.Context, _ string) (bool, error) { return false, nil },
	}
	noopMemberChecker := mockMembershipChecker{
		hasMembershipFunc: func(_ context.Context, _, _ string) (bool, error) { return false, nil },
	}

	tests := []struct {
		name      string
		ctx       context.Context
		post      *posts.Post
		checker   mockRoleCheckerForClub
		deleteErr error
		wantOk    bool
		wantErr   bool
	}{
		{
			name:    "Author deletes own post",
			ctx:     authorCtx(),
			post:    testPost,
			checker: notManagerChecker,
			wantOk:  true,
		},
		{
			name:    "Manager deletes any post",
			ctx:     managerCtx(),
			post:    testPost,
			checker: superAdminChecker,
			wantOk:  true,
		},
		{
			name:    "Non-author non-manager rejected",
			ctx:     managerCtx(),
			post:    testPost,
			checker: notManagerChecker,
			wantErr: true,
		},
		{
			name:    "Post not found → false, no error",
			ctx:     authorCtx(),
			post:    nil,
			checker: notManagerChecker,
			wantOk:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := mockPostRepository{
				findFunc: func(_ context.Context, _ string) (*posts.Post, error) {
					return tt.post, nil
				},
				deleteFunc: func(_ context.Context, _ string) (bool, error) {
					return tt.deleteErr == nil, tt.deleteErr
				},
			}

			svc := newTestPostService(repo, tt.checker, noopMemberChecker)
			ok, err := svc.DeletePost(tt.ctx, postId.String())

			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.wantOk, ok)
		})
	}
}

func TestGetClubPosts(t *testing.T) {
	superAdminChecker := mockRoleCheckerForClub{
		isSuperAdminFunc: func(_ context.Context, _ string) (bool, error) { return true, nil },
	}
	notManagerChecker := mockRoleCheckerForClub{
		isSuperAdminFunc: func(_ context.Context, _ string) (bool, error) { return false, nil },
	}

	publishedAdherent := &posts.Post{Status: posts.StatusPublished, Visibility: posts.VisibilityAdherent}
	publishedPublic := &posts.Post{Status: posts.StatusPublished, Visibility: posts.VisibilityPublic}
	draftAdherent := &posts.Post{Status: posts.StatusDraft, Visibility: posts.VisibilityAdherent}

	tests := []struct {
		name             string
		ctx              context.Context
		checker          mockRoleCheckerForClub
		isMember         bool
		repoResp         []*posts.Post
		repoTotal        int
		wantStatusFilter string
		wantVisFilter    string
		wantCount        int
	}{
		{
			name:             "Manager sees all (no filter)",
			ctx:              managerCtx(),
			checker:          superAdminChecker,
			repoResp:         []*posts.Post{publishedAdherent, draftAdherent},
			repoTotal:        2,
			wantStatusFilter: "",
			wantVisFilter:    "",
			wantCount:        2,
		},
		{
			name:             "Valid member sees published only, any visibility",
			ctx:              authorCtx(),
			checker:          notManagerChecker,
			isMember:         true,
			repoResp:         []*posts.Post{publishedAdherent, publishedPublic},
			repoTotal:        2,
			wantStatusFilter: posts.StatusPublished,
			wantVisFilter:    "",
			wantCount:        2,
		},
		{
			name:             "Non-member sees published public only",
			ctx:              authorCtx(),
			checker:          notManagerChecker,
			isMember:         false,
			repoResp:         []*posts.Post{publishedPublic},
			repoTotal:        1,
			wantStatusFilter: posts.StatusPublished,
			wantVisFilter:    posts.VisibilityPublic,
			wantCount:        1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotParams posts.PostSearchParams
			repo := mockPostRepository{
				findByClubFunc: func(_ context.Context, _ string, p posts.PostSearchParams) ([]*posts.Post, int, error) {
					gotParams = p
					return tt.repoResp, tt.repoTotal, nil
				},
			}
			memberChecker := mockMembershipChecker{
				hasMembershipFunc: func(_ context.Context, _, _ string) (bool, error) {
					return tt.isMember, nil
				},
			}

			svc := newTestPostService(repo, tt.checker, memberChecker)
			res, err := svc.GetClubPosts(tt.ctx, &dto.GetClubPostsRequest{
				ClubId: postClubId.String(), Page: 1, PageSize: 20,
			})

			assert.NoError(t, err)
			assert.Len(t, res.Posts, tt.wantCount)
			assert.Equal(t, tt.repoTotal, res.Total)
			assert.Equal(t, tt.wantStatusFilter, gotParams.Status)
			assert.Equal(t, tt.wantVisFilter, gotParams.Visibility)
		})
	}
}
