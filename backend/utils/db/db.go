package db

import (
	"context"
	"errors"
	"fmt"
  "strconv"

	"github.com/jackc/pgx/v5"
	"google.golang.org/grpc/metadata"
)

func SetMetadataLog(db *pgx.Conn, ctx context.Context, id string) error {
  md, ok := metadata.FromIncomingContext(ctx)
  if !ok {
    return errors.New("No metadata provided.")
  }
  
  _, err := db.Exec(ctx, fmt.Sprintf(`
    SET LOCAL current_user_id = '%s';
    SET LOCAL client_ip = '%s';
    SET LOCAL user_agent = '%s';
  `, id, md.Get("client-ip")[0], md.Get("user-agent")[0]))

  if err != nil {
    return errors.New("Pb with sql logs.")
  }

  return nil
}

func GenerateUpdateQuery(table string, data map[string]string) (string, []any) {
  query := "UPDATE " + table + " SET "
  args := make([]any, 0)
  i := 1
  id := data["id"]
  delete(data, "id")

  for k, v := range data {
    args = append(args, v)
    query += k 
    query += " = $"
    query += strconv.Itoa(i)
    i++
    if i < len(data) {
      query += ", "
    } 
  }

  query += " WHERE id = $"
  query += strconv.Itoa(i)
  args = append(args, id)

  return query, args
}
