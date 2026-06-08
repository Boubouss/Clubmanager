package postgres

import (
	"clubmanager/internal/domain/events"
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type carpoolRepository struct {
	db *pgxpool.Pool
}

func NewCarpoolRepository(db *pgxpool.Pool) *carpoolRepository {
	return &carpoolRepository{db: db}
}

func (r carpoolRepository) CreateOffer(ctx context.Context, o *events.CarpoolOffer) (*events.CarpoolOffer, error) {
	row := r.db.QueryRow(ctx, `
		INSERT INTO carpool_offers (event_id, member_id, departure_address, available_seats, departure_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, event_id, member_id, departure_address, available_seats, departure_at::text
	`, o.EventId, o.MemberId, o.DepartureAddress, o.AvailableSeats, o.DepartureAt)

	if err := row.Scan(&o.Id, &o.EventId, &o.MemberId, &o.DepartureAddress, &o.AvailableSeats, &o.DepartureAt); err != nil {
		return nil, err
	}
	return o, nil
}

func (r carpoolRepository) FindOfferById(ctx context.Context, offerId string) (*events.CarpoolOffer, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, event_id, member_id, departure_address, available_seats, departure_at::text
		FROM carpool_offers WHERE id = $1
	`, offerId)
	o := &events.CarpoolOffer{}
	if err := row.Scan(&o.Id, &o.EventId, &o.MemberId, &o.DepartureAddress, &o.AvailableSeats, &o.DepartureAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return o, nil
}

func (r carpoolRepository) FindOffersByEvent(ctx context.Context, eventId string) ([]*events.CarpoolOffer, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, event_id, member_id, departure_address, available_seats, departure_at::text
		FROM carpool_offers WHERE event_id = $1 ORDER BY departure_at
	`, eventId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*events.CarpoolOffer
	for rows.Next() {
		o := &events.CarpoolOffer{}
		if err := rows.Scan(&o.Id, &o.EventId, &o.MemberId, &o.DepartureAddress, &o.AvailableSeats, &o.DepartureAt); err != nil {
			return nil, err
		}
		list = append(list, o)
	}
	return list, nil
}

func (r carpoolRepository) FindOfferByMember(ctx context.Context, eventId, memberId string) (*events.CarpoolOffer, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, event_id, member_id, departure_address, available_seats, departure_at::text
		FROM carpool_offers WHERE event_id = $1 AND member_id = $2
	`, eventId, memberId)
	o := &events.CarpoolOffer{}
	if err := row.Scan(&o.Id, &o.EventId, &o.MemberId, &o.DepartureAddress, &o.AvailableSeats, &o.DepartureAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return o, nil
}

func (r carpoolRepository) DeleteOffer(ctx context.Context, offerId string) (bool, error) {
	_, err := r.db.Exec(ctx, `DELETE FROM carpool_offers WHERE id = $1`, offerId)
	return err == nil, err
}

// AddPassenger inserts atomically only if the member is not already a passenger on the event
// and there is at least one seat available. Returns (false, nil) if blocked.
func (r carpoolRepository) AddPassenger(ctx context.Context, offerId, memberId string) (bool, error) {
	tag, err := r.db.Exec(ctx, `
		INSERT INTO carpool_passengers (offer_id, member_id)
		SELECT $1, $2
		WHERE NOT EXISTS (
			SELECT 1 FROM carpool_passengers cp
			JOIN carpool_offers co ON co.id = cp.offer_id
			WHERE co.event_id = (SELECT event_id FROM carpool_offers WHERE id = $1)
			AND cp.member_id = $2
		)
		AND (
			SELECT COUNT(*) FROM carpool_passengers WHERE offer_id = $1
		) < (
			SELECT available_seats FROM carpool_offers WHERE id = $1
		)
	`, offerId, memberId)
	if err != nil {
		return false, err
	}
	return tag.RowsAffected() == 1, nil
}

func (r carpoolRepository) RemovePassenger(ctx context.Context, offerId, memberId string) (bool, error) {
	_, err := r.db.Exec(ctx, `
		DELETE FROM carpool_passengers WHERE offer_id = $1 AND member_id = $2
	`, offerId, memberId)
	return err == nil, err
}

func (r carpoolRepository) CountPassengers(ctx context.Context, offerId string) (int, error) {
	var count int
	err := r.db.QueryRow(ctx, `
		SELECT COUNT(*) FROM carpool_passengers WHERE offer_id = $1
	`, offerId).Scan(&count)
	return count, err
}

func (r carpoolRepository) IsPassenger(ctx context.Context, eventId, memberId string) (bool, error) {
	var count int
	err := r.db.QueryRow(ctx, `
		SELECT COUNT(*) FROM carpool_passengers cp
		JOIN carpool_offers co ON co.id = cp.offer_id
		WHERE co.event_id = $1 AND cp.member_id = $2
	`, eventId, memberId).Scan(&count)
	return count > 0, err
}
