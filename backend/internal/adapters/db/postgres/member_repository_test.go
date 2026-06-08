//go:build integration

package postgres

import (
	"context"
	"testing"

	"clubmanager/internal/domain"
	"clubmanager/internal/domain/members"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupMemberFixtures(t *testing.T) (userId, clubId uuid.UUID) {
	t.Helper()
	userId = seedUser(t, "user_mb", "user_mb@example.com")
	clubId = seedClub(t, "123456789", "Judo Club Test")
	return
}

func TestMemberRepository_Insert(t *testing.T) {
	truncateAll(t)
	userId, clubId := setupMemberFixtures(t)
	repo := NewMemberRepository(testPool)

	m := &members.Member{
		UserId:    userId,
		ClubId:    clubId,
		Firstname: "Jean",
		Lastname:  "Dupont",
		Birthdate: "2000-06-15",
		Gender:    "man",
	}
	saved, err := repo.Save(context.Background(), m)
	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, saved.Id)
	assert.Equal(t, "Jean", saved.Firstname)
	assert.Equal(t, "Dupont", saved.Lastname)
	assert.False(t, saved.IsValid)
}

func TestMemberRepository_Find(t *testing.T) {
	truncateAll(t)
	userId, clubId := setupMemberFixtures(t)
	memberId := seedMember(t, userId, clubId, "Marie", "Martin")
	repo := NewMemberRepository(testPool)

	found, err := repo.Find(context.Background(), memberId.String())
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, memberId, found.Id)
	assert.Equal(t, "Marie", found.Firstname)
}

func TestMemberRepository_Find_NotFound(t *testing.T) {
	truncateAll(t)
	repo := NewMemberRepository(testPool)

	found, err := repo.Find(context.Background(), uuid.New().String())
	require.NoError(t, err)
	assert.Nil(t, found)
}

func TestMemberRepository_Update(t *testing.T) {
	truncateAll(t)
	userId, clubId := setupMemberFixtures(t)
	memberId := seedMember(t, userId, clubId, "Pierre", "Durand")
	repo := NewMemberRepository(testPool)

	m := &members.Member{
		Id:        memberId,
		UserId:    userId,
		ClubId:    clubId,
		Firstname: "Pierre-Updated",
		Lastname:  "Durand",
		Birthdate: "2000-06-15",
		Gender:    "man",
		IsValid:   true,
	}
	updated, err := repo.Save(context.Background(), m)
	require.NoError(t, err)
	assert.Equal(t, "Pierre-Updated", updated.Firstname)
	assert.True(t, updated.IsValid)

	// Verify persisted
	found, err := repo.Find(context.Background(), memberId.String())
	require.NoError(t, err)
	assert.Equal(t, "Pierre-Updated", found.Firstname)
	assert.True(t, found.IsValid)
}

func TestMemberRepository_Search_ByClub(t *testing.T) {
	truncateAll(t)
	userId, clubId := setupMemberFixtures(t)
	userId2 := seedUser(t, "user_mb2", "user_mb2@example.com")
	clubId2 := seedClub(t, "987654321", "Other Club")

	seedMember(t, userId, clubId, "Alice", "A")
	seedMember(t, userId2, clubId2, "Bob", "B")
	repo := NewMemberRepository(testPool)

	list, err := repo.Search(context.Background(), &domain.SearchParams{
		Fields:    map[string]any{"club_id": clubId.String()},
		Keys:      map[string]bool{"club_id": true},
		Connector: "AND",
	})
	require.NoError(t, err)
	require.Len(t, list, 1)
	assert.Equal(t, "Alice", list[0].Firstname)
}

func TestMemberRepository_Delete(t *testing.T) {
	truncateAll(t)
	userId, clubId := setupMemberFixtures(t)
	memberId := seedMember(t, userId, clubId, "Tom", "X")
	repo := NewMemberRepository(testPool)

	ok, err := repo.Delete(context.Background(), memberId.String())
	require.NoError(t, err)
	assert.True(t, ok)

	found, err := repo.Find(context.Background(), memberId.String())
	require.NoError(t, err)
	assert.Nil(t, found)
}
