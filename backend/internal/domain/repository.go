package domain

import (
	"context"
	"strconv"
)

type Repository[T any, ID string | int] interface {
  Save(context.Context, *T) (*T, error)
  Find(context.Context, ID) (*T, error)
  Search(context.Context, *SearchParams) ([]*T, error)
  Delete(context.Context, ID) (bool, error)
}

type SearchParams struct {
  Fields map[string]any
  Connector string
}

func (s SearchParams) GetWhereClauses() (string, []any) {
  where := ""
  args := make([]any, len(s.Fields))
  i := 1
  
  for k, v := range s.Fields {
    args = append(args, v)
    where += k + " = $" + strconv.Itoa(i) + " "
    if i < len(s.Fields) {
      where += s.Connector + " "
    }
    i++
  }

  return where, args
}
