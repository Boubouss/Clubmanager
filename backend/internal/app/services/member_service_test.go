package services

import (
	"clubmanager/internal/adapters/api/grpc/dto"
	"clubmanager/internal/domain"
	"clubmanager/internal/domain/members"
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

type mockMemberRepository struct {
	saveFunc   func(context.Context, *members.Member) (*members.Member, error)
	findFunc   func(context.Context, string) (*members.Member, error)
	searchFunc func(context.Context, *domain.SearchParams) ([]*members.Member, error)
	deleteFunc func(context.Context, string) (bool, error)
}

func (m mockMemberRepository) Save(ctx context.Context, mb *members.Member) (*members.Member, error) {
	return m.saveFunc(ctx, mb)
}
func (m mockMemberRepository) Find(ctx context.Context, id string) (*members.Member, error) {
	return m.findFunc(ctx, id)
}
func (m mockMemberRepository) Search(ctx context.Context, p *domain.SearchParams) ([]*members.Member, error) {
	return m.searchFunc(ctx, p)
}
func (m mockMemberRepository) Delete(ctx context.Context, id string) (bool, error) {
	return m.deleteFunc(ctx, id)
}

var (
	memberUUID = uuid.MustParse("00000000-0000-0000-0000-000000000020")
	memberUser = uuid.MustParse("00000000-0000-0000-0000-000000000001")
	memberClub = uuid.MustParse("00000000-0000-0000-0000-000000000002")
	testMember = &members.Member{
		Id:        memberUUID,
		UserId:    memberUser,
		ClubId:    memberClub,
		Firstname: "Jean",
		Lastname:  "Dupont",
		Birthdate: "2000-06-15",
		Gender:    "man",
		IsValid:   false,
	}
)

func newTestMemberService(repo domain.Repository[members.Member, string]) *memberService {
	return NewMemberService(MemberServiceConfig{Repository: repo})
}

func TestAddMember(t *testing.T) {
	tests := []struct {
		name       string
		req        *dto.AddMemberRequest
		saveErr    error
		wantErrors bool
		wantErr    bool
	}{
		{
			name: "Valid member",
			req: &dto.AddMemberRequest{
				UserId: memberUser.String(), ClubId: memberClub.String(),
				Firstname: "Jean", Lastname: "Dupont", Birthdate: "2000-06-15", Gender: "man",
			},
		},
		{
			name: "Missing firstname",
			req: &dto.AddMemberRequest{
				UserId: memberUser.String(), ClubId: memberClub.String(),
				Lastname: "Dupont", Birthdate: "2000-06-15", Gender: "man",
			},
			wantErrors: true,
		},
		{
			name: "Invalid gender",
			req: &dto.AddMemberRequest{
				UserId: memberUser.String(), ClubId: memberClub.String(),
				Firstname: "Jean", Lastname: "Dupont", Birthdate: "2000-06-15", Gender: "unknown",
			},
			wantErrors: true,
		},
		{
			name: "Invalid birthdate (future)",
			req: &dto.AddMemberRequest{
				UserId: memberUser.String(), ClubId: memberClub.String(),
				Firstname: "Jean", Lastname: "Dupont", Birthdate: "2090-06-15", Gender: "man",
			},
			wantErrors: true,
		},
		{
			name: "Repository error",
			req: &dto.AddMemberRequest{
				UserId: memberUser.String(), ClubId: memberClub.String(),
				Firstname: "Jean", Lastname: "Dupont", Birthdate: "2000-06-15", Gender: "man",
			},
			saveErr: errors.New("db error"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := mockMemberRepository{
				saveFunc: func(_ context.Context, m *members.Member) (*members.Member, error) {
					if tt.saveErr != nil {
						return nil, tt.saveErr
					}
					m.Id = memberUUID
					return m, nil
				},
			}

			svc := newTestMemberService(repo)
			res, err := svc.AddMember(context.Background(), tt.req)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			if tt.wantErrors {
				assert.True(t, len(res.Errors) > 0)
				assert.Nil(t, res.Member)
			} else {
				assert.Empty(t, res.Errors)
				assert.NotNil(t, res.Member)
				assert.False(t, res.Member.IsValid)
			}
		})
	}
}

func TestValidateMember(t *testing.T) {
	tests := []struct {
		name       string
		findResp   *members.Member
		findErr    error
		saveErr    error
		wantErrors bool
		wantErr    bool
	}{
		{
			name:     "Valid — is_valid set to true",
			findResp: testMember,
		},
		{
			name:       "Member not found",
			findResp:   nil,
			wantErrors: true,
		},
		{
			name:    "Repository error on Find",
			findErr: errors.New("db error"),
			wantErr: true,
		},
		{
			name:     "Repository error on Save",
			findResp: testMember,
			saveErr:  errors.New("db error"),
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := mockMemberRepository{
				findFunc: func(_ context.Context, _ string) (*members.Member, error) {
					return tt.findResp, tt.findErr
				},
				saveFunc: func(_ context.Context, m *members.Member) (*members.Member, error) {
					if tt.saveErr != nil {
						return nil, tt.saveErr
					}
					return m, nil
				},
			}

			svc := newTestMemberService(repo)
			res, err := svc.ValidateMember(context.Background(), &dto.ValidateMemberRequest{Id: memberUUID.String()})

			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			if tt.wantErrors {
				assert.True(t, len(res.Errors) > 0)
			} else {
				assert.Empty(t, res.Errors)
				assert.NotNil(t, res.Member)
				assert.True(t, res.Member.IsValid)
			}
		})
	}
}

