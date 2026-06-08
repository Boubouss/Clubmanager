package middlewares

import (
	"clubmanager/internal/adapters/api/grpc/dto"
	"clubmanager/internal/app/services"
	"context"
	"fmt"
	"time"
)

type licenceLoggingService struct {
	next services.LicenceService
}

func NewLicenceLoggingService(next services.LicenceService) services.LicenceService {
	return &licenceLoggingService{next: next}
}

func (s *licenceLoggingService) CreateLicence(ctx context.Context, data *dto.CreateLicenceRequest) (res *dto.CreateLicenceResponse, err error) {
	defer func(begin time.Time) {
		fmt.Printf("=> type: '%s'; took: '%v'; err: '%v'.\n", "CreateLicence", time.Since(begin), err)
	}(time.Now())
	return s.next.CreateLicence(ctx, data)
}

func (s *licenceLoggingService) GetMemberLicences(ctx context.Context, data *dto.GetMemberLicencesRequest) (res *dto.GetMemberLicencesResponse, err error) {
	defer func(begin time.Time) {
		fmt.Printf("=> type: '%s'; took: '%v'; err: '%v'.\n", "GetMemberLicences", time.Since(begin), err)
	}(time.Now())
	return s.next.GetMemberLicences(ctx, data)
}

func (s *licenceLoggingService) UpdateLicenceStatus(ctx context.Context, data *dto.UpdateLicenceStatusRequest) (res *dto.UpdateLicenceStatusResponse, err error) {
	defer func(begin time.Time) {
		fmt.Printf("=> type: '%s'; took: '%v'; err: '%v'.\n", "UpdateLicenceStatus", time.Since(begin), err)
	}(time.Now())
	return s.next.UpdateLicenceStatus(ctx, data)
}
