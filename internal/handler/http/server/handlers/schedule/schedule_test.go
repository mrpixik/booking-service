package schedule

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/avito-internships/test-backend-1-mrpixik/internal/domain/dto"
	"github.com/avito-internships/test-backend-1-mrpixik/internal/domain/dto/errors/service"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func jsonBody(v any) *bytes.Buffer {
	b, _ := json.Marshal(v)
	return bytes.NewBuffer(b)
}

func withChiParam(r *http.Request, key, value string) *http.Request {
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add(key, value)
	return r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
}

func TestCreate(t *testing.T) {
	t.Parallel()

	validRoomID := uuid.New().String()
	validReq := dto.ScheduleCreateRequest{
		DaysOfWeek: []int{1, 2, 3, 4, 5},
		StartTime:  "09:00",
		EndTime:    "18:00",
	}
	shortReq := dto.ScheduleCreateRequest{
		DaysOfWeek: []int{1}, StartTime: "09:00", EndTime: "18:00",
	}

	tests := []struct {
		name       string
		roomID     string
		body       any
		rawBody    string
		setupSvc   func(*mockScheduleService)
		wantStatus int
		wantRoomID string
	}{
		{
			name:   "success",
			roomID: validRoomID,
			body:   validReq,
			setupSvc: func(s *mockScheduleService) {
				s.On("Create", mock.Anything, validRoomID, validReq).Return(&dto.ScheduleResponse{
					ID: uuid.New().String(), RoomID: validRoomID,
					DaysOfWeek: []int{1, 2, 3, 4, 5}, StartTime: "09:00", EndTime: "18:00",
				}, nil)
			},
			wantStatus: http.StatusCreated,
			wantRoomID: validRoomID,
		},
		{
			name:       "invalid json",
			roomID:     uuid.New().String(),
			rawBody:    "{bad",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:   "invalid room id",
			roomID: "bad-id",
			body:   shortReq,
			setupSvc: func(s *mockScheduleService) {
				s.On("Create", mock.Anything, "bad-id", mock.Anything).Return(nil, service.ErrInvalidRequest)
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:   "room not found",
			roomID: uuid.New().String(),
			body:   shortReq,
			setupSvc: func(s *mockScheduleService) {
				s.On("Create", mock.Anything, mock.Anything, mock.Anything).Return(nil, service.ErrRoomNotFound)
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:   "schedule exists",
			roomID: uuid.New().String(),
			body:   shortReq,
			setupSvc: func(s *mockScheduleService) {
				s.On("Create", mock.Anything, mock.Anything, mock.Anything).Return(nil, service.ErrScheduleExists)
			},
			wantStatus: http.StatusConflict,
		},
		{
			name:   "internal error",
			roomID: uuid.New().String(),
			body:   shortReq,
			setupSvc: func(s *mockScheduleService) {
				s.On("Create", mock.Anything, mock.Anything, mock.Anything).Return(nil, service.ErrInternalError)
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			svc := newMockScheduleService(t)
			if tt.setupSvc != nil {
				tt.setupSvc(svc)
			}
			h := NewScheduleHandler(svc)

			var reqBody *bytes.Buffer
			if tt.rawBody != "" {
				reqBody = bytes.NewBufferString(tt.rawBody)
			} else {
				reqBody = jsonBody(tt.body)
			}

			req := httptest.NewRequest(http.MethodPost, "/rooms/"+tt.roomID+"/schedule/create", reqBody)
			req = withChiParam(req, "roomId", tt.roomID)
			w := httptest.NewRecorder()
			h.Create(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantRoomID != "" {
				var resp map[string]dto.ScheduleResponse
				json.NewDecoder(w.Body).Decode(&resp)
				assert.Equal(t, tt.wantRoomID, resp["schedule"].RoomID)
			}
		})
	}
}
