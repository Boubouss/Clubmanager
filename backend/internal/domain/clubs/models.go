package clubs

import (
	"maps"

	"github.com/google/uuid"
)

const (
	StatusPending   = "pending"
	StatusActive    = "active"
	StatusSuspended = "suspended"
)

type Club struct {
	Id          uuid.UUID
	Siren       string
	Name        string
	Address     string
	City        string
	PostalCode  string
	Country     string
	Phonenumber string
	Status      string
	CreatorId   uuid.UUID
}

func NewClub(data map[string]string) (*Club, map[string]string) {
	errs := make(map[string]string)

	if ok, err := IsValidSiren(data["siren"]); !ok {
		errs["siren"] = err
	}
	if ok, err := IsNotBlank(data["name"], "Name"); !ok {
		errs["name"] = err
	}
	if ok, err := IsNotBlank(data["address"], "Address"); !ok {
		errs["address"] = err
	}
	if ok, err := IsNotBlank(data["city"], "City"); !ok {
		errs["city"] = err
	}
	if ok, err := IsValidPostalCode(data["postal_code"]); !ok {
		errs["postal_code"] = err
	}
	if ok, err := IsValidPhone(data["phonenumber"]); !ok {
		errs["phonenumber"] = err
	}

	country := data["country"]
	if country == "" {
		country = "FRANCE"
	}

	return &Club{
		Siren:       data["siren"],
		Name:        data["name"],
		Address:     data["address"],
		City:        data["city"],
		PostalCode:  data["postal_code"],
		Country:     country,
		Phonenumber: data["phonenumber"],
		Status:      StatusPending,
	}, errs
}

func (c Club) Update(data map[string]string) (*Club, map[string]string) {
	club := make(map[string]string)
	maps.Copy(club, data)

	if _, ok := club["name"]; !ok {
		club["name"] = c.Name
	}
	if _, ok := club["address"]; !ok {
		club["address"] = c.Address
	}
	if _, ok := club["city"]; !ok {
		club["city"] = c.City
	}
	if _, ok := club["postal_code"]; !ok {
		club["postal_code"] = c.PostalCode
	}
	if _, ok := club["country"]; !ok {
		club["country"] = c.Country
	}
	if _, ok := club["phonenumber"]; !ok {
		club["phonenumber"] = c.Phonenumber
	}
	// siren is immutable
	club["siren"] = c.Siren

	updated, errs := NewClub(club)
	if len(errs) == 0 {
		updated.Id = c.Id
		updated.Status = c.Status // status is not changed via Update
	}
	return updated, errs
}
