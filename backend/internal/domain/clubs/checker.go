package clubs

import "context"

// ClubStatusChecker is the port that services depend on to verify a club is active.
// Implemented by the club repository adapter.
type ClubStatusChecker interface {
	IsActive(ctx context.Context, clubId string) (bool, error)
}
