package booking

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/avito-internships/test-backend-1-mrpixik/internal/adapters"
	"github.com/avito-internships/test-backend-1-mrpixik/internal/domain/dto"
	"github.com/avito-internships/test-backend-1-mrpixik/internal/domain/dto/errors/server"
	"github.com/avito-internships/test-backend-1-mrpixik/internal/handler/http/middleware"
	"github.com/go-chi/chi/v5"
)

//go:generate mockery --name=bookingService --inpackage --testonly
type bookingService interface {
	Create(ctx context.Context, userID string, req dto.CreateBookingRequest) (*dto.CreateBookingResponse, error)
	List(ctx context.Context, req dto.AllBookingsRequest) (*dto.BookingListResponse, error)
	ListByUserID(ctx context.Context, userID string) (*dto.MyBookingsResponse, error)
	Cancel(ctx context.Context, userID string, bookingID string) (*dto.CancelBookingResponse, error)
}

type BookingHandler struct {
	service bookingService
}

func NewBookingHandler(service bookingService) *BookingHandler {
	return &BookingHandler{service: service}
}

// @Summary Создать бронь на слот
// @Description Забронировать слот (только user). userId берётся из JWT-токена.
// @Tags Bookings
// @Accept json
// @Produce json
// @Param input body dto.CreateBookingRequest true "Данные бронирования"
// @Success 201 {object} dto.CreateBookingResponse
// @Failure 400 {object} server.ErrorResponse
// @Failure 401 {object} server.ErrorResponse
// @Failure 403 {object} server.ErrorResponse
// @Failure 404 {object} server.ErrorResponse
// @Failure 409 {object} server.ErrorResponse
// @Failure 500 {object} server.InternalErrorResponse
// @Security BearerAuth
// @Router /bookings/create [post]
func (h *BookingHandler) Create(w http.ResponseWriter, r *http.Request) {

	_, userID := middleware.UserInfoFromCtx(r.Context())

	var req dto.CreateBookingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		adapters.WriteError(w, server.CodeInvalidRequest, server.InvalidRequestMsg, http.StatusBadRequest)
		return
	}

	result, err := h.service.Create(r.Context(), userID, req)
	if err != nil {
		adapters.WriteServiceError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(result)
}

// @Summary Список всех броней с пагинацией
// @Description Получить список всех бронирований (только admin)
// @Tags Bookings
// @Produce json
// @Param page query integer false "Номер страницы (по умолчанию 1)" minimum(1) default(1)
// @Param pageSize query integer false "Размер страницы (по умолчанию 20, максимум 100)" minimum(1) maximum(100) default(20)
// @Success 200 {object} dto.BookingListResponse
// @Failure 400 {object} server.ErrorResponse
// @Failure 401 {object} server.ErrorResponse
// @Failure 403 {object} server.ErrorResponse
// @Failure 500 {object} server.InternalErrorResponse
// @Security BearerAuth
// @Router /bookings/list [get]
func (h *BookingHandler) List(w http.ResponseWriter, r *http.Request) {

	var page, pageSize int

	if v := r.URL.Query().Get("page"); v != "" {
		p, err := strconv.Atoi(v)
		if err != nil {
			adapters.WriteError(w, server.CodeInvalidRequest, server.InvalidRequestMsg, http.StatusBadRequest)
			return
		}
		page = p
	}

	if v := r.URL.Query().Get("pageSize"); v != "" {
		ps, err := strconv.Atoi(v)
		if err != nil {
			adapters.WriteError(w, server.CodeInvalidRequest, server.InvalidRequestMsg, http.StatusBadRequest)
			return
		}
		pageSize = ps
	}

	result, err := h.service.List(r.Context(), dto.AllBookingsRequest{
		Page:     page,
		PageSize: pageSize,
	})
	if err != nil {
		adapters.WriteServiceError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(result)
}

// @Summary Список броней текущего пользователя
// @Description Возвращает брони пользователя из JWT. Только будущие слоты (start >= now).
// @Tags Bookings
// @Produce json
// @Success 200 {object} dto.MyBookingsResponse
// @Failure 401 {object} server.ErrorResponse
// @Failure 403 {object} server.ErrorResponse
// @Failure 500 {object} server.InternalErrorResponse
// @Security BearerAuth
// @Router /bookings/my [get]
func (h *BookingHandler) ListMy(w http.ResponseWriter, r *http.Request) {
	_, userID := middleware.UserInfoFromCtx(r.Context())

	result, err := h.service.ListByUserID(r.Context(), userID)
	if err != nil {
		adapters.WriteServiceError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(result)
}

// @Summary Отменить бронь
// @Description Отменить свою бронь (только user). Идемпотентно — повторный вызов не ошибка.
// @Tags Bookings
// @Produce json
// @Param bookingId path string true "ID бронирования" format(uuid)
// @Success 200 {object} dto.CancelBookingResponse
// @Failure 401 {object} server.ErrorResponse
// @Failure 403 {object} server.ErrorResponse
// @Failure 404 {object} server.ErrorResponse
// @Failure 500 {object} server.InternalErrorResponse
// @Security BearerAuth
// @Router /bookings/{bookingId}/cancel [post]
func (h *BookingHandler) Cancel(w http.ResponseWriter, r *http.Request) {
	_, userID := middleware.UserInfoFromCtx(r.Context())

	bookingID := chi.URLParam(r, "bookingId")

	result, err := h.service.Cancel(r.Context(), userID, bookingID)
	if err != nil {
		adapters.WriteServiceError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(result)
}
