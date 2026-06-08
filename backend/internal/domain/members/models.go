package members

import (
	"maps"

	"github.com/google/uuid"
)

type Member struct {
	Id        uuid.UUID
	UserId    uuid.UUID
	ClubId    uuid.UUID
	Firstname string
	Lastname  string
	Birthdate string
	Gender    string
	IsValid   bool
}

func NewMember(userId, clubId string, data map[string]string) (*Member, map[string]string) {
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
	clubUUID, err := uuid.Parse(clubId)
	if err != nil {
		errs["club_id"] = "Invalid club ID."
	}

	return &Member{
		UserId:    userUUID,
		ClubId:    clubUUID,
		Firstname: data["firstname"],
		Lastname:  data["lastname"],
		Birthdate: data["birthdate"],
		Gender:    data["gender"],
		IsValid:   false,
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

	member, errs := NewMember(m.UserId.String(), m.ClubId.String(), updated)
	if len(errs) == 0 {
		member.Id = m.Id
		member.IsValid = m.IsValid
	}
	return member, errs
}
