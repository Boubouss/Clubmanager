package services

import (
	"clubmanager/internal/adapters/api/grpc/dto"
	"clubmanager/internal/domain"
	"clubmanager/internal/domain/clubs"
	"clubmanager/internal/domain/members"
	"context"
	"fmt"

	"github.com/google/uuid"
)

// MemberRepository extends the generic repository with a batch lookup by IDs.
type MemberRepository interface {
	domain.Repository[members.Member, string]
	FindByIds(ctx context.Context, ids []string) ([]*members.Member, error)
}

// ClubMembershipRepository defines the persistence port for club memberships.
type ClubMembershipRepository interface {
	Save(ctx context.Context, cm *members.ClubMembership) (*members.ClubMembership, error)
	Find(ctx context.Context, id string) (*members.ClubMembership, error)
	FindByMember(ctx context.Context, memberId string) ([]*members.ClubMembership, error)
	FindByClub(ctx context.Context, clubId string) ([]*members.ClubMembership, error)
	FindByClubWithContacts(ctx context.Context, clubId string, page, pageSize int) ([]*members.MemberContact, int, error)
	Delete(ctx context.Context, id string) (bool, error)
}

// StaffContactsPort is the port for fetching staff/coach contact info for a club.
type StaffContactsPort interface {
	FindStaffContactsByClub(ctx context.Context, clubId string) ([]dto.StaffContact, error)
}

// UserClubsPort is the port for fetching clubs a user belongs to.
type UserClubsPort interface {
	FindClubsByUser(ctx context.Context, userId string) ([]*clubs.Club, error)
}

type MemberService interface {
	CreateMember(context.Context, *dto.CreateMemberRequest) (*dto.CreateMemberResponse, error)
	RequestMembership(context.Context, *dto.RequestMembershipRequest) (*dto.RequestMembershipResponse, error)
	AddMembership(context.Context, string, string) (*members.ClubMembership, error)
	AddMember(context.Context, *dto.AddMemberRequest) (*dto.AddMemberResponse, error)
	ValidateMember(context.Context, *dto.ValidateMemberRequest) (*dto.ValidateMemberResponse, error)
	GetMember(context.Context, string) (*members.Member, error)
	GetMembersByIds(context.Context, []string) (map[string]*members.Member, error)
	GetMembersByUser(context.Context, *dto.GetMembersByUserRequest) (*dto.GetMembersByUserResponse, error)
	GetMembersByClub(context.Context, *dto.GetMembersByClubRequest) (*dto.GetMembersByClubResponse, error)
	GetStaffContacts(context.Context, string) (*dto.GetStaffContactsResponse, error)
	GetUserClubs(context.Context, *dto.GetUserClubsRequest) (*dto.GetUserClubsResponse, error)
	UpdateMember(context.Context, *dto.UpdateMemberRequest) (*dto.UpdateMemberResponse, error)
	RemoveMember(context.Context, string) (bool, error)
	RemoveMembership(context.Context, string) (bool, error)
}

type MemberServiceConfig struct {
	Repository           MemberRepository
	MembershipRepository ClubMembershipRepository
	UserClubsPort        UserClubsPort
	StaffContactsPort    StaffContactsPort
	ClubStatusChecker    clubs.ClubStatusChecker
}

type memberService struct {
	repo              MemberRepository
	membershipRepo    ClubMembershipRepository
	userClubsPort     UserClubsPort
	staffContactsPort StaffContactsPort
	clubStatusChecker clubs.ClubStatusChecker
}

func NewMemberService(config MemberServiceConfig) *memberService {
	return &memberService{
		repo:              config.Repository,
		membershipRepo:    config.MembershipRepository,
		userClubsPort:     config.UserClubsPort,
		staffContactsPort: config.StaffContactsPort,
		clubStatusChecker: config.ClubStatusChecker,
	}
}

func (s *memberService) CreateMember(ctx context.Context, data *dto.CreateMemberRequest) (*dto.CreateMemberResponse, error) {
	m, errs := members.NewMember(data.UserId, data.Map(), data.IsPrimary)
	if len(errs) > 0 {
		return &dto.CreateMemberResponse{Errors: errs}, nil
	}
	created, err := s.repo.Save(ctx, m)
	if err != nil {
		return nil, err
	}
	return &dto.CreateMemberResponse{Member: created, Errors: make(map[string]string)}, nil
}

