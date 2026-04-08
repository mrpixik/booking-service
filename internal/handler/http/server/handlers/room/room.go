package room

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/avito-internships/test-backend-1-mrpixik/internal/adapters"
	"github.com/avito-internships/test-backend-1-mrpixik/internal/domain/dto"
	"github.com/avito-internships/test-backend-1-mrpixik/internal/domain/dto/errors/server"
)

//go:generate mockery --name=roomService --inpackage --testonly
type roomService interface {
	List(ctx context.Context) (*dto.RoomListResponse, error)
	Create(ctx context.Context, req dto.RoomCreateRequest) (*dto.RoomResponse, error)
}

type RoomHandler struct {
	service roomService
}

func NewRoomHandler(service roomService) *RoomHandler {
	return &RoomHandler{service: service}
}

// @Summary Список переговорок
// @Description Получить список всех переговорных комнат (admin и user)
// @Tags Rooms
// @Produce json
// @Success 200 {object} dto.RoomListResponse
// @Failure 401 {object} server.ErrorResponse
// @Failure 500 {object} server.InternalErrorResponse
// @Security BearerAuth
// @Router /rooms/list [get]
func (h *RoomHandler) List(w http.ResponseWriter, r *http.Request) {
	resp, err := h.service.List(r.Context())
	if err != nil {
		adapters.WriteServiceError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

// @Summary Создать переговорку
// @Description Создать новую переговорную комнату (только admin)
// @Tags Rooms
// @Accept json
// @Produce json
// @Param input body dto.RoomCreateRequest true "Данные комнаты"
// @Success 201 {object} dto.RoomCreateResponse
// @Failure 400 {object} server.ErrorResponse
// @Failure 401 {object} server.ErrorResponse
// @Failure 403 {object} server.ErrorResponse
// @Failure 500 {object} server.InternalErrorResponse
// @Security BearerAuth
// @Router /rooms/create [post]
func (h *RoomHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req dto.RoomCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		adapters.WriteError(w, server.CodeInvalidRequest, server.InvalidRequestMsg, http.StatusBadRequest)
		return
	}

	resp, err := h.service.Create(r.Context(), req)
	if err != nil {
		adapters.WriteServiceError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(dto.RoomCreateResponse{Room: resp})
}
