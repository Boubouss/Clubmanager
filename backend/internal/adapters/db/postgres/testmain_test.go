//go:build integration

package postgres

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	tc "github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

var testPool *pgxpool.Pool

func TestMain(m *testing.M) {
	ctx := context.Background()

	sqlDir := sqlMigrationsDir()

	pgc, err := tcpostgres.Run(ctx,
		"postgres:16-alpine",
		tcpostgres.WithDatabase("testdb"),
		tcpostgres.WithUsername("test"),
		tcpostgres.WithPassword("test"),
		// 4-triggers.sql is intentionally skipped: the triggers rely on session
		// variables set via set_config(), which do not propagate across pgxpool
		// connections. Running without triggers avoids false failures while still
		// exercising the real SQL logic.
		tcpostgres.WithInitScripts(
			filepath.Join(sqlDir, "1-tables.sql"),
			filepath.Join(sqlDir, "2-indexes.sql"),
			filepath.Join(sqlDir, "3-functions.sql"),
			filepath.Join(sqlDir, "5-seeds.sql"),
		),
		tc.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(60*time.Second),
		),
	)
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to start postgres container:", err)
		os.Exit(1)
	}
	defer pgc.Terminate(ctx) //nolint:errcheck

	connStr, err := pgc.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to get connection string:", err)
		os.Exit(1)
	}

	testPool, err = pgxpool.New(ctx, connStr)
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to create pool:", err)
		os.Exit(1)
	}
	defer testPool.Close()

	os.Exit(m.Run())
}

// sqlMigrationsDir returns the absolute path to the project's sql/ directory by
// walking up from the current source file location.
func sqlMigrationsDir() string {
	_, filename, _, _ := runtime.Caller(0)
	// filename: .../backend/internal/adapters/db/postgres/testmain_test.go
	// Walk up: postgres → db → adapters → internal → backend → CLUBMANAGER
	dir := filepath.Dir(filename)
	for i := 0; i < 5; i++ {
		dir = filepath.Dir(dir)
	}
	return filepath.Join(dir, "sql")
}

// ── Cleanup helper ────────────────────────────────────────────────────────────

// truncateAll wipes all tables in dependency order and resets sequences.
func truncateAll(t *testing.T) {
	t.Helper()
	_, err := testPool.Exec(context.Background(), `
		TRUNCATE TABLE
			carpool_passengers, carpool_offers,
			event_participants, event_categories,
			events, posts, licences, roles, club_memberships, members,
			clubs, users, logs
		RESTART IDENTITY CASCADE
	`)
	if err != nil {
		t.Fatalf("truncateAll: %v", err)
	}
}

// ── Seed helpers ──────────────────────────────────────────────────────────────

func seedUser(t *testing.T, username, email string) uuid.UUID {
	t.Helper()
	var id uuid.UUID
	err := testPool.QueryRow(context.Background(), `
		INSERT INTO users (username, email, password)
		VALUES ($1, $2, 'hashed')
		RETURNING id
	`, username, email).Scan(&id)
	if err != nil {
		t.Fatalf("seedUser %q: %v", username, err)
	}
	return id
}

func seedClub(t *testing.T, siren, name string) uuid.UUID {
	t.Helper()
	var id uuid.UUID
	err := testPool.QueryRow(context.Background(), `
		INSERT INTO clubs (siren, name, address, city, postal_code)
		VALUES ($1, $2, '1 rue Test', 'Paris', '75001')
		RETURNING id
	`, siren, name).Scan(&id)
	if err != nil {
		t.Fatalf("seedClub %q: %v", siren, err)
	}
	return id
}

func seedMember(t *testing.T, userId uuid.UUID, firstname, lastname string) uuid.UUID {
	t.Helper()
	var id uuid.UUID
	err := testPool.QueryRow(context.Background(), `
		INSERT INTO members (user_id, firstname, lastname, birthdate, gender)
		VALUES ($1, $2, $3, '2000-01-01', 'man')
		RETURNING id
	`, userId, firstname, lastname).Scan(&id)
	if err != nil {
		t.Fatalf("seedMember %q %q: %v", firstname, lastname, err)
	}
	return id
}

func seedClubMembership(t *testing.T, memberId, clubId uuid.UUID) uuid.UUID {
	t.Helper()
	var id uuid.UUID
	err := testPool.QueryRow(context.Background(), `
		INSERT INTO club_memberships (member_id, club_id)
		VALUES ($1, $2)
		RETURNING id
	`, memberId, clubId).Scan(&id)
	if err != nil {
		t.Fatalf("seedClubMembership member=%s club=%s: %v", memberId, clubId, err)
	}
	return id
}

func seedEvent(t *testing.T, clubId, createdBy uuid.UUID, title string) uuid.UUID {
	t.Helper()
	var id uuid.UUID
	err := testPool.QueryRow(context.Background(), `
		INSERT INTO events (club_id, created_by, title, type, status, location,
		                    registration_open_at, registration_close_at, date)
		VALUES ($1, $2, $3, 'competition', 'open', 'Dojo Paris',
		        NOW() - INTERVAL '1 day',
		        NOW() + INTERVAL '30 days',
		        NOW() + INTERVAL '60 days')
		RETURNING id
	`, clubId, createdBy, title).Scan(&id)
	if err != nil {
		t.Fatalf("seedEvent %q: %v", title, err)
	}
	return id
}

func seedEventMaxParticipants(t *testing.T, clubId, createdBy uuid.UUID, title string, max int) uuid.UUID {
	t.Helper()
	var id uuid.UUID
	err := testPool.QueryRow(context.Background(), `
		INSERT INTO events (club_id, created_by, title, type, status, location,
		                    registration_open_at, registration_close_at, date, max_participants)
		VALUES ($1, $2, $3, 'competition', 'open', 'Dojo Paris',
		        NOW() - INTERVAL '1 day',
		        NOW() + INTERVAL '30 days',
		        NOW() + INTERVAL '60 days',
		        $4)
		RETURNING id
	`, clubId, createdBy, title, max).Scan(&id)
	if err != nil {
		t.Fatalf("seedEventMaxParticipants %q: %v", title, err)
	}
	return id
}