func (s *memberService) RequestMembership(ctx context.Context, data *dto.RequestMembershipRequest) (*dto.RequestMembershipResponse, error) {
	// Verify the member belongs to the requesting user
	member, err := s.repo.Find(ctx, data.MemberId)
	if err != nil {
		return nil, err
	}
	if member == nil || member.UserId.String() != data.UserId {
		return &dto.RequestMembershipResponse{
			Errors: map[string]string{"member": "Profil membre introuvable ou non autorisé."},
		}, nil
	}

	// Check club is active
	if s.clubStatusChecker != nil {
		active, err := s.clubStatusChecker.IsActive(ctx, data.ClubId)
		if err != nil {
			return nil, fmt.Errorf("check club status: %w", err)
		}
		if !active {
			return &dto.RequestMembershipResponse{
				Errors: map[string]string{"club": "Ce club n'accepte pas de nouvelles adhésions pour le moment."},
			}, nil
		}
	}

	cm, cmErrs := members.NewClubMembership(data.MemberId, data.ClubId)
	if len(cmErrs) > 0 {
		return &dto.RequestMembershipResponse{Errors: cmErrs}, nil
	}
	saved, err := s.membershipRepo.Save(ctx, cm)
	if err != nil {
		return nil, err
	}
	return &dto.RequestMembershipResponse{Membership: saved, Errors: make(map[string]string)}, nil
}

func (s *memberService) AddMember(ctx context.Context, data *dto.AddMemberRequest) (*dto.AddMemberResponse, error) {
	// Validate all inputs before any DB write to avoid orphaned records.
	m, errs := members.NewMember(data.UserId, data.Map(), data.IsPrimary)

	// Validate ClubId early — NewClubMembership would catch it, but only after
	// the member is already saved. Catching it here prevents orphaned members.
	if _, parseErr := uuid.Parse(data.ClubId); parseErr != nil {
		errs["club_id"] = "Invalid club ID."
	}

	if len(errs) > 0 {
		return &dto.AddMemberResponse{Errors: errs}, nil
	}

	// Block member creation if the club is not active.
	if s.clubStatusChecker != nil {
		active, err := s.clubStatusChecker.IsActive(ctx, data.ClubId)
		if err != nil {
			return nil, fmt.Errorf("check club status: %w", err)
		}
		if !active {
			return &dto.AddMemberResponse{
				Errors: map[string]string{"club": "Club is not active. Members cannot be added to a pending or suspended club."},
			}, nil
		}
	}

	created, err := s.repo.Save(ctx, m)
	if err != nil {
		return nil, err
	}

	cm, cmErrs := members.NewClubMembership(created.Id.String(), data.ClubId)
	if len(cmErrs) > 0 {
		// Should not happen — we validated ClubId above — but guard anyway.
		return &dto.AddMemberResponse{Errors: cmErrs}, nil
	}

	membership, err := s.membershipRepo.Save(ctx, cm)
	if err != nil {
		return nil, err
	}

	return &dto.AddMemberResponse{
		Member:     created,
		Membership: membership,
		Errors:     make(map[string]string),
	}, nil
}

// ValidateMember approves a club membership by its ID.
func (s *memberService) ValidateMember(ctx context.Context, data *dto.ValidateMemberRequest) (*dto.ValidateMemberResponse, error) {
	cm, err := s.membershipRepo.Find(ctx, data.Id)
	if err != nil {
		return nil, err
	}
	if cm == nil {
		return &dto.ValidateMemberResponse{Errors: map[string]string{"membership": "Club membership not found."}}, nil
	}

	cm.IsValid = true
	updated, err := s.membershipRepo.Save(ctx, cm)
	if err != nil {
		return nil, err
	}

	member, err := s.repo.Find(ctx, cm.MemberId.String())
	if err != nil {
		return nil, err
	}

	return &dto.ValidateMemberResponse{
		Member:     member,
		Membership: updated,
		Errors:     make(map[string]string),
	}, nil
}

func (s *memberService) GetMember(ctx context.Context, memberId string) (*members.Member, error) {
	return s.repo.Find(ctx, memberId)
}

// GetMembersByIds fetches multiple members in a single query and returns them
// indexed by their ID string. Missing IDs are silently absent from the map.
func (s *memberService) GetMembersByIds(ctx context.Context, ids []string) (map[string]*members.Member, error) {
	if len(ids) == 0 {
		return map[string]*members.Member{}, nil
	}
	list, err := s.repo.FindByIds(ctx, ids)
	if err != nil {
		return nil, err
	}
	result := make(map[string]*members.Member, len(list))
	for _, m := range list {
		result[m.Id.String()] = m
	}
	return result, nil
}

