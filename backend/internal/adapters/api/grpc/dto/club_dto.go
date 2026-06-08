package dto

import c "clubmanager/internal/domain/clubs"

type CreateClubRequest struct {
	Siren       string
	Name        string
	Address     string
	City        string
	PostalCode  string
	Country     string
	Phonenumber string
}

func (r CreateClubRequest) Map() map[string]string {
	return map[string]string{
		"siren":       r.Siren,
		"name":        r.Name,
		"address":     r.Address,
		"city":        r.City,
		"postal_code": r.PostalCode,
		"country":     r.Country,
		"phonenumber": r.Phonenumber,
	}
}

type CreateClubResponse struct {
	Club   *c.Club
	Errors map[string]string
}

type ReadClubRequest struct {
	Params map[string]any
}

type ReadClubResponse struct {
	Clubs  []*c.Club
	Errors map[string]string
}

type UpdateClubRequest struct {
	Id          string
	Name        string
	Address     string
	City        string
	PostalCode  string
	Country     string
	Phonenumber string
}

func (r UpdateClubRequest) Map() map[string]string {
	m := make(map[string]string)
	if r.Name != "" {
		m["name"] = r.Name
	}
	if r.Address != "" {
		m["address"] = r.Address
	}
	if r.City != "" {
		m["city"] = r.City
	}
	if r.PostalCode != "" {
		m["postal_code"] = r.PostalCode
	}
	if r.Country != "" {
		m["country"] = r.Country
	}
	if r.Phonenumber != "" {
		m["phonenumber"] = r.Phonenumber
	}
	return m
}

type UpdateClubResponse struct {
	Club   *c.Club
	Errors map[string]string
}

type DeleteClubRequest struct {
	Id string
}

type DeleteClubResponse struct {
	Ok bool
}
