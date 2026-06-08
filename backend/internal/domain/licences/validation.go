package licences

import "time"

func IsNotBlank(field, label string) (bool, string) {
	if len(field) == 0 {
		return false, label + " is required."
	}
	return true, ""
}

func IsValidDateRange(from, until string) (bool, string) {
	f, err := time.Parse("2006-01-02", from)
	if err != nil {
		return false, "valid_from must be in YYYY-MM-DD format."
	}
	u, err := time.Parse("2006-01-02", until)
	if err != nil {
		return false, "valid_until must be in YYYY-MM-DD format."
	}
	if !u.After(f) {
		return false, "valid_until must be after valid_from."
	}
	return true, ""
}

func IsValidStatus(status string) (bool, string) {
	switch status {
	case "active", "pending", "expired":
		return true, ""
	}
	return false, "Status must be 'active', 'pending' or 'expired'."
}
