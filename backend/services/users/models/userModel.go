package models

import (
	"clubmanager/validation"
	"clubmanager/api/grpc/proto"

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
  Members []Member
}

func (m Member) Proto() (*proto.Member) {
  return &proto.Member{
    Id: m.Id.String(),
    Firstname: m.Fristname,
    Lastname: m.Lastname,
    Birthdate: m.Birthdate,
    Gender: m.Gender,
    Club: m.Club,
  }
}

func ArrayMemberProto(arr []Member) ([]*proto.Member) {
  var members []*proto.Member

  for _, m := range arr {
    members = append(members, m.Proto())
  }

  return members
}

func (u User) Proto() (*proto.User) {
  return &proto.User{
    Id: u.Id.String(),
    Username: u.Username,
    Email: u.Email,
    Phonenumber: u.Phonenumber,
    Members: ArrayMemberProto(u.Members),
  }
}

func ArrayUserProto(arr []User) ([]*proto.User) {
  var users []*proto.User

  for _, u := range arr {
    users = append(users, u.Proto())
  }

  return users
}

type CreateUserRequest struct {
  Username string
  Email string
  Phonenumber string
  Password string
}

type CreateUserResponse struct {
  User User
  Token string
  Errors map[string]string
}

type ReadUserRequest struct {
  Params map[string]string
}

type ReadUserResponse struct {
  Users []User
  Errors map[string]string
}

type UpdateUserRequest struct {
  Id string
  Email string
  Phonenumber string
  Password string
}

type UpdateUserResponse struct {
  User User
  Errors map[string]string
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

func (u ReadUserRequest) Validate() map[string]string {
  errs := make(map[string]string)

  return errs
}

func (u UpdateUserRequest) Validate() map[string]string {
  errs := make(map[string]string)
  
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

