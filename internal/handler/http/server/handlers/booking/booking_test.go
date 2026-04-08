package booking

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/avito-internships/test-backend-1-mrpixik/internal/domain/dto"
	"github.com/avito-internships/test-backend-1-mrpixik/internal/domain/dto/errors/service"
	"github.com/avito-internships/test-backend-1-mrpixik/internal/domain/model"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func jsonBody(v any) *bytes.Buffer {
	b, _ := json.Marshal(v)
	return bytes.NewBuffer(b)
}

func withUserCtx(r *http.Request, userID string, role model.Role) *http.Request {
	ctx := context.WithValue(r.Context(), model.CtxUserID, userID)
	ctx = context.WithValue(ctx, model.CtxRole, role)
	return r.WithContext(ctx)
}

func withChiParam(r *http.Request, key, value string) *http.Request {
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add(key, value)
	return r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
}

func TestCreate(t *testing.T) {
	t.Parallel()

	userID := uuid.New().String()
	slotID := uuid.New().String()

	tests := []struct {
		name       string
		body       any
		rawBody    string
		userID     string
		setupSvc   func(*mockBookingService)
		wantStatus int
		wantField  string
	}{
		{
			name:   "success",
			body:   dto.CreateBookingRequest{SlotID: slotID},
			userID: userID,
			setupSvc: func(s *mockBookingService) {
				s.On("Create", mock.Anything, userID, dto.CreateBookingRequest{SlotID: slotID}).
					Return(&dto.CreateBookingResponse{
						Booking: dto.BookingResponse{ID: uuid.New().String(), SlotID: slotID, UserID: userID, Status: "active"},
					}, nil)
			},
			wantStatus: http.StatusCreated,
			wantField:  "active",
		},
		{
			name:       "invalid json",
			rawBody:    "{bad",
			userID:     uuid.New().String(),
			wantStatus: http.StatusBadRequest,
		},
		{
			name:   "slot not found",
			body:   dto.CreateBookingRequest{SlotID: uuid.New().String()},
			userID: uuid.New().String(),
			setupSvc: func(s *mockBookingService) {
				s.On("Create", mock.Anything, mock.Anything, mock.Anything).Return(nil, service.ErrSlotNotFound)
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:   "already booked",
			body:   dto.CreateBookingRequest{SlotID: uuid.New().String()},
			userID: uuid.New().String(),
			setupSvc: func(s *mockBookingService) {
				s.On("Create", mock.Anything, mock.Anything, mock.Anything).Return(nil, service.ErrSlotAlreadyBooked)
			},
			wantStatus: http.StatusConflict,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			svc := newMockBookingService(t)
			if tt.setupSvc != nil {
				tt.setupSvc(svc)
			}
			h := NewBookingHandler(svc)

			var reqBody *bytes.Buffer
			if tt.rawBody != "" {
				reqBody = bytes.NewBufferString(tt.rawBody)
			} else {
				reqBody = jsonBody(tt.body)
			}

			req := httptest.NewRequest(http.MethodPost, "/bookings/create", reqBody)
			req = withUserCtx(req, tt.userID, model.RoleUser)
			w := httptest.NewRecorder()
			h.Create(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantField != "" {
				var resp dto.CreateBookingResponse
				json.NewDecoder(w.Body).Decode(&resp)
				assert.Equal(t, tt.wantField, resp.Booking.Status)
			}
		})
	}
}

func TestList(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		query      string
		setupSvc   func(*mockBookingService)
		wantStatus int
		wantCount  int
	}{
		{
			name:  "success",
			query: "?page=1&pageSize=20",
			setupSvc: func(s *mockBookingService) {
				s.On("List", mock.Anything, dto.AllBookingsRequest{Page: 1, PageSize: 20}).
					Return(&dto.BookingListResponse{
						Bookings:   []dto.BookingResponse{{ID: "1", Status: "active"}},
						Pagination: dto.PaginationResponse{Page: 1, PageSize: 20, TotalItems: 1, TotalPages: 1},
					}, nil)
			},
			wantStatus: http.StatusOK,
			wantCount:  1,
		},
		{
			name:  "default params error",
			query: "",
			setupSvc: func(s *mockBookingService) {
				s.On("List", mock.Anything, dto.AllBookingsRequest{Page: 0, PageSize: 0}).
					Return(nil, service.ErrInvalidRequest)
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "invalid page",
			query:      "?page=abc",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "invalid page size",
			query:      "?page=1&pageSize=abc",
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			svc := newMockBookingService(t)
			if tt.setupSvc != nil {
				tt.setupSvc(svc)
			}
			h := NewBookingHandler(svc)

			req := httptest.NewRequest(http.MethodGet, "/bookings/list"+tt.query, nil)
			w := httptest.NewRecorder()
			h.List(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantCount > 0 {
				var resp dto.BookingListResponse
				json.NewDecoder(w.Body).Decode(&resp)
				assert.Len(t, resp.Bookings, tt.wantCount)
			}
		})
	}
}

func TestListMy(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		setupSvc   func(*mockBookingService, string)
		wantStatus int
		wantCount  int
	}{
		{
			name: "success",
			setupSvc: func(s *mockBookingService, uid string) {
				s.On("ListByUserID", mock.Anything, uid).Return(&dto.MyBookingsResponse{
					Bookings: []dto.BookingResponse{{ID: "1", UserID: uid, Status: "active"}},
				}, nil)
			},
			wantStatus: http.StatusOK,
			wantCount:  1,
		},
		{
			name: "empty",
			setupSvc: func(s *mockBookingService, uid string) {
				s.On("ListByUserID", mock.Anything, uid).Return(&dto.MyBookingsResponse{
					Bookings: []dto.BookingResponse{},
				}, nil)
			},
			wantStatus: http.StatusOK,
			wantCount:  0,
		},
		{
			name: "service error",
			setupSvc: func(s *mockBookingService, uid string) {
				s.On("ListByUserID", mock.Anything, mock.Anything).Return(nil, service.ErrInternalError)
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			userID := uuid.New().String()
			svc := newMockBookingService(t)
			if tt.setupSvc != nil {
				tt.setupSvc(svc, userID)
			}
			h := NewBookingHandler(svc)

			req := httptest.NewRequest(http.MethodGet, "/bookings/my", nil)
			req = withUserCtx(req, userID, model.RoleUser)
			w := httptest.NewRecorder()
			h.ListMy(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantStatus == http.StatusOK {
				var resp dto.MyBookingsResponse
				json.NewDecoder(w.Body).Decode(&resp)
				assert.Len(t, resp.Bookings, tt.wantCount)
			}
		})
	}
}

func TestCancel(t *testing.T) {
	t.Parallel()

	ownerID := uuid.New().String()
	bookingID := uuid.New().String()

	tests := []struct {
		name       string
		userID     string
		bookingID  string
		setupSvc   func(*mockBookingService)
		wantStatus int
		wantField  string
	}{
		{
			name:      "success",
			userID:    ownerID,
			bookingID: bookingID,
			setupSvc: func(s *mockBookingService) {
				s.On("Cancel", mock.Anything, ownerID, bookingID).Return(&dto.CancelBookingResponse{
					Booking: dto.BookingResponse{ID: bookingID, UserID: ownerID, Status: "cancelled"},
				}, nil)
			},
			wantStatus: http.StatusOK,
			wantField:  "cancelled",
		},
		{
			name:      "already cancelled",
			userID:    ownerID,
			bookingID: bookingID,
			setupSvc: func(s *mockBookingService) {
				s.On("Cancel", mock.Anything, ownerID, bookingID).Return(&dto.CancelBookingResponse{
					Booking: dto.BookingResponse{ID: bookingID, UserID: ownerID, Status: "cancelled", CreatedAt: time.Now()},
				}, nil)
			},
			wantStatus: http.StatusOK,
			wantField:  "cancelled",
		},
		{
			name:      "not found",
			userID:    uuid.New().String(),
			bookingID: uuid.New().String(),
			setupSvc: func(s *mockBookingService) {
				s.On("Cancel", mock.Anything, mock.Anything, mock.Anything).Return(nil, service.ErrBookingNotFound)
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:      "another user",
			userID:    uuid.New().String(),
			bookingID: uuid.New().String(),
			setupSvc: func(s *mockBookingService) {
				s.On("Cancel", mock.Anything, mock.Anything, mock.Anything).Return(nil, service.ErrCancelBookFromAnotherUser)
			},
			wantStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			svc := newMockBookingService(t)
			if tt.setupSvc != nil {
				tt.setupSvc(svc)
			}
			h := NewBookingHandler(svc)

			req := httptest.NewRequest(http.MethodPost, "/bookings/"+tt.bookingID+"/cancel", nil)
			req = withUserCtx(req, tt.userID, model.RoleUser)
			req = withChiParam(req, "bookingId", tt.bookingID)
			w := httptest.NewRecorder()
			h.Cancel(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantField != "" {
				var resp dto.CancelBookingResponse
				json.NewDecoder(w.Body).Decode(&resp)
				assert.Equal(t, tt.wantField, resp.Booking.Status)
			}
		})
	}
}
