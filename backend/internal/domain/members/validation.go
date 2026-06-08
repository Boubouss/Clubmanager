package members

import (
	"regexp"
	"time"
)

var genderRegex = regexp.MustCompile(`^(man|woman)$`).MatchString

func IsNotBlank(field, label string) (bool, string) {
	if len(field) == 0 {
		return false, label + " is required."
	}
	return true, ""
}

func IsValidGender(gender string) (bool, string) {
	if !genderRegex(gender) {
		return false, "Gender must be 'man' or 'woman'."
	}
	return true, ""
}

func IsValidBirthdate(birthdate string) (bool, string) {
	t, err := time.Parse("2006-01-02", birthdate)
	if err != nil {
		return false, "Birthdate must be in YYYY-MM-DD format."
	}
	now := time.Now()
	if t.After(now) {
		return false, "Birthdate cannot be in the future."
	}
	if t.Before(now.AddDate(-120, 0, 0)) {
		return false, "Birthdate is too far in the past."
	}
	return true, ""
}
