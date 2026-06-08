package services

import (
	"clubmanager/internal/adapters/api/grpc/dto"
	"clubmanager/internal/domain"
	"clubmanager/internal/domain/members"
	"context"
)

type MemberService interface {
	AddMember(context.Context, *dto.AddMemberRequest) (*dto.AddMemberResponse, error)
	ValidateMember(context.Context, *dto.ValidateMemberRequest) (*dto.ValidateMemberResponse, error)
	GetMembersByUser(context.Context, *dto.GetMembersByUserRequest) (*dto.GetMembersByUserResponse, error)
	UpdateMember(context.Context, *dto.UpdateMemberRequest) (*dto.UpdateMemberResponse, error)
	RemoveMember(context.Context, string) (bool, error)
}

type MemberServiceConfig struct {
	Repository domain.Repository[members.Member, string]
}

type memberService struct {
	repo domain.Repository[members.Member, string]
}

func NewMemberService(config MemberServiceConfig) *memberService {
	return &memberService{repo: config.Repository}
}

func (s *memberService) AddMember(ctx context.Context, data *dto.AddMemberRequest) (*dto.AddMemberResponse, error) {
	m, errs := members.NewMember(data.UserId, data.ClubId, data.Map())
	if len(errs) > 0 {
		return &dto.AddMemberResponse{Errors: errs}, nil
	}

	created, err := s.repo.Save(ctx, m)
	if err != nil {
		return nil, err
	}
	return &dto.AddMemberResponse{Member: created, Errors: make(map[string]string)}, nil
}

func (s *memberService) ValidateMember(ctx context.Context, data *dto.ValidateMemberRequest) (*dto.ValidateMemberResponse, error) {
	m, err := s.repo.Find(ctx, data.Id)
	if err != nil {
		return nil, err
	}
	if m == nil {
		return &dto.ValidateMemberResponse{Errors: map[string]string{"member": "Member not found."}}, nil
	}

	m.IsValid = true
	updated, err := s.repo.Save(ctx, m)
	if err != nil {
		return nil, err
	}
	return &dto.ValidateMemberResponse{Member: updated, Errors: make(map[string]string)}, nil
}

func (s *memberService) GetMembersByUser(ctx context.Context, data *dto.GetMembersByUserRequest) (*dto.GetMembersByUserResponse, error) {
	fields := map[string]any{"user_id": data.UserId}
	keys := map[string]bool{"user_id": true}

	if data.ClubId != "" {
		fields["club_id"] = data.ClubId
		keys["club_id"] = true
	}

	list, err := s.repo.Search(ctx, &domain.SearchParams{
		Fields:    fields,
		Keys:      keys,
		Connector: "AND",
	})
	if err != nil {
		return nil, err
	}
	return &dto.GetMembersByUserResponse{Members: list, Errors: make(map[string]string)}, nil
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

func (s *memberService) RemoveMember(ctx context.Context, id string) (bool, error) {
	return s.repo.Delete(ctx, id)
}
