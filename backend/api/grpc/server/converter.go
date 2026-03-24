package server

import (
	"clubmanager/services/users"
	"clubmanager/api/grpc/proto"
)

func MemberProto(m *users.Member) (*proto.Member) {
  return &proto.Member{
    Id: m.Id.String(),
    Firstname: m.Fristname,
    Lastname: m.Lastname,
    Birthdate: m.Birthdate,
    Gender: m.Gender,
    Club: m.Club,
  }
}

func ArrayMemberProto(arr []users.Member) ([]*proto.Member) {
  var members []*proto.Member

  for _, m := range arr {
    members = append(members, MemberProto(&m))
  }

  return members
}

func UserProto(u *users.User) (*proto.User) {
  return &proto.User{
    Id: u.Id.String(),
    Username: u.Username,
    Email: u.Email,
    Phonenumber: u.Phonenumber,
    Members: ArrayMemberProto(u.Members),
  }
}

func ArrayUserProto(arr []users.User) ([]*proto.User) {
  var users []*proto.User

  for _, u := range arr {
    users = append(users, UserProto(&u))
  }

  return users
}


