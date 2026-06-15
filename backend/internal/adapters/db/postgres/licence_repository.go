package postgres

import (
	"clubmanager/internal/domain"
	"clubmanager/internal/domain/licences"
	"context"

	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type licenceRepository struct {
	db *pgxpool.Pool
}

func NewLicenceRepository(db *pgxpool.Pool) *licenceRepository {
	return &licenceRepository{db: db}
}

func (r licenceRepository) Save(ctx context.Context, l *licences.Licence) (*licences.Licence, error) {
	if err := setMetadataLog(ctx, r.db, l.Id.String()); err != nil {
		return nil, err
	}
	if l.Id == uuid.Nil {
		return r.insert(ctx, l)
	}
	return r.update(ctx, l)
}

func (r licenceRepository) insert(ctx context.Context, l *licences.Licence) (*licences.Licence, error) {
	row := r.db.QueryRow(ctx, `
		INSERT INTO licences (member_id, licence_number, valid_from, valid_until, status)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, member_id, licence_number, valid_from::text, valid_until::text, status
	`, l.MemberId, l.LicenceNumber, l.ValidFrom, l.ValidUntil, l.Status)

	if err := row.Scan(&l.Id, &l.MemberId, &l.LicenceNumber, &l.ValidFrom, &l.ValidUntil, &l.Status); err != nil {
		return nil, err
	}
	return l, nil
}

func (r licenceRepository) update(ctx context.Context, l *licences.Licence) (*licences.Licence, error) {
	_, err := r.db.Exec(ctx, `
		UPDATE licences SET status = $1 WHERE id = $2
	`, l.Status, l.Id)
	if err != nil {
		return nil, err
	}
	return l, nil
}

func (r licenceRepository) Find(ctx context.Context, id string) (*licences.Licence, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, member_id, licence_number, valid_from::text, valid_until::text, status
		FROM licences WHERE id = $1
	`, id)

	l := licences.Licence{}
	if err := row.Scan(&l.Id, &l.MemberId, &l.LicenceNumber, &l.ValidFrom, &l.ValidUntil, &l.Status); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &l, nil
}

func (r licenceRepository) Search(ctx context.Context, params *domain.SearchParams) ([]*licences.Licence, error) {
	where, args, err := params.GetWhereClauses()
	if err != nil {
		return nil, err
	}

	rows, err := r.db.Query(ctx,
		"SELECT id, member_id, licence_number, valid_from::text, valid_until::text, status FROM licences WHERE "+where,
		args...,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*licences.Licence
	for rows.Next() {
		l := licences.Licence{}
		if err := rows.Scan(&l.Id, &l.MemberId, &l.LicenceNumber, &l.ValidFrom, &l.ValidUntil, &l.Status); err != nil {
			return nil, err
		}
		list = append(list, &l)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return list, nil
}

func (r licenceRepository) Delete(ctx context.Context, id string) (bool, error) {
	if err := setMetadataLog(ctx, r.db, id); err != nil {
		return false, err
	}
	_, err := r.db.Exec(ctx, `DELETE FROM licences WHERE id = $1`, id)
	if err != nil {
		return false, err
	}
	return true, nil
}
