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

// ── Mocks ─────────────────────────────────────────────────────────────────────

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

type mockClubMembershipRepository struct {
	saveFunc        func(context.Context, *members.ClubMembership) (*members.ClubMembership, error)
	findFunc        func(context.Context, string) (*members.ClubMembership, error)
	findByMemberFunc func(context.Context, string) ([]*members.ClubMembership, error)
	findByClubFunc  func(context.Context, string) ([]*members.ClubMembership, error)
	deleteFunc      func(context.Context, string) (bool, error)
}

func (m mockClubMembershipRepository) Save(ctx context.Context, cm *members.ClubMembership) (*members.ClubMembership, error) {
	return m.saveFunc(ctx, cm)
}
func (m mockClubMembershipRepository) Find(ctx context.Context, id string) (*members.ClubMembership, error) {
	return m.findFunc(ctx, id)
}
func (m mockClubMembershipRepository) FindByMember(ctx context.Context, memberId string) ([]*members.ClubMembership, error) {
	return m.findByMemberFunc(ctx, memberId)
}
func (m mockClubMembershipRepository) FindByClub(ctx context.Context, clubId string) ([]*members.ClubMembership, error) {
	return m.findByClubFunc(ctx, clubId)
}
func (m mockClubMembershipRepository) Delete(ctx context.Context, id string) (bool, error) {
	return m.deleteFunc(ctx, id)
}
func (m mockClubMembershipRepository) FindByClubWithContacts(ctx context.Context, clubId string, page, pageSize int) ([]*members.MemberContact, int, error) {
	return nil, 0, nil
}

// ── Fixtures ──────────────────────────────────────────────────────────────────

var (
	memberUUID      = uuid.MustParse("00000000-0000-0000-0000-000000000020")
	memberUser      = uuid.MustParse("00000000-0000-0000-0000-000000000001")
	memberClub      = uuid.MustParse("00000000-0000-0000-0000-000000000002")
	membershipUUID  = uuid.MustParse("00000000-0000-0000-0000-000000000030")
	testMember      = &members.Member{
		Id:        memberUUID,
		UserId:    memberUser,
		Firstname: "Jean",
		Lastname:  "Dupont",
		Birthdate: "2000-06-15",
		Gender:    "man",
		IsPrimary: false,
	}
	testMembership = &members.ClubMembership{
		Id:       membershipUUID,
		MemberId: memberUUID,
		ClubId:   memberClub,
		IsValid:  false,
	}
)

func newTestMemberService(memberRepo domain.Repository[members.Member, string], msRepo ClubMembershipRepository) *memberService {
	return NewMemberService(MemberServiceConfig{
		Repository:           memberRepo,
		MembershipRepository: msRepo,
	})
}

func newTestMemberServiceWithChecker(memberRepo domain.Repository[members.Member, string], msRepo ClubMembershipRepository, checker mockClubStatusChecker) *memberService {
	return NewMemberService(MemberServiceConfig{
		Repository:           memberRepo,
		MembershipRepository: msRepo,
		ClubStatusChecker:    checker,
	})
}

func noopMembershipRepo() mockClubMembershipRepository {
	return mockClubMembershipRepository{
		saveFunc: func(_ context.Context, cm *members.ClubMembership) (*members.ClubMembership, error) {
			cm.Id = membershipUUID
			return cm, nil
		},
		findFunc:        func(_ context.Context, _ string) (*members.ClubMembership, error) { return nil, nil },
		findByMemberFunc: func(_ context.Context, _ string) ([]*members.ClubMembership, error) { return nil, nil },
		findByClubFunc:  func(_ context.Context, _ string) ([]*members.ClubMembership, error) { return nil, nil },
		deleteFunc:      func(_ context.Context, _ string) (bool, error) { return true, nil },
	}
}

// ── Tests ─────────────────────────────────────────────────────────────────────

func TestAddMember(t *testing.T) {
	tests := []struct {
		name        string
		req         *dto.AddMemberRequest
		memberSave  error
		membershipSave error
		wantErrors  bool
		wantErr     bool
	}{
		{
			name: "Valid member",
			req: &dto.AddMemberRequest{
				UserId: memberUser.String(), ClubId: memberClub.String(),
				Firstname: "Jean", Lastname: "Dupont", Birthdate: "2000-06-15", Gender: "man",
			},
		},
		{
			name: "Valid primary member",
			req: &dto.AddMemberRequest{
				UserId: memberUser.String(), ClubId: memberClub.String(),
				Firstname: "Jean", Lastname: "Dupont", Birthdate: "2000-06-15", Gender: "man",
				IsPrimary: true,
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
			name: "Invalid club ID — rejected before DB write",
			req: &dto.AddMemberRequest{
				UserId: memberUser.String(), ClubId: "not-a-uuid",
				Firstname: "Jean", Lastname: "Dupont", Birthdate: "2000-06-15", Gender: "man",
			},
			wantErrors: true,
		},
		{
			name: "Repository error on member save",
			req: &dto.AddMemberRequest{
				UserId: memberUser.String(), ClubId: memberClub.String(),
				Firstname: "Jean", Lastname: "Dupont", Birthdate: "2000-06-15", Gender: "man",
			},
			memberSave: errors.New("db error"),
			wantErr:    true,
		},
		{
			name: "Repository error on membership save",
			req: &dto.AddMemberRequest{
				UserId: memberUser.String(), ClubId: memberClub.String(),
				Firstname: "Jean", Lastname: "Dupont", Birthdate: "2000-06-15", Gender: "man",
			},
			membershipSave: errors.New("db error"),
			wantErr:        true,
		},
	}

	// Test pending-club blocking separately (requires ClubStatusChecker injection).
	t.Run("Pending club blocks member creation", func(t *testing.T) {
		memberRepo := mockMemberRepository{}
		checker := mockClubStatusChecker{
			isActiveFunc: func(_ context.Context, _ string) (bool, error) { return false, nil },
		}
		svc := newTestMemberServiceWithChecker(memberRepo, noopMembershipRepo(), checker)
		res, err := svc.AddMember(context.Background(), &dto.AddMemberRequest{
			UserId: memberUser.String(), ClubId: memberClub.String(),
			Firstname: "Jean", Lastname: "Dupont", Birthdate: "2000-06-15", Gender: "man",
		})
		assert.NoError(t, err)
		assert.True(t, len(res.Errors) > 0)
		assert.Nil(t, res.Member)
	})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			memberRepo := mockMemberRepository{
				saveFunc: func(_ context.Context, m *members.Member) (*members.Member, error) {
					if tt.memberSave != nil {
						return nil, tt.memberSave
					}
					m.Id = memberUUID
					return m, nil
				},
			}
			msRepo := noopMembershipRepo()
			if tt.membershipSave != nil {
				msRepo.saveFunc = func(_ context.Context, _ *members.ClubMembership) (*members.ClubMembership, error) {
					return nil, tt.membershipSave
				}
			}

			svc := newTestMemberService(memberRepo, msRepo)
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
				assert.NotNil(t, res.Membership)
				assert.False(t, res.Membership.IsValid)
				assert.Equal(t, tt.req.IsPrimary, res.Member.IsPrimary)
			}
		})
	}
}

