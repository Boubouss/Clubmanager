package services

import (
	"clubmanager/internal/adapters/api/grpc/dto"
	"clubmanager/internal/domain"
	"clubmanager/internal/domain/users"
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// --- Mocks ---

type mockUserRepository struct {
	saveFunc   func(context.Context, *users.User) (*users.User, error)
	findFunc   func(context.Context, string) (*users.User, error)
	searchFunc func(context.Context, *domain.SearchParams) ([]*users.User, error)
	deleteFunc func(context.Context, string) (bool, error)
}

func (m mockUserRepository) Save(ctx context.Context, u *users.User) (*users.User, error) {
	return m.saveFunc(ctx, u)
}
func (m mockUserRepository) Find(ctx context.Context, id string) (*users.User, error) {
	return m.findFunc(ctx, id)
}
func (m mockUserRepository) Search(ctx context.Context, p *domain.SearchParams) ([]*users.User, error) {
	return m.searchFunc(ctx, p)
}
func (m mockUserRepository) Delete(ctx context.Context, id string) (bool, error) {
	return m.deleteFunc(ctx, id)
}

type mockHasher struct {
	hashFunc    func(string) (string, error)
	compareFunc func(plain, hash string) error
}

func (m mockHasher) Hash(s string) (string, error)            { return m.hashFunc(s) }
func (m mockHasher) Compare(plain, hash string) error         { return m.compareFunc(plain, hash) }

type mockTokenManager struct {
	generateFunc func(string) (string, error)
	parseFunc    func(string) (string, error)
}

func (m mockTokenManager) GenerateToken(id string) (string, error) {
	return m.generateFunc(id)
}

func (m mockTokenManager) ParseToken(token string) (string, error) {
	return m.parseFunc(token)
}

// --- Helpers ---

func newTestService(repo domain.Repository[users.User, string], hasher users.PasswordHasher, tkm TokenManager) *userService {
	return NewUserService(UserServiceConfig{
		Repository:   repo,
		Hasher:       hasher,
		TokenManager: tkm,
	})
}

var (
	testUUID    = uuid.MustParse("00000000-0000-0000-0000-000000000001")
	testUser    = &users.User{Id: testUUID, Username: "rbouselh", Email: "test@mail.com", Phonenumber: "0606060606", Password: "hashed"}
	errDatabase = errors.New("database error")
)

// --- TestCreateUser ---

func TestCreateUser(t *testing.T) {
	tests := []struct {
		name       string
		req        *dto.CreateUserRequest
		searchResp []*users.User
		searchErr  error
		wantErrors bool
		wantToken  bool
		wantErr    bool
	}{
		{
			name: "Valid user",
			req:  &dto.CreateUserRequest{Username: "rbouselh", Email: "test@mail.com", Phonenumber: "0606060606", Password: "Password123!"},
			searchResp: []*users.User{},
			wantToken:  true,
		},
		{
			name:       "Invalid email",
			req:        &dto.CreateUserRequest{Username: "rbouselh", Email: "invalid", Phonenumber: "0606060606", Password: "Password123!"},
			wantErrors: true,
		},
		{
			name:       "Invalid password",
			req:        &dto.CreateUserRequest{Username: "rbouselh", Email: "test@mail.com", Phonenumber: "0606060606", Password: "weak"},
			wantErrors: true,
		},
		{
			name:       "Email already exists",
			req:        &dto.CreateUserRequest{Username: "other", Email: "test@mail.com", Phonenumber: "0606060606", Password: "Password123!"},
			searchResp: []*users.User{testUser},
			wantErrors: true,
		},
		{
			name:       "Username already exists",
			req:        &dto.CreateUserRequest{Username: "rbouselh", Email: "other@mail.com", Phonenumber: "0606060606", Password: "Password123!"},
			searchResp: []*users.User{testUser},
			wantErrors: true,
		},
		{
			name:      "Repository error on search",
			req:       &dto.CreateUserRequest{Username: "rbouselh", Email: "test@mail.com", Phonenumber: "0606060606", Password: "Password123!"},
			searchErr: errDatabase,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := mockUserRepository{
				searchFunc: func(_ context.Context, _ *domain.SearchParams) ([]*users.User, error) {
					return tt.searchResp, tt.searchErr
				},
				saveFunc: func(_ context.Context, u *users.User) (*users.User, error) {
					u.Id = testUUID
					return u, nil
				},
			}
			hasher := mockHasher{
				hashFunc: func(s string) (string, error) { return "hashed", nil },
			}
			tkm := mockTokenManager{
				generateFunc: func(id string) (string, error) { return "jwt-token", nil },
			}

			svc := newTestService(repo, hasher, tkm)
			res, err := svc.CreateUser(context.Background(), tt.req)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			if tt.wantErrors {
				assert.True(t, len(res.Errors) > 0)
				assert.Nil(t, res.User)
			} else {
				assert.Empty(t, res.Errors)
				assert.NotNil(t, res.User)
			}
			if tt.wantToken {
				assert.Equal(t, "jwt-token", res.Token)
			}
		})
	}
}

