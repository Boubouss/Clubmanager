package postgres

import (
	"clubmanager/internal/domain"
	"clubmanager/internal/domain/events"
	"context"

	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type eventRepository struct {
	db *pgxpool.Pool
}

func NewEventRepository(db *pgxpool.Pool) *eventRepository {
	return &eventRepository{db: db}
}

func (r eventRepository) Save(ctx context.Context, e *events.Event) (*events.Event, error) {
	if err := setMetadataLog(ctx, r.db, e.Id.String()); err != nil {
		return nil, err
	}
	if e.Id == uuid.Nil {
		return r.insert(ctx, e)
	}
	return r.update(ctx, e)
}

func (r eventRepository) insert(ctx context.Context, e *events.Event) (*events.Event, error) {
	row := r.db.QueryRow(ctx, `
		INSERT INTO events (club_id, created_by, title, description, type, status, location,
		                    registration_open_at, registration_close_at, date, max_participants)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id, club_id, created_by, title, description, type, status, location,
		          registration_open_at::text, registration_close_at::text, date::text,
		          COALESCE(max_participants, 0)
	`, e.ClubId, e.CreatedBy, e.Title, e.Description, e.Type, e.Status, e.Location,
		e.RegistrationOpenAt, e.RegistrationCloseAt, e.Date, nullableInt(e.MaxParticipants))

	return scanEvent(row.Scan, e)
}

func (r eventRepository) update(ctx context.Context, e *events.Event) (*events.Event, error) {
	_, err := r.db.Exec(ctx, `
		UPDATE events SET title=$1, description=$2, status=$3, location=$4,
		                  registration_open_at=$5, registration_close_at=$6,
		                  date=$7, max_participants=$8
		WHERE id=$9
	`, e.Title, e.Description, e.Status, e.Location,
		e.RegistrationOpenAt, e.RegistrationCloseAt, e.Date,
		nullableInt(e.MaxParticipants), e.Id)
	if err != nil {
		return nil, err
	}
	return e, nil
}

func (r eventRepository) Find(ctx context.Context, id string) (*events.Event, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, club_id, created_by, title, description, type, status, location,
		       registration_open_at::text, registration_close_at::text, date::text,
		       COALESCE(max_participants, 0)
		FROM events WHERE id = $1
	`, id)
	e := &events.Event{}
	result, err := scanEvent(row.Scan, e)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return result, nil
}

func (r eventRepository) Search(ctx context.Context, params *domain.SearchParams) ([]*events.Event, error) {
	where, args, err := params.GetWhereClauses()
	if err != nil {
		return nil, err
	}
	rows, err := r.db.Query(ctx, `
		SELECT id, club_id, created_by, title, description, type, status, location,
		       registration_open_at::text, registration_close_at::text, date::text,
		       COALESCE(max_participants, 0)
		FROM events WHERE `+where, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*events.Event
	for rows.Next() {
		e := &events.Event{}
		if _, err := scanEvent(rows.Scan, e); err != nil {
			return nil, err
		}
		list = append(list, e)
	}
	return list, nil
}

func (r eventRepository) Delete(ctx context.Context, id string) (bool, error) {
	if err := setMetadataLog(ctx, r.db, id); err != nil {
		return false, err
	}
	_, err := r.db.Exec(ctx, `DELETE FROM events WHERE id = $1`, id)
	return err == nil, err
}

func scanEvent(scan func(...any) error, e *events.Event) (*events.Event, error) {
	if err := scan(&e.Id, &e.ClubId, &e.CreatedBy, &e.Title, &e.Description,
		&e.Type, &e.Status, &e.Location,
		&e.RegistrationOpenAt, &e.RegistrationCloseAt, &e.Date, &e.MaxParticipants); err != nil {
		return nil, err
	}
	return e, nil
}

func nullableInt(n int) *int {
	if n <= 0 {
		return nil
	}
	return &n
}

// EventCategoryRepository

type eventCategoryRepository struct {
	db *pgxpool.Pool
}

func NewEventCategoryRepository(db *pgxpool.Pool) *eventCategoryRepository {
	return &eventCategoryRepository{db: db}
}

func (r eventCategoryRepository) Save(ctx context.Context, c *events.EventCategory) (*events.EventCategory, error) {
	row := r.db.QueryRow(ctx, `
		INSERT INTO event_categories (event_id, judo_category_id, weigh_in_at, starts_at, status)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, event_id, judo_category_id, weigh_in_at::text, starts_at::text, status
	`, c.EventId, c.JudoCategoryId, c.WeighInAt, c.StartsAt, c.Status)

	if err := row.Scan(&c.Id, &c.EventId, &c.JudoCategoryId, &c.WeighInAt, &c.StartsAt, &c.Status); err != nil {
		return nil, err
	}
	return c, nil
}

func (r eventCategoryRepository) FindByEvent(ctx context.Context, eventId string) ([]*events.EventCategory, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, event_id, judo_category_id, weigh_in_at::text, starts_at::text, status
		FROM event_categories WHERE event_id = $1 ORDER BY starts_at
	`, eventId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*events.EventCategory
	for rows.Next() {
		c := &events.EventCategory{}
		if err := rows.Scan(&c.Id, &c.EventId, &c.JudoCategoryId, &c.WeighInAt, &c.StartsAt, &c.Status); err != nil {
			return nil, err
		}
		list = append(list, c)
	}
	return list, nil
}

func (r eventCategoryRepository) FindById(ctx context.Context, id string) (*events.EventCategory, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, event_id, judo_category_id, weigh_in_at::text, starts_at::text, status
		FROM event_categories WHERE id = $1
	`, id)
	c := &events.EventCategory{}
	if err := row.Scan(&c.Id, &c.EventId, &c.JudoCategoryId, &c.WeighInAt, &c.StartsAt, &c.Status); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return c, nil
}

func (r eventCategoryRepository) UpdateStatus(ctx context.Context, id, status string) (*events.EventCategory, error) {
	row := r.db.QueryRow(ctx, `
		UPDATE event_categories SET status = $1 WHERE id = $2
		RETURNING id, event_id, judo_category_id, weigh_in_at::text, starts_at::text, status
	`, status, id)
	c := &events.EventCategory{}
	if err := row.Scan(&c.Id, &c.EventId, &c.JudoCategoryId, &c.WeighInAt, &c.StartsAt, &c.Status); err != nil {
		return nil, err
	}
	return c, nil
}

func (r eventCategoryRepository) DeleteByEvent(ctx context.Context, eventId string) error {
	_, err := r.db.Exec(ctx, `DELETE FROM event_categories WHERE event_id = $1`, eventId)
	return err
}
