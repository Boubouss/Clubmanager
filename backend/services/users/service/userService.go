package service

import (
	"clubmanager/services/users/models"
	"clubmanager/services/users/repositories"

	"context"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

type UserService interface {
  CreateUser(context.Context, *models.CreateUserRequest) (*models.CreateUserResponse, error)
  ReadUser(context.Context, *models.ReadUserRequest) (*models.ReadUserResponse, error)
  UpdateUser(context.Context, *models.UpdateUserRequest) (*models.UpdateUserResponse, error)
  DeleteUser(context.Context, string) (bool, error)
}

type userService struct {
  repo repositories.UserRepository
}

func NewUserService(db *pgx.Conn) *userService {
  return &userService{
    repo: repositories.NewUserRepository(db),
  }
}

func (s *userService) CreateUser(ctx context.Context, data *models.CreateUserRequest) (*models.CreateUserResponse, error) {
  // Validate data 
  if errs := data.Validate(); len(errs) > 0 {
    return &models.CreateUserResponse{
      User: models.User{},
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
    return &models.CreateUserResponse{
      User: models.User{},
      Token: "",
      Errors: errs,
    }, nil
  }
  
  // Encrypt password
  hash, err := bcrypt.GenerateFromPassword([]byte(data.Password), bcrypt.DefaultCost)
  if err != nil {
    return nil, err
  }
  data.Password = string(hash)
  
  // Store data with the repo method
  created, err := s.repo.CreateUser(ctx, data)
  if err != nil {
    return nil, err
  }

  // Create and sign JWT token
  token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
    "userId": created.Id.String(),
    "ttl": time.Now().Add(time.Hour * 24 * 30).Unix(),
  })

  tokenStr, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))

  if err != nil {
    return nil, err
  }

  return &models.CreateUserResponse{
    User: *created,
    Token: tokenStr,
    Errors: make(map[string]string, 0),
  }, nil
}

func (s *userService) ReadUser(ctx context.Context, data *models.ReadUserRequest) (*models.ReadUserResponse, error) {
  // Validate data 
  if errs := data.Validate(); len(errs) > 0 {
    return &models.ReadUserResponse{
      Users: make([]models.User, 0),
      Errors: errs,
    }, nil
  }

  // Fetch users with the repo method
  users, err := s.repo.ReadUser(ctx, data)

  if err != nil {
    return nil, err
  }

  return &models.ReadUserResponse{
    Users: users,
    Errors: make(map[string]string, 0),
  }, nil
}

func (s *userService) UpdateUser(ctx context.Context, data *models.UpdateUserRequest) (*models.UpdateUserResponse, error) {
  // Validate data 
  if errs := data.Validate(); len(errs) > 0 {
    return &models.UpdateUserResponse{
      User: models.User{},
      Errors: errs,
    }, nil
  }

  // Update user with repo method
  updated, err := s.repo.UpdateUser(ctx, data)

  if err != nil {
    return nil, err
  }

  return &models.UpdateUserResponse{
    User: *updated,
    Errors: make(map[string]string, 0),
  }, nil
}

func (s *userService) DeleteUser(ctx context.Context, token string) (bool, error) {
  return false, nil
}
