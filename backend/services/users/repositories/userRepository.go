package repositories

import (
	"clubmanager/services/users/models"
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"google.golang.org/grpc/metadata"
)

type UserRepository interface {
  CreateUser(context.Context, *models.CreateUserRequest) (*models.UserCreated, error)
  IsUserExist(context.Context, string, string) (map[string]string, error)
}

type userRepository struct {
  db *pgx.Conn
}

func NewUserRepository(db *pgx.Conn) *userRepository {
  return &userRepository{
    db: db,
  }
}

func (r userRepository) CreateUser(ctx context.Context, data *models.CreateUserRequest) (*models.UserCreated, error) {
  var (
    id uuid.UUID
    username string
  )

  md, ok := metadata.FromIncomingContext(ctx)
  if !ok {
    return nil, errors.New("No metadata provided.")
  }

  fmt.Println(md.Get("client-ip")[0])
  fmt.Println(md.Get("user-agent")[0])
  
  _, err := r.db.Exec(ctx, fmt.Sprintf(`
    SET LOCAL current_user_id = '%s';
    SET LOCAL client_ip = '%s';
    SET LOCAL user_agent = '%s';
  `,"NULL", md.Get("client-ip")[0], md.Get("user-agent")[0]))

  if err != nil {
    return nil, errors.New("Pb with sql logs.")
  }

  rows, err := r.db.Query(ctx, `
    INSERT INTO users (username, email, phonenumber, password) 
    VALUES ($1, $2, $3, $4)
    RETURNING id, username
  `, data.Username, data.Email, data.Phonenumber, data.Password)
  
  if err != nil {
    return nil, err
  }

  defer rows.Close()

  for rows.Next() {
    if err := rows.Scan(&id, &username); err != nil {
      return nil, err
    }   
  }

  return &models.UserCreated{ Id: id, Username: username }, nil
}

func (r userRepository) IsUserExist(ctx context.Context, email, username string) (map[string]string, error) {
  errs := make(map[string]string, 2)

  rows, err := r.db.Query(ctx, `
    SELECT email, username FROM users WHERE email = $1 OR username = $2
  `, email, username)

  if err != nil {
    return nil, err
  }

  defer rows.Close()

  for rows.Next() {
    var e string
    var u string

    if err := rows.Scan(&e, &u); err != nil {
      return nil, err
    }
    
    if email == e {
      errs["email"] = "Email already exist."
    } 

    if username == u {
      errs["username"] = "Username already exist."
    }
  }

  return errs, nil
}
