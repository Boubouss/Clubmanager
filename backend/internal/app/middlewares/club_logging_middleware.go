package middlewares

import (
	"clubmanager/internal/adapters/api/grpc/dto"
	"clubmanager/internal/app/services"
	"context"
	"fmt"
	"time"
)

type clubLoggingService struct {
	next services.ClubService
}

func NewClubLoggingService(next services.ClubService) services.ClubService {
	return &clubLoggingService{next: next}
}

func (s *clubLoggingService) CreateClub(ctx context.Context, data *dto.CreateClubRequest) (club *dto.CreateClubResponse, err error) {
	defer func(begin time.Time) {
		fmt.Printf("=> type: '%s'; took: '%v'; err: '%v'.\n", "CreateClub", time.Since(begin), err)
	}(time.Now())
	return s.next.CreateClub(ctx, data)
}

func (s *clubLoggingService) ReadClub(ctx context.Context, data *dto.ReadClubRequest) (club *dto.ReadClubResponse, err error) {
	defer func(begin time.Time) {
		fmt.Printf("=> type: '%s'; took: '%v'; err: '%v'.\n", "ReadClub", time.Since(begin), err)
	}(time.Now())
	return s.next.ReadClub(ctx, data)
}

func (s *clubLoggingService) UpdateClub(ctx context.Context, data *dto.UpdateClubRequest) (club *dto.UpdateClubResponse, err error) {
	defer func(begin time.Time) {
		fmt.Printf("=> type: '%s'; took: '%v'; err: '%v'.\n", "UpdateClub", time.Since(begin), err)
	}(time.Now())
	return s.next.UpdateClub(ctx, data)
}

func (s *clubLoggingService) DeleteClub(ctx context.Context, id string) (ok bool, err error) {
	defer func(begin time.Time) {
		fmt.Printf("=> type: '%s'; took: '%v'; err: '%v'.\n", "DeleteClub", time.Since(begin), err)
	}(time.Now())
	return s.next.DeleteClub(ctx, id)
}
