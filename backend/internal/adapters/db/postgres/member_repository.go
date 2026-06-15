package postgres

import (
	"clubmanager/internal/domain"
	"clubmanager/internal/domain/members"
	"context"
	"errors"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type memberRepository struct {
	db *pgxpool.Pool
}

func NewMemberRepository(db *pgxpool.Pool) *memberRepository {
	return &memberRepository{db: db}
}

func (r memberRepository) Save(ctx context.Context, m *members.Member) (*members.Member, error) {
	if err := setMetadataLog(ctx, r.db, m.Id.String()); err != nil {
		return nil, err
	}
	if m.Id == uuid.Nil {
		return r.insert(ctx, m)
	}
	return r.update(ctx, m)
}

func (r memberRepository) insert(ctx context.Context, m *members.Member) (*members.Member, error) {
	row := r.db.QueryRow(ctx, `
		INSERT INTO members (user_id, firstname, lastname, birthdate, gender, is_primary)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, user_id, firstname, lastname, birthdate::text, gender, is_primary
	`, m.UserId, m.Firstname, m.Lastname, m.Birthdate, m.Gender, m.IsPrimary)

	if err := row.Scan(&m.Id, &m.UserId, &m.Firstname, &m.Lastname, &m.Birthdate, &m.Gender, &m.IsPrimary); err != nil {
		return nil, err
	}
	return m, nil
}

func (r memberRepository) update(ctx context.Context, m *members.Member) (*members.Member, error) {
	_, err := r.db.Exec(ctx, `
		UPDATE members SET firstname = $1, lastname = $2, birthdate = $3, gender = $4, is_primary = $5
		WHERE id = $6
	`, m.Firstname, m.Lastname, m.Birthdate, m.Gender, m.IsPrimary, m.Id)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func (r memberRepository) Find(ctx context.Context, id string) (*members.Member, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, user_id, firstname, lastname, birthdate::text, gender, is_primary
		FROM members WHERE id = $1
	`, id)

	m := members.Member{}
	if err := row.Scan(&m.Id, &m.UserId, &m.Firstname, &m.Lastname, &m.Birthdate, &m.Gender, &m.IsPrimary); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &m, nil
}

func (r memberRepository) Search(ctx context.Context, params *domain.SearchParams) ([]*members.Member, error) {
	where, args, err := params.GetWhereClauses()
	if err != nil {
		return nil, err
	}

	rows, err := r.db.Query(ctx,
		"SELECT id, user_id, firstname, lastname, birthdate::text, gender, is_primary FROM members WHERE "+where,
		args...,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*members.Member
	for rows.Next() {
		m := members.Member{}
		if err := rows.Scan(&m.Id, &m.UserId, &m.Firstname, &m.Lastname, &m.Birthdate, &m.Gender, &m.IsPrimary); err != nil {
			return nil, err
		}
		list = append(list, &m)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return list, nil
}

// FindByIds fetches all members whose ID is in the given list in a single query.
// IDs that don't match any row are silently absent from the result.
func (r memberRepository) FindByIds(ctx context.Context, ids []string) ([]*members.Member, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	// Build $1, $2, ... placeholders
	placeholders := make([]string, len(ids))
	args := make([]any, len(ids))
	for i, id := range ids {
		placeholders[i] = "$" + strconv.Itoa(i+1)
		args[i] = id
	}
	query := "SELECT id, user_id, firstname, lastname, birthdate::text, gender, is_primary FROM members WHERE id IN (" + strings.Join(placeholders, ", ") + ")"
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*members.Member
	for rows.Next() {
		m := members.Member{}
		if err := rows.Scan(&m.Id, &m.UserId, &m.Firstname, &m.Lastname, &m.Birthdate, &m.Gender, &m.IsPrimary); err != nil {
			return nil, err
		}
		list = append(list, &m)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return list, nil
}

func (r memberRepository) Delete(ctx context.Context, id string) (bool, error) {
	if err := setMetadataLog(ctx, r.db, id); err != nil {
		return false, err
	}
	_, err := r.db.Exec(ctx, `DELETE FROM members WHERE id = $1`, id)
	if err != nil {
		return false, err
	}
	return true, nil
}
