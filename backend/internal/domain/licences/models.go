package licences

import "github.com/google/uuid"

type Licence struct {
	Id            uuid.UUID
	MemberId      uuid.UUID
	LicenceNumber string
	ValidFrom     string
	ValidUntil    string
	Status        string
}

func NewLicence(memberId string, data map[string]string) (*Licence, map[string]string) {
	errs := make(map[string]string)

	if ok, err := IsNotBlank(data["licence_number"], "Licence number"); !ok {
		errs["licence_number"] = err
	}
	if ok, err := IsValidDateRange(data["valid_from"], data["valid_until"]); !ok {
		errs["dates"] = err
	}

	memberUUID, err := uuid.Parse(memberId)
	if err != nil {
		errs["member_id"] = "Invalid member ID."
	}

	return &Licence{
		MemberId:      memberUUID,
		LicenceNumber: data["licence_number"],
		ValidFrom:     data["valid_from"],
		ValidUntil:    data["valid_until"],
		Status:        "pending",
	}, errs
}

func (l Licence) UpdateStatus(status string) (*Licence, map[string]string) {
	errs := make(map[string]string)
	if ok, err := IsValidStatus(status); !ok {
		errs["status"] = err
		return nil, errs
	}
	updated := l
	updated.Status = status
	return &updated, errs
}
