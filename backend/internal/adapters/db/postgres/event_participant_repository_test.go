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

// setupParticipantFixtures creates a user, club, event, and two members for participant tests.
func setupParticipantFixtures(t *testing.T) (eventId, member1Id, member2Id uuid.UUID) {
	t.Helper()
	userId := seedUser(t, "user_part", "user_part@example.com")
	clubId := seedClub(t, "222222222", "Part Club")
	member1Id = seedMember(t, userId, "Part1", "One")
	userId2 := seedUser(t, "user_part2", "user_part2@example.com")
	member2Id = seedMember(t, userId2, "Part2", "Two")
	eventId = seedEvent(t, clubId, userId, "Test Competition")
	return
}

func TestEventParticipantRepository_Register_Success(t *testing.T) {
	truncateAll(t)
	eventId, member1Id, _ := setupParticipantFixtures(t)
	repo := NewEventParticipantRepository(testPool)

	p := &events.EventParticipant{
		EventId:  eventId,
		MemberId: member1Id,
		Role:     events.RoleAccompanist,
	}
	registered, err := repo.Register(context.Background(), p, 0)
	require.NoError(t, err)
	require.NotNil(t, registered, "expected registration to succeed")
	assert.NotEqual(t, uuid.Nil, registered.Id)
	assert.Equal(t, events.RoleAccompanist, registered.Role)
}

func TestEventParticipantRepository_Register_DuplicateReturnsNil(t *testing.T) {
	truncateAll(t)
	eventId, member1Id, _ := setupParticipantFixtures(t)
	repo := NewEventParticipantRepository(testPool)

	p := &events.EventParticipant{EventId: eventId, MemberId: member1Id, Role: events.RoleAccompanist}

	first, err := repo.Register(context.Background(), p, 0)
	require.NoError(t, err)
	require.NotNil(t, first)

	// Second registration for same member on same event → blocked
	second, err := repo.Register(context.Background(), p, 0)
	require.NoError(t, err)
	assert.Nil(t, second, "duplicate registration must return nil")
}

func TestEventParticipantRepository_Register_CapacityReturnsNil(t *testing.T) {
	truncateAll(t)
	userId := seedUser(t, "user_cap", "user_cap@example.com")
	clubId := seedClub(t, "333333333", "Cap Club")
	member1Id := seedMember(t, userId, "Cap1", "One")
	userId2 := seedUser(t, "user_cap2", "user_cap2@example.com")
	member2Id := seedMember(t, userId2, "Cap2", "Two")
	// Event with max 1 participant
	eventId := seedEventMaxParticipants(t, clubId, userId, "Capped Event", 1)
	repo := NewEventParticipantRepository(testPool)

	p1 := &events.EventParticipant{EventId: eventId, MemberId: member1Id, Role: events.RoleAccompanist}
	first, err := repo.Register(context.Background(), p1, 1)
	require.NoError(t, err)
	require.NotNil(t, first, "first participant should be registered")

	// Second participant → capacity exceeded
	p2 := &events.EventParticipant{EventId: eventId, MemberId: member2Id, Role: events.RoleAccompanist}
	second, err := repo.Register(context.Background(), p2, 1)
	require.NoError(t, err)
	assert.Nil(t, second, "registration beyond capacity must return nil")
}

func TestEventParticipantRepository_FindByMember(t *testing.T) {
	truncateAll(t)
	eventId, member1Id, _ := setupParticipantFixtures(t)
	repo := NewEventParticipantRepository(testPool)

	p := &events.EventParticipant{EventId: eventId, MemberId: member1Id, Role: events.RoleCoach}
	_, err := repo.Register(context.Background(), p, 0)
	require.NoError(t, err)

	found, err := repo.FindByMember(context.Background(), eventId.String(), member1Id.String())
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, member1Id, found.MemberId)
	assert.Equal(t, events.RoleCoach, found.Role)
}

func TestEventParticipantRepository_FindByMember_NotFound(t *testing.T) {
	truncateAll(t)
	eventId, _, _ := setupParticipantFixtures(t)
	repo := NewEventParticipantRepository(testPool)

	found, err := repo.FindByMember(context.Background(), eventId.String(), uuid.New().String())
	require.NoError(t, err)
	assert.Nil(t, found)
}

func TestEventParticipantRepository_FindByEvent(t *testing.T) {
	truncateAll(t)
	eventId, member1Id, member2Id := setupParticipantFixtures(t)
	repo := NewEventParticipantRepository(testPool)

	_, err := repo.Register(context.Background(), &events.EventParticipant{EventId: eventId, MemberId: member1Id, Role: events.RoleAccompanist}, 0)
	require.NoError(t, err)
	_, err = repo.Register(context.Background(), &events.EventParticipant{EventId: eventId, MemberId: member2Id, Role: events.RoleAccompanist}, 0)
	require.NoError(t, err)

	list, err := repo.FindByEvent(context.Background(), eventId.String())
	require.NoError(t, err)
	assert.Len(t, list, 2)
}

func TestEventParticipantRepository_CountByEvent(t *testing.T) {
	truncateAll(t)
	eventId, member1Id, member2Id := setupParticipantFixtures(t)
	repo := NewEventParticipantRepository(testPool)

	_, err := repo.Register(context.Background(), &events.EventParticipant{EventId: eventId, MemberId: member1Id, Role: events.RoleAccompanist}, 0)
	require.NoError(t, err)
	_, err = repo.Register(context.Background(), &events.EventParticipant{EventId: eventId, MemberId: member2Id, Role: events.RoleAccompanist}, 0)
	require.NoError(t, err)

	count, err := repo.CountByEvent(context.Background(), eventId.String())
	require.NoError(t, err)
	assert.Equal(t, 2, count)
}

func TestEventParticipantRepository_Unregister(t *testing.T) {
	truncateAll(t)
	eventId, member1Id, _ := setupParticipantFixtures(t)
	repo := NewEventParticipantRepository(testPool)

	_, err := repo.Register(context.Background(), &events.EventParticipant{EventId: eventId, MemberId: member1Id, Role: events.RoleAccompanist}, 0)
	require.NoError(t, err)

	ok, err := repo.Unregister(context.Background(), eventId.String(), member1Id.String())
	require.NoError(t, err)
	assert.True(t, ok)

	found, err := repo.FindByMember(context.Background(), eventId.String(), member1Id.String())
	require.NoError(t, err)
	assert.Nil(t, found)
}
