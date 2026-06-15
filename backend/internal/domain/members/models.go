package members

import (
	"maps"

	"github.com/google/uuid"
)

// Member represents a physical person linked to a user account.
// A member can belong to multiple clubs via ClubMembership.
type Member struct {
	Id        uuid.UUID
	UserId    uuid.UUID
	Firstname string
	Lastname  string
	Birthdate string
	Gender    string
	// IsPrimary indicates this member profile inherits the user's club roles.
	// A user should have at most one primary member per club.
	IsPrimary bool
}

// ClubMembership represents the association between a member and a club.
// IsValid is set to true once a club staff member approves the membership.
type ClubMembership struct {
	Id       uuid.UUID
	MemberId uuid.UUID
	ClubId   uuid.UUID
	IsValid  bool
}

func NewMember(userId string, data map[string]string, isPrimary bool) (*Member, map[string]string) {
	errs := make(map[string]string)

	if ok, err := IsNotBlank(data["firstname"], "Firstname"); !ok {
		errs["firstname"] = err
	}
	if ok, err := IsNotBlank(data["lastname"], "Lastname"); !ok {
		errs["lastname"] = err
	}
	if ok, err := IsValidBirthdate(data["birthdate"]); !ok {
		errs["birthdate"] = err
	}
	if ok, err := IsValidGender(data["gender"]); !ok {
		errs["gender"] = err
	}

	userUUID, err := uuid.Parse(userId)
	if err != nil {
		errs["user_id"] = "Invalid user ID."
	}

	return &Member{
		UserId:    userUUID,
		Firstname: data["firstname"],
		Lastname:  data["lastname"],
		Birthdate: data["birthdate"],
		Gender:    data["gender"],
		IsPrimary: isPrimary,
	}, errs
}

// MemberContact enriches a Member + ClubMembership with user contact info
// obtained via a JOIN on the users table.
type MemberContact struct {
	Member     Member
	Membership ClubMembership
	Email      string
	Phone      string
}

func NewClubMembership(memberId, clubId string) (*ClubMembership, map[string]string) {
	errs := make(map[string]string)

	memberUUID, err := uuid.Parse(memberId)
	if err != nil {
		errs["member_id"] = "Invalid member ID."
	}
	clubUUID, err := uuid.Parse(clubId)
	if err != nil {
		errs["club_id"] = "Invalid club ID."
	}

	if len(errs) > 0 {
		return nil, errs
	}

	return &ClubMembership{
		MemberId: memberUUID,
		ClubId:   clubUUID,
		IsValid:  false,
	}, errs
}

func (m Member) Update(data map[string]string) (*Member, map[string]string) {
	updated := make(map[string]string)
	maps.Copy(updated, data)

	if _, ok := updated["firstname"]; !ok {
		updated["firstname"] = m.Firstname
	}
	if _, ok := updated["lastname"]; !ok {
		updated["lastname"] = m.Lastname
	}
	if _, ok := updated["birthdate"]; !ok {
		updated["birthdate"] = m.Birthdate
	}
	if _, ok := updated["gender"]; !ok {
		updated["gender"] = m.Gender
	}

	member, errs := NewMember(m.UserId.String(), updated, m.IsPrimary)
	if len(errs) == 0 {
		member.Id = m.Id
	}
	return member, errs
}
