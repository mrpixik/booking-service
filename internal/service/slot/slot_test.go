package slot

import (
	"context"
	"testing"
	"time"

	"github.com/avito-internships/test-backend-1-mrpixik/internal/domain/dto/errors/repository"
	"github.com/avito-internships/test-backend-1-mrpixik/internal/domain/dto/errors/service"
	"github.com/avito-internships/test-backend-1-mrpixik/internal/domain/model"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// не получится применять автогенерированный MocktxManager, поэтому написал свой маленький мок, чтобы он вызывал переданную функцию
type mockTxManager struct{}

func (m *mockTxManager) Begin(ctx context.Context, fn func(ctx context.Context) error) error {
	return fn(ctx)
}

func TestListAvailable(t *testing.T) {
	t.Parallel()

	validRoomID := uuid.New().String()
	date := time.Date(2026, 3, 25, 0, 0, 0, 0, time.UTC)
	weekdaySchedule := &model.Schedule{
		DaysOfWeek: []int{1, 2, 3, 4, 5},
		StartTime:  "09:00:00",
		EndTime:    "10:00:00",
	}

	tests := []struct {
		name          string
		roomID        string
		dateStr       string
		setupSlot     func(*mockSlotRepo)
		setupRoom     func(*mockRoomRepo)
		setupSchedule func(*mockScheduleRepo)
		wantErr       error
		wantSlots     int
		wantBulk      bool
	}{
		{
			name:    "success slots exist",
			roomID:  validRoomID,
			dateStr: "2026-03-25",
			setupRoom: func(r *mockRoomRepo) {
				r.On("GetById", mock.Anything, validRoomID).Return(&model.Room{ID: validRoomID}, nil)
			},
			setupSchedule: func(r *mockScheduleRepo) {
				r.On("GetByRoomID", mock.Anything, validRoomID).Return(weekdaySchedule, nil)
			},
			setupSlot: func(r *mockSlotRepo) {
				r.On("CountByRoomAndDate", mock.Anything, validRoomID, date).Return(2, nil)
				r.On("GetAvailableByRoomAndDate", mock.Anything, validRoomID, date).Return([]model.Slot{
					{ID: "s1", RoomID: validRoomID, Start: date.Add(9 * time.Hour), End: date.Add(9*time.Hour + 30*time.Minute)},
					{ID: "s2", RoomID: validRoomID, Start: date.Add(9*time.Hour + 30*time.Minute), End: date.Add(10 * time.Hour)},
				}, nil)
			},
			wantSlots: 2,
			wantBulk:  false,
		},
		{
			name:    "success generates slots",
			roomID:  validRoomID,
			dateStr: "2026-03-25",
			setupRoom: func(r *mockRoomRepo) {
				r.On("GetById", mock.Anything, validRoomID).Return(&model.Room{ID: validRoomID}, nil)
			},
			setupSchedule: func(r *mockScheduleRepo) {
				r.On("GetByRoomID", mock.Anything, validRoomID).Return(weekdaySchedule, nil)
			},
			setupSlot: func(r *mockSlotRepo) {
				r.On("CountByRoomAndDate", mock.Anything, validRoomID, date).Return(0, nil)
				r.On("BulkCreate", mock.Anything, mock.Anything).Return(nil)
				r.On("GetAvailableByRoomAndDate", mock.Anything, validRoomID, date).Return([]model.Slot{
					{ID: "s1", RoomID: validRoomID},
				}, nil)
			},
			wantSlots: 1,
			wantBulk:  true,
		},
		{
			name:    "invalid uuid",
			roomID:  "bad",
			dateStr: "2026-03-25",
			wantErr: service.ErrInvalidRequest,
		},
		{
			name:    "empty date",
			roomID:  uuid.New().String(),
			dateStr: "",
			wantErr: service.ErrInvalidRequest,
		},
		{
			name:    "invalid date",
			roomID:  uuid.New().String(),
			dateStr: "not-a-date",
			wantErr: service.ErrInvalidRequest,
		},
		{
			name:    "room not found",
			roomID:  uuid.New().String(),
			dateStr: "2026-03-25",
			setupRoom: func(r *mockRoomRepo) {
				r.On("GetById", mock.Anything, mock.Anything).Return(nil, repository.ErrNotFound)
			},
			wantErr: service.ErrNotFound,
		},
		{
			name:    "no schedule",
			roomID:  uuid.New().String(),
			dateStr: "2026-03-25",
			setupRoom: func(r *mockRoomRepo) {
				r.On("GetById", mock.Anything, mock.Anything).Return(&model.Room{}, nil)
			},
			setupSchedule: func(r *mockScheduleRepo) {
				r.On("GetByRoomID", mock.Anything, mock.Anything).Return(nil, repository.ErrNotFound)
			},
			wantSlots: 0,
		},
		{
			name:    "day not in schedule",
			roomID:  uuid.New().String(),
			dateStr: "2026-03-25",
			setupRoom: func(r *mockRoomRepo) {
				r.On("GetById", mock.Anything, mock.Anything).Return(&model.Room{}, nil)
			},
			setupSchedule: func(r *mockScheduleRepo) {
				r.On("GetByRoomID", mock.Anything, mock.Anything).Return(&model.Schedule{
					DaysOfWeek: []int{1, 2},
					StartTime:  "09:00:00",
					EndTime:    "18:00:00",
				}, nil)
			},
			wantSlots: 0,
		},
		{
			name:    "schedule repo error",
			roomID:  uuid.New().String(),
			dateStr: "2026-03-25",
			setupRoom: func(r *mockRoomRepo) {
				r.On("GetById", mock.Anything, mock.Anything).Return(&model.Room{}, nil)
			},
			setupSchedule: func(r *mockScheduleRepo) {
				r.On("GetByRoomID", mock.Anything, mock.Anything).Return(nil, repository.ErrInternalError)
			},
			wantErr: service.ErrInternalError,
		},
		{
			name:    "count error",
			roomID:  uuid.New().String(),
			dateStr: "2026-03-25",
			setupRoom: func(r *mockRoomRepo) {
				r.On("GetById", mock.Anything, mock.Anything).Return(&model.Room{}, nil)
			},
			setupSchedule: func(r *mockScheduleRepo) {
				r.On("GetByRoomID", mock.Anything, mock.Anything).Return(weekdaySchedule, nil)
			},
			setupSlot: func(r *mockSlotRepo) {
				r.On("CountByRoomAndDate", mock.Anything, mock.Anything, mock.Anything).Return(0, repository.ErrInternalError)
			},
			wantErr: service.ErrInternalError,
		},
		{
			name:    "bulk create error",
			roomID:  uuid.New().String(),
			dateStr: "2026-03-25",
			setupRoom: func(r *mockRoomRepo) {
				r.On("GetById", mock.Anything, mock.Anything).Return(&model.Room{}, nil)
			},
			setupSchedule: func(r *mockScheduleRepo) {
				r.On("GetByRoomID", mock.Anything, mock.Anything).Return(weekdaySchedule, nil)
			},
			setupSlot: func(r *mockSlotRepo) {
				r.On("CountByRoomAndDate", mock.Anything, mock.Anything, mock.Anything).Return(0, nil)
				r.On("BulkCreate", mock.Anything, mock.Anything).Return(repository.ErrInternalError)
			},
			wantErr: service.ErrInternalError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var slotR *mockSlotRepo
			var roomR *mockRoomRepo
			var schedR *mockScheduleRepo

			if tt.setupSlot != nil {
				slotR = newMockSlotRepo(t)
				tt.setupSlot(slotR)
			}
			if tt.setupRoom != nil {
				roomR = newMockRoomRepo(t)
				tt.setupRoom(roomR)
			}
			if tt.setupSchedule != nil {
				schedR = newMockScheduleRepo(t)
				tt.setupSchedule(schedR)
			}

			svc := NewSlotService(slotR, schedR, roomR, &mockTxManager{})
			result, err := svc.ListAvailable(context.Background(), tt.roomID, tt.dateStr)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}

			assert.NoError(t, err)
			assert.Len(t, result.Slots, tt.wantSlots)

			if slotR != nil {
				if tt.wantBulk {
					slotR.AssertCalled(t, "BulkCreate", mock.Anything, mock.Anything)
				} else {
					slotR.AssertNotCalled(t, "BulkCreate", mock.Anything, mock.Anything)
				}
			}
		})
	}
}

func TestGenerateSlots(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		startTime string
		endTime   string
		wantCount int
	}{
		{
			name:      "divisible by 30",
			startTime: "09:00:00",
			endTime:   "18:00:00",
			wantCount: 18,
		},
		{
			name:      "not divisible by 30",
			startTime: "09:00:00",
			endTime:   "09:45:00",
			wantCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			schedule := &model.Schedule{StartTime: tt.startTime, EndTime: tt.endTime}
			date := time.Date(2026, 3, 25, 0, 0, 0, 0, time.UTC)

			slots := generateSlots("room1", schedule, date)
			assert.Len(t, slots, tt.wantCount)
		})
	}
}
