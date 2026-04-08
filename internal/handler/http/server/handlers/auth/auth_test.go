package auth

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/avito-internships/test-backend-1-mrpixik/internal/domain/dto"
	"github.com/avito-internships/test-backend-1-mrpixik/internal/domain/dto/errors/service"
	"github.com/avito-internships/test-backend-1-mrpixik/internal/domain/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func jsonBody(v any) *bytes.Buffer {
	b, _ := json.Marshal(v)
	return bytes.NewBuffer(b)
}

func TestDummyLogin(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		body       any
		rawBody    string
		setupJWT   func(*mockJwtManager)
		wantStatus int
		wantToken  string
	}{
		{
			name: "admin",
			body: map[string]string{"role": "admin"},
			setupJWT: func(j *mockJwtManager) {
				j.On("GenerateToken", model.DummyAdminID, "admin").Return("admin-token", nil)
			},
			wantStatus: http.StatusOK,
			wantToken:  "admin-token",
		},
		{
			name: "user",
			body: map[string]string{"role": "user"},
			setupJWT: func(j *mockJwtManager) {
				j.On("GenerateToken", model.DummyUserID, "user").Return("user-token", nil)
			},
			wantStatus: http.StatusOK,
			wantToken:  "user-token",
		},
		{
			name:       "invalid role",
			body:       map[string]string{"role": "superadmin"},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "invalid json",
			rawBody:    "not json",
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "jwt error",
			body: map[string]string{"role": "admin"},
			setupJWT: func(j *mockJwtManager) {
				j.On("GenerateToken", mock.Anything, mock.Anything).Return("", assert.AnError)
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var jwt *mockJwtManager
			if tt.setupJWT != nil {
				jwt = newMockJwtManager(t)
				tt.setupJWT(jwt)
			} else {
				jwt = newMockJwtManager(t)
			}

			h := NewAuthHandler(jwt, nil)

			var reqBody *bytes.Buffer
			if tt.rawBody != "" {
				reqBody = bytes.NewBufferString(tt.rawBody)
			} else {
				reqBody = jsonBody(tt.body)
			}

			req := httptest.NewRequest(http.MethodPost, "/dummyLogin", reqBody)
			w := httptest.NewRecorder()
			h.DummyLogin(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantToken != "" {
				var resp map[string]string
				json.NewDecoder(w.Body).Decode(&resp)
				assert.Equal(t, tt.wantToken, resp["token"])
			}
		})
	}
}

func TestRegister(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		body       any
		rawBody    string
		setupAuth  func(*mockAuthService)
		wantStatus int
		wantEmail  string
	}{
		{
			name: "success",
			body: dto.UserRegisterRequest{Email: "test@test.com", Password: "pass123", Role: "user"},
			setupAuth: func(s *mockAuthService) {
				s.On("Register", mock.Anything, dto.UserRegisterRequest{
					Email: "test@test.com", Password: "pass123", Role: "user",
				}).Return(&dto.UserRegisterResponse{
					ID: "id-1", Email: "test@test.com", Role: "user", CreatedAt: time.Now(),
				}, nil)
			},
			wantStatus: http.StatusCreated,
			wantEmail:  "test@test.com",
		},
		{
			name:       "invalid json",
			rawBody:    "{bad",
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "empty credentials",
			body: dto.UserRegisterRequest{Email: "", Password: "", Role: "user"},
			setupAuth: func(s *mockAuthService) {
				s.On("Register", mock.Anything, mock.Anything).Return(nil, service.ErrEmptyCredentials)
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "email taken",
			body: dto.UserRegisterRequest{Email: "taken@test.com", Password: "pass", Role: "user"},
			setupAuth: func(s *mockAuthService) {
				s.On("Register", mock.Anything, mock.Anything).Return(nil, service.ErrEmailTaken)
			},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var authSvc *mockAuthService
			if tt.setupAuth != nil {
				authSvc = newMockAuthService(t)
				tt.setupAuth(authSvc)
			} else {
				authSvc = newMockAuthService(t)
			}

			h := NewAuthHandler(nil, authSvc)

			var reqBody *bytes.Buffer
			if tt.rawBody != "" {
				reqBody = bytes.NewBufferString(tt.rawBody)
			} else {
				reqBody = jsonBody(tt.body)
			}

			req := httptest.NewRequest(http.MethodPost, "/register", reqBody)
			w := httptest.NewRecorder()
			h.Register(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantEmail != "" {
				var resp dto.UserRegisterResponse
				json.NewDecoder(w.Body).Decode(&resp)
				assert.Equal(t, tt.wantEmail, resp.Email)
			}
		})
	}
}

func TestLogin(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		body       any
		rawBody    string
		setupAuth  func(*mockAuthService)
		wantStatus int
		wantToken  string
	}{
		{
			name: "success",
			body: dto.UserLoginRequest{Email: "test@test.com", Password: "pass123"},
			setupAuth: func(s *mockAuthService) {
				s.On("Login", mock.Anything, dto.UserLoginRequest{
					Email: "test@test.com", Password: "pass123",
				}).Return(&dto.UserLoginResponse{Token: "jwt-token"}, nil)
			},
			wantStatus: http.StatusOK,
			wantToken:  "jwt-token",
		},
		{
			name:       "invalid json",
			rawBody:    "nope",
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "invalid credentials",
			body: dto.UserLoginRequest{Email: "test@test.com", Password: "wrong"},
			setupAuth: func(s *mockAuthService) {
				s.On("Login", mock.Anything, mock.Anything).Return(nil, service.ErrInvalidCredentials)
			},
			wantStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var authSvc *mockAuthService
			if tt.setupAuth != nil {
				authSvc = newMockAuthService(t)
				tt.setupAuth(authSvc)
			} else {
				authSvc = newMockAuthService(t)
			}

			h := NewAuthHandler(nil, authSvc)

			var reqBody *bytes.Buffer
			if tt.rawBody != "" {
				reqBody = bytes.NewBufferString(tt.rawBody)
			} else {
				reqBody = jsonBody(tt.body)
			}

			req := httptest.NewRequest(http.MethodPost, "/login", reqBody)
			w := httptest.NewRecorder()
			h.Login(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantToken != "" {
				var resp dto.UserLoginResponse
				json.NewDecoder(w.Body).Decode(&resp)
				assert.Equal(t, tt.wantToken, resp.Token)
			}
		})
	}
}
