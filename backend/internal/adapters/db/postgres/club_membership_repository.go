package postgres

import (
	"clubmanager/internal/domain/clubs"
	"clubmanager/internal/domain/members"
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type clubMembershipRepository struct {
	db *pgxpool.Pool
}

func NewClubMembershipRepository(db *pgxpool.Pool) *clubMembershipRepository {
	return &clubMembershipRepository{db: db}
}

func (r clubMembershipRepository) Save(ctx context.Context, cm *members.ClubMembership) (*members.ClubMembership, error) {
	if err := setMetadataLog(ctx, r.db, cm.Id.String()); err != nil {
		return nil, err
	}
	if cm.Id == uuid.Nil {
		return r.insert(ctx, cm)
	}
	return r.update(ctx, cm)
}

func (r clubMembershipRepository) insert(ctx context.Context, cm *members.ClubMembership) (*members.ClubMembership, error) {
	row := r.db.QueryRow(ctx, `
		INSERT INTO club_memberships (member_id, club_id, is_valid)
		VALUES ($1, $2, $3)
		RETURNING id, member_id, club_id, is_valid
	`, cm.MemberId, cm.ClubId, cm.IsValid)

	if err := row.Scan(&cm.Id, &cm.MemberId, &cm.ClubId, &cm.IsValid); err != nil {
		return nil, err
	}
	return cm, nil
}

func (r clubMembershipRepository) update(ctx context.Context, cm *members.ClubMembership) (*members.ClubMembership, error) {
	_, err := r.db.Exec(ctx, `
		UPDATE club_memberships SET is_valid = $1 WHERE id = $2
	`, cm.IsValid, cm.Id)
	if err != nil {
		return nil, err
	}
	return cm, nil
}

func (r clubMembershipRepository) Find(ctx context.Context, id string) (*members.ClubMembership, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, member_id, club_id, is_valid
		FROM club_memberships WHERE id = $1
	`, id)

	cm := members.ClubMembership{}
	if err := row.Scan(&cm.Id, &cm.MemberId, &cm.ClubId, &cm.IsValid); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &cm, nil
}

func (r clubMembershipRepository) FindByMember(ctx context.Context, memberId string) ([]*members.ClubMembership, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, member_id, club_id, is_valid
		FROM club_memberships WHERE member_id = $1
	`, memberId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*members.ClubMembership
	for rows.Next() {
		cm := members.ClubMembership{}
		if err := rows.Scan(&cm.Id, &cm.MemberId, &cm.ClubId, &cm.IsValid); err != nil {
			return nil, err
		}
		list = append(list, &cm)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return list, nil
}

func (r clubMembershipRepository) FindByClub(ctx context.Context, clubId string) ([]*members.ClubMembership, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, member_id, club_id, is_valid
		FROM club_memberships WHERE club_id = $1
	`, clubId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*members.ClubMembership
	for rows.Next() {
		cm := members.ClubMembership{}
		if err := rows.Scan(&cm.Id, &cm.MemberId, &cm.ClubId, &cm.IsValid); err != nil {
			return nil, err
		}
		list = append(list, &cm)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return list, nil
}

func (r clubMembershipRepository) Delete(ctx context.Context, id string) (bool, error) {
	if err := setMetadataLog(ctx, r.db, id); err != nil {
		return false, err
	}
	_, err := r.db.Exec(ctx, `DELETE FROM club_memberships WHERE id = $1`, id)
	if err != nil {
		return false, err
	}
	return true, nil
}

// FindByClubWithContacts returns a paginated list of memberships for a club enriched with user contact info.
func (r clubMembershipRepository) FindByClubWithContacts(ctx context.Context, clubId string, page, pageSize int) ([]*members.MemberContact, int, error) {
	var total int
	if err := r.db.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM club_memberships cm
		JOIN members m ON m.id = cm.member_id
		WHERE cm.club_id = $1
	`, clubId).Scan(&total); err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	rows, err := r.db.Query(ctx, `
		SELECT cm.id, cm.member_id, cm.club_id, cm.is_valid,
		       m.id, m.user_id, m.firstname, m.lastname, m.birthdate::text, m.gender, m.is_primary,
		       u.email, COALESCE(u.phonenumber, '')
		FROM club_memberships cm
		JOIN members m ON m.id = cm.member_id
		JOIN users u ON u.id = m.user_id
		WHERE cm.club_id = $1
		ORDER BY m.lastname, m.firstname
		LIMIT $2 OFFSET $3
	`, clubId, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var list []*members.MemberContact
	for rows.Next() {
		var mc members.MemberContact
		var msId, memberId, clubUUID, userId uuid.UUID
		if err := rows.Scan(
			&msId, &memberId, &clubUUID, &mc.Membership.IsValid,
			&mc.Member.Id, &userId, &mc.Member.Firstname, &mc.Member.Lastname,
			&mc.Member.Birthdate, &mc.Member.Gender, &mc.Member.IsPrimary,
			&mc.Email, &mc.Phone,
		); err != nil {
			return nil, 0, err
		}
		mc.Membership.Id = msId
		mc.Membership.MemberId = memberId
		mc.Membership.ClubId = clubUUID
		mc.Member.UserId = userId
		list = append(list, &mc)
	}
	return list, total, rows.Err()
}

// FindClubsByUser returns all clubs the user belongs to (via membership or as creator).
// Implements services.UserClubsPort.
func (r clubMembershipRepository) FindClubsByUser(ctx context.Context, userId string) ([]*clubs.Club, error) {
	rows, err := r.db.Query(ctx, `
		SELECT c.id, c.siren, c.name, c.address, c.city, c.postal_code,
		       c.country, c.phonenumber, c.status, c.creator_id
		FROM clubs c
		JOIN club_memberships cm ON cm.club_id = c.id
		JOIN members m ON m.id = cm.member_id
		WHERE m.user_id = $1
		UNION
		SELECT c.id, c.siren, c.name, c.address, c.city, c.postal_code,
		       c.country, c.phonenumber, c.status, c.creator_id
		FROM clubs c
		WHERE c.creator_id = $1
		UNION
		SELECT c.id, c.siren, c.name, c.address, c.city, c.postal_code,
		       c.country, c.phonenumber, c.status, c.creator_id
		FROM clubs c
		JOIN roles r ON r.club_id = c.id
		WHERE r.user_id = $1
		ORDER BY name
	`, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*clubs.Club
	for rows.Next() {
		cl := &clubs.Club{}
		var creatorId *uuid.UUID
		if err := rows.Scan(&cl.Id, &cl.Siren, &cl.Name, &cl.Address, &cl.City,
			&cl.PostalCode, &cl.Country, &cl.Phonenumber, &cl.Status, &creatorId); err != nil {
			return nil, err
		}
		if creatorId != nil {
			cl.CreatorId = *creatorId
		}
		list = append(list, cl)
	}
	return list, rows.Err()
}
