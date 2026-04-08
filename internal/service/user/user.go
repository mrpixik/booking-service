package user

import (
	"context"

	"github.com/avito-internships/test-backend-1-mrpixik/internal/adapters"
	"github.com/avito-internships/test-backend-1-mrpixik/internal/domain/dto"
	"github.com/avito-internships/test-backend-1-mrpixik/internal/domain/dto/errors/service"
	"github.com/avito-internships/test-backend-1-mrpixik/internal/domain/model"
	"golang.org/x/crypto/bcrypt"
)

//go:generate mockery --name=userRepo --inpackage --testonly
//go:generate mockery --name=jwtManager --inpackage --testonly
type jwtManager interface {
	GenerateToken(userID, role string) (string, error)
}

type AuthService struct {
	repo       userRepo
	jwtManager jwtManager
}

type userRepo interface {
	Create(ctx context.Context, user *model.User) (*model.User, error)
	GetByEmail(ctx context.Context, email string) (*model.User, error)
}

func NewAuthService(repo userRepo, jwtManager jwtManager) *AuthService {
	return &AuthService{repo: repo, jwtManager: jwtManager}
}

func (s *AuthService) Register(ctx context.Context, req dto.UserRegisterRequest) (*dto.UserRegisterResponse, error) {
	// Валидация (сделал без дополнительных условий на проверку реального емейла и сложного пароля)
	if req.Email == "" || req.Password == "" || (req.Role != "admin" && req.Role != "user") {
		return nil, service.ErrEmptyCredentials
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, service.ErrInternalError
	}

	userDB := &model.User{
		Email:        req.Email,
		Role:         model.Role(req.Role),
		PasswordHash: string(hash),
	}

	created, err := s.repo.Create(ctx, userDB)
	if err != nil {
		return nil, adapters.ErrUnwrapRepoToService(err)
	}

	return &dto.UserRegisterResponse{
		ID:        created.ID,
		Email:     created.Email,
		Role:      created.Role,
		CreatedAt: created.CreatedAt,
	}, nil
}

func (s *AuthService) Login(ctx context.Context, req dto.UserLoginRequest) (*dto.UserLoginResponse, error) {
	// Валидация
	if req.Email == "" || req.Password == "" {
		return nil, service.ErrEmptyCredentials
	}

	// Проверка существования пользователя
	user, err := s.repo.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, service.ErrInvalidCredentials
	}

	// Проверка пароля
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, service.ErrInvalidCredentials
	}

	// JWT
	token, err := s.jwtManager.GenerateToken(user.ID, string(user.Role))
	if err != nil {
		return nil, service.ErrInternalError
	}

	return &dto.UserLoginResponse{Token: token}, nil
}
