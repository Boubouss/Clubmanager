//go:build integration

package postgres

import (
	"clubmanager/internal/domain/posts"
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupPostFixtures(t *testing.T) (clubId, authorId uuid.UUID) {
	t.Helper()
	truncateAll(t)
	authorId = seedUser(t, "postauthor", "postauthor@test.com")
	clubId = seedClub(t, "123456789", "Judo Club Post")
	return clubId, authorId
}

func seedPost(t *testing.T, clubId, authorId uuid.UUID, title, status, visibility string) uuid.UUID {
	t.Helper()
	var id uuid.UUID
	err := testPool.QueryRow(context.Background(), `
		INSERT INTO posts (club_id, author_id, title, content, status, visibility)
		VALUES ($1, $2, $3, 'Contenu test', $4, $5)
		RETURNING id
	`, clubId, authorId, title, status, visibility).Scan(&id)
	if err != nil {
		t.Fatalf("seedPost %q: %v", title, err)
	}
	return id
}

func TestPostRepository_Insert(t *testing.T) {
	clubId, authorId := setupPostFixtures(t)
	repo := NewPostRepository(testPool)

	p := &posts.Post{
		ClubId:     clubId,
		AuthorId:   authorId,
		Title:      "Mon actu",
		Content:    "Corps de l'actu",
		Status:     posts.StatusDraft,
		Visibility: posts.VisibilityAdherent,
	}

	created, err := repo.Save(context.Background(), p)
	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, created.Id)
	assert.Equal(t, "Mon actu", created.Title)
	assert.Equal(t, posts.StatusDraft, created.Status)
	assert.Equal(t, posts.VisibilityAdherent, created.Visibility)
	assert.NotEmpty(t, created.CreatedAt)
	assert.NotEmpty(t, created.UpdatedAt)
}

func TestPostRepository_Find(t *testing.T) {
	clubId, authorId := setupPostFixtures(t)
	repo := NewPostRepository(testPool)

	postId := seedPost(t, clubId, authorId, "Titre findable", posts.StatusPublished, posts.VisibilityPublic)

	found, err := repo.Find(context.Background(), postId.String())
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, "Titre findable", found.Title)
	assert.Equal(t, posts.StatusPublished, found.Status)
	assert.Equal(t, posts.VisibilityPublic, found.Visibility)
}

func TestPostRepository_Find_NotFound(t *testing.T) {
	setupPostFixtures(t)
	repo := NewPostRepository(testPool)

	found, err := repo.Find(context.Background(), uuid.New().String())
	require.NoError(t, err)
	assert.Nil(t, found)
}

func TestPostRepository_Update_Publish(t *testing.T) {
	clubId, authorId := setupPostFixtures(t)
	repo := NewPostRepository(testPool)

	postId := seedPost(t, clubId, authorId, "Brouillon", posts.StatusDraft, posts.VisibilityAdherent)

	found, err := repo.Find(context.Background(), postId.String())
	require.NoError(t, err)
	require.NotNil(t, found)

	found.Status = posts.StatusPublished
	found.Visibility = posts.VisibilityPublic
	updated, err := repo.Save(context.Background(), found)
	require.NoError(t, err)
	assert.Equal(t, posts.StatusPublished, updated.Status)
	assert.Equal(t, posts.VisibilityPublic, updated.Visibility)
}

