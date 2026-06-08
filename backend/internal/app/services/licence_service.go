package services

import (
	"clubmanager/internal/adapters/api/grpc/dto"
	"clubmanager/internal/domain"
	"clubmanager/internal/domain/licences"
	"context"
)

type LicenceService interface {
	CreateLicence(context.Context, *dto.CreateLicenceRequest) (*dto.CreateLicenceResponse, error)
	GetMemberLicences(context.Context, *dto.GetMemberLicencesRequest) (*dto.GetMemberLicencesResponse, error)
	UpdateLicenceStatus(context.Context, *dto.UpdateLicenceStatusRequest) (*dto.UpdateLicenceStatusResponse, error)
}

type LicenceServiceConfig struct {
	Repository domain.Repository[licences.Licence, string]
}

type licenceService struct {
	repo domain.Repository[licences.Licence, string]
}

func NewLicenceService(config LicenceServiceConfig) *licenceService {
	return &licenceService{repo: config.Repository}
}

func (s *licenceService) CreateLicence(ctx context.Context, data *dto.CreateLicenceRequest) (*dto.CreateLicenceResponse, error) {
	l, errs := licences.NewLicence(data.MemberId, data.Map())
	if len(errs) > 0 {
		return &dto.CreateLicenceResponse{Errors: errs}, nil
	}

	created, err := s.repo.Save(ctx, l)
	if err != nil {
		return nil, err
	}
	return &dto.CreateLicenceResponse{Licence: created, Errors: make(map[string]string)}, nil
}

func (s *licenceService) GetMemberLicences(ctx context.Context, data *dto.GetMemberLicencesRequest) (*dto.GetMemberLicencesResponse, error) {
	list, err := s.repo.Search(ctx, &domain.SearchParams{
		Fields:    map[string]any{"member_id": data.MemberId},
		Keys:      map[string]bool{"member_id": true},
		Connector: "AND",
	})
	if err != nil {
		return nil, err
	}
	return &dto.GetMemberLicencesResponse{Licences: list, Errors: make(map[string]string)}, nil
}

func (s *licenceService) UpdateLicenceStatus(ctx context.Context, data *dto.UpdateLicenceStatusRequest) (*dto.UpdateLicenceStatusResponse, error) {
	current, err := s.repo.Find(ctx, data.Id)
	if err != nil {
		return nil, err
	}
	if current == nil {
		return &dto.UpdateLicenceStatusResponse{Errors: map[string]string{"licence": "Licence not found."}}, nil
	}

	l, errs := current.UpdateStatus(data.Status)
	if len(errs) > 0 {
		return &dto.UpdateLicenceStatusResponse{Errors: errs}, nil
	}

	updated, err := s.repo.Save(ctx, l)
	if err != nil {
		return nil, err
	}
	return &dto.UpdateLicenceStatusResponse{Licence: updated, Errors: make(map[string]string)}, nil
}
