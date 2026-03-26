package users

import (
	"maps"

	"github.com/google/uuid"
)

type Member struct {
  Id uuid.UUID
  Fristname string
  Lastname string
  Birthdate string
  Gender string
  Club string
}

type User struct {
  Id uuid.UUID
  Username string
  Email string
  Phonenumber string
  Password string
  IsValid bool
  Members []*Member
}

func NewUser(data map[string]string) (*User, map[string]string) {
  errs := make(map[string]string, 4)
  
  if ok, err := IsValidUsername(data["username"]); !ok {
    errs["username"] = err
  }

  if ok, err := IsEmail(data["email"]); !ok {
    errs["email"] = err
  }

  if ok, err := IsPhoneNumber(data["phonenumber"]); !ok {
    errs["phonenumber"] = err
  }

  if ok, err := IsValidPassword(data["password"]); !ok {
    errs["password"] = err
  }

  return &User{
    Username: data["username"],
    Email: data["email"],
    Phonenumber: data["phonenumber"],
    Password: data["password"],
  }, errs
}

func (u User) Update(data map[string]string) (*User, map[string]string) {
  user := make(map[string]string, 4)

  maps.Copy(user, data)

  if _, ok := user["username"]; !ok {
    user["username"] = u.Username
  }
  
  if _, ok := user["email"]; !ok {
    user["email"] = u.Email
  }

  if _, ok := user["phonenumber"]; !ok {
    user["phonenumber"] = u.Phonenumber
  } 

  if _, ok := user["password"]; !ok {
    user["password"] = u.Password
  }

  return NewUser(user)
}