// --- TestDeleteUser ---

func TestDeleteUser(t *testing.T) {
	tests := []struct {
		name       string
		token      string
		parseResp  string
		parseErr   error
		deleteResp bool
		deleteErr  error
		wantOk     bool
		wantErr    bool
	}{
		{
			name:       "Valid token and user deleted",
			token:      "valid-token",
			parseResp:  testUUID.String(),
			deleteResp: true,
			wantOk:     true,
		},
		{
			name:     "Invalid token",
			token:    "bad-token",
			parseErr: errors.New("invalid token"),
			wantErr:  true,
		},
		{
			name:      "Repository error on delete",
			token:     "valid-token",
			parseResp: testUUID.String(),
			deleteErr: errDatabase,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := mockUserRepository{
				deleteFunc: func(_ context.Context, id string) (bool, error) {
					assert.Equal(t, testUUID.String(), id)
					return tt.deleteResp, tt.deleteErr
				},
			}
			tkm := mockTokenManager{
				parseFunc: func(token string) (string, error) {
					return tt.parseResp, tt.parseErr
				},
			}

			svc := newTestService(repo, mockHasher{}, tkm)
			ok, err := svc.DeleteUser(context.Background(), tt.token)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.wantOk, ok)
		})
	}
}

// --- TestLoginUser ---

func TestLoginUser(t *testing.T) {
	tests := []struct {
		name        string
		req         *dto.LoginUserRequest
		searchResp  []*users.User
		searchErr   error
		compareErr  error
		wantErrors  bool
		wantErrorKey string
		wantToken   bool
		wantErr     bool
	}{
		{
			name:       "Valid credentials",
			req:        &dto.LoginUserRequest{Email: "test@mail.com", Password: "Password123!"},
			searchResp: []*users.User{testUser},
			wantToken:  true,
		},
		{
			name:         "Email not found",
			req:          &dto.LoginUserRequest{Email: "unknown@mail.com", Password: "Password123!"},
			searchResp:   []*users.User{},
			wantErrors:   true,
			wantErrorKey: "email",
		},
		{
			name:         "Wrong password",
			req:          &dto.LoginUserRequest{Email: "test@mail.com", Password: "wrongpassword"},
			searchResp:   []*users.User{testUser},
			compareErr:   errors.New("crypto/bcrypt: hashedPassword is not the hash of the given password"),
			wantErrors:   true,
			wantErrorKey: "password",
		},
		{
			name:      "Repository error",
			req:       &dto.LoginUserRequest{Email: "test@mail.com", Password: "Password123!"},
			searchErr: errDatabase,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := mockUserRepository{
				searchFunc: func(_ context.Context, _ *domain.SearchParams) ([]*users.User, error) {
					return tt.searchResp, tt.searchErr
				},
			}
			hasher := mockHasher{
				compareFunc: func(plain, hash string) error { return tt.compareErr },
			}
			tkm := mockTokenManager{
				generateFunc: func(id string) (string, error) { return "jwt-token", nil },
			}

			svc := newTestService(repo, hasher, tkm)
			res, err := svc.LoginUser(context.Background(), tt.req)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			if tt.wantErrors {
				assert.True(t, len(res.Errors) > 0)
				if tt.wantErrorKey != "" {
					_, ok := res.Errors[tt.wantErrorKey]
					assert.True(t, ok, "expected error key %q", tt.wantErrorKey)
				}
				assert.Nil(t, res.User)
			} else {
				assert.Empty(t, res.Errors)
				assert.NotNil(t, res.User)
				assert.Equal(t, testUser.Username, res.User.Username)
			}
			if tt.wantToken {
				assert.Equal(t, "jwt-token", res.Token)
			}
		})
	}
}
