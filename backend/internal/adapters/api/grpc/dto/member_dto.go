package dto

import m "clubmanager/internal/domain/members"

type AddMemberRequest struct {
	UserId    string
	ClubId    string
	Firstname string
	Lastname  string
	Birthdate string
	Gender    string
}

func (r AddMemberRequest) Map() map[string]string {
	return map[string]string{
		"firstname": r.Firstname,
		"lastname":  r.Lastname,
		"birthdate": r.Birthdate,
		"gender":    r.Gender,
	}
}

type AddMemberResponse struct {
	Member *m.Member
	Errors map[string]string
}

type ValidateMemberRequest struct {
	Id string
}

type ValidateMemberResponse struct {
	Member *m.Member
	Errors map[string]string
}

type GetMembersByUserRequest struct {
	UserId string
	ClubId string
}

type GetMembersByUserResponse struct {
	Members []*m.Member
	Errors  map[string]string
}

type UpdateMemberRequest struct {
	Id        string
	Firstname string
	Lastname  string
	Birthdate string
	Gender    string
}

func (r UpdateMemberRequest) Map() map[string]string {
	m := make(map[string]string)
	if r.Firstname != "" {
		m["firstname"] = r.Firstname
	}
	if r.Lastname != "" {
		m["lastname"] = r.Lastname
	}
	if r.Birthdate != "" {
		m["birthdate"] = r.Birthdate
	}
	if r.Gender != "" {
		m["gender"] = r.Gender
	}
	return m
}

type UpdateMemberResponse struct {
	Member *m.Member
	Errors map[string]string
}

type RemoveMemberRequest struct {
	Id string
}
