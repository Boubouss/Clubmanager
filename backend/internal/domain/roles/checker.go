package roles

import "context"

// RoleChecker is the port that any service needing authorization can depend on.
// The implementation (SQL today, gRPC tomorrow) is injected as an adapter.
type RoleChecker interface {
	HasRole(ctx context.Context, userId, clubId string, roles ...string) (bool, error)
	IsSuperAdmin(ctx context.Context, userId string) (bool, error)
}
