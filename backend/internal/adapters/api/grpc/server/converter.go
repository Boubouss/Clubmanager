package server

import (
	"clubmanager/internal/domain/users"
	"clubmanager/internal/adapters/api/grpc/proto"
)

func memberProto(m *users.Member) (*proto.Member) {
  return &proto.Member{
    Id: m.Id.String(),
    Firstname: m.Fristname,
    Lastname: m.Lastname,
    Birthdate: m.Birthdate,
    Gender: m.Gender,
    Club: m.Club,
  }
}

func arrayMemberProto(arr []*users.Member) ([]*proto.Member) {
  var members []*proto.Member

  for _, m := range arr {
    members = append(members, memberProto(m))
  }

  return members
}

func userProto(u *users.User) (*proto.User) {
  return &proto.User{
    Id: u.Id.String(),
    Username: u.Username,
    Email: u.Email,
    Phonenumber: u.Phonenumber,
    Members: arrayMemberProto(u.Members),
  }
}

func arrayUserProto(arr []*users.User) ([]*proto.User) {
  var users []*proto.User

  for _, u := range arr {
    users = append(users, userProto(u))
  }

  return users
}
