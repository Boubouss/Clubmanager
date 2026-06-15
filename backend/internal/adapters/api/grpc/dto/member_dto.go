package dto

import (
	c "clubmanager/internal/domain/clubs"
	m "clubmanager/internal/domain/members"
)

type AddMemberRequest struct {
	UserId    string
	ClubId    string
	Firstname string
	Lastname  string
	Birthdate string
	Gender    string
	IsPrimary bool
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
	Member     *m.Member
	Membership *m.ClubMembership
	Errors     map[string]string
}

// ValidateMemberRequest validates a ClubMembership by its ID.
type ValidateMemberRequest struct {
	Id string // club_membership ID
}

type ValidateMemberResponse struct {
	Member     *m.Member
	Membership *m.ClubMembership
	Errors     map[string]string
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

type CreateMemberRequest struct {
	UserId    string
	Firstname string
	Lastname  string
	Birthdate string
	Gender    string
	IsPrimary bool
}

func (r CreateMemberRequest) Map() map[string]string {
	return map[string]string{
		"firstname": r.Firstname,
		"lastname":  r.Lastname,
		"birthdate": r.Birthdate,
		"gender":    r.Gender,
	}
}

type CreateMemberResponse struct {
	Member *m.Member
	Errors map[string]string
}

type RequestMembershipRequest struct {
	MemberId string
	ClubId   string
	UserId   string // pour vérifier que le membre appartient bien à cet utilisateur
}

type RequestMembershipResponse struct {
	Membership *m.ClubMembership
	Errors     map[string]string
}

type GetMembersByClubRequest struct {
	ClubId   string
	Page     int
	PageSize int
}

type MemberWithMembership struct {
	Member     *m.Member
	Membership *m.ClubMembership
	Email      string
	Phone      string
}

type GetMembersByClubResponse struct {
	Members []MemberWithMembership
	Total   int
	Errors  map[string]string
}

// StaffContact holds contact info for a club staff/coach member.
type StaffContact struct {
	Firstname string
	Lastname  string
	Role      string
	Email     string
	Phone     string
}

type GetStaffContactsResponse struct {
	Contacts []StaffContact
}

type GetUserClubsRequest struct {
	UserId string
}

type GetUserClubsResponse struct {
	Clubs  []*c.Club
	Errors map[string]string
}
