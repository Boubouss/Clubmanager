package users

import (
	dbutils "clubmanager/utils/db"
	"context"

	"github.com/jackc/pgx/v5"
)

type UserRepository interface {
  CreateUser(context.Context, *CreateUserRequest) (*User, error)
  IsUserExist(context.Context, string, string) (map[string]string, error)
  ReadUser(context.Context, *ReadUserRequest) ([]User, error)
  UpdateUser(context.Context, *UpdateUserRequest) (*User, error)
  DeleteUser(context.Context, string) (bool, error)
}

type userRepository struct {
  db *pgx.Conn
}

func NewUserRepository(db *pgx.Conn) *userRepository {
  return &userRepository{
    db: db,
  }
}

func (r userRepository) CreateUser(ctx context.Context, data *CreateUserRequest) (*User, error) {
  if err := dbutils.SetMetadataLog(r.db, ctx, "NULL"); err != nil {
    return nil, err
  }

  row := r.db.QueryRow(ctx, `
    INSERT INTO users (username, email, phonenumber, password) 
    VALUES ($1, $2, $3, $4)
    RETURNING id, username, email, phonenumber
  `, data.Username, data.Email, data.Phonenumber, data.Password)
  
  var user User  
  if err := row.Scan(&user.Id, &user.Username, &user.Email, &user.Phonenumber); err != nil {
    return nil, err
  }   

  return &user, nil
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

func (r userRepository) ReadUser(ctx context.Context, data *ReadUserRequest) ([]User, error) {

  return nil, nil
}

func (r userRepository) UpdateUser(ctx context.Context, data *UpdateUserRequest) (*User, error) {
  if err := dbutils.SetMetadataLog(r.db, ctx, data.Id); err != nil {
    return nil, err
  }
  
  query, args := dbutils.GenerateUpdateQuery("users", data.Map())

  query += " RETURNING id, username, email, phonenumber"

  row := r.db.QueryRow(ctx, query, args...)

  var user User
  if err := row.Scan(&user.Id, &user.Username, &user.Email, &user.Phonenumber); err != nil {
    return nil, err
  }

  return &user, nil
}


func (r userRepository) DeleteUser(ctx context.Context, id string) (bool, error) {
  
  if err := dbutils.SetMetadataLog(r.db, ctx, id); err != nil {
    return false, err
  }

  _, err := r.db.Exec(ctx, `
    DELETE FROM users WHERE id = $1
  `, id)

  if err != nil {
    return false, err 
  }

  return true, nil
}
