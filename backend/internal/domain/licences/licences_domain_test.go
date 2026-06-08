package licences

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

var validMemberId = uuid.New().String()

func TestNewLicence(t *testing.T) {
	tests := []struct {
		name    string
		data    map[string]string
		wantErr bool
	}{
		{"Valid licence", map[string]string{
			"licence_number": "FF123456",
			"valid_from":     "2024-09-01",
			"valid_until":    "2025-08-31",
		}, false},
		{"Blank licence number", map[string]string{
			"licence_number": "",
			"valid_from":     "2024-09-01",
			"valid_until":    "2025-08-31",
		}, true},
		{"valid_until before valid_from", map[string]string{
			"licence_number": "FF123456",
			"valid_from":     "2025-08-31",
			"valid_until":    "2024-09-01",
		}, true},
		{"Same dates", map[string]string{
			"licence_number": "FF123456",
			"valid_from":     "2024-09-01",
			"valid_until":    "2024-09-01",
		}, true},
		{"Invalid date format", map[string]string{
			"licence_number": "FF123456",
			"valid_from":     "01/09/2024",
			"valid_until":    "2025-08-31",
		}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, errs := NewLicence(validMemberId, tt.data)
			if tt.wantErr {
				assert.True(t, len(errs) > 0)
			} else {
				assert.Empty(t, errs)
			}
		})
	}
}

func TestNewLicenceDefaultStatus(t *testing.T) {
	l, errs := NewLicence(validMemberId, map[string]string{
		"licence_number": "FF123456",
		"valid_from":     "2024-09-01",
		"valid_until":    "2025-08-31",
	})
	assert.Empty(t, errs)
	assert.Equal(t, "pending", l.Status)
}

func TestUpdateLicenceStatus(t *testing.T) {
	base := Licence{Id: uuid.New(), Status: "pending"}

	tests := []struct {
		name    string
		status  string
		wantErr bool
	}{
		{"To active", "active", false},
		{"To expired", "expired", false},
		{"To pending", "pending", false},
		{"Invalid status", "unknown", true},
		{"Empty status", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			updated, errs := base.UpdateStatus(tt.status)
			if tt.wantErr {
				assert.True(t, len(errs) > 0)
				assert.Nil(t, updated)
			} else {
				assert.Empty(t, errs)
				assert.Equal(t, tt.status, updated.Status)
			}
		})
	}
}
