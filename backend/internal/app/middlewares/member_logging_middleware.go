package middlewares

import (
	"clubmanager/internal/adapters/api/grpc/dto"
	"clubmanager/internal/app/services"
	"clubmanager/internal/domain/members"
	"context"
	"fmt"
	"time"
)

type memberLoggingService struct {
	next services.MemberService
}

func NewMemberLoggingService(next services.MemberService) services.MemberService {
	return &memberLoggingService{next: next}
}

func (s *memberLoggingService) CreateMember(ctx context.Context, data *dto.CreateMemberRequest) (res *dto.CreateMemberResponse, err error) {
	defer func(begin time.Time) {
		fmt.Printf("=> type: '%s'; took: '%v'; err: '%v'.\n", "CreateMember", time.Since(begin), err)
	}(time.Now())
	return s.next.CreateMember(ctx, data)
}

func (s *memberLoggingService) RequestMembership(ctx context.Context, data *dto.RequestMembershipRequest) (res *dto.RequestMembershipResponse, err error) {
	defer func(begin time.Time) {
		fmt.Printf("=> type: '%s'; took: '%v'; err: '%v'.\n", "RequestMembership", time.Since(begin), err)
	}(time.Now())
	return s.next.RequestMembership(ctx, data)
}

func (s *memberLoggingService) AddMembership(ctx context.Context, memberId, clubId string) (res *members.ClubMembership, err error) {
	defer func(begin time.Time) {
		fmt.Printf("=> type: '%s'; took: '%v'; err: '%v'.\n", "AddMembership", time.Since(begin), err)
	}(time.Now())
	return s.next.AddMembership(ctx, memberId, clubId)
}

func (s *memberLoggingService) AddMember(ctx context.Context, data *dto.AddMemberRequest) (res *dto.AddMemberResponse, err error) {
	defer func(begin time.Time) {
		fmt.Printf("=> type: '%s'; took: '%v'; err: '%v'.\n", "AddMember", time.Since(begin), err)
	}(time.Now())
	return s.next.AddMember(ctx, data)
}

func (s *memberLoggingService) ValidateMember(ctx context.Context, data *dto.ValidateMemberRequest) (res *dto.ValidateMemberResponse, err error) {
	defer func(begin time.Time) {
		fmt.Printf("=> type: '%s'; took: '%v'; err: '%v'.\n", "ValidateMember", time.Since(begin), err)
	}(time.Now())
	return s.next.ValidateMember(ctx, data)
}

func (s *memberLoggingService) GetMember(ctx context.Context, memberId string) (res *members.Member, err error) {
	defer func(begin time.Time) {
		fmt.Printf("=> type: '%s'; took: '%v'; err: '%v'.\n", "GetMember", time.Since(begin), err)
	}(time.Now())
	return s.next.GetMember(ctx, memberId)
}

func (s *memberLoggingService) GetMembersByIds(ctx context.Context, ids []string) (res map[string]*members.Member, err error) {
	defer func(begin time.Time) {
		fmt.Printf("=> type: '%s'; count: %d; took: '%v'; err: '%v'.\n", "GetMembersByIds", len(ids), time.Since(begin), err)
	}(time.Now())
	return s.next.GetMembersByIds(ctx, ids)
}

func (s *memberLoggingService) GetMembersByUser(ctx context.Context, data *dto.GetMembersByUserRequest) (res *dto.GetMembersByUserResponse, err error) {
	defer func(begin time.Time) {
		fmt.Printf("=> type: '%s'; took: '%v'; err: '%v'.\n", "GetMembersByUser", time.Since(begin), err)
	}(time.Now())
	return s.next.GetMembersByUser(ctx, data)
}

func (s *memberLoggingService) UpdateMember(ctx context.Context, data *dto.UpdateMemberRequest) (res *dto.UpdateMemberResponse, err error) {
	defer func(begin time.Time) {
		fmt.Printf("=> type: '%s'; took: '%v'; err: '%v'.\n", "UpdateMember", time.Since(begin), err)
	}(time.Now())
	return s.next.UpdateMember(ctx, data)
}

func (s *memberLoggingService) RemoveMember(ctx context.Context, id string) (ok bool, err error) {
	defer func(begin time.Time) {
		fmt.Printf("=> type: '%s'; took: '%v'; err: '%v'.\n", "RemoveMember", time.Since(begin), err)
	}(time.Now())
	return s.next.RemoveMember(ctx, id)
}

func (s *memberLoggingService) RemoveMembership(ctx context.Context, id string) (ok bool, err error) {
	defer func(begin time.Time) {
		fmt.Printf("=> type: '%s'; took: '%v'; err: '%v'.\n", "RemoveMembership", time.Since(begin), err)
	}(time.Now())
	return s.next.RemoveMembership(ctx, id)
}

func (s *memberLoggingService) GetMembersByClub(ctx context.Context, data *dto.GetMembersByClubRequest) (res *dto.GetMembersByClubResponse, err error) {
	defer func(begin time.Time) {
		fmt.Printf("=> type: '%s'; took: '%v'; err: '%v'.\n", "GetMembersByClub", time.Since(begin), err)
	}(time.Now())
	return s.next.GetMembersByClub(ctx, data)
}

func (s *memberLoggingService) GetStaffContacts(ctx context.Context, clubId string) (res *dto.GetStaffContactsResponse, err error) {
	defer func(begin time.Time) {
		fmt.Printf("=> type: '%s'; took: '%v'; err: '%v'.\n", "GetStaffContacts", time.Since(begin), err)
	}(time.Now())
	return s.next.GetStaffContacts(ctx, clubId)
}

func (s *memberLoggingService) GetUserClubs(ctx context.Context, data *dto.GetUserClubsRequest) (res *dto.GetUserClubsResponse, err error) {
	defer func(begin time.Time) {
		fmt.Printf("=> type: '%s'; took: '%v'; err: '%v'.\n", "GetUserClubs", time.Since(begin), err)
	}(time.Now())
	return s.next.GetUserClubs(ctx, data)
}
