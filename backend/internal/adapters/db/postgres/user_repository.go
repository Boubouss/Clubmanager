package postgres

import (
	"clubmanager/internal/domain"
	"clubmanager/internal/domain/users"
	"context"

	"github.com/jackc/pgx/v5"
)

type userRepository struct {
  db *pgx.Conn
}

func NewUserRepository(db *pgx.Conn) *userRepository {
  return &userRepository{
    db: db,
  }
}

func (r userRepository) Save(ctx context.Context, u *users.User) (*users.User, error) {
  if err := setMetadataLog(ctx, r.db, u.Id.String()); err != nil {
    return nil, err
  }

  if u.Id.String() == "" {
    return r.insert(ctx, u)
  }
  return r.update(ctx, u)
}

func (r userRepository) insert(ctx context.Context, u *users.User) (*users.User, error) {
  
  row := r.db.QueryRow(ctx, `
    INSERT INTO users (username, email, phonenumber, password) 
    VALUES ($1, $2, $3, $4)
    RETURNING id, username, email, phonenumber
  `, u.Username, u.Email, u.Phonenumber, u.Password)
  
  if err := row.Scan(&u.Id, &u.Username, &u.Email, &u.Phonenumber); err != nil {
    return nil, err
  }   

  return u, nil
}

func (r userRepository) update(ctx context.Context, u *users.User) (*users.User, error) {

  _, err := r.db.Exec(ctx, `
    UPDATE users SET email = '$1', phonenumber = '$2', password = '$3'
    WHERE id = '$4'
  `, u.Email, u.Phonenumber, u.Password, u.Id.String())

  if err != nil {
    return nil, err
  }

  return u, nil
}

func (r userRepository) Find(ctx context.Context, id string) (*users.User, error) {
  row := r.db.QueryRow(ctx, `
    SELECT id, username, email, phonenumber FROM users
    WHERE id = '$1'
  `, id)

  u := users.User{}
  if err := row.Scan(&u.Id, &u.Username, &u.Email, &u.Phonenumber); err != nil {
    return nil, err
  }
  return &u, nil
}

func (r userRepository) Search(ctx context.Context, params *domain.SearchParams) ([]*users.User, error) {

  where, args := params.GetWhereClauses()
  query := "SELECT id, username, email, phonenumber FROM users WHERE " + where

  rows, err := r.db.Query(ctx, query, args...)

  if err != nil {
    return nil, err
  }

  defer rows.Close()

  var list []*users.User
  for rows.Next() {
    u := users.User{}
    if err := rows.Scan(&u.Id, &u.Username, &u.Email, &u.Phonenumber); err != nil {
      return nil, err
    }
    list = append(list, &u)
  }

  return list, nil
}

func (r userRepository) Delete(ctx context.Context, id string) (bool, error) {
  
  if err := setMetadataLog(ctx, r.db, id); err != nil {
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


