package members

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

var (
	validUserId = uuid.New().String()
	validClubId = uuid.New().String()
)

func TestNewMember(t *testing.T) {
	tests := []struct {
		name    string
		data    map[string]string
		wantErr bool
	}{
		{"Valid member", map[string]string{
			"firstname": "Jean", "lastname": "Dupont",
			"birthdate": "1990-06-15", "gender": "man",
		}, false},
		{"Blank firstname", map[string]string{
			"firstname": "", "lastname": "Dupont",
			"birthdate": "1990-06-15", "gender": "man",
		}, true},
		{"Blank lastname", map[string]string{
			"firstname": "Jean", "lastname": "",
			"birthdate": "1990-06-15", "gender": "man",
		}, true},
		{"Invalid birthdate format", map[string]string{
			"firstname": "Jean", "lastname": "Dupont",
			"birthdate": "15/06/1990", "gender": "man",
		}, true},
		{"Future birthdate", map[string]string{
			"firstname": "Jean", "lastname": "Dupont",
			"birthdate": "2099-01-01", "gender": "man",
		}, true},
		{"Invalid gender", map[string]string{
			"firstname": "Jean", "lastname": "Dupont",
			"birthdate": "1990-06-15", "gender": "other",
		}, true},
		{"Woman gender valid", map[string]string{
			"firstname": "Marie", "lastname": "Dupont",
			"birthdate": "1995-03-20", "gender": "woman",
		}, false},
		{"Invalid user_id", map[string]string{
			"firstname": "Jean", "lastname": "Dupont",
			"birthdate": "1990-06-15", "gender": "man",
		}, true},
	}

	for i, tt := range tests {
		userId := validUserId
		// last test uses invalid userId
		if i == len(tests)-1 {
			userId = "not-a-uuid"
		}
		t.Run(tt.name, func(t *testing.T) {
			_, errs := NewMember(userId, tt.data, false)
			if tt.wantErr {
				assert.True(t, len(errs) > 0)
			} else {
				assert.Empty(t, errs)
			}
		})
	}
}

func TestNewMember_IsPrimary(t *testing.T) {
	data := map[string]string{
		"firstname": "Jean", "lastname": "Dupont",
		"birthdate": "1990-06-15", "gender": "man",
	}

	m, errs := NewMember(validUserId, data, true)
	assert.Empty(t, errs)
	assert.True(t, m.IsPrimary)

	m2, errs2 := NewMember(validUserId, data, false)
	assert.Empty(t, errs2)
	assert.False(t, m2.IsPrimary)
}

func TestMemberUpdate(t *testing.T) {
	base := Member{
		Id:        uuid.New(),
		UserId:    uuid.MustParse(validUserId),
		Firstname: "Jean",
		Lastname:  "Dupont",
		Birthdate: "1990-06-15",
		Gender:    "man",
		IsPrimary: true,
	}

	t.Run("Partial update keeps existing values", func(t *testing.T) {
		updated, errs := base.Update(map[string]string{"firstname": "Paul"})
		assert.Empty(t, errs)
		assert.Equal(t, "Paul", updated.Firstname)
		assert.Equal(t, base.Lastname, updated.Lastname)
		assert.Equal(t, base.IsPrimary, updated.IsPrimary)
	})

	t.Run("Invalid gender in update", func(t *testing.T) {
		_, errs := base.Update(map[string]string{"gender": "alien"})
		assert.True(t, len(errs) > 0)
	})
}

func TestNewClubMembership(t *testing.T) {
	memberId := uuid.New().String()

	t.Run("Valid membership", func(t *testing.T) {
		cm, errs := NewClubMembership(memberId, validClubId)
		assert.Empty(t, errs)
		assert.NotNil(t, cm)
		assert.False(t, cm.IsValid)
		assert.Equal(t, validClubId, cm.ClubId.String())
	})

	t.Run("Invalid member ID", func(t *testing.T) {
		_, errs := NewClubMembership("not-a-uuid", validClubId)
		assert.True(t, len(errs) > 0)
	})

	t.Run("Invalid club ID", func(t *testing.T) {
		_, errs := NewClubMembership(memberId, "not-a-uuid")
		assert.True(t, len(errs) > 0)
	})
}

func TestIsValidBirthdate(t *testing.T) {
	tests := []struct {
		name    string
		date    string
		wantErr bool
	}{
		{"Valid date", "1990-06-15", false},
		{"Future date", "2099-01-01", true},
		{"Wrong format", "15/06/1990", true},
		{"Too old", "1800-01-01", true},
		{"Empty", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ok, _ := IsValidBirthdate(tt.date)
			assert.Equal(t, !tt.wantErr, ok)
		})
	}
}
