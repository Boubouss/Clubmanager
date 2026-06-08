package clubs

import (
	"maps"

	"github.com/google/uuid"
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
	// siren n'est pas modifiable, on le force depuis le club existant
	club["siren"] = c.Siren

	return NewClub(club)
}
