package logging

import (
	u "clubmanager/services/users"
	"context"
	"fmt"
	"time"
)

type userLoggingService struct {
  next u.UserService
} 

func NewUserLoggingService(next u.UserService) u.UserService {
  return &userLoggingService{
    next: next,
  }
}

func (s *userLoggingService) CreateUser(ctx context.Context, data *u.CreateUserRequest) (user *u.CreateUserResponse, err error) {
  defer func(begin time.Time){
    fmt.Printf("=> type: '%s'; took: '%v'; err: '%v'.\n", "CreateUser", time.Since(begin), err)
  }(time.Now())

  return s.next.CreateUser(ctx, data)
}

func (s *userLoggingService) ReadUser(ctx context.Context, data *u.ReadUserRequest) (user *u.ReadUserResponse, err error) {
  defer func(begin time.Time){
    fmt.Printf("=> type: '%s'; took: '%v'; err: '%v'.\n", "ReadUser", time.Since(begin), err)
  }(time.Now())

  return s.next.ReadUser(ctx, data)
}

func (s *userLoggingService) UpdateUser(ctx context.Context, data *u.UpdateUserRequest) (user *u.UpdateUserResponse, err error) {
  defer func(begin time.Time){
    fmt.Printf("=> type: '%s'; took: '%v'; err: '%v'.\n", "UpdateUser", time.Since(begin), err)
  }(time.Now())

  return s.next.UpdateUser(ctx, data)
}

func (s *userLoggingService) DeleteUser(ctx context.Context, token string) (res bool, err error) {
  defer func(begin time.Time){
    fmt.Printf("=> type: '%s'; took: '%v'; err: '%v'.\n", "DeleteUser", time.Since(begin), err)
  }(time.Now())

  return s.next.DeleteUser(ctx, token)
}
