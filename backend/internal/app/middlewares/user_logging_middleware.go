package middlewares

import (
	"clubmanager/internal/adapters/api/grpc/dto"
	"clubmanager/internal/app/services"
	"context"
	"fmt"
	"time"
)

type userLoggingService struct {
  next services.UserService
} 

func NewUserLoggingService(next services.UserService) services.UserService {
  return &userLoggingService{
    next: next,
  }
}

func (s *userLoggingService) CreateUser(ctx context.Context, data *dto.CreateUserRequest) (user *dto.CreateUserResponse, err error) {
  defer func(begin time.Time){
    fmt.Printf("=> type: '%s'; took: '%v'; err: '%v'.\n", "CreateUser", time.Since(begin), err)
  }(time.Now())

  return s.next.CreateUser(ctx, data)
}

func (s *userLoggingService) ReadUser(ctx context.Context, data *dto.ReadUserRequest) (user *dto.ReadUserResponse, err error) {
  defer func(begin time.Time){
    fmt.Printf("=> type: '%s'; took: '%v'; err: '%v'.\n", "ReadUser", time.Since(begin), err)
  }(time.Now())

  return s.next.ReadUser(ctx, data)
}

func (s *userLoggingService) UpdateUser(ctx context.Context, data *dto.UpdateUserRequest) (user *dto.UpdateUserResponse, err error) {
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
