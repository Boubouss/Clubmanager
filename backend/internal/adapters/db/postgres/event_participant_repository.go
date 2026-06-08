package postgres

import (
	"clubmanager/internal/domain/events"
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type eventParticipantRepository struct {
	db *pgxpool.Pool
}

func NewEventParticipantRepository(db *pgxpool.Pool) *eventParticipantRepository {
	return &eventParticipantRepository{db: db}
}

// Register inserts atomically: blocks if member is already registered OR capacity is exceeded.
// Returns (nil, nil) when the insert was blocked (service must distinguish the cause).
func (r eventParticipantRepository) Register(ctx context.Context, p *events.EventParticipant, maxParticipants int) (*events.EventParticipant, error) {
	row := r.db.QueryRow(ctx, `
		INSERT INTO event_participants (event_id, member_id, role, event_category_id)
		SELECT $1, $2, $3, $4
		WHERE NOT EXISTS (
			SELECT 1 FROM event_participants WHERE event_id = $1 AND member_id = $2
		)
		AND ($5 = 0 OR (SELECT COUNT(*) FROM event_participants WHERE event_id = $1) < $5)
		RETURNING id, event_id, member_id, role, event_category_id, registered_at::text
	`, p.EventId, p.MemberId, p.Role, p.EventCategoryId, maxParticipants)

	result, err := scanParticipant(row.Scan, p)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil // blocked: duplicate or capacity exceeded
		}
		return nil, err
	}
	return result, nil
}

func (r eventParticipantRepository) Unregister(ctx context.Context, eventId, memberId string) (bool, error) {
	_, err := r.db.Exec(ctx, `
		DELETE FROM event_participants WHERE event_id = $1 AND member_id = $2
	`, eventId, memberId)
	return err == nil, err
}

func (r eventParticipantRepository) FindByEvent(ctx context.Context, eventId string) ([]*events.EventParticipant, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, event_id, member_id, role, event_category_id, registered_at::text
		FROM event_participants WHERE event_id = $1 ORDER BY registered_at
	`, eventId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*events.EventParticipant
	for rows.Next() {
		p := &events.EventParticipant{}
		if _, err := scanParticipant(rows.Scan, p); err != nil {
			return nil, err
		}
		list = append(list, p)
	}
	return list, nil
}

func (r eventParticipantRepository) FindByMember(ctx context.Context, eventId, memberId string) (*events.EventParticipant, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, event_id, member_id, role, event_category_id, registered_at::text
		FROM event_participants WHERE event_id = $1 AND member_id = $2
	`, eventId, memberId)
	p := &events.EventParticipant{}
	result, err := scanParticipant(row.Scan, p)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return result, nil
}

func (r eventParticipantRepository) CountByEvent(ctx context.Context, eventId string) (int, error) {
	var count int
	err := r.db.QueryRow(ctx, `
		SELECT COUNT(*) FROM event_participants WHERE event_id = $1
	`, eventId).Scan(&count)
	return count, err
}

func scanParticipant(scan func(...any) error, p *events.EventParticipant) (*events.EventParticipant, error) {
	var catId *uuid.UUID
	if err := scan(&p.Id, &p.EventId, &p.MemberId, &p.Role, &catId, &p.RegisteredAt); err != nil {
		return nil, err
	}
	p.EventCategoryId = catId
	return p, nil
}
