package posts

import (
	"context"
	"maps"
	"strings"

	"github.com/google/uuid"
)

const (
	StatusDraft     = "draft"
	StatusPublished = "published"

	VisibilityPublic   = "public"
	VisibilityAdherent = "adherent"
)

type Post struct {
	Id         uuid.UUID
	ClubId     uuid.UUID
	AuthorId   uuid.UUID
	EventId    uuid.UUID // zero value means no linked event
	EventDate  string    // populated via JOIN, empty if no linked event
	Title      string
	Content    string
	Status     string
	Visibility string
	CreatedAt  string
	UpdatedAt  string
}

// PostSearchParams controls filtering and pagination for FindByClub.
type PostSearchParams struct {
	// Status filters by post status. Empty string means no filter.
	Status string
	// Visibility filters by visibility. Empty string means no filter.
	Visibility string
	// Page is 1-indexed.
	Page     int
	PageSize int
}

func (p *PostSearchParams) Offset() int {
	if p.Page < 1 {
		p.Page = 1
	}
	return (p.Page - 1) * p.PageSize
}

func (p *PostSearchParams) Limit() int {
	if p.PageSize <= 0 {
		p.PageSize = 20
	}
	if p.PageSize > 100 {
		p.PageSize = 100
	}
	return p.PageSize
}

// PostRepository is the persistence port for posts.
type PostRepository interface {
	Save(ctx context.Context, p *Post) (*Post, error)
	Find(ctx context.Context, id string) (*Post, error)
	// FindByClub returns posts matching the params and the total matching count.
	FindByClub(ctx context.Context, clubId string, params PostSearchParams) ([]*Post, int, error)
	// FindByEventId returns the post linked to the given event, or nil if none.
	FindByEventId(ctx context.Context, eventId string) (*Post, error)
	Delete(ctx context.Context, id string) (bool, error)
}

func NewPost(clubId, authorId string, data map[string]string) (*Post, map[string]string) {
	errs := make(map[string]string)

	cid, err := uuid.Parse(clubId)
	if err != nil {
		errs["club_id"] = "Invalid club ID."
	}
	aid, err := uuid.Parse(authorId)
	if err != nil {
		errs["author_id"] = "Invalid author ID."
	}

	var eid uuid.UUID
	if rawEid := strings.TrimSpace(data["event_id"]); rawEid != "" {
		eid, _ = uuid.Parse(rawEid)
	}

	title := strings.TrimSpace(data["title"])
	if title == "" {
		errs["title"] = "Title is required."
	} else if len(title) > 200 {
		errs["title"] = "Title must not exceed 200 characters."
	}

	content := strings.TrimSpace(data["content"])
	if content == "" {
		errs["content"] = "Content is required."
	}

	visibility := data["visibility"]
	if visibility == "" {
		visibility = VisibilityAdherent
	} else if visibility != VisibilityPublic && visibility != VisibilityAdherent {
		errs["visibility"] = "Visibility must be 'public' or 'adherent'."
	}

	status := data["status"]
	if status == "" {
		status = StatusDraft
	} else if status != StatusDraft && status != StatusPublished {
		errs["status"] = "Status must be 'draft' or 'published'."
	}

	if len(errs) > 0 {
		return &Post{}, errs
	}

	return &Post{
		ClubId:     cid,
		AuthorId:   aid,
		EventId:    eid,
		Title:      title,
		Content:    content,
		Status:     status,
		Visibility: visibility,
	}, errs
}

// Update returns a new Post with the provided fields merged in.
// Only title, content and visibility can be updated — status is managed separately.
func (p Post) Update(data map[string]string) (*Post, map[string]string) {
	merged := make(map[string]string)
	maps.Copy(merged, map[string]string{
		"title":      p.Title,
		"content":    p.Content,
		"visibility": p.Visibility,
	})
	for k, v := range data {
		if v != "" {
			merged[k] = strings.TrimSpace(v)
		}
	}

	errs := make(map[string]string)

	if merged["title"] == "" {
		errs["title"] = "Title is required."
	} else if len(merged["title"]) > 200 {
		errs["title"] = "Title must not exceed 200 characters."
	}
	if merged["content"] == "" {
		errs["content"] = "Content is required."
	}
	if merged["visibility"] != VisibilityPublic && merged["visibility"] != VisibilityAdherent {
		errs["visibility"] = "Visibility must be 'public' or 'adherent'."
	}

	if len(errs) > 0 {
		return &Post{}, errs
	}

	updated := p
	updated.Title = merged["title"]
	updated.Content = merged["content"]
	updated.Visibility = merged["visibility"]
	return &updated, errs
}
