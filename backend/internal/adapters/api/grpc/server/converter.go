package server

import (
	"clubmanager/internal/adapters/api/grpc/proto"
	"clubmanager/internal/domain/clubs"
	"clubmanager/internal/domain/events"
	"clubmanager/internal/domain/licences"
	"clubmanager/internal/domain/members"
	"clubmanager/internal/domain/roles"
	"clubmanager/internal/domain/users"
)

func userProto(u *users.User) *proto.User {
	return &proto.User{
		Id:          u.Id.String(),
		Username:    u.Username,
		Email:       u.Email,
		Phonenumber: u.Phonenumber,
	}
}

func memberProto(m *members.Member) *proto.Member {
	return &proto.Member{
		Id:        m.Id.String(),
		UserId:    m.UserId.String(),
		ClubId:    m.ClubId.String(),
		Firstname: m.Firstname,
		Lastname:  m.Lastname,
		Birthdate: m.Birthdate,
		Gender:    m.Gender,
		IsValid:   m.IsValid,
	}
}

func arrayMemberProto(arr []*members.Member) []*proto.Member {
	var result []*proto.Member
	for _, m := range arr {
		result = append(result, memberProto(m))
	}
	return result
}

func licenceProto(l *licences.Licence) *proto.Licence {
	return &proto.Licence{
		Id:            l.Id.String(),
		MemberId:      l.MemberId.String(),
		LicenceNumber: l.LicenceNumber,
		ValidFrom:     l.ValidFrom,
		ValidUntil:    l.ValidUntil,
		Status:        l.Status,
	}
}

func arrayLicenceProto(arr []*licences.Licence) []*proto.Licence {
	var result []*proto.Licence
	for _, l := range arr {
		result = append(result, licenceProto(l))
	}
	return result
}

func arrayUserProto(arr []*users.User) []*proto.User {
	var result []*proto.User
	for _, u := range arr {
		result = append(result, userProto(u))
	}
	return result
}

func clubProto(c *clubs.Club) *proto.Club {
	return &proto.Club{
		Id:          c.Id.String(),
		Siren:       c.Siren,
		Name:        c.Name,
		Address:     c.Address,
		City:        c.City,
		PostalCode:  c.PostalCode,
		Country:     c.Country,
		Phonenumber: c.Phonenumber,
	}
}

func arrayClubProto(arr []*clubs.Club) []*proto.Club {
	var result []*proto.Club
	for _, c := range arr {
		result = append(result, clubProto(c))
	}
	return result
}

func roleProto(r *roles.Role) *proto.Role {
	return &proto.Role{
		Id:     r.Id.String(),
		UserId: r.UserId.String(),
		ClubId: r.ClubId.String(),
		Role:   r.Role,
	}
}

func arrayRoleProto(arr []*roles.Role) []*proto.Role {
	var result []*proto.Role
	for _, r := range arr {
		result = append(result, roleProto(r))
	}
	return result
}

func eventProto(e *events.Event) *proto.Event {
	if e == nil {
		return nil
	}
	return &proto.Event{
		Id:                  e.Id.String(),
		ClubId:              e.ClubId.String(),
		CreatedBy:           e.CreatedBy.String(),
		Title:               e.Title,
		Description:         e.Description,
		Type:                e.Type,
		Status:              e.Status,
		Location:            e.Location,
		RegistrationOpenAt:  e.RegistrationOpenAt,
		RegistrationCloseAt: e.RegistrationCloseAt,
		Date:                e.Date,
		MaxParticipants:     int32(e.MaxParticipants),
	}
}

func arrayEventProto(arr []*events.Event) []*proto.Event {
	var result []*proto.Event
	for _, e := range arr {
		result = append(result, eventProto(e))
	}
	return result
}

func eventCategoryProto(c *events.EventCategory) *proto.EventCategory {
	if c == nil {
		return nil
	}
	return &proto.EventCategory{
		Id:             c.Id.String(),
		EventId:        c.EventId.String(),
		JudoCategoryId: c.JudoCategoryId.String(),
		WeighInAt:      c.WeighInAt,
		StartsAt:       c.StartsAt,
		Status:         c.Status,
	}
}

func arrayEventCategoryProto(arr []*events.EventCategory) []*proto.EventCategory {
	var result []*proto.EventCategory
	for _, c := range arr {
		result = append(result, eventCategoryProto(c))
	}
	return result
}

func participantProto(p *events.EventParticipant) *proto.EventParticipant {
	if p == nil {
		return nil
	}
	catId := ""
	if p.EventCategoryId != nil {
		catId = p.EventCategoryId.String()
	}
	return &proto.EventParticipant{
		Id:              p.Id.String(),
		EventId:         p.EventId.String(),
		MemberId:        p.MemberId.String(),
		Role:            p.Role,
		EventCategoryId: catId,
		RegisteredAt:    p.RegisteredAt,
	}
}

func arrayParticipantProto(arr []*events.EventParticipant) []*proto.EventParticipant {
	var result []*proto.EventParticipant
	for _, p := range arr {
		result = append(result, participantProto(p))
	}
	return result
}

func carpoolOfferProto(o *events.CarpoolOffer) *proto.CarpoolOffer {
	if o == nil {
		return nil
	}
	return &proto.CarpoolOffer{
		Id:               o.Id.String(),
		EventId:          o.EventId.String(),
		MemberId:         o.MemberId.String(),
		DepartureAddress: o.DepartureAddress,
		AvailableSeats:   int32(o.AvailableSeats),
		DepartureAt:      o.DepartureAt,
	}
}

func arrayCarpoolOfferProto(arr []*events.CarpoolOffer) []*proto.CarpoolOffer {
	var result []*proto.CarpoolOffer
	for _, o := range arr {
		result = append(result, carpoolOfferProto(o))
	}
	return result
}

func judoCategoryProto(c *events.JudoCategory) *proto.JudoCategory {
	if c == nil {
		return nil
	}
	p := &proto.JudoCategory{
		Id:          c.Id.String(),
		AgeGroup:    c.AgeGroup,
		Gender:      c.Gender,
		AgeMin:      int32(c.AgeMin),
		WeightLabel: c.WeightLabel,
	}
	if c.AgeMax != nil {
		p.AgeMax = int32(*c.AgeMax)
	}
	if c.WeightMax != nil {
		p.WeightMax = *c.WeightMax
	}
	return p
}

func arrayJudoCategoryProto(arr []*events.JudoCategory) []*proto.JudoCategory {
	var result []*proto.JudoCategory
	for _, c := range arr {
		result = append(result, judoCategoryProto(c))
	}
	return result
}
