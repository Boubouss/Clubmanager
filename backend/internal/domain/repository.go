package domain

import "context"

type Repository[T any, ID string | int] interface {
  Save(context.Context, *T) (*T, error)
  Find(context.Context, ID) (*T, error)
  Search(context.Context, map[string]string) ([]*T, error)
  Delete(context.Context, ID) (bool, error)
}
