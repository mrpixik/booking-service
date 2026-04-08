package user

import (
	"context"
	"testing"
	"time"

	"github.com/avito-internships/test-backend-1-mrpixik/internal/domain/dto"
	"github.com/avito-internships/test-backend-1-mrpixik/internal/domain/dto/errors/repository"
	"github.com/avito-internships/test-backend-1-mrpixik/internal/domain/dto/errors/service"
	"github.com/avito-internships/test-backend-1-mrpixik/internal/domain/model"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

func TestRegister(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		req       dto.UserRegisterRequest
		setupRepo func(*mockUserRepo)
		wantErr   error
		wantEmail string
		wantRole  model.Role
	}{
		{
			name: "success user role",
			req:  dto.UserRegisterRequest{Email: "test@test.com", Password: "pass123", Role: "user"},
			setupRepo: func(r *mockUserRepo) {
				r.On("Create", mock.Anything, mock.Anything).Return(&model.User{
					ID: uuid.New().String(), Email: "test@test.com", Role: "user", CreatedAt: time.Now(),
				}, nil)
			},
			wantEmail: "test@test.com",
			wantRole:  "user",
		},
		{
			name: "success admin role",
			req:  dto.UserRegisterRequest{Email: "admin@test.com", Password: "pass123", Role: "admin"},
			setupRepo: func(r *mockUserRepo) {
				r.On("Create", mock.Anything, mock.Anything).Return(&model.User{
					ID: uuid.New().String(), Email: "admin@test.com", Role: "admin",
				}, nil)
			},
			wantEmail: "admin@test.com",
			wantRole:  "admin",
		},
		{
			name:    "empty email",
			req:     dto.UserRegisterRequest{Email: "", Password: "pass123", Role: "user"},
			wantErr: service.ErrEmptyCredentials,
		},
		{
			name:    "empty password",
			req:     dto.UserRegisterRequest{Email: "test@test.com", Password: "", Role: "user"},
			wantErr: service.ErrEmptyCredentials,
		},
		{
			name:    "invalid role",
			req:     dto.UserRegisterRequest{Email: "test@test.com", Password: "pass123", Role: "superadmin"},
			wantErr: service.ErrEmptyCredentials,
		},
		{
			name: "email taken",
			req:  dto.UserRegisterRequest{Email: "taken@test.com", Password: "pass123", Role: "user"},
			setupRepo: func(r *mockUserRepo) {
				r.On("Create", mock.Anything, mock.Anything).Return(nil, repository.ErrEmailExists)
			},
			wantErr: service.ErrEmailTaken,
		},
		{
			name: "repo internal error",
			req:  dto.UserRegisterRequest{Email: "test@test.com", Password: "pass123", Role: "user"},
			setupRepo: func(r *mockUserRepo) {
				r.On("Create", mock.Anything, mock.Anything).Return(nil, repository.ErrInternalError)
			},
			wantErr: service.ErrInternalError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var repo *mockUserRepo
			if tt.setupRepo != nil {
				repo = newMockUserRepo(t)
				tt.setupRepo(repo)
			}

			svc := NewAuthService(repo, nil)
			result, err := svc.Register(context.Background(), tt.req)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantEmail, result.Email)
				assert.Equal(t, tt.wantRole, result.Role)
			}
		})
	}
}

func TestLogin(t *testing.T) {
	t.Parallel()

	correctHash, _ := bcrypt.GenerateFromPassword([]byte("pass123"), bcrypt.DefaultCost)
	wrongHash, _ := bcrypt.GenerateFromPassword([]byte("correct"), bcrypt.DefaultCost)

	tests := []struct {
		name      string
		req       dto.UserLoginRequest
		setupRepo func(*mockUserRepo)
		setupJWT  func(*mockJwtManager)
		wantErr   error
		wantToken string
	}{
		{
			name: "success",
			req:  dto.UserLoginRequest{Email: "test@test.com", Password: "pass123"},
			setupRepo: func(r *mockUserRepo) {
				r.On("GetByEmail", mock.Anything, "test@test.com").Return(&model.User{
					ID: uuid.New().String(), Email: "test@test.com", Role: "user", PasswordHash: string(correctHash),
				}, nil)
			},
			setupJWT: func(j *mockJwtManager) {
				j.On("GenerateToken", mock.Anything, "user").Return("jwt-token", nil)
			},
			wantToken: "jwt-token",
		},
		{
			name:    "empty email",
			req:     dto.UserLoginRequest{Email: "", Password: "pass123"},
			wantErr: service.ErrEmptyCredentials,
		},
		{
			name:    "empty password",
			req:     dto.UserLoginRequest{Email: "test@test.com", Password: ""},
			wantErr: service.ErrEmptyCredentials,
		},
		{
			name: "user not found",
			req:  dto.UserLoginRequest{Email: "noone@test.com", Password: "pass123"},
			setupRepo: func(r *mockUserRepo) {
				r.On("GetByEmail", mock.Anything, mock.Anything).Return(nil, repository.ErrNotFound)
			},
			wantErr: service.ErrInvalidCredentials,
		},
		{
			name: "wrong password",
			req:  dto.UserLoginRequest{Email: "test@test.com", Password: "wrong"},
			setupRepo: func(r *mockUserRepo) {
				r.On("GetByEmail", mock.Anything, mock.Anything).Return(&model.User{
					PasswordHash: string(wrongHash),
				}, nil)
			},
			wantErr: service.ErrInvalidCredentials,
		},
		{
			name: "jwt error",
			req:  dto.UserLoginRequest{Email: "test@test.com", Password: "pass123"},
			setupRepo: func(r *mockUserRepo) {
				r.On("GetByEmail", mock.Anything, mock.Anything).Return(&model.User{
					ID: "id", Role: "user", PasswordHash: string(correctHash),
				}, nil)
			},
			setupJWT: func(j *mockJwtManager) {
				j.On("GenerateToken", mock.Anything, mock.Anything).Return("", assert.AnError)
			},
			wantErr: service.ErrInternalError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var repo *mockUserRepo
			var jwt *mockJwtManager

			if tt.setupRepo != nil {
				repo = newMockUserRepo(t)
				tt.setupRepo(repo)
			}
			if tt.setupJWT != nil {
				jwt = newMockJwtManager(t)
				tt.setupJWT(jwt)
			}

			svc := NewAuthService(repo, jwt)
			result, err := svc.Login(context.Background(), tt.req)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantToken, result.Token)
			}
		})
	}
}
