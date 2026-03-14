package models

import (
	"clubmanager/validation"

	"github.com/google/uuid"
)

type UserCreated struct {
  Id uuid.UUID
  Username string
}

type CreateUserResponse struct {
  Id uuid.UUID
  Username string
  Token string
  Errors map[string]string
}

type CreateUserRequest struct {
  Username string
  Email string
  Phonenumber string
  Password string
}

func (u CreateUserRequest) Validate() map[string]string {
  errs := make(map[string]string, 4)
  
  if ok, err := validation.IsValidUsername(u.Username); !ok {
    errs["username"] = err
  }
  
  if ok, err := validation.IsEmail(u.Email); !ok {
    errs["email"] = err
  }
  
  if ok, err := validation.IsPhoneNumber(u.Phonenumber); !ok {
    errs["phonenumber"] = err
  }
  
  if ok, err := validation.IsValidPassword(u.Password); !ok {
    errs["password"] = err
  }

  return errs
}

