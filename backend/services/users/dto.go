package users

import "clubmanager/utils/validation"

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

