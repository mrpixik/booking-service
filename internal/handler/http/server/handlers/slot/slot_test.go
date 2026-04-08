package slot

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/avito-internships/test-backend-1-mrpixik/internal/domain/dto"
	"github.com/avito-internships/test-backend-1-mrpixik/internal/domain/dto/errors/service"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func withChiParam(r *http.Request, key, value string) *http.Request {
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add(key, value)
	return r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
}

func TestList(t *testing.T) {
	t.Parallel()

	validRoomID := uuid.New().String()

	tests := []struct {
		name       string
		roomID     string
		date       string
		setupSvc   func(*mockSlotService)
		wantStatus int
		wantCount  int
	}{
		{
			name:   "success",
			roomID: validRoomID,
			date:   "2026-03-25",
			setupSvc: func(s *mockSlotService) {
				s.On("ListAvailable", mock.Anything, validRoomID, "2026-03-25").Return(&dto.SlotListResponse{
					Slots: []dto.SlotResponse{
						{ID: "s1", RoomID: validRoomID, Start: time.Now(), End: time.Now().Add(30 * time.Minute)},
						{ID: "s2", RoomID: validRoomID, Start: time.Now(), End: time.Now().Add(30 * time.Minute)},
					},
				}, nil)
			},
			wantStatus: http.StatusOK,
			wantCount:  2,
		},
		{
			name:   "empty",
			roomID: validRoomID,
			date:   "2026-03-30",
			setupSvc: func(s *mockSlotService) {
				s.On("ListAvailable", mock.Anything, validRoomID, "2026-03-30").Return(&dto.SlotListResponse{
					Slots: []dto.SlotResponse{},
				}, nil)
			},
			wantStatus: http.StatusOK,
			wantCount:  0,
		},
		{
			name:   "invalid room id",
			roomID: "bad-id",
			date:   "2026-03-25",
			setupSvc: func(s *mockSlotService) {
				s.On("ListAvailable", mock.Anything, "bad-id", mock.Anything).Return(nil, service.ErrInvalidRequest)
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:   "missing date",
			roomID: validRoomID,
			date:   "",
			setupSvc: func(s *mockSlotService) {
				s.On("ListAvailable", mock.Anything, validRoomID, "").Return(nil, service.ErrInvalidRequest)
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:   "invalid date",
			roomID: validRoomID,
			date:   "not-a-date",
			setupSvc: func(s *mockSlotService) {
				s.On("ListAvailable", mock.Anything, validRoomID, "not-a-date").Return(nil, service.ErrInvalidRequest)
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:   "room not found",
			roomID: validRoomID,
			date:   "2026-03-25",
			setupSvc: func(s *mockSlotService) {
				s.On("ListAvailable", mock.Anything, validRoomID, "2026-03-25").Return(nil, service.ErrNotFound)
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:   "internal error",
			roomID: validRoomID,
			date:   "2026-03-25",
			setupSvc: func(s *mockSlotService) {
				s.On("ListAvailable", mock.Anything, validRoomID, "2026-03-25").Return(nil, service.ErrInternalError)
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			svc := newMockSlotService(t)
			tt.setupSvc(svc)
			h := NewSlotHandler(svc)

			url := "/rooms/" + tt.roomID + "/slots/list"
			if tt.date != "" {
				url += "?date=" + tt.date
			}

			req := httptest.NewRequest(http.MethodGet, url, nil)
			req = withChiParam(req, "roomId", tt.roomID)
			w := httptest.NewRecorder()
			h.List(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantStatus == http.StatusOK {
				var resp dto.SlotListResponse
				json.NewDecoder(w.Body).Decode(&resp)
				assert.Len(t, resp.Slots, tt.wantCount)
			}
		})
	}
}
