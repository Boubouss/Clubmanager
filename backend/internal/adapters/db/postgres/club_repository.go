package postgres

import (
	"clubmanager/internal/domain"
	"clubmanager/internal/domain/clubs"
	"context"

	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type clubRepository struct {
	db *pgxpool.Pool
}

func NewClubRepository(db *pgxpool.Pool) *clubRepository {
	return &clubRepository{db: db}
}

func (r clubRepository) Save(ctx context.Context, c *clubs.Club) (*clubs.Club, error) {
	if err := setMetadataLog(ctx, r.db, c.Id.String()); err != nil {
		return nil, err
	}

	if c.Id == uuid.Nil {
		return r.insert(ctx, c)
	}
	return r.update(ctx, c)
}

func (r clubRepository) insert(ctx context.Context, c *clubs.Club) (*clubs.Club, error) {
	row := r.db.QueryRow(ctx, `
		INSERT INTO clubs (siren, name, address, city, postal_code, country, phonenumber)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, siren, name, address, city, postal_code, country, phonenumber
	`, c.Siren, c.Name, c.Address, c.City, c.PostalCode, c.Country, c.Phonenumber)

	if err := row.Scan(&c.Id, &c.Siren, &c.Name, &c.Address, &c.City, &c.PostalCode, &c.Country, &c.Phonenumber); err != nil {
		return nil, err
	}
	return c, nil
}

func (r clubRepository) update(ctx context.Context, c *clubs.Club) (*clubs.Club, error) {
	_, err := r.db.Exec(ctx, `
		UPDATE clubs SET name = $1, address = $2, city = $3, postal_code = $4, country = $5, phonenumber = $6
		WHERE id = $7
	`, c.Name, c.Address, c.City, c.PostalCode, c.Country, c.Phonenumber, c.Id.String())

	if err != nil {
		return nil, err
	}
	return c, nil
}

func (r clubRepository) Find(ctx context.Context, id string) (*clubs.Club, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, siren, name, address, city, postal_code, country, phonenumber
		FROM clubs WHERE id = $1
	`, id)

	c := clubs.Club{}
	if err := row.Scan(&c.Id, &c.Siren, &c.Name, &c.Address, &c.City, &c.PostalCode, &c.Country, &c.Phonenumber); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &c, nil
}

func (r clubRepository) Search(ctx context.Context, params *domain.SearchParams) ([]*clubs.Club, error) {
	where, args, err := params.GetWhereClauses()
	if err != nil {
		return nil, err
	}

	rows, err := r.db.Query(ctx,
		"SELECT id, siren, name, address, city, postal_code, country, phonenumber FROM clubs WHERE "+where,
		args...,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*clubs.Club
	for rows.Next() {
		c := clubs.Club{}
		if err := rows.Scan(&c.Id, &c.Siren, &c.Name, &c.Address, &c.City, &c.PostalCode, &c.Country, &c.Phonenumber); err != nil {
			return nil, err
		}
		list = append(list, &c)
	}
	return list, nil
}

func (r clubRepository) Delete(ctx context.Context, id string) (bool, error) {
	if err := setMetadataLog(ctx, r.db, id); err != nil {
		return false, err
	}

	_, err := r.db.Exec(ctx, `DELETE FROM clubs WHERE id = $1`, id)
	if err != nil {
		return false, err
	}
	return true, nil
}
