package users

import (
	"regexp"
	"strconv"
)

var (
  emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
  phoneRegex = regexp.MustCompile(`^(\+33|0)[1-9](\d{2}){4}$`)
  usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_]{3,20}$`)
  hasUpper     = regexp.MustCompile(`[A-Z]`).MatchString         
	hasLower     = regexp.MustCompile(`[a-z]`).MatchString         
	hasDigit     = regexp.MustCompile(`[0-9]`).MatchString         
	hasSpecial   = regexp.MustCompile(`[^a-zA-Z0-9]`).MatchString
) 


func IsEmail(email string) (bool, string) {
  if !emailRegex.MatchString(email) {
    return false, "Invalid email."
  }
  return true, ""
}

func IsLengthBetween(s string, mini, maxi int) (bool, string) {
  
  if len(s) < mini {
    return false, "Length has to be more than " + strconv.Itoa(mini) + " charateres."
  }

  if len(s) > maxi {
    return false, "Length has to be less than " + strconv.Itoa(maxi) + " charateres."
  }

  return true, ""
}

func IsPhoneNumber(phone string) (bool, string) {
  if !phoneRegex.MatchString(phone) {
    return false, "Invalid phonenumber."
  }
  return true, ""
}

func IsValidPassword(pswd string) (bool, string) {
  
  if ok, err := IsLengthBetween(pswd, 8, 20); !ok {
    return false, err
  }
  
  if hasDigit(pswd) && hasLower(pswd) && hasUpper(pswd) && hasSpecial(pswd) {
    return true, ""
  }

  return false, "Password required 1 lowercase, 1 uppercase, 1 number and 1 special."

}

func IsValidUsername(username string) (bool, string) {
  if ok, err := IsLengthBetween(username, 5, 20); !ok {
    return false, err
  }


  if !usernameRegex.MatchString(username) {
    return false, "Invalid username."
  }

  return true, ""
}
