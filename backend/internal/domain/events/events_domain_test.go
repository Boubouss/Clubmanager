package events

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewEvent(t *testing.T) {
	validData := map[string]string{
		"title":                "Championnat départemental",
		"type":                 TypeCompetition,
		"location":             "Dojo Paris",
		"registration_open_at": "2026-09-01T00:00:00Z",
		"registration_close_at": "2026-09-20T00:00:00Z",
		"date":                 "2026-10-01T09:00:00Z",
	}

	tests := []struct {
		name      string
		clubId    string
		createdBy string
		data      map[string]string
		wantErr   bool
	}{
		{"Valid competition", "00000000-0000-0000-0000-000000000001", "00000000-0000-0000-0000-000000000002", validData, false},
		{"Missing title", "00000000-0000-0000-0000-000000000001", "00000000-0000-0000-0000-000000000002", func() map[string]string {
			d := copyMap(validData); delete(d, "title"); return d
		}(), true},
		{"Invalid type", "00000000-0000-0000-0000-000000000001", "00000000-0000-0000-0000-000000000002", func() map[string]string {
			d := copyMap(validData); d["type"] = "invalid"; return d
		}(), true},
		{"Close before open", "00000000-0000-0000-0000-000000000001", "00000000-0000-0000-0000-000000000002", func() map[string]string {
			d := copyMap(validData)
			d["registration_close_at"] = "2026-08-01T00:00:00Z"
			return d
		}(), true},
		{"Invalid club_id", "not-a-uuid", "00000000-0000-0000-0000-000000000002", validData, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event, errs := NewEvent(tt.clubId, tt.createdBy, tt.data)
			if tt.wantErr {
				assert.True(t, len(errs) > 0)
				assert.Nil(t, event)
			} else {
				assert.Empty(t, errs)
				assert.NotNil(t, event)
				assert.Equal(t, StatusDraft, event.Status)
			}
		})
	}
}

func TestIsRoleAllowed(t *testing.T) {
	tests := []struct {
		eventType string
		role      string
		allowed   bool
	}{
		{TypeCompetition, RoleCompetitor, true},
		{TypeCompetition, RoleCoach, true},
		{TypeCompetition, RoleReferee, true},
		{TypeCompetition, RoleAccompanist, true},
		{TypeStage, RoleCompetitor, true},
		{TypeStage, RoleCoach, true},
		{TypeStage, RoleReferee, false},
		{TypeStage, RoleAccompanist, true},
		{TypeAutre, RoleAccompanist, true},
		{TypeAutre, RoleCompetitor, false},
		{TypeAutre, RoleCoach, false},
	}

	for _, tt := range tests {
		t.Run(tt.eventType+"/"+tt.role, func(t *testing.T) {
			assert.Equal(t, tt.allowed, IsRoleAllowed(tt.eventType, tt.role))
		})
	}
}

func TestRequiresLicence(t *testing.T) {
	assert.True(t, RequiresLicence(RoleCompetitor))
	assert.True(t, RequiresLicence(RoleReferee))
	assert.False(t, RequiresLicence(RoleCoach))
	assert.False(t, RequiresLicence(RoleAccompanist))
}

func TestEventUpdate(t *testing.T) {
	base := &Event{
		Title:               "Original",
		Type:                TypeCompetition,
		Status:              StatusDraft,
		Location:            "Paris",
		RegistrationOpenAt:  "2026-09-01T00:00:00Z",
		RegistrationCloseAt: "2026-09-20T00:00:00Z",
		Date:                "2026-10-01T09:00:00Z",
		MaxParticipants:     0,
	}

	tests := []struct {
		name      string
		data      map[string]string
		wantErr   bool
		checkFunc func(t *testing.T, e *Event)
	}{
		{
			name: "Partial update title only",
			data: map[string]string{"title": "New Title"},
			checkFunc: func(t *testing.T, e *Event) {
				assert.Equal(t, "New Title", e.Title)
				assert.Equal(t, base.Location, e.Location)
				assert.Equal(t, base.Type, e.Type)
			},
		},
		{
			name: "Update type to stage",
			data: map[string]string{"type": TypeStage},
			checkFunc: func(t *testing.T, e *Event) {
				assert.Equal(t, TypeStage, e.Type)
			},
		},
		{
			name:    "Invalid type",
			data:    map[string]string{"type": "invalid"},
			wantErr: true,
		},
		{
			name:    "Invalid max_participants (negative)",
			data:    map[string]string{"max_participants": "-5"},
			wantErr: true,
		},
		{
			name:    "Invalid max_participants (text)",
			data:    map[string]string{"max_participants": "abc"},
			wantErr: true,
		},
		{
			name:    "Close date before open date",
			data:    map[string]string{"registration_close_at": "2026-08-01T00:00:00Z"},
			wantErr: true,
		},
		{
			name: "Empty data keeps all existing values",
			data: map[string]string{},
			checkFunc: func(t *testing.T, e *Event) {
				assert.Equal(t, base.Title, e.Title)
				assert.Equal(t, base.Type, e.Type)
				assert.Equal(t, base.Location, e.Location)
			},
		},
		{
			name: "Update max_participants",
			data: map[string]string{"max_participants": "50"},
			checkFunc: func(t *testing.T, e *Event) {
				assert.Equal(t, 50, e.MaxParticipants)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			updated, errs := base.Update(tt.data)
			if tt.wantErr {
				assert.True(t, len(errs) > 0)
				assert.Nil(t, updated)
			} else {
				assert.Empty(t, errs)
				assert.NotNil(t, updated)
				if tt.checkFunc != nil {
					tt.checkFunc(t, updated)
				}
			}
		})
	}
}

func copyMap(m map[string]string) map[string]string {
	out := make(map[string]string, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}
