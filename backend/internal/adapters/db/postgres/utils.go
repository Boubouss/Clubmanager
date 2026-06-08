package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

func setMetadataLog(ctx context.Context, db *pgxpool.Pool, id string) error {
  clientIP, _ := ctx.Value("client_ip").(string)
  userAgent, _ := ctx.Value("user_agent").(string)

  _, err := db.Exec(ctx, `
    SELECT
      set_config('app.current_user_id', $1, true),
      set_config('app.client_ip',       $2, true),
      set_config('app.user_agent',      $3, true)
  `, id, clientIP, userAgent)

  if err != nil {
    return fmt.Errorf("failed to set session metadata for logs: %w", err)
  }

  return nil
}

