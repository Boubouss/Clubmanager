package service

import (
	"clubmanager/services/users/models"
	"clubmanager/services/users/repositories"
	"fmt"

	"context"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

type UserService interface {
  CreateUser(context.Context, *models.CreateUserRequest) (*models.CreateUserResponse, error)
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
      Id: uuid.UUID{},
      Username: "",
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
      Id: uuid.UUID{},
      Username: "",
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
    Id: created.Id,
    Username: created.Username,
    Token: tokenStr,
    Errors: make(map[string]string, 0),
  }, nil
}
