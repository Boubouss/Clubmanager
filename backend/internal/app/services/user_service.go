package services

import (
	"clubmanager/internal/adapters/api/grpc/dto"
	"clubmanager/internal/domain"
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
  Repository      domain.Repository[users.User, string]
  Hasher          users.PasswordHasher
  TokenManager    TokenManager
}

type userService struct {
  repo    domain.Repository[users.User, string]
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
  u, errs := users.NewUser(data.Map())

  // Validate data 
  if len(errs) > 0 {
    return &dto.CreateUserResponse{
      User: nil,
      Token: "",
      Errors: errs,
    }, nil
  }

  // Check if the user already exist
  list, err := s.repo.Search(ctx, &domain.SearchParams{
    Fields: map[string]any{"email": u.Email, "username": u.Username},
    Connector: "OR",
  })

  if err != nil {
    return nil, err
  }

  if len(list) > 0 {
    errs := make(map[string]string, 2)
    for _, v := range list {
      if u.Email == v.Email {
        errs["email"] = "Email already exist."
      }

      if u.Username == v.Username {
        errs["username"] = "Username already exist."
      }
    }

    return &dto.CreateUserResponse{
      User: nil,
      Token: "",
      Errors: errs,
    }, nil
  }
  
  // Encrypt password
  hash, err :=  s.hasher.Hash(u.Password)
  if err != nil {
    return nil, err
  }
  u.Password = hash

  // Store data with the repo method
  created, err := s.repo.Save(ctx, u)
  if err != nil {
    return nil, err
  }

  // Create and sign JWT token
  token, err := s.tkm.GenerateToken(created.Id.String())
  if err != nil {
    return nil, err
  }

  return &dto.CreateUserResponse{
    User: created,
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
  list, err := s.repo.Search(ctx, &domain.SearchParams{
    Fields: data.Params,
    Connector: "AND",
  })

  if err != nil {
    return nil, err
  }

  return &dto.ReadUserResponse{
    Users: list,
    Errors: make(map[string]string, 0),
  }, nil
}

func (s *userService) UpdateUser(ctx context.Context, data *dto.UpdateUserRequest) (*dto.UpdateUserResponse, error) {
  current, err := s.repo.Find(ctx, data.Id)
  d := data.Map()

  if err != nil {
    return nil, err
  }

  u, errs := current.Update(d)
  
  // Validate data 
  if len(errs) > 0 {
    return &dto.UpdateUserResponse{
      User: nil,
      Errors: errs,
    }, nil
  }
  
  // Encrypt password if exist
  if _, ok := d["password"]; ok {
    hash, err := s.hasher.Hash(u.Password)
    if err != nil {
      return nil, err
    }
    u.Password = hash
  }
  
  // Update user with repo method
  updated, err := s.repo.Save(ctx, u)

  if err != nil {
    return nil, err
  }

  return &dto.UpdateUserResponse{
    User: updated,
    Errors: make(map[string]string, 0),
  }, nil
}

func (s *userService) DeleteUser(ctx context.Context, token string) (bool, error) {
  return false, nil
}