func TestValidateMember(t *testing.T) {
	tests := []struct {
		name       string
		findMs     *members.ClubMembership
		findMsErr  error
		saveMsErr  error
		findMember *members.Member
		wantErrors bool
		wantErr    bool
	}{
		{
			name:       "Valid — membership approved",
			findMs:     testMembership,
			findMember: testMember,
		},
		{
			name:       "Membership not found",
			findMs:     nil,
			wantErrors: true,
		},
		{
			name:      "Repository error on Find",
			findMsErr: errors.New("db error"),
			wantErr:   true,
		},
		{
			name:      "Repository error on Save",
			findMs:    testMembership,
			saveMsErr: errors.New("db error"),
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			memberRepo := mockMemberRepository{
				findFunc: func(_ context.Context, _ string) (*members.Member, error) {
					return tt.findMember, nil
				},
			}
			msRepo := mockClubMembershipRepository{
				findFunc: func(_ context.Context, _ string) (*members.ClubMembership, error) {
					return tt.findMs, tt.findMsErr
				},
				saveFunc: func(_ context.Context, cm *members.ClubMembership) (*members.ClubMembership, error) {
					if tt.saveMsErr != nil {
						return nil, tt.saveMsErr
					}
					return cm, nil
				},
				findByMemberFunc: func(_ context.Context, _ string) ([]*members.ClubMembership, error) { return nil, nil },
				findByClubFunc:   func(_ context.Context, _ string) ([]*members.ClubMembership, error) { return nil, nil },
				deleteFunc:       func(_ context.Context, _ string) (bool, error) { return true, nil },
			}

			svc := newTestMemberService(memberRepo, msRepo)
			res, err := svc.ValidateMember(context.Background(), &dto.ValidateMemberRequest{Id: membershipUUID.String()})

			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			if tt.wantErrors {
				assert.True(t, len(res.Errors) > 0)
			} else {
				assert.Empty(t, res.Errors)
				assert.NotNil(t, res.Membership)
				assert.True(t, res.Membership.IsValid)
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
			memberRepo := mockMemberRepository{
				findFunc: func(_ context.Context, _ string) (*members.Member, error) {
					return tt.findResp, tt.findErr
				},
				saveFunc: func(_ context.Context, m *members.Member) (*members.Member, error) {
					return m, nil
				},
			}

			svc := newTestMemberService(memberRepo, noopMembershipRepo())
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
			memberRepo := mockMemberRepository{
				deleteFunc: func(_ context.Context, _ string) (bool, error) {
					return tt.deleteResp, tt.deleteErr
				},
			}

			svc := newTestMemberService(memberRepo, noopMembershipRepo())
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
		name              string
		req               *dto.GetMembersByUserRequest
		searchResp        []*members.Member
		searchErr         error
		findByMemberResp  []*members.ClubMembership
		wantCount         int
		wantErr           bool
	}{
		{
			name:       "Without club filter — returns all members",
			req:        &dto.GetMembersByUserRequest{UserId: memberUser.String()},
			searchResp: []*members.Member{testMember},
			wantCount:  1,
		},
		{
			name:       "With club filter — member has matching membership",
			req:        &dto.GetMembersByUserRequest{UserId: memberUser.String(), ClubId: memberClub.String()},
			searchResp: []*members.Member{testMember},
			findByMemberResp: []*members.ClubMembership{testMembership},
			wantCount:  1,
		},
		{
			name:             "With club filter — member has no matching membership",
			req:              &dto.GetMembersByUserRequest{UserId: memberUser.String(), ClubId: uuid.New().String()},
			searchResp:       []*members.Member{testMember},
			findByMemberResp: []*members.ClubMembership{testMembership},
			wantCount:        0,
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
			memberRepo := mockMemberRepository{
				searchFunc: func(_ context.Context, _ *domain.SearchParams) ([]*members.Member, error) {
					return tt.searchResp, tt.searchErr
				},
			}
			msRepo := noopMembershipRepo()
			msRepo.findByMemberFunc = func(_ context.Context, _ string) ([]*members.ClubMembership, error) {
				return tt.findByMemberResp, nil
			}

			svc := newTestMemberService(memberRepo, msRepo)
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
