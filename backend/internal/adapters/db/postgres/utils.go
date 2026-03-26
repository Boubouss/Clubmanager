package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
)

func setMetadataLog(ctx context.Context, db *pgx.Conn, id string) error {
   
  _, err := db.Exec(ctx, fmt.Sprintf(`
    SET LOCAL current_user_id = '%s';
    SET LOCAL client_ip = '%s';
    SET LOCAL user_agent = '%s';
  `, id, ctx.Value("client_ip"), ctx.Value("user_agent")))

  if err != nil {
    return errors.New("Pb with sql logs.")
  }

  return nil
}

