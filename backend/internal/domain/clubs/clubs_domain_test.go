package clubs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewClub(t *testing.T) {
	tests := []struct {
		name     string
		data     map[string]string
		wantErr  bool
	}{
		{"Valid club", map[string]string{
			"siren": "123456789", "name": "Judo Club Paris",
			"address": "1 rue de la Paix", "city": "Paris",
			"postal_code": "75001", "phonenumber": "0606060606",
		}, false},
		{"Invalid SIREN (too short)", map[string]string{
			"siren": "12345", "name": "Judo Club",
			"address": "1 rue", "city": "Paris", "postal_code": "75001",
		}, true},
		{"Invalid SIREN (letters)", map[string]string{
			"siren": "12345678a", "name": "Judo Club",
			"address": "1 rue", "city": "Paris", "postal_code": "75001",
		}, true},
		{"Blank name", map[string]string{
			"siren": "123456789", "name": "",
			"address": "1 rue", "city": "Paris", "postal_code": "75001",
		}, true},
		{"Blank address", map[string]string{
			"siren": "123456789", "name": "Judo Club",
			"address": "", "city": "Paris", "postal_code": "75001",
		}, true},
		{"Blank city", map[string]string{
			"siren": "123456789", "name": "Judo Club",
			"address": "1 rue", "city": "", "postal_code": "75001",
		}, true},
		{"Invalid postal code", map[string]string{
			"siren": "123456789", "name": "Judo Club",
			"address": "1 rue", "city": "Paris", "postal_code": "7500",
		}, true},
		{"Invalid phonenumber", map[string]string{
			"siren": "123456789", "name": "Judo Club",
			"address": "1 rue", "city": "Paris", "postal_code": "75001",
			"phonenumber": "123",
		}, true},
		{"Empty phonenumber is valid (optional)", map[string]string{
			"siren": "123456789", "name": "Judo Club",
			"address": "1 rue", "city": "Paris", "postal_code": "75001",
			"phonenumber": "",
		}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, errs := NewClub(tt.data)
			if tt.wantErr {
				assert.True(t, len(errs) > 0)
			} else {
				assert.Empty(t, errs)
			}
		})
	}
}

func TestIsValidSiren(t *testing.T) {
	tests := []struct {
		name    string
		siren   string
		wantErr bool
	}{
		{"Valid SIREN", "123456789", false},
		{"Too short", "12345678", true},
		{"Too long", "1234567890", true},
		{"Contains letter", "12345678a", true},
		{"Empty", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ok, _ := IsValidSiren(tt.siren)
			assert.Equal(t, !tt.wantErr, ok)
		})
	}
}

func TestIsValidPostalCode(t *testing.T) {
	tests := []struct {
		name    string
		code    string
		wantErr bool
	}{
		{"Valid code", "75001", false},
		{"Too short", "7500", true},
		{"Too long", "750011", true},
		{"Contains letter", "7500a", true},
		{"Empty", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ok, _ := IsValidPostalCode(tt.code)
			assert.Equal(t, !tt.wantErr, ok)
		})
	}
}

func TestClubUpdate(t *testing.T) {
	base := Club{
		Siren: "123456789", Name: "Old Name", Address: "Old Address",
		City: "Old City", PostalCode: "75001", Country: "FRANCE",
	}

	t.Run("Partial update keeps existing values", func(t *testing.T) {
		updated, errs := base.Update(map[string]string{"name": "New Name"})
		assert.Empty(t, errs)
		assert.Equal(t, "New Name", updated.Name)
		assert.Equal(t, base.Address, updated.Address)
		assert.Equal(t, base.Siren, updated.Siren)
	})

	t.Run("SIREN cannot be changed via Update", func(t *testing.T) {
		updated, errs := base.Update(map[string]string{"siren": "999999999"})
		assert.Empty(t, errs)
		assert.Equal(t, base.Siren, updated.Siren)
	})

	t.Run("Invalid postal code in update", func(t *testing.T) {
		_, errs := base.Update(map[string]string{"postal_code": "bad"})
		assert.True(t, len(errs) > 0)
	})
}
