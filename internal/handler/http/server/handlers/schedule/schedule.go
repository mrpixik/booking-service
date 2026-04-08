package schedule

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/avito-internships/test-backend-1-mrpixik/internal/adapters"
	"github.com/go-chi/chi/v5"

	"github.com/avito-internships/test-backend-1-mrpixik/internal/domain/dto"
	"github.com/avito-internships/test-backend-1-mrpixik/internal/domain/dto/errors/server"
)

//go:generate mockery --name=scheduleService --inpackage --testonly
type scheduleService interface {
	Create(ctx context.Context, roomID string, req dto.ScheduleCreateRequest) (*dto.ScheduleResponse, error)
}

type ScheduleHandler struct {
	service scheduleService
}

func NewScheduleHandler(service scheduleService) *ScheduleHandler {
	return &ScheduleHandler{service: service}
}

// @Summary Создать расписание переговорки
// @Description Создать расписание для переговорки (только admin, только один раз). Длительность слота 30 мин.
// @Tags Schedules
// @Accept json
// @Produce json
// @Param roomId path string true "ID комнаты" format(uuid)
// @Param input body dto.ScheduleCreateRequest true "Данные расписания"
// @Success 201 {object} dto.ScheduleCreateResponse
// @Failure 400 {object} server.ErrorResponse
// @Failure 401 {object} server.ErrorResponse
// @Failure 403 {object} server.ErrorResponse
// @Failure 404 {object} server.ErrorResponse
// @Failure 409 {object} server.ErrorResponse
// @Failure 500 {object} server.InternalErrorResponse
// @Security BearerAuth
// @Router /rooms/{roomId}/schedule/create [post]
func (h *ScheduleHandler) Create(w http.ResponseWriter, r *http.Request) {
	roomID := chi.URLParam(r, "roomId")

	var req dto.ScheduleCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		adapters.WriteError(w, server.CodeInvalidRequest, server.InvalidRequestMsg, http.StatusBadRequest)
		return
	}

	resp, err := h.service.Create(r.Context(), roomID, req)
	if err != nil {
		adapters.WriteServiceError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(dto.ScheduleCreateResponse{Schedule: resp})
}
