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

func TestMemberRepository_Insert(t *testing.T) {
	truncateAll(t)
	userId := seedUser(t, "user_mb", "user_mb@example.com")
	repo := NewMemberRepository(testPool)

	m := &members.Member{
		UserId:    userId,
		Firstname: "Jean",
		Lastname:  "Dupont",
		Birthdate: "2000-06-15",
		Gender:    "man",
		IsPrimary: true,
	}
	saved, err := repo.Save(context.Background(), m)
	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, saved.Id)
	assert.Equal(t, "Jean", saved.Firstname)
	assert.Equal(t, "Dupont", saved.Lastname)
	assert.True(t, saved.IsPrimary)
}

func TestMemberRepository_Find(t *testing.T) {
	truncateAll(t)
	userId := seedUser(t, "user_mb2", "user_mb2@example.com")
	memberId := seedMember(t, userId, "Marie", "Martin")
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
	userId := seedUser(t, "user_mb3", "user_mb3@example.com")
	memberId := seedMember(t, userId, "Pierre", "Durand")
	repo := NewMemberRepository(testPool)

	m := &members.Member{
		Id:        memberId,
		UserId:    userId,
		Firstname: "Pierre-Updated",
		Lastname:  "Durand",
		Birthdate: "2000-06-15",
		Gender:    "man",
		IsPrimary: true,
	}
	updated, err := repo.Save(context.Background(), m)
	require.NoError(t, err)
	assert.Equal(t, "Pierre-Updated", updated.Firstname)
	assert.True(t, updated.IsPrimary)

	// Verify persisted
	found, err := repo.Find(context.Background(), memberId.String())
	require.NoError(t, err)
	assert.Equal(t, "Pierre-Updated", found.Firstname)
	assert.True(t, found.IsPrimary)
}

func TestMemberRepository_Search_ByUser(t *testing.T) {
	truncateAll(t)
	userId1 := seedUser(t, "user_mb4", "user_mb4@example.com")
	userId2 := seedUser(t, "user_mb5", "user_mb5@example.com")

	seedMember(t, userId1, "Alice", "A")
	seedMember(t, userId2, "Bob", "B")
	repo := NewMemberRepository(testPool)

	list, err := repo.Search(context.Background(), &domain.SearchParams{
		Fields:    map[string]any{"user_id": userId1.String()},
		Keys:      map[string]bool{"user_id": true},
		Connector: "AND",
	})
	require.NoError(t, err)
	require.Len(t, list, 1)
	assert.Equal(t, "Alice", list[0].Firstname)
}

func TestMemberRepository_Delete(t *testing.T) {
	truncateAll(t)
	userId := seedUser(t, "user_mb6", "user_mb6@example.com")
	memberId := seedMember(t, userId, "Tom", "X")
	repo := NewMemberRepository(testPool)

	ok, err := repo.Delete(context.Background(), memberId.String())
	require.NoError(t, err)
	assert.True(t, ok)

	found, err := repo.Find(context.Background(), memberId.String())
	require.NoError(t, err)
	assert.Nil(t, found)
}

// ── ClubMembership repository tests ──────────────────────────────────────────

func TestClubMembershipRepository_Insert(t *testing.T) {
	truncateAll(t)
	userId := seedUser(t, "user_cms1", "user_cms1@example.com")
	clubId := seedClub(t, "111111111", "Club CMS 1")
	memberId := seedMember(t, userId, "Jean", "Test")
	repo := NewClubMembershipRepository(testPool)

	cm := &members.ClubMembership{
		MemberId: memberId,
		ClubId:   clubId,
	}
	saved, err := repo.Save(context.Background(), cm)
	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, saved.Id)
	assert.Equal(t, clubId, saved.ClubId)
	assert.False(t, saved.IsValid)
}

func TestClubMembershipRepository_Find(t *testing.T) {
	truncateAll(t)
	userId := seedUser(t, "user_cms2", "user_cms2@example.com")
	clubId := seedClub(t, "222222222", "Club CMS 2")
	memberId := seedMember(t, userId, "Marie", "Test")
	msId := seedClubMembership(t, memberId, clubId)
	repo := NewClubMembershipRepository(testPool)

	found, err := repo.Find(context.Background(), msId.String())
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, msId, found.Id)
	assert.Equal(t, memberId, found.MemberId)
	assert.Equal(t, clubId, found.ClubId)
	assert.False(t, found.IsValid)
}

func TestClubMembershipRepository_Find_NotFound(t *testing.T) {
	truncateAll(t)
	repo := NewClubMembershipRepository(testPool)

	found, err := repo.Find(context.Background(), uuid.New().String())
	require.NoError(t, err)
	assert.Nil(t, found)
}

func TestClubMembershipRepository_Update_Validate(t *testing.T) {
	truncateAll(t)
	userId := seedUser(t, "user_cms3", "user_cms3@example.com")
	clubId := seedClub(t, "333333333", "Club CMS 3")
	memberId := seedMember(t, userId, "Paul", "Test")
	msId := seedClubMembership(t, memberId, clubId)
	repo := NewClubMembershipRepository(testPool)

	cm := &members.ClubMembership{
		Id:       msId,
		MemberId: memberId,
		ClubId:   clubId,
		IsValid:  true,
	}
	updated, err := repo.Save(context.Background(), cm)
	require.NoError(t, err)
	assert.True(t, updated.IsValid)

	// Verify persisted
	found, err := repo.Find(context.Background(), msId.String())
	require.NoError(t, err)
	assert.True(t, found.IsValid)
}

func TestClubMembershipRepository_FindByMember(t *testing.T) {
	truncateAll(t)
	userId := seedUser(t, "user_cms4", "user_cms4@example.com")
	clubId1 := seedClub(t, "444444441", "Club A")
	clubId2 := seedClub(t, "444444442", "Club B")
	memberId := seedMember(t, userId, "Alice", "Test")
	seedClubMembership(t, memberId, clubId1)
	seedClubMembership(t, memberId, clubId2)
	repo := NewClubMembershipRepository(testPool)

	list, err := repo.FindByMember(context.Background(), memberId.String())
	require.NoError(t, err)
	assert.Len(t, list, 2)
}

func TestClubMembershipRepository_FindByClub(t *testing.T) {
	truncateAll(t)
	userId1 := seedUser(t, "user_cms5a", "user_cms5a@example.com")
	userId2 := seedUser(t, "user_cms5b", "user_cms5b@example.com")
	clubId := seedClub(t, "555555555", "Club CMS 5")
	m1 := seedMember(t, userId1, "Bob", "Test")
	m2 := seedMember(t, userId2, "Eve", "Test")
	seedClubMembership(t, m1, clubId)
	seedClubMembership(t, m2, clubId)
	repo := NewClubMembershipRepository(testPool)

	list, err := repo.FindByClub(context.Background(), clubId.String())
	require.NoError(t, err)
	assert.Len(t, list, 2)
}

func TestClubMembershipRepository_Delete(t *testing.T) {
	truncateAll(t)
	userId := seedUser(t, "user_cms6", "user_cms6@example.com")
	clubId := seedClub(t, "666666666", "Club CMS 6")
	memberId := seedMember(t, userId, "Frank", "Test")
	msId := seedClubMembership(t, memberId, clubId)
	repo := NewClubMembershipRepository(testPool)

	ok, err := repo.Delete(context.Background(), msId.String())
	require.NoError(t, err)
	assert.True(t, ok)

	found, err := repo.Find(context.Background(), msId.String())
	require.NoError(t, err)
	assert.Nil(t, found)
}
