//go:build integration

package postgres

import (
	"context"
	"testing"

	"clubmanager/internal/domain/events"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupCarpoolFixtures creates the full prerequisite chain:
// users → clubs → members → event → registrations (so members are registered participants).
func setupCarpoolFixtures(t *testing.T) (eventId, member1Id, member2Id uuid.UUID) {
	t.Helper()
	userId := seedUser(t, "user_carp", "user_carp@example.com")
	clubId := seedClub(t, "444444444", "Carp Club")
	member1Id = seedMember(t, userId, clubId, "Carp1", "One")
	userId2 := seedUser(t, "user_carp2", "user_carp2@example.com")
	member2Id = seedMember(t, userId2, clubId, "Carp2", "Two")
	eventId = seedEvent(t, clubId, userId, "Carpool Event")

	partRepo := NewEventParticipantRepository(testPool)
	_, err := partRepo.Register(context.Background(), &events.EventParticipant{
		EventId: eventId, MemberId: member1Id, Role: events.RoleAccompanist,
	}, 0)
	require.NoError(t, err)
	_, err = partRepo.Register(context.Background(), &events.EventParticipant{
		EventId: eventId, MemberId: member2Id, Role: events.RoleAccompanist,
	}, 0)
	require.NoError(t, err)
	return
}

func TestCarpoolRepository_CreateOffer(t *testing.T) {
	truncateAll(t)
	eventId, member1Id, _ := setupCarpoolFixtures(t)
	repo := NewCarpoolRepository(testPool)

	offer := &events.CarpoolOffer{
		EventId:          eventId,
		MemberId:         member1Id,
		DepartureAddress: "10 rue de la Paix, Paris",
		AvailableSeats:   3,
		DepartureAt:      "2027-01-15T07:00:00Z",
	}
	created, err := repo.CreateOffer(context.Background(), offer)
	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, created.Id)
	assert.Equal(t, 3, created.AvailableSeats)
	assert.Equal(t, "10 rue de la Paix, Paris", created.DepartureAddress)
}

func TestCarpoolRepository_FindOfferById(t *testing.T) {
	truncateAll(t)
	eventId, member1Id, _ := setupCarpoolFixtures(t)
	repo := NewCarpoolRepository(testPool)

	offer := &events.CarpoolOffer{EventId: eventId, MemberId: member1Id, DepartureAddress: "X", AvailableSeats: 2, DepartureAt: "2027-01-15T07:00:00Z"}
	created, err := repo.CreateOffer(context.Background(), offer)
	require.NoError(t, err)

	found, err := repo.FindOfferById(context.Background(), created.Id.String())
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, created.Id, found.Id)
}

func TestCarpoolRepository_FindOfferById_NotFound(t *testing.T) {
	truncateAll(t)
	repo := NewCarpoolRepository(testPool)

	found, err := repo.FindOfferById(context.Background(), uuid.New().String())
	require.NoError(t, err)
	assert.Nil(t, found)
}

func TestCarpoolRepository_FindOfferByMember(t *testing.T) {
	truncateAll(t)
	eventId, member1Id, _ := setupCarpoolFixtures(t)
	repo := NewCarpoolRepository(testPool)

	offer := &events.CarpoolOffer{EventId: eventId, MemberId: member1Id, DepartureAddress: "X", AvailableSeats: 2, DepartureAt: "2027-01-15T07:00:00Z"}
	_, err := repo.CreateOffer(context.Background(), offer)
	require.NoError(t, err)

	found, err := repo.FindOfferByMember(context.Background(), eventId.String(), member1Id.String())
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, member1Id, found.MemberId)
}

func TestCarpoolRepository_FindOffersByEvent(t *testing.T) {
	truncateAll(t)
	eventId, member1Id, member2Id := setupCarpoolFixtures(t)
	repo := NewCarpoolRepository(testPool)

	_, err := repo.CreateOffer(context.Background(), &events.CarpoolOffer{EventId: eventId, MemberId: member1Id, DepartureAddress: "A", AvailableSeats: 2, DepartureAt: "2027-01-15T07:00:00Z"})
	require.NoError(t, err)
	_, err = repo.CreateOffer(context.Background(), &events.CarpoolOffer{EventId: eventId, MemberId: member2Id, DepartureAddress: "B", AvailableSeats: 1, DepartureAt: "2027-01-15T08:00:00Z"})
	require.NoError(t, err)

	list, err := repo.FindOffersByEvent(context.Background(), eventId.String())
	require.NoError(t, err)
	assert.Len(t, list, 2)
}

