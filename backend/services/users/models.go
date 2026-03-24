package users

import "github.com/google/uuid"

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
  Members []Member
}


