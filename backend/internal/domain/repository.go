package domain

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"
)

type Repository[T any, ID string | int] interface {
  Save(context.Context, *T) (*T, error)
  Find(context.Context, ID) (*T, error)
  Search(context.Context, *SearchParams) ([]*T, error)
  Delete(context.Context, ID) (bool, error)
}

type SearchParams struct {
  Fields map[string]any
  Keys  map[string]bool
  Connector string
}

func (s SearchParams) GetWhereClauses() (string, []any, error) {
  keys := make([]string, 0, len(s.Fields))
  for k := range s.Fields {
    if _, ok := s.Keys[k]; !ok {
      return "", nil, fmt.Errorf("invalid key in search params: %s", k)
    }
    keys = append(keys, k)
  }
  sort.Strings(keys)

  clauses := make([]string, 0, len(keys))
  args := make([]any, 0, len(keys))
  for i, k := range keys {
    clauses = append(clauses, k+" = $"+strconv.Itoa(i+1))
    args = append(args, s.Fields[k])
  }

  if len(clauses) == 0 {
    return "1=1", args, nil
  }
  return strings.Join(clauses, " "+s.Connector+" "), args, nil
}
