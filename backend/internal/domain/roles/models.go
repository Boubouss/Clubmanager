package roles

import "github.com/google/uuid"

const (
	RolePresident = "president"
	RoleStaff     = "staff"
	RoleCoach     = "coach"
	RoleAssociate = "associate"
)

type Role struct {
	Id     uuid.UUID
	UserId uuid.UUID
	ClubId uuid.UUID
	Role   string
}

func IsValid(role string) bool {
	switch role {
	case RolePresident, RoleStaff, RoleCoach, RoleAssociate:
		return true
	}
	return false
}

// CanAssign returns true if the assigner's role is allowed to assign the target role.
func CanAssign(assignerRole, targetRole string) bool {
	if assignerRole == RolePresident {
		return true
	}
	// staff can assign coach and associate, not president or staff
	if assignerRole == RoleStaff {
		return targetRole == RoleCoach || targetRole == RoleAssociate
	}
	return false
}
