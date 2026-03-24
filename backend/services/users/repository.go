package users

import "context"

type UserRepository interface {
  CreateUser(context.Context, *CreateUserRequest) (*User, error)
  IsUserExist(context.Context, string, string) (map[string]string, error)
  ReadUser(context.Context, *ReadUserRequest) ([]User, error)
  UpdateUser(context.Context, *UpdateUserRequest) (*User, error)
  DeleteUser(context.Context, string) (bool, error)
}
