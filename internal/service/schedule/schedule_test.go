package schedule

import (
	"context"
	"testing"

	"github.com/avito-internships/test-backend-1-mrpixik/internal/domain/dto"
	"github.com/avito-internships/test-backend-1-mrpixik/internal/domain/dto/errors/repository"
	"github.com/avito-internships/test-backend-1-mrpixik/internal/domain/dto/errors/service"
	"github.com/avito-internships/test-backend-1-mrpixik/internal/domain/model"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// не получится применять автогенерированный MocktxManager, поэтому написал свой маленький мок, чтобы он вызывал переданную функцию
// P.S. возможно стоило бы перенести его в отдельный пакет, чтобы переиспользовать,
// но думаю это не критично и для понимания кода возможно есть смысл оставить 5 доп строчек в файле
type mockTxManager struct{}

func (m *mockTxManager) Begin(ctx context.Context, fn func(ctx context.Context) error) error {
	return fn(ctx)
}

func validRequest() dto.ScheduleCreateRequest {
	return dto.ScheduleCreateRequest{
		DaysOfWeek: []int{1, 2, 3, 4, 5},
		StartTime:  "09:00",
		EndTime:    "18:00",
	}
}

func TestCreate(t *testing.T) {
	t.Parallel()

	validRoomID := uuid.New().String()

	tests := []struct {
		name          string
		roomID        string
		req           dto.ScheduleCreateRequest
		setupRoom     func(*mockRoomRepo)
		setupSchedule func(*mockScheduleRepo)
		wantErr       error
		checkResult   func(*testing.T, *dto.ScheduleResponse)
	}{
		{
			name:   "success",
			roomID: validRoomID,
			req:    validRequest(),
			setupRoom: func(r *mockRoomRepo) {
				r.On("GetById", mock.Anything, validRoomID).Return(&model.Room{ID: validRoomID}, nil)
			},
			setupSchedule: func(r *mockScheduleRepo) {
				r.On("Create", mock.Anything, mock.Anything).Return(&model.Schedule{
					ID: uuid.New().String(), RoomID: validRoomID,
					DaysOfWeek: []int{1, 2, 3, 4, 5}, StartTime: "09:00", EndTime: "18:00",
				}, nil)
			},
			checkResult: func(t *testing.T, res *dto.ScheduleResponse) {
				assert.Equal(t, validRoomID, res.RoomID)
				assert.Equal(t, "09:00", res.StartTime)
				assert.Equal(t, "18:00", res.EndTime)
			},
		},
		{
			name:    "invalid room id",
			roomID:  "bad-id",
			req:     validRequest(),
			wantErr: service.ErrInvalidRequest,
		},
		{
			name:    "invalid start time",
			roomID:  uuid.New().String(),
			req:     func() dto.ScheduleCreateRequest { r := validRequest(); r.StartTime = "25:00"; return r }(),
			wantErr: service.ErrInvalidRequest,
		},
		{
			name:    "invalid end time",
			roomID:  uuid.New().String(),
			req:     func() dto.ScheduleCreateRequest { r := validRequest(); r.EndTime = "abc"; return r }(),
			wantErr: service.ErrInvalidRequest,
		},
		{
			name:   "start after end",
			roomID: uuid.New().String(),
			req: func() dto.ScheduleCreateRequest {
				r := validRequest()
				r.StartTime = "18:00"
				r.EndTime = "09:00"
				return r
			}(),
			wantErr: service.ErrInvalidRequest,
		},
		{
			name:   "start equals end",
			roomID: uuid.New().String(),
			req: func() dto.ScheduleCreateRequest {
				r := validRequest()
				r.StartTime = "09:00"
				r.EndTime = "09:00"
				return r
			}(),
			wantErr: service.ErrInvalidRequest,
		},
		{
			name:    "empty days",
			roomID:  uuid.New().String(),
			req:     func() dto.ScheduleCreateRequest { r := validRequest(); r.DaysOfWeek = []int{}; return r }(),
			wantErr: service.ErrInvalidRequest,
		},
		{
			name:    "invalid day zero",
			roomID:  uuid.New().String(),
			req:     func() dto.ScheduleCreateRequest { r := validRequest(); r.DaysOfWeek = []int{0}; return r }(),
			wantErr: service.ErrInvalidRequest,
		},
		{
			name:    "invalid day eight",
			roomID:  uuid.New().String(),
			req:     func() dto.ScheduleCreateRequest { r := validRequest(); r.DaysOfWeek = []int{8}; return r }(),
			wantErr: service.ErrInvalidRequest,
		},
		{
			name:   "room not found",
			roomID: uuid.New().String(),
			req:    validRequest(),
			setupRoom: func(r *mockRoomRepo) {
				r.On("GetById", mock.Anything, mock.Anything).Return(nil, repository.ErrRoomNotFound)
			},
			wantErr: service.ErrRoomNotFound,
		},
		{
			name:   "schedule already exists",
			roomID: uuid.New().String(),
			req:    validRequest(),
			setupRoom: func(r *mockRoomRepo) {
				r.On("GetById", mock.Anything, mock.Anything).Return(&model.Room{}, nil)
			},
			setupSchedule: func(r *mockScheduleRepo) {
				r.On("Create", mock.Anything, mock.Anything).Return(nil, repository.ErrScheduleExists)
			},
			wantErr: service.ErrScheduleExists,
		},
		{
			name:   "repo internal error",
			roomID: uuid.New().String(),
			req:    validRequest(),
			setupRoom: func(r *mockRoomRepo) {
				r.On("GetById", mock.Anything, mock.Anything).Return(&model.Room{}, nil)
			},
			setupSchedule: func(r *mockScheduleRepo) {
				r.On("Create", mock.Anything, mock.Anything).Return(nil, repository.ErrInternalError)
			},
			wantErr: service.ErrInternalError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var roomR *mockRoomRepo
			var schedR *mockScheduleRepo

			if tt.setupRoom != nil {
				roomR = newMockRoomRepo(t)
				tt.setupRoom(roomR)
			}
			if tt.setupSchedule != nil {
				schedR = newMockScheduleRepo(t)
				tt.setupSchedule(schedR)
			}

			svc := NewScheduleService(schedR, roomR, &mockTxManager{})
			result, err := svc.Create(context.Background(), tt.roomID, tt.req)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, result)
				return
			}

			assert.NoError(t, err)
			if tt.checkResult != nil {
				tt.checkResult(t, result)
			}
		})
	}
}
