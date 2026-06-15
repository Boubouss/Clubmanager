package postgres

import (
	"clubmanager/internal/adapters/api/grpc/dto"
	"clubmanager/internal/domain"
	"clubmanager/internal/domain/roles"
	"context"

	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type roleRepository struct {
	db *pgxpool.Pool
}

func NewRoleRepository(db *pgxpool.Pool) *roleRepository {
	return &roleRepository{db: db}
}

func (r roleRepository) Save(ctx context.Context, role *roles.Role) (*roles.Role, error) {
	if err := setMetadataLog(ctx, r.db, role.Id.String()); err != nil {
		return nil, err
	}
	if role.Id == uuid.Nil {
		return r.insert(ctx, role)
	}
	return role, nil
}

func (r roleRepository) insert(ctx context.Context, role *roles.Role) (*roles.Role, error) {
	row := r.db.QueryRow(ctx, `
		INSERT INTO roles (user_id, club_id, role)
		VALUES ($1, $2, $3)
		RETURNING id, user_id, club_id, role
	`, role.UserId, role.ClubId, role.Role)

	if err := row.Scan(&role.Id, &role.UserId, &role.ClubId, &role.Role); err != nil {
		return nil, err
	}
	return role, nil
}

func (r roleRepository) Find(ctx context.Context, id string) (*roles.Role, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, user_id, club_id, role FROM roles WHERE id = $1
	`, id)

	role := roles.Role{}
	if err := row.Scan(&role.Id, &role.UserId, &role.ClubId, &role.Role); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &role, nil
}

func (r roleRepository) Search(ctx context.Context, params *domain.SearchParams) ([]*roles.Role, error) {
	where, args, err := params.GetWhereClauses()
	if err != nil {
		return nil, err
	}

	rows, err := r.db.Query(ctx,
		"SELECT id, user_id, club_id, role FROM roles WHERE "+where,
		args...,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*roles.Role
	for rows.Next() {
		role := roles.Role{}
		if err := rows.Scan(&role.Id, &role.UserId, &role.ClubId, &role.Role); err != nil {
			return nil, err
		}
		list = append(list, &role)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return list, nil
}

func (r roleRepository) Delete(ctx context.Context, id string) (bool, error) {
	if err := setMetadataLog(ctx, r.db, id); err != nil {
		return false, err
	}
	_, err := r.db.Exec(ctx, `DELETE FROM roles WHERE id = $1`, id)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (r roleRepository) HasRole(ctx context.Context, userId, clubId string, roleList ...string) (bool, error) {
	var count int
	err := r.db.QueryRow(ctx, `
		SELECT COUNT(*) FROM roles
		WHERE user_id = $1 AND club_id = $2 AND role = ANY($3)
	`, userId, clubId, roleList).Scan(&count)
	return count > 0, err
}

func (r roleRepository) IsSuperAdmin(ctx context.Context, userId string) (bool, error) {
	var isSuperAdmin bool
	err := r.db.QueryRow(ctx, `
		SELECT is_superadmin FROM users WHERE id = $1
	`, userId).Scan(&isSuperAdmin)
	return isSuperAdmin, err
}

// FindStaffContactsByClub returns contact info for all members with a staff/coach role in a club.
// Uses a single JOIN across roles → members → users, filtering on the primary member profile.
func (r roleRepository) FindStaffContactsByClub(ctx context.Context, clubId string) ([]dto.StaffContact, error) {
	rows, err := r.db.Query(ctx, `
		SELECT m.firstname, m.lastname, ro.role,
		       u.email, COALESCE(u.phonenumber, '')
		FROM roles ro
		JOIN users u ON u.id = ro.user_id
		JOIN members m ON m.user_id = ro.user_id AND m.is_primary = true
		WHERE ro.club_id = $1
		  AND ro.role IN ('president', 'staff', 'coach')
		ORDER BY ro.role, m.lastname, m.firstname
	`, clubId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []dto.StaffContact
	for rows.Next() {
		var sc dto.StaffContact
		if err := rows.Scan(&sc.Firstname, &sc.Lastname, &sc.Role, &sc.Email, &sc.Phone); err != nil {
			return nil, err
		}
		list = append(list, sc)
	}
	return list, rows.Err()
}
