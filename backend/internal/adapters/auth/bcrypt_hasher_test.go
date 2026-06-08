package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBcryptHasher_Hash(t *testing.T) {
	h := NewBcryptHasher()

	hash, err := h.Hash("Password123!")

	assert.NoError(t, err)
	assert.NotEmpty(t, hash)
	assert.NotEqual(t, "Password123!", hash)
}

func TestBcryptHasher_Compare(t *testing.T) {
	h := NewBcryptHasher()

	hash, _ := h.Hash("Password123!")

	tests := []struct {
		name    string
		plain   string
		hash    string
		wantErr bool
	}{
		{"Correct password", "Password123!", hash, false},
		{"Wrong password", "wrongpassword", hash, true},
		{"Empty password", "", hash, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := h.Compare(tt.plain, tt.hash)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
