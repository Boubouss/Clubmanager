package dto

import l "clubmanager/internal/domain/licences"

type CreateLicenceRequest struct {
	MemberId      string
	LicenceNumber string
	ValidFrom     string
	ValidUntil    string
}

func (r CreateLicenceRequest) Map() map[string]string {
	return map[string]string{
		"licence_number": r.LicenceNumber,
		"valid_from":     r.ValidFrom,
		"valid_until":    r.ValidUntil,
	}
}

type CreateLicenceResponse struct {
	Licence *l.Licence
	Errors  map[string]string
}

type GetMemberLicencesRequest struct {
	MemberId string
}

type GetMemberLicencesResponse struct {
	Licences []*l.Licence
	Errors   map[string]string
}

type UpdateLicenceStatusRequest struct {
	Id     string
	Status string
}

type UpdateLicenceStatusResponse struct {
	Licence *l.Licence
	Errors  map[string]string
}
