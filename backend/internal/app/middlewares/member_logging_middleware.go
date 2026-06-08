package middlewares

import (
	"clubmanager/internal/adapters/api/grpc/dto"
	"clubmanager/internal/app/services"
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
