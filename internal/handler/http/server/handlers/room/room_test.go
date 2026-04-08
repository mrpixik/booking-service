package room

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/avito-internships/test-backend-1-mrpixik/internal/domain/dto"
	"github.com/avito-internships/test-backend-1-mrpixik/internal/domain/dto/errors/service"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func jsonBody(v any) *bytes.Buffer {
	b, _ := json.Marshal(v)
	return bytes.NewBuffer(b)
}

func intPtr(v int) *int { return &v }

func TestList(t *testing.T) {
	t.Parallel()

	now := time.Now()

	tests := []struct {
		name       string
		setupSvc   func(*mockRoomService)
		wantStatus int
		wantCount  int
	}{
		{
			name: "success",
			setupSvc: func(s *mockRoomService) {
				s.On("List", mock.Anything).Return(&dto.RoomListResponse{
					Rooms: []dto.RoomResponse{
						{ID: uuid.New().String(), Name: "Room A", CreatedAt: &now},
						{ID: uuid.New().String(), Name: "Room B", CreatedAt: &now},
					},
				}, nil)
			},
			wantStatus: http.StatusOK,
			wantCount:  2,
		},
		{
			name: "empty",
			setupSvc: func(s *mockRoomService) {
				s.On("List", mock.Anything).Return(&dto.RoomListResponse{Rooms: []dto.RoomResponse{}}, nil)
			},
			wantStatus: http.StatusOK,
			wantCount:  0,
		},
		{
			name: "service error",
			setupSvc: func(s *mockRoomService) {
				s.On("List", mock.Anything).Return(nil, service.ErrInternalError)
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			svc := newMockRoomService(t)
			tt.setupSvc(svc)
			h := NewRoomHandler(svc)

			req := httptest.NewRequest(http.MethodGet, "/rooms/list", nil)
			w := httptest.NewRecorder()
			h.List(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantStatus == http.StatusOK {
				var resp dto.RoomListResponse
				json.NewDecoder(w.Body).Decode(&resp)
				assert.Len(t, resp.Rooms, tt.wantCount)
			}
		})
	}
}

func TestCreate(t *testing.T) {
	t.Parallel()

	now := time.Now()

	tests := []struct {
		name       string
		body       any
		rawBody    string
		setupSvc   func(*mockRoomService)
		wantStatus int
		wantName   string
	}{
		{
			name: "success",
			body: dto.RoomCreateRequest{Name: "New Room", Capacity: intPtr(10)},
			setupSvc: func(s *mockRoomService) {
				s.On("Create", mock.Anything, dto.RoomCreateRequest{Name: "New Room", Capacity: intPtr(10)}).
					Return(&dto.RoomResponse{
						ID: uuid.New().String(), Name: "New Room", Capacity: intPtr(10), CreatedAt: &now,
					}, nil)
			},
			wantStatus: http.StatusCreated,
			wantName:   "New Room",
		},
		{
			name:       "invalid json",
			rawBody:    "{bad",
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "empty name",
			body: dto.RoomCreateRequest{Name: ""},
			setupSvc: func(s *mockRoomService) {
				s.On("Create", mock.Anything, mock.Anything).Return(nil, service.ErrInvalidRequest)
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "service error",
			body: dto.RoomCreateRequest{Name: "Room"},
			setupSvc: func(s *mockRoomService) {
				s.On("Create", mock.Anything, mock.Anything).Return(nil, service.ErrInternalError)
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			svc := newMockRoomService(t)
			if tt.setupSvc != nil {
				tt.setupSvc(svc)
			}
			h := NewRoomHandler(svc)

			var reqBody *bytes.Buffer
			if tt.rawBody != "" {
				reqBody = bytes.NewBufferString(tt.rawBody)
			} else {
				reqBody = jsonBody(tt.body)
			}

			req := httptest.NewRequest(http.MethodPost, "/rooms/create", reqBody)
			w := httptest.NewRecorder()
			h.Create(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantName != "" {
				var resp map[string]dto.RoomResponse
				json.NewDecoder(w.Body).Decode(&resp)
				assert.Equal(t, tt.wantName, resp["room"].Name)
			}
		})
	}
}
