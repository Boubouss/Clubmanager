package logging

import (
	"clubmanager/services/users/models"
	"clubmanager/services/users/service"
	"context"
	"fmt"
	"time"
)

type userLoggingService struct {
  next service.UserService
} 

func NewUserLoggingService(next service.UserService) service.UserService {
  return &userLoggingService{
    next: next,
  }
}

func (s *userLoggingService) CreateUser(ctx context.Context, data *models.CreateUserRequest) (user *models.CreateUserResponse, err error) {
  defer func(begin time.Time){
    fmt.Printf("=> type: '%s'; took: '%v'; err: '%v'.\n", "CreateUser", time.Since(begin), err)
  }(time.Now())

  return s.next.CreateUser(ctx, data)
}
