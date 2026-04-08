package slot

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/avito-internships/test-backend-1-mrpixik/internal/adapters"
	"github.com/avito-internships/test-backend-1-mrpixik/internal/domain/dto"
	"github.com/go-chi/chi/v5"
)

//go:generate mockery --name=slotService --inpackage --testonly
type slotService interface {
	ListAvailable(ctx context.Context, roomID string, date string) (*dto.SlotListResponse, error)
}

type SlotHandler struct {
	service slotService
}

func NewSlotHandler(service slotService) *SlotHandler {
	return &SlotHandler{service: service}
}

// @Summary Список доступных слотов
// @Description Возвращает слоты, не занятые активной бронью, для указанной переговорки на указанную дату
// @Tags Slots
// @Produce json
// @Param roomId path string true "ID комнаты" format(uuid)
// @Param date query string true "Дата в формате ISO 8601 (например: 2024-06-10)" format(date)
// @Success 200 {object} dto.SlotListResponse
// @Failure 400 {object} server.ErrorResponse
// @Failure 401 {object} server.ErrorResponse
// @Failure 404 {object} server.ErrorResponse
// @Failure 500 {object} server.InternalErrorResponse
// @Security BearerAuth
// @Router /rooms/{roomId}/slots/list [get]
func (h *SlotHandler) List(w http.ResponseWriter, r *http.Request) {
	roomID := chi.URLParam(r, "roomId")
	date := r.URL.Query().Get("date")

	resp, err := h.service.ListAvailable(r.Context(), roomID, date)
	if err != nil {
		adapters.WriteServiceError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}
