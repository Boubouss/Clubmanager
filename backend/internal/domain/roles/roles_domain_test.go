package roles

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsValid(t *testing.T) {
	tests := []struct {
		name  string
		role  string
		valid bool
	}{
		{"President", RolePresident, true},
		{"Staff", RoleStaff, true},
		{"Coach", RoleCoach, true},
		{"Associate", RoleAssociate, true},
		{"Unknown role", "member", false},
		{"Empty string", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.valid, IsValid(tt.role))
		})
	}
}

func TestCanAssign(t *testing.T) {
	tests := []struct {
		name        string
		assigner    string
		target      string
		canAssign   bool
	}{
		{"President can assign president", RolePresident, RolePresident, true},
		{"President can assign staff", RolePresident, RoleStaff, true},
		{"President can assign coach", RolePresident, RoleCoach, true},
		{"President can assign associate", RolePresident, RoleAssociate, true},
		{"Staff can assign coach", RoleStaff, RoleCoach, true},
		{"Staff can assign associate", RoleStaff, RoleAssociate, true},
		{"Staff cannot assign president", RoleStaff, RolePresident, false},
		{"Staff cannot assign staff", RoleStaff, RoleStaff, false},
		{"Coach cannot assign anyone", RoleCoach, RoleAssociate, false},
		{"Associate cannot assign anyone", RoleAssociate, RoleCoach, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.canAssign, CanAssign(tt.assigner, tt.target))
		})
	}
}
