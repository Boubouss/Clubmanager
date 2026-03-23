package models

import (
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