func TestCarpoolRepository_AddPassenger_Success(t *testing.T) {
	truncateAll(t)
	eventId, member1Id, member2Id := setupCarpoolFixtures(t)
	repo := NewCarpoolRepository(testPool)

	// member1 creates offer
	offer, err := repo.CreateOffer(context.Background(), &events.CarpoolOffer{
		EventId: eventId, MemberId: member1Id, DepartureAddress: "X", AvailableSeats: 2, DepartureAt: "2027-01-15T07:00:00Z",
	})
	require.NoError(t, err)

	// member2 joins
	added, err := repo.AddPassenger(context.Background(), offer.Id.String(), member2Id.String())
	require.NoError(t, err)
	assert.True(t, added)

	count, err := repo.CountPassengers(context.Background(), offer.Id.String())
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestCarpoolRepository_AddPassenger_DuplicateReturnsFalse(t *testing.T) {
	truncateAll(t)
	eventId, member1Id, member2Id := setupCarpoolFixtures(t)
	repo := NewCarpoolRepository(testPool)

	offer, err := repo.CreateOffer(context.Background(), &events.CarpoolOffer{
		EventId: eventId, MemberId: member1Id, DepartureAddress: "X", AvailableSeats: 3, DepartureAt: "2027-01-15T07:00:00Z",
	})
	require.NoError(t, err)

	first, err := repo.AddPassenger(context.Background(), offer.Id.String(), member2Id.String())
	require.NoError(t, err)
	assert.True(t, first)

	// Same passenger on same event → blocked
	second, err := repo.AddPassenger(context.Background(), offer.Id.String(), member2Id.String())
	require.NoError(t, err)
	assert.False(t, second, "duplicate passenger must return false")
}

func TestCarpoolRepository_AddPassenger_NoSeatsReturnsFalse(t *testing.T) {
	truncateAll(t)
	userId := seedUser(t, "user_seat", "user_seat@example.com")
	clubId := seedClub(t, "555555555", "Seat Club")
	member1Id := seedMember(t, userId, clubId, "Seat1", "One")
	userId2 := seedUser(t, "user_seat2", "user_seat2@example.com")
	member2Id := seedMember(t, userId2, clubId, "Seat2", "Two")
	userId3 := seedUser(t, "user_seat3", "user_seat3@example.com")
	member3Id := seedMember(t, userId3, clubId, "Seat3", "Three")
	eventId := seedEvent(t, clubId, userId, "Full Carpool Event")

	partRepo := NewEventParticipantRepository(testPool)
	for _, mId := range []uuid.UUID{member1Id, member2Id, member3Id} {
		_, err := partRepo.Register(context.Background(), &events.EventParticipant{EventId: eventId, MemberId: mId, Role: events.RoleAccompanist}, 0)
		require.NoError(t, err)
	}

	repo := NewCarpoolRepository(testPool)

	// Offer with only 1 seat
	offer, err := repo.CreateOffer(context.Background(), &events.CarpoolOffer{
		EventId: eventId, MemberId: member1Id, DepartureAddress: "X", AvailableSeats: 1, DepartureAt: "2027-01-15T07:00:00Z",
	})
	require.NoError(t, err)

	// member2 fills the seat
	added, err := repo.AddPassenger(context.Background(), offer.Id.String(), member2Id.String())
	require.NoError(t, err)
	assert.True(t, added)

	// member3 → no seats left
	blocked, err := repo.AddPassenger(context.Background(), offer.Id.String(), member3Id.String())
	require.NoError(t, err)
	assert.False(t, blocked, "no seats available must return false")
}

func TestCarpoolRepository_IsPassenger(t *testing.T) {
	truncateAll(t)
	eventId, member1Id, member2Id := setupCarpoolFixtures(t)
	repo := NewCarpoolRepository(testPool)

	offer, err := repo.CreateOffer(context.Background(), &events.CarpoolOffer{
		EventId: eventId, MemberId: member1Id, DepartureAddress: "X", AvailableSeats: 2, DepartureAt: "2027-01-15T07:00:00Z",
	})
	require.NoError(t, err)

	// Before joining
	is, err := repo.IsPassenger(context.Background(), eventId.String(), member2Id.String())
	require.NoError(t, err)
	assert.False(t, is)

	// After joining
	_, err = repo.AddPassenger(context.Background(), offer.Id.String(), member2Id.String())
	require.NoError(t, err)

	is, err = repo.IsPassenger(context.Background(), eventId.String(), member2Id.String())
	require.NoError(t, err)
	assert.True(t, is)
}

func TestCarpoolRepository_RemovePassenger(t *testing.T) {
	truncateAll(t)
	eventId, member1Id, member2Id := setupCarpoolFixtures(t)
	repo := NewCarpoolRepository(testPool)

	offer, err := repo.CreateOffer(context.Background(), &events.CarpoolOffer{
		EventId: eventId, MemberId: member1Id, DepartureAddress: "X", AvailableSeats: 2, DepartureAt: "2027-01-15T07:00:00Z",
	})
	require.NoError(t, err)

	_, err = repo.AddPassenger(context.Background(), offer.Id.String(), member2Id.String())
	require.NoError(t, err)

	ok, err := repo.RemovePassenger(context.Background(), offer.Id.String(), member2Id.String())
	require.NoError(t, err)
	assert.True(t, ok)

	count, err := repo.CountPassengers(context.Background(), offer.Id.String())
	require.NoError(t, err)
	assert.Equal(t, 0, count)
}

func TestCarpoolRepository_DeleteOffer(t *testing.T) {
	truncateAll(t)
	eventId, member1Id, _ := setupCarpoolFixtures(t)
	repo := NewCarpoolRepository(testPool)

	offer, err := repo.CreateOffer(context.Background(), &events.CarpoolOffer{
		EventId: eventId, MemberId: member1Id, DepartureAddress: "X", AvailableSeats: 2, DepartureAt: "2027-01-15T07:00:00Z",
	})
	require.NoError(t, err)

	ok, err := repo.DeleteOffer(context.Background(), offer.Id.String())
	require.NoError(t, err)
	assert.True(t, ok)

	found, err := repo.FindOfferById(context.Background(), offer.Id.String())
	require.NoError(t, err)
	assert.Nil(t, found)
}
