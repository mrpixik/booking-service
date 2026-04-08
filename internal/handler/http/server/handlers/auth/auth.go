package auth

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/avito-internships/test-backend-1-mrpixik/internal/adapters"
	"github.com/avito-internships/test-backend-1-mrpixik/internal/domain/dto"
	"github.com/avito-internships/test-backend-1-mrpixik/internal/domain/dto/errors/server"
	"github.com/avito-internships/test-backend-1-mrpixik/internal/domain/model"
)

//go:generate mockery --name=jwtManager --inpackage --testonly
type jwtManager interface {
	GenerateToken(userID, role string) (string, error)
}

//go:generate mockery --name=authService --inpackage --testonly
type authService interface {
	Register(ctx context.Context, req dto.UserRegisterRequest) (*dto.UserRegisterResponse, error)
	Login(ctx context.Context, req dto.UserLoginRequest) (*dto.UserLoginResponse, error)
}

type AuthHandler struct {
	jwtManager  jwtManager
	authService authService
}

func NewAuthHandler(jwtManager jwtManager, authService authService) *AuthHandler {
	return &AuthHandler{jwtManager: jwtManager, authService: authService}
}

// @Summary Получить тестовый JWT по роли
// @Description Выдаёт тестовый JWT для указанной роли (admin / user). Доступен без авторизации.
// @Tags Auth
// @Accept json
// @Produce json
// @Param input body dto.DummyLoginRequest true "Роль"
// @Success 200 {object} dto.TokenResponse
// @Failure 400 {object} server.ErrorResponse
// @Failure 500 {object} server.InternalErrorResponse
// @Router /dummyLogin [post]
func (h *AuthHandler) DummyLogin(w http.ResponseWriter, r *http.Request) {
	var req dto.DummyLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		adapters.WriteError(w, server.CodeInvalidRequest, server.InvalidRequestMsg, http.StatusBadRequest)
		return
	}

	var userID string
	switch model.Role(req.Role) {
	case model.RoleAdmin:
		userID = model.DummyAdminID
	case model.RoleUser:
		userID = model.DummyUserID
	default:
		adapters.WriteError(w, server.CodeInvalidRequest, "role must be 'admin' or 'user'", http.StatusBadRequest)
		return
	}

	token, err := h.jwtManager.GenerateToken(userID, req.Role)
	if err != nil {
		adapters.WriteError(w, server.CodeInternalError, server.InternalErrorMsg, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(dto.TokenResponse{Token: token})
}

// @Summary Регистрация пользователя
// @Description Создать нового пользователя. Доступен без авторизации.
// @Tags Auth
// @Accept json
// @Produce json
// @Param input body dto.UserRegisterRequest true "Данные регистрации"
// @Success 201 {object} dto.UserRegisterResponse
// @Failure 400 {object} server.ErrorResponse
// @Failure 500 {object} server.InternalErrorResponse
// @Router /register [post]
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req dto.UserRegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		adapters.WriteError(w, server.CodeInvalidRequest, server.InvalidRequestMsg, http.StatusBadRequest)
		return
	}

	user, err := h.authService.Register(r.Context(), req)
	if err != nil {
		adapters.WriteServiceError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(user)
}

// @Summary Авторизация по email и паролю
// @Description Авторизует пользователя, возвращает JWT. Доступен без авторизации.
// @Tags Auth
// @Accept json
// @Produce json
// @Param input body dto.UserLoginRequest true "Данные входа"
// @Success 200 {object} dto.TokenResponse
// @Failure 401 {object} server.ErrorResponse
// @Failure 500 {object} server.InternalErrorResponse
// @Router /login [post]
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req dto.UserLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		adapters.WriteError(w, server.CodeInvalidRequest, server.InvalidRequestMsg, http.StatusBadRequest)
		return
	}

	token, err := h.authService.Login(r.Context(), req)
	if err != nil {
		adapters.WriteServiceError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(token)
}
