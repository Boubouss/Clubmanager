package clubs

import "regexp"

var (
	sirenRegex      = regexp.MustCompile(`^[0-9]{9}$`).MatchString
	postalCodeRegex = regexp.MustCompile(`^[0-9]{5}$`).MatchString
	phoneRegex      = regexp.MustCompile(`^0[1-9]\d{8}$`).MatchString
)

func IsValidSiren(siren string) (bool, string) {
	if !sirenRegex(siren) {
		return false, "SIREN must be exactly 9 digits."
	}
	return true, ""
}

func IsValidPostalCode(code string) (bool, string) {
	if !postalCodeRegex(code) {
		return false, "Postal code must be exactly 5 digits."
	}
	return true, ""
}

func IsNotBlank(field, label string) (bool, string) {
	if len(field) == 0 {
		return false, label + " is required."
	}
	return true, ""
}

func IsValidPhone(phone string) (bool, string) {
	if phone == "" {
		return true, ""
	}
	if !phoneRegex(phone) {
		return false, "Invalid phonenumber."
	}
	return true, ""
}