func (s *memberService) GetMembersByUser(ctx context.Context, data *dto.GetMembersByUserRequest) (*dto.GetMembersByUserResponse, error) {
	memberList, err := s.repo.Search(ctx, &domain.SearchParams{
		Fields:    map[string]any{"user_id": data.UserId},
		Keys:      map[string]bool{"user_id": true},
		Connector: "AND",
	})
	if err != nil {
		return nil, err
	}

	if data.ClubId == "" {
		return &dto.GetMembersByUserResponse{Members: memberList, Errors: make(map[string]string)}, nil
	}

	// Filter members that have an active membership in the requested club.
	var filtered []*members.Member
	for _, m := range memberList {
		memberships, err := s.membershipRepo.FindByMember(ctx, m.Id.String())
		if err != nil {
			return nil, err
		}
		for _, ms := range memberships {
			if ms.ClubId.String() == data.ClubId {
				filtered = append(filtered, m)
				break
			}
		}
	}

	return &dto.GetMembersByUserResponse{Members: filtered, Errors: make(map[string]string)}, nil
}

func (s *memberService) UpdateMember(ctx context.Context, data *dto.UpdateMemberRequest) (*dto.UpdateMemberResponse, error) {
	current, err := s.repo.Find(ctx, data.Id)
	if err != nil {
		return nil, err
	}
	if current == nil {
		return &dto.UpdateMemberResponse{Errors: map[string]string{"member": "Member not found."}}, nil
	}

	m, errs := current.Update(data.Map())
	if len(errs) > 0 {
		return &dto.UpdateMemberResponse{Errors: errs}, nil
	}

	updated, err := s.repo.Save(ctx, m)
	if err != nil {
		return nil, err
	}
	return &dto.UpdateMemberResponse{Member: updated, Errors: make(map[string]string)}, nil
}

// AddMembership creates a validated ClubMembership directly (e.g. for a president at club creation).
func (s *memberService) AddMembership(ctx context.Context, memberId, clubId string) (*members.ClubMembership, error) {
	cm, errs := members.NewClubMembership(memberId, clubId)
	if len(errs) > 0 {
		return nil, fmt.Errorf("invalid membership params: %v", errs)
	}
	cm.IsValid = true
	return s.membershipRepo.Save(ctx, cm)
}

func (s *memberService) RemoveMember(ctx context.Context, id string) (bool, error) {
	return s.repo.Delete(ctx, id)
}

func (s *memberService) RemoveMembership(ctx context.Context, id string) (bool, error) {
	return s.membershipRepo.Delete(ctx, id)
}

func (s *memberService) GetMembersByClub(ctx context.Context, data *dto.GetMembersByClubRequest) (*dto.GetMembersByClubResponse, error) {
	page := data.Page
	if page <= 0 {
		page = 1
	}
	pageSize := data.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	contacts, total, err := s.membershipRepo.FindByClubWithContacts(ctx, data.ClubId, page, pageSize)
	if err != nil {
		return nil, err
	}

	result := make([]dto.MemberWithMembership, 0, len(contacts))
	for _, mc := range contacts {
		result = append(result, dto.MemberWithMembership{
			Member:     &mc.Member,
			Membership: &mc.Membership,
			Email:      mc.Email,
			Phone:      mc.Phone,
		})
	}

	return &dto.GetMembersByClubResponse{Members: result, Total: total, Errors: make(map[string]string)}, nil
}

func (s *memberService) GetStaffContacts(ctx context.Context, clubId string) (*dto.GetStaffContactsResponse, error) {
	if s.staffContactsPort == nil {
		return &dto.GetStaffContactsResponse{}, nil
	}
	contacts, err := s.staffContactsPort.FindStaffContactsByClub(ctx, clubId)
	if err != nil {
		return nil, err
	}
	return &dto.GetStaffContactsResponse{Contacts: contacts}, nil
}

func (s *memberService) GetUserClubs(ctx context.Context, data *dto.GetUserClubsRequest) (*dto.GetUserClubsResponse, error) {
	if s.userClubsPort == nil {
		return &dto.GetUserClubsResponse{Clubs: nil, Errors: make(map[string]string)}, nil
	}
	list, err := s.userClubsPort.FindClubsByUser(ctx, data.UserId)
	if err != nil {
		return nil, err
	}
	return &dto.GetUserClubsResponse{Clubs: list, Errors: make(map[string]string)}, nil
}
