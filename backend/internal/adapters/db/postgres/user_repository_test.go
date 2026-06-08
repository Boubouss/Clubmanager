//go:build integration

package postgres

import (
	"context"
	"testing"

	"clubmanager/internal/domain"
	"clubmanager/internal/domain/users"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserRepository_Insert(t *testing.T) {
	truncateAll(t)
	repo := NewUserRepository(testPool)

	u := &users.User{
		Username:    "alice",
		Email:       "alice@example.com",
		Phonenumber: "+33612345678",
		Password:    "hashed_password",
	}
	saved, err := repo.Save(context.Background(), u)
	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, saved.Id)
	assert.Equal(t, "alice", saved.Username)
	assert.Equal(t, "alice@example.com", saved.Email)
}

func TestUserRepository_Find(t *testing.T) {
	truncateAll(t)
	repo := NewUserRepository(testPool)

	userId := seedUser(t, "bob", "bob@example.com")

	found, err := repo.Find(context.Background(), userId.String())
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, userId, found.Id)
	assert.Equal(t, "bob", found.Username)
}

func TestUserRepository_Find_NotFound(t *testing.T) {
	truncateAll(t)
	repo := NewUserRepository(testPool)

	found, err := repo.Find(context.Background(), uuid.New().String())
	require.NoError(t, err)
	assert.Nil(t, found)
}

func TestUserRepository_Update(t *testing.T) {
	truncateAll(t)
	repo := NewUserRepository(testPool)

	userId := seedUser(t, "charlie", "charlie@example.com")

	u := &users.User{
		Id:          userId,
		Username:    "charlie",
		Email:       "charlie-new@example.com",
		Phonenumber: "+33612345678",
		Password:    "new_hash",
	}
	updated, err := repo.Save(context.Background(), u)
	require.NoError(t, err)
	assert.Equal(t, "charlie-new@example.com", updated.Email)

	// Verify persisted
	found, err := repo.Find(context.Background(), userId.String())
	require.NoError(t, err)
	assert.Equal(t, "charlie-new@example.com", found.Email)
}

func TestUserRepository_Search_ByUsername(t *testing.T) {
	truncateAll(t)
	repo := NewUserRepository(testPool)

	seedUser(t, "dave", "dave@example.com")
	seedUser(t, "eve", "eve@example.com")

	list, err := repo.Search(context.Background(), &domain.SearchParams{
		Fields:    map[string]any{"username": "dave"},
		Keys:      map[string]bool{"username": true},
		Connector: "AND",
	})
	require.NoError(t, err)
	require.Len(t, list, 1)
	assert.Equal(t, "dave", list[0].Username)
}

func TestUserRepository_Delete(t *testing.T) {
	truncateAll(t)
	repo := NewUserRepository(testPool)

	userId := seedUser(t, "frank", "frank@example.com")

	ok, err := repo.Delete(context.Background(), userId.String())
	require.NoError(t, err)
	assert.True(t, ok)

	found, err := repo.Find(context.Background(), userId.String())
	require.NoError(t, err)
	assert.Nil(t, found)
}