func TestPostRepository_FindByClub_Pagination(t *testing.T) {
	clubId, authorId := setupPostFixtures(t)
	repo := NewPostRepository(testPool)

	// Create 5 published posts and 2 drafts.
	for i := 0; i < 5; i++ {
		seedPost(t, clubId, authorId, "Published", posts.StatusPublished, posts.VisibilityPublic)
	}
	for i := 0; i < 2; i++ {
		seedPost(t, clubId, authorId, "Draft", posts.StatusDraft, posts.VisibilityAdherent)
	}

	// Page 1, size 3 — no filters → 7 total.
	page1, total, err := repo.FindByClub(context.Background(), clubId.String(), posts.PostSearchParams{
		Page: 1, PageSize: 3,
	})
	require.NoError(t, err)
	assert.Equal(t, 7, total)
	assert.Len(t, page1, 3)

	// Page 3, size 3 — should return the last 1.
	page3, total, err := repo.FindByClub(context.Background(), clubId.String(), posts.PostSearchParams{
		Page: 3, PageSize: 3,
	})
	require.NoError(t, err)
	assert.Equal(t, 7, total)
	assert.Len(t, page3, 1)
}

func TestPostRepository_FindByClub_StatusFilter(t *testing.T) {
	clubId, authorId := setupPostFixtures(t)
	repo := NewPostRepository(testPool)

	seedPost(t, clubId, authorId, "Published 1", posts.StatusPublished, posts.VisibilityPublic)
	seedPost(t, clubId, authorId, "Published 2", posts.StatusPublished, posts.VisibilityAdherent)
	seedPost(t, clubId, authorId, "Draft 1", posts.StatusDraft, posts.VisibilityAdherent)

	published, total, err := repo.FindByClub(context.Background(), clubId.String(), posts.PostSearchParams{
		Status: posts.StatusPublished, Page: 1, PageSize: 20,
	})
	require.NoError(t, err)
	assert.Equal(t, 2, total)
	assert.Len(t, published, 2)
}

func TestPostRepository_FindByClub_VisibilityFilter(t *testing.T) {
	clubId, authorId := setupPostFixtures(t)
	repo := NewPostRepository(testPool)

	seedPost(t, clubId, authorId, "Public 1", posts.StatusPublished, posts.VisibilityPublic)
	seedPost(t, clubId, authorId, "Public 2", posts.StatusPublished, posts.VisibilityPublic)
	seedPost(t, clubId, authorId, "Adherent 1", posts.StatusPublished, posts.VisibilityAdherent)

	public, total, err := repo.FindByClub(context.Background(), clubId.String(), posts.PostSearchParams{
		Status: posts.StatusPublished, Visibility: posts.VisibilityPublic, Page: 1, PageSize: 20,
	})
	require.NoError(t, err)
	assert.Equal(t, 2, total)
	assert.Len(t, public, 2)
}

func TestPostRepository_Delete(t *testing.T) {
	clubId, authorId := setupPostFixtures(t)
	repo := NewPostRepository(testPool)

	postId := seedPost(t, clubId, authorId, "A supprimer", posts.StatusDraft, posts.VisibilityAdherent)

	ok, err := repo.Delete(context.Background(), postId.String())
	require.NoError(t, err)
	assert.True(t, ok)

	found, err := repo.Find(context.Background(), postId.String())
	require.NoError(t, err)
	assert.Nil(t, found)
}

func TestPostRepository_HasMembership(t *testing.T) {
	clubId, authorId := setupPostFixtures(t)
	repo := NewPostRepository(testPool)

	memberId := seedMember(t, authorId, "Post", "Author")

	// No membership yet → false
	hasMembership, err := repo.HasMembership(context.Background(), authorId.String(), clubId.String())
	require.NoError(t, err)
	assert.False(t, hasMembership)

	// Add membership but not validated → false
	seedClubMembership(t, memberId, clubId)
	hasMembership, err = repo.HasMembership(context.Background(), authorId.String(), clubId.String())
	require.NoError(t, err)
	assert.False(t, hasMembership)

	// Validate membership → true
	_, err = testPool.Exec(context.Background(), `
		UPDATE club_memberships SET is_valid = true WHERE member_id = $1 AND club_id = $2
	`, memberId, clubId)
	require.NoError(t, err)
	hasMembership, err = repo.HasMembership(context.Background(), authorId.String(), clubId.String())
	require.NoError(t, err)
	assert.True(t, hasMembership)
}
