package postgres

import (
	"clubmanager/internal/domain/posts"
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type postRepository struct {
	db *pgxpool.Pool
}

func NewPostRepository(db *pgxpool.Pool) *postRepository {
	return &postRepository{db: db}
}

func (r postRepository) Save(ctx context.Context, p *posts.Post) (*posts.Post, error) {
	if p.Id == uuid.Nil {
		return r.insert(ctx, p)
	}
	return r.update(ctx, p)
}

func (r postRepository) insert(ctx context.Context, p *posts.Post) (*posts.Post, error) {
	var eid *uuid.UUID
	if p.EventId != uuid.Nil {
		eid = &p.EventId
	}
	row := r.db.QueryRow(ctx, `
		INSERT INTO posts (club_id, author_id, event_id, title, content, status, visibility)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, club_id, author_id, event_id, title, content, status, visibility,
		          created_at::text, updated_at::text
	`, p.ClubId, p.AuthorId, eid, p.Title, p.Content, p.Status, p.Visibility)

	return scanPost(row.Scan)
}

func (r postRepository) update(ctx context.Context, p *posts.Post) (*posts.Post, error) {
	row := r.db.QueryRow(ctx, `
		UPDATE posts
		SET title = $1, content = $2, status = $3, visibility = $4
		WHERE id = $5
		RETURNING id, club_id, author_id, event_id, title, content, status, visibility,
		          created_at::text, updated_at::text
	`, p.Title, p.Content, p.Status, p.Visibility, p.Id)

	return scanPost(row.Scan)
}

func (r postRepository) Find(ctx context.Context, id string) (*posts.Post, error) {
	row := r.db.QueryRow(ctx, `
		SELECT p.id, p.club_id, p.author_id, p.event_id,
		       p.title, p.content, p.status, p.visibility,
		       p.created_at::text, p.updated_at::text,
		       COALESCE(e.date::text, '')
		FROM posts p
		LEFT JOIN events e ON e.id = p.event_id
		WHERE p.id = $1
	`, id)

	p, err := scanPostFull(row.Scan)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return p, nil
}

func (r postRepository) FindByClub(ctx context.Context, clubId string, params posts.PostSearchParams) ([]*posts.Post, int, error) {
	limit := params.Limit()
	offset := params.Offset()

	// Count total matching rows.
	var total int
	err := r.db.QueryRow(ctx, `
		SELECT COUNT(*) FROM posts
		WHERE club_id = $1
		  AND ($2 = '' OR status = $2)
		  AND ($3 = '' OR visibility = $3)
	`, clubId, params.Status, params.Visibility).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.db.Query(ctx, `
		SELECT p.id, p.club_id, p.author_id, p.event_id,
		       p.title, p.content, p.status, p.visibility,
		       p.created_at::text, p.updated_at::text,
		       COALESCE(e.date::text, '')
		FROM posts p
		LEFT JOIN events e ON e.id = p.event_id
		WHERE p.club_id = $1
		  AND ($2 = '' OR p.status = $2)
		  AND ($3 = '' OR p.visibility = $3)
		ORDER BY p.created_at DESC
		LIMIT $4 OFFSET $5
	`, clubId, params.Status, params.Visibility, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var list []*posts.Post
	for rows.Next() {
		p, err := scanPostFull(rows.Scan)
		if err != nil {
			return nil, 0, err
		}
		list = append(list, p)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r postRepository) FindByEventId(ctx context.Context, eventId string) (*posts.Post, error) {
	row := r.db.QueryRow(ctx, `
		SELECT p.id, p.club_id, p.author_id, p.event_id,
		       p.title, p.content, p.status, p.visibility,
		       p.created_at::text, p.updated_at::text,
		       COALESCE(e.date::text, '')
		FROM posts p
		LEFT JOIN events e ON e.id = p.event_id
		WHERE p.event_id = $1
		LIMIT 1
	`, eventId)

	p, err := scanPostFull(row.Scan)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return p, nil
}

func (r postRepository) Delete(ctx context.Context, id string) (bool, error) {
	_, err := r.db.Exec(ctx, `DELETE FROM posts WHERE id = $1`, id)
	return err == nil, err
}

// scanPost is used for INSERT/UPDATE RETURNING (no event JOIN).
func scanPost(scan func(...any) error) (*posts.Post, error) {
	p := &posts.Post{}
	var eid *uuid.UUID
	if err := scan(
		&p.Id, &p.ClubId, &p.AuthorId, &eid,
		&p.Title, &p.Content, &p.Status, &p.Visibility,
		&p.CreatedAt, &p.UpdatedAt,
	); err != nil {
		return nil, err
	}
	if eid != nil {
		p.EventId = *eid
	}
	return p, nil
}

// scanPostFull is used for SELECT with LEFT JOIN events (includes EventDate).
func scanPostFull(scan func(...any) error) (*posts.Post, error) {
	p := &posts.Post{}
	var eid *uuid.UUID
	if err := scan(
		&p.Id, &p.ClubId, &p.AuthorId, &eid,
		&p.Title, &p.Content, &p.Status, &p.Visibility,
		&p.CreatedAt, &p.UpdatedAt, &p.EventDate,
	); err != nil {
		return nil, err
	}
	if eid != nil {
		p.EventId = *eid
	}
	return p, nil
}

// HasMembership returns true when the user has at least one valid ClubMembership
// in the given club. Used by the post service for visibility enforcement.
func (r postRepository) HasMembership(ctx context.Context, userId, clubId string) (bool, error) {
	var count int
	err := r.db.QueryRow(ctx, `
		SELECT COUNT(*) FROM club_memberships cm
		JOIN members m ON m.id = cm.member_id
		WHERE m.user_id = $1 AND cm.club_id = $2 AND cm.is_valid = true
	`, userId, clubId).Scan(&count)
	return count > 0, err
}