func TestUpdateMember(t *testing.T) {
	tests := []struct {
		name       string
		req        *dto.UpdateMemberRequest
		findResp   *members.Member
		findErr    error
		wantErrors bool
		wantErr    bool
	}{
		{
			name:     "Valid update",
			req:      &dto.UpdateMemberRequest{Id: memberUUID.String(), Firstname: "Pierre"},
			findResp: testMember,
		},
		{
			name:       "Invalid gender",
			req:        &dto.UpdateMemberRequest{Id: memberUUID.String(), Gender: "invalid"},
			findResp:   testMember,
			wantErrors: true,
		},
		{
			name:       "Member not found",
			req:        &dto.UpdateMemberRequest{Id: memberUUID.String()},
			findResp:   nil,
			wantErrors: true,
		},
		{
			name:    "Repository error",
			req:     &dto.UpdateMemberRequest{Id: memberUUID.String()},
			findErr: errors.New("db error"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := mockMemberRepository{
				findFunc: func(_ context.Context, _ string) (*members.Member, error) {
					return tt.findResp, tt.findErr
				},
				saveFunc: func(_ context.Context, m *members.Member) (*members.Member, error) {
					return m, nil
				},
			}

			svc := newTestMemberService(repo)
			res, err := svc.UpdateMember(context.Background(), tt.req)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			if tt.wantErrors {
				assert.True(t, len(res.Errors) > 0)
			} else {
				assert.Empty(t, res.Errors)
				assert.NotNil(t, res.Member)
			}
		})
	}
}

func TestRemoveMember(t *testing.T) {
	tests := []struct {
		name       string
		deleteResp bool
		deleteErr  error
		wantOk     bool
		wantErr    bool
	}{
		{"Valid delete", true, nil, true, false},
		{"Repository error", false, errors.New("db error"), false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := mockMemberRepository{
				deleteFunc: func(_ context.Context, _ string) (bool, error) {
					return tt.deleteResp, tt.deleteErr
				},
			}

			svc := newTestMemberService(repo)
			ok, err := svc.RemoveMember(context.Background(), memberUUID.String())

			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.wantOk, ok)
		})
	}
}

func TestGetMembersByUser(t *testing.T) {
	tests := []struct {
		name       string
		req        *dto.GetMembersByUserRequest
		searchResp []*members.Member
		searchErr  error
		wantCount  int
		wantErr    bool
	}{
		{
			name:       "Without club filter",
			req:        &dto.GetMembersByUserRequest{UserId: memberUser.String()},
			searchResp: []*members.Member{testMember},
			wantCount:  1,
		},
		{
			name:       "With club filter",
			req:        &dto.GetMembersByUserRequest{UserId: memberUser.String(), ClubId: memberClub.String()},
			searchResp: []*members.Member{testMember},
			wantCount:  1,
		},
		{
			name:       "No results",
			req:        &dto.GetMembersByUserRequest{UserId: memberUser.String()},
			searchResp: []*members.Member{},
			wantCount:  0,
		},
		{
			name:      "Repository error",
			req:       &dto.GetMembersByUserRequest{UserId: memberUser.String()},
			searchErr: errors.New("db error"),
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := mockMemberRepository{
				searchFunc: func(_ context.Context, _ *domain.SearchParams) ([]*members.Member, error) {
					return tt.searchResp, tt.searchErr
				},
			}

			svc := newTestMemberService(repo)
			res, err := svc.GetMembersByUser(context.Background(), tt.req)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Len(t, res.Members, tt.wantCount)
		})
	}
}
