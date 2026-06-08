package services

import (
	"clubmanager/internal/adapters/api/grpc/dto"
	"clubmanager/internal/domain"
	"clubmanager/internal/domain/clubs"
	"context"
)

type ClubService interface {
	CreateClub(context.Context, *dto.CreateClubRequest) (*dto.CreateClubResponse, error)
	ReadClub(context.Context, *dto.ReadClubRequest) (*dto.ReadClubResponse, error)
	UpdateClub(context.Context, *dto.UpdateClubRequest) (*dto.UpdateClubResponse, error)
	DeleteClub(context.Context, string) (bool, error)
}

type ClubServiceConfig struct {
	Repository domain.Repository[clubs.Club, string]
}

type clubService struct {
	repo domain.Repository[clubs.Club, string]
}

func NewClubService(config ClubServiceConfig) *clubService {
	return &clubService{repo: config.Repository}
}

func (s *clubService) CreateClub(ctx context.Context, data *dto.CreateClubRequest) (*dto.CreateClubResponse, error) {
	c, errs := clubs.NewClub(data.Map())
	if len(errs) > 0 {
		return &dto.CreateClubResponse{Errors: errs}, nil
	}

	list, err := s.repo.Search(ctx, &domain.SearchParams{
		Fields:    map[string]any{"siren": c.Siren},
		Keys:      map[string]bool{"siren": true},
		Connector: "AND",
	})
	if err != nil {
		return nil, err
	}
	if len(list) > 0 {
		return &dto.CreateClubResponse{
			Errors: map[string]string{"siren": "A club with this SIREN already exists."},
		}, nil
	}

	created, err := s.repo.Save(ctx, c)
	if err != nil {
		return nil, err
	}
	return &dto.CreateClubResponse{Club: created, Errors: make(map[string]string)}, nil
}

func (s *clubService) ReadClub(ctx context.Context, data *dto.ReadClubRequest) (*dto.ReadClubResponse, error) {
	list, err := s.repo.Search(ctx, &domain.SearchParams{
		Fields:    data.Params,
		Keys:      map[string]bool{"siren": true, "name": true, "city": true, "postal_code": true},
		Connector: "AND",
	})
	if err != nil {
		return nil, err
	}
	return &dto.ReadClubResponse{Clubs: list, Errors: make(map[string]string)}, nil
}

func (s *clubService) UpdateClub(ctx context.Context, data *dto.UpdateClubRequest) (*dto.UpdateClubResponse, error) {
	current, err := s.repo.Find(ctx, data.Id)
	if err != nil {
		return nil, err
	}

	c, errs := current.Update(data.Map())
	if len(errs) > 0 {
		return &dto.UpdateClubResponse{Errors: errs}, nil
	}

	updated, err := s.repo.Save(ctx, c)
	if err != nil {
		return nil, err
	}
	return &dto.UpdateClubResponse{Club: updated, Errors: make(map[string]string)}, nil
}

func (s *clubService) DeleteClub(ctx context.Context, id string) (bool, error) {
	return s.repo.Delete(ctx, id)
}
