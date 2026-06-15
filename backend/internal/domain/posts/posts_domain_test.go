package posts

import (
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

var (
	validClubId   = uuid.New().String()
	validAuthorId = uuid.New().String()
)

func TestNewPost(t *testing.T) {
	tests := []struct {
		name       string
		clubId     string
		authorId   string
		data       map[string]string
		wantErrors []string
		wantStatus string
		wantVis    string
	}{
		{
			name:       "Valid post — defaults to draft/adherent",
			clubId:     validClubId,
			authorId:   validAuthorId,
			data:       map[string]string{"title": "Mon titre", "content": "Mon contenu"},
			wantStatus: StatusDraft,
			wantVis:    VisibilityAdherent,
		},
		{
			name:       "Valid post — published + public",
			clubId:     validClubId,
			authorId:   validAuthorId,
			data:       map[string]string{"title": "Actu", "content": "Corps", "status": "published", "visibility": "public"},
			wantStatus: StatusPublished,
			wantVis:    VisibilityPublic,
		},
		{
			name:       "Missing title",
			clubId:     validClubId,
			authorId:   validAuthorId,
			data:       map[string]string{"content": "Corps"},
			wantErrors: []string{"title"},
		},
		{
			name:       "Title too long (201 chars)",
			clubId:     validClubId,
			authorId:   validAuthorId,
			data:       map[string]string{"title": strings.Repeat("a", 201), "content": "Corps"},
			wantErrors: []string{"title"},
		},
		{
			name:       "Missing content",
			clubId:     validClubId,
			authorId:   validAuthorId,
			data:       map[string]string{"title": "Titre"},
			wantErrors: []string{"content"},
		},
		{
			name:       "Whitespace-only title rejected",
			clubId:     validClubId,
			authorId:   validAuthorId,
			data:       map[string]string{"title": "   ", "content": "Corps"},
			wantErrors: []string{"title"},
		},
		{
			name:       "Whitespace-only content rejected",
			clubId:     validClubId,
			authorId:   validAuthorId,
			data:       map[string]string{"title": "Titre", "content": "\t\n "},
			wantErrors: []string{"content"},
		},
		{
			name:       "Title is trimmed before save",
			clubId:     validClubId,
			authorId:   validAuthorId,
			data:       map[string]string{"title": "  Mon actu  ", "content": "Corps"},
			wantStatus: StatusDraft,
			wantVis:    VisibilityAdherent,
		},
		{
			name:       "Invalid visibility",
			clubId:     validClubId,
			authorId:   validAuthorId,
			data:       map[string]string{"title": "T", "content": "C", "visibility": "private"},
			wantErrors: []string{"visibility"},
		},
		{
			name:       "Invalid status",
			clubId:     validClubId,
			authorId:   validAuthorId,
			data:       map[string]string{"title": "T", "content": "C", "status": "archived"},
			wantErrors: []string{"status"},
		},
		{
			name:       "Invalid club_id",
			clubId:     "not-a-uuid",
			authorId:   validAuthorId,
			data:       map[string]string{"title": "T", "content": "C"},
			wantErrors: []string{"club_id"},
		},
		{
			name:       "Invalid author_id",
			clubId:     validClubId,
			authorId:   "not-a-uuid",
			data:       map[string]string{"title": "T", "content": "C"},
			wantErrors: []string{"author_id"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			post, errs := NewPost(tt.clubId, tt.authorId, tt.data)
			if len(tt.wantErrors) > 0 {
				for _, key := range tt.wantErrors {
					assert.Contains(t, errs, key, "expected error key %q", key)
				}
				return
			}
			assert.Empty(t, errs)
			assert.Equal(t, tt.wantStatus, post.Status)
			assert.Equal(t, tt.wantVis, post.Visibility)
			assert.Equal(t, strings.TrimSpace(tt.data["title"]), post.Title)
			assert.Equal(t, strings.TrimSpace(tt.data["content"]), post.Content)
		})
	}
}

func TestPostUpdate(t *testing.T) {
	original := Post{
		Title:      "Titre original",
		Content:    "Contenu original",
		Visibility: VisibilityAdherent,
		Status:     StatusDraft,
	}

	tests := []struct {
		name       string
		data       map[string]string
		wantErrors []string
		wantTitle  string
		wantVis    string
	}{
		{
			name:      "Update title only",
			data:      map[string]string{"title": "Nouveau titre"},
			wantTitle: "Nouveau titre",
			wantVis:   VisibilityAdherent,
		},
		{
			name:      "Update visibility to public",
			data:      map[string]string{"visibility": "public"},
			wantTitle: "Titre original",
			wantVis:   VisibilityPublic,
		},
		{
			name:       "Whitespace-only title rejected",
			data:       map[string]string{"title": "   "},
			wantErrors: []string{"title"},
		},
		{
			name:       "Whitespace-only content rejected",
			data:       map[string]string{"content": "  \t  "},
			wantErrors: []string{"content"},
		},
		{
			name:      "Title is trimmed before save",
			data:      map[string]string{"title": "  Nouveau titre  "},
			wantTitle: "Nouveau titre",
			wantVis:   VisibilityAdherent,
		},
		{
			name:       "Invalid visibility",
			data:       map[string]string{"visibility": "private"},
			wantErrors: []string{"visibility"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			updated, errs := original.Update(tt.data)
			if len(tt.wantErrors) > 0 {
				for _, key := range tt.wantErrors {
					assert.Contains(t, errs, key)
				}
				return
			}
			assert.Empty(t, errs)
			assert.Equal(t, tt.wantTitle, updated.Title)
			assert.Equal(t, tt.wantVis, updated.Visibility)
			// Status is not changed by Update
			assert.Equal(t, original.Status, updated.Status)
		})
	}
}

func TestPostSearchParams(t *testing.T) {
	p := PostSearchParams{Page: 3, PageSize: 10}
	assert.Equal(t, 20, p.Offset())
	assert.Equal(t, 10, p.Limit())

	p2 := PostSearchParams{Page: 0, PageSize: 0}
	assert.Equal(t, 0, p2.Offset())
	assert.Equal(t, 20, p2.Limit()) // default

	p3 := PostSearchParams{Page: 1, PageSize: 200}
	assert.Equal(t, 100, p3.Limit()) // capped at 100
}
