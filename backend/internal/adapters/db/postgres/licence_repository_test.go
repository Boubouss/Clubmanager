//go:build integration

package postgres

import (
	"context"
	"testing"

	"clubmanager/internal/domain"
	"clubmanager/internal/domain/licences"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupLicenceFixtures(t *testing.T) (memberId uuid.UUID) {
	t.Helper()
	userId := seedUser(t, "user_lic", "user_lic@example.com")
	clubId := seedClub(t, "111111111", "Licence Club")
	memberId = seedMember(t, userId, clubId, "Lic", "User")
	return
}

func TestLicenceRepository_Insert(t *testing.T) {
	truncateAll(t)
	memberId := setupLicenceFixtures(t)
	repo := NewLicenceRepository(testPool)

	l := &licences.Licence{
		MemberId:      memberId,
		LicenceNumber: "LIC-2026-001",
		ValidFrom:     "2026-01-01",
		ValidUntil:    "2026-12-31",
		Status:        "pending",
	}
	saved, err := repo.Save(context.Background(), l)
	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, saved.Id)
	assert.Equal(t, "LIC-2026-001", saved.LicenceNumber)
	assert.Equal(t, "pending", saved.Status)
}

func TestLicenceRepository_Find(t *testing.T) {
	truncateAll(t)
	memberId := setupLicenceFixtures(t)
	repo := NewLicenceRepository(testPool)

	l := &licences.Licence{
		MemberId: memberId, LicenceNumber: "LIC-FIND-001",
		ValidFrom: "2026-01-01", ValidUntil: "2026-12-31", Status: "pending",
	}
	saved, err := repo.Save(context.Background(), l)
	require.NoError(t, err)

	found, err := repo.Find(context.Background(), saved.Id.String())
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, saved.Id, found.Id)
	assert.Equal(t, "LIC-FIND-001", found.LicenceNumber)
}

func TestLicenceRepository_Find_NotFound(t *testing.T) {
	truncateAll(t)
	repo := NewLicenceRepository(testPool)

	found, err := repo.Find(context.Background(), uuid.New().String())
	require.NoError(t, err)
	assert.Nil(t, found)
}

func TestLicenceRepository_UpdateStatus(t *testing.T) {
	truncateAll(t)
	memberId := setupLicenceFixtures(t)
	repo := NewLicenceRepository(testPool)

	l := &licences.Licence{
		MemberId: memberId, LicenceNumber: "LIC-UPD-001",
		ValidFrom: "2026-01-01", ValidUntil: "2026-12-31", Status: "pending",
	}
	saved, err := repo.Save(context.Background(), l)
	require.NoError(t, err)

	saved.Status = "active"
	updated, err := repo.Save(context.Background(), saved)
	require.NoError(t, err)
	assert.Equal(t, "active", updated.Status)

	found, err := repo.Find(context.Background(), saved.Id.String())
	require.NoError(t, err)
	assert.Equal(t, "active", found.Status)
}

func TestLicenceRepository_Search_ByMemberAndStatus(t *testing.T) {
	truncateAll(t)
	memberId := setupLicenceFixtures(t)
	repo := NewLicenceRepository(testPool)

	// Insert active and pending licences
	active := &licences.Licence{
		MemberId: memberId, LicenceNumber: "LIC-SRCH-ACT",
		ValidFrom: "2026-01-01", ValidUntil: "2026-12-31", Status: "pending",
	}
	saved, err := repo.Save(context.Background(), active)
	require.NoError(t, err)
	saved.Status = "active"
	_, err = repo.Save(context.Background(), saved)
	require.NoError(t, err)

	pending := &licences.Licence{
		MemberId: memberId, LicenceNumber: "LIC-SRCH-PND",
		ValidFrom: "2027-01-01", ValidUntil: "2027-12-31", Status: "pending",
	}
	_, err = repo.Save(context.Background(), pending)
	require.NoError(t, err)

	// Search active only
	list, err := repo.Search(context.Background(), &domain.SearchParams{
		Fields:    map[string]any{"member_id": memberId.String(), "status": "active"},
		Keys:      map[string]bool{"member_id": true, "status": true},
		Connector: "AND",
	})
	require.NoError(t, err)
	require.Len(t, list, 1)
	assert.Equal(t, "active", list[0].Status)
}

func TestLicenceRepository_Delete(t *testing.T) {
	truncateAll(t)
	memberId := setupLicenceFixtures(t)
	repo := NewLicenceRepository(testPool)

	l := &licences.Licence{
		MemberId: memberId, LicenceNumber: "LIC-DEL-001",
		ValidFrom: "2026-01-01", ValidUntil: "2026-12-31", Status: "pending",
	}
	saved, err := repo.Save(context.Background(), l)
	require.NoError(t, err)

	ok, err := repo.Delete(context.Background(), saved.Id.String())
	require.NoError(t, err)
	assert.True(t, ok)

	found, err := repo.Find(context.Background(), saved.Id.String())
	require.NoError(t, err)
	assert.Nil(t, found)
}
