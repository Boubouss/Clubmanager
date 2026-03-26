package dto

import u "clubmanager/internal/domain/users"


type CreateUserRequest struct {
  Username string
  Email string
  Phonenumber string
  Password string
}

type CreateUserResponse struct {
  User *u.User
  Token string
  Errors map[string]string
}

func (u CreateUserRequest) Map() map[string]string {
  m := make(map[string]string, 4)

  if u.Username != "" {
    m["username"] = u.Username
  }
  
  if u.Email != "" {
    m["email"] = u.Email
  }

  if u.Phonenumber != "" {
    m["phonumber"] = u.Phonenumber
  }

  if u.Password != "" {
    m["password"] = u.Password
  }

  return m
}

type ReadUserRequest struct {
  Params map[string]any
}

type ReadUserResponse struct {
  Users  []*u.User
  Errors map[string]string
}

type UpdateUserRequest struct {
  Id string
  Email string
  Phonenumber string
  Password string
}

type UpdateUserResponse struct {
  User *u.User
  Errors map[string]string
}

func (u UpdateUserRequest) Map() map[string]string {
  m := make(map[string]string, 4)

  if u.Id != "" {
    m["id"] = u.Id
  }
  
  if u.Email != "" {
    m["email"] = u.Email
  }

  if u.Phonenumber != "" {
    m["phonumber"] = u.Phonenumber
  }

  if u.Password != "" {
    m["password"] = u.Password
  }

  return m
}

