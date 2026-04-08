package slot

import (
	"context"
	"errors"
	"time"

	"github.com/avito-internships/test-backend-1-mrpixik/internal/domain/dto"
	"github.com/avito-internships/test-backend-1-mrpixik/internal/domain/dto/errors/repository"
	"github.com/avito-internships/test-backend-1-mrpixik/internal/domain/dto/errors/service"
	"github.com/avito-internships/test-backend-1-mrpixik/internal/domain/model"
	"github.com/google/uuid"
)

//go:generate mockery --name=roomRepo --inpackage --testonly
//go:generate mockery --name=slotRepo --inpackage --testonly
//go:generate mockery --name=scheduleRepo --inpackage --testonly
type txManager interface {
	Begin(ctx context.Context, fn func(ctx context.Context) error) error
}

type roomRepo interface {
	GetAll(ctx context.Context) ([]model.Room, error)
	Create(ctx context.Context, room *model.Room) (*model.Room, error)
	GetById(ctx context.Context, roomID string) (*model.Room, error)
}

type slotRepo interface {
	BulkCreate(ctx context.Context, slots []model.Slot) error
	GetAvailableByRoomAndDate(ctx context.Context, roomID string, date time.Time) ([]model.Slot, error)
	CountByRoomAndDate(ctx context.Context, roomID string, date time.Time) (int, error)
}

type scheduleRepo interface {
	Create(ctx context.Context, schedule *model.Schedule) (*model.Schedule, error)
	GetByRoomID(ctx context.Context, roomID string) (*model.Schedule, error)
}

type SlotService struct {
	slotRepo     slotRepo
	scheduleRepo scheduleRepo
	roomRepo     roomRepo
	txManager    txManager
}

func NewSlotService(slotRepo slotRepo, scheduleRepo scheduleRepo, roomRepo roomRepo, txManager txManager) *SlotService {
	return &SlotService{
		slotRepo:     slotRepo,
		scheduleRepo: scheduleRepo,
		roomRepo:     roomRepo,
		txManager:    txManager,
	}
}

func (s *SlotService) ListAvailable(ctx context.Context, roomID string, dateStr string) (*dto.SlotListResponse, error) {
	// Валидация UUID
	if _, err := uuid.Parse(roomID); err != nil {
		return nil, service.ErrInvalidRequest
	}

	// Валидация даты
	if dateStr == "" {
		return nil, service.ErrInvalidRequest
	}

	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return nil, service.ErrInvalidRequest
	}

	var result *dto.SlotListResponse

	err = s.txManager.Begin(ctx, func(txCtx context.Context) error {
		// Проверка существования комнаты
		_, err := s.roomRepo.GetById(txCtx, roomID)
		if err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				return service.ErrNotFound
			}
			return service.ErrInternalError
		}

		// Получение расписания
		schedule, err := s.scheduleRepo.GetByRoomID(txCtx, roomID)
		if err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				// Нет расписания — пустой список
				result = &dto.SlotListResponse{Slots: []dto.SlotResponse{}}
				return nil
			}
			return service.ErrInternalError
		}

		// Проверка ксть ли день недели в расписании
		weekday := isoWeekday(date)
		if !containsDay(schedule.DaysOfWeek, weekday) {
			result = &dto.SlotListResponse{Slots: []dto.SlotResponse{}}
			return nil
		}

		// Проверяем, сгенерированы ли слоты
		// (это ускорило работу ручки с 18ms до 8-9ms при запросах на дату с уже созданными слотами)
		count, err := s.slotRepo.CountByRoomAndDate(txCtx, roomID, date)
		if err != nil {
			return service.ErrInternalError
		}
		if count == 0 {
			slots := generateSlots(roomID, schedule, date)
			if err := s.slotRepo.BulkCreate(txCtx, slots); err != nil {
				return service.ErrInternalError
			}
		}

		// Получение доступных
		available, err := s.slotRepo.GetAvailableByRoomAndDate(txCtx, roomID, date)
		if err != nil {
			return service.ErrInternalError
		}

		resp := make([]dto.SlotResponse, 0, len(available))
		for _, sl := range available {
			resp = append(resp, dto.SlotResponse{
				ID:     sl.ID,
				RoomID: sl.RoomID,
				Start:  sl.Start,
				End:    sl.End,
			})
		}

		result = &dto.SlotListResponse{Slots: resp}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return result, nil
}

func isoWeekday(date time.Time) int {
	day := int(date.Weekday())
	if day == 0 {
		return 7
	}
	return day
}

func containsDay(days []int, day int) bool {
	for _, d := range days {
		if d == day {
			return true
		}
	}
	return false
}

func generateSlots(roomID string, schedule *model.Schedule, date time.Time) []model.Slot {
	startTime, _ := time.Parse("15:04:05", schedule.StartTime)
	endTime, _ := time.Parse("15:04:05", schedule.EndTime)

	var slots []model.Slot
	current := time.Date(date.Year(), date.Month(), date.Day(),
		startTime.Hour(), startTime.Minute(), 0, 0, time.UTC)
	end := time.Date(date.Year(), date.Month(), date.Day(),
		endTime.Hour(), endTime.Minute(), 0, 0, time.UTC)

	for current.Add(30*time.Minute).Before(end) || current.Add(30*time.Minute).Equal(end) {
		slots = append(slots, model.Slot{
			RoomID: roomID,
			Start:  current,
			End:    current.Add(30 * time.Minute),
		})
		current = current.Add(30 * time.Minute)
	}

	return slots
}
