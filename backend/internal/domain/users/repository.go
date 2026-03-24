package users

import "context"

type UserRepository interface {
  CreateUser(context.Context, map[string]string) (*User, error)
  IsUserExist(context.Context, string, string) (map[string]string, error)
  ReadUser(context.Context) ([]User, error)
  UpdateUser(context.Context, map[string]string) (*User, error)
  DeleteUser(context.Context, string) (bool, error)
}
