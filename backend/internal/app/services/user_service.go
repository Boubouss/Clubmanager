package services

import (
	"clubmanager/internal/adapters/api/grpc/dto"
	"clubmanager/internal/domain/users"
	"context"
)

type UserService interface {
  CreateUser(context.Context, *dto.CreateUserRequest) (*dto.CreateUserResponse, error)
  ReadUser(context.Context, *dto.ReadUserRequest) (*dto.ReadUserResponse, error)
  UpdateUser(context.Context, *dto.UpdateUserRequest) (*dto.UpdateUserResponse, error)
  DeleteUser(context.Context, string) (bool, error)
}

type UserServiceConfig struct {
  Repository      users.UserRepository
  Hasher          users.PasswordHasher
  TokenManager    TokenManager
}

type userService struct {
  repo    users.UserRepository
  hasher  users.PasswordHasher
  tkm     TokenManager
}

func NewUserService(config UserServiceConfig) *userService {
  return &userService{
    repo: config.Repository,
    hasher: config.Hasher,
    tkm: config.TokenManager,
  }
}

func (s *userService) CreateUser(ctx context.Context, data *dto.CreateUserRequest) (*dto.CreateUserResponse, error) {
  d := data.Map()

  // Validate data 
  if _, errs := users.NewUser(d); len(errs) > 0 {
    return &dto.CreateUserResponse{
      User: users.User{},
      Token: "",
      Errors: errs,
    }, nil
  }

  // Check if the user already exist
  errs, err := s.repo.IsUserExist(ctx, data.Email, data.Username)

  if err != nil {
    return nil, err
  }

  if len(errs) > 0 {
    return &dto.CreateUserResponse{
      User: users.User{},
      Token: "",
      Errors: errs,
    }, nil
  }
  
  // Encrypt password
  hash, err :=  s.hasher.Hash(data.Password)
  if err != nil {
    return nil, err
  }
  data.Password = hash
  
  // Store data with the repo method
  created, err := s.repo.CreateUser(ctx, d)
  if err != nil {
    return nil, err
  }

  // Create and sign JWT token
  token, err := s.tkm.GenerateToken(created.Id.String())
  if err != nil {
    return nil, err
  }

  return &dto.CreateUserResponse{
    User: *created,
    Token: token,
    Errors: make(map[string]string, 0),
  }, nil
}

func (s *userService) ReadUser(ctx context.Context, data *dto.ReadUserRequest) (*dto.ReadUserResponse, error) {
  // Validate data 
  // if errs := data.Validate(); len(errs) > 0 {
  //   return &users.ReadUserResponse{
  //     Users: make([]users.User, 0),
  //     Errors: errs,
  //   }, nil
  // }

  // Fetch users with the repo method
  list, err := s.repo.ReadUser(ctx)

  if err != nil {
    return nil, err
  }

  return &dto.ReadUserResponse{
    Users: list,
    Errors: make(map[string]string, 0),
  }, nil
}

func (s *userService) UpdateUser(ctx context.Context, data *dto.UpdateUserRequest) (*dto.UpdateUserResponse, error) {
  d := data.Map()

  // Validate data 
  if _, errs := users.NewUser(d); len(errs) > 0 {
    return &dto.UpdateUserResponse{
      User: users.User{},
      Errors: errs,
    }, nil
  }
  
  // Encrypt password if exist
  if data.Password != "" {
    hash, err := s.hasher.Hash(data.Password)
    if err != nil {
      return nil, err
    }
    data.Password = hash
  }
  
  // Update user with repo method
  updated, err := s.repo.UpdateUser(ctx, d)

  if err != nil {
    return nil, err
  }

  return &dto.UpdateUserResponse{
    User: *updated,
    Errors: make(map[string]string, 0),
  }, nil
}

func (s *userService) DeleteUser(ctx context.Context, token string) (bool, error) {
  return false, nil
}
