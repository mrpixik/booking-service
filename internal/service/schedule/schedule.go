package schedule

import (
	"context"
	"regexp"

	"github.com/avito-internships/test-backend-1-mrpixik/internal/adapters"
	"github.com/avito-internships/test-backend-1-mrpixik/internal/domain/dto"
	"github.com/avito-internships/test-backend-1-mrpixik/internal/domain/dto/errors/service"
	"github.com/avito-internships/test-backend-1-mrpixik/internal/domain/model"
	"github.com/google/uuid"
)

var timeRegex = regexp.MustCompile(`^([01]?[0-9]|2[0-3]):[0-5][0-9]$`)

//go:generate mockery --name=roomRepo --inpackage --testonly
//go:generate mockery --name=scheduleRepo --inpackage --testonly
type scheduleRepo interface {
	Create(ctx context.Context, schedule *model.Schedule) (*model.Schedule, error)
}

type roomRepo interface {
	GetById(ctx context.Context, roomID string) (*model.Room, error)
}

type txManager interface {
	Begin(ctx context.Context, fn func(ctx context.Context) error) error
}

type ScheduleService struct {
	scheduleRepo scheduleRepo
	roomRepo     roomRepo
	txManager    txManager
}

func NewScheduleService(scheduleRepo scheduleRepo, roomRepo roomRepo, txManager txManager) *ScheduleService {
	return &ScheduleService{
		scheduleRepo: scheduleRepo,
		roomRepo:     roomRepo,
		txManager:    txManager,
	}
}

func (s *ScheduleService) Create(ctx context.Context, roomID string, req dto.ScheduleCreateRequest) (*dto.ScheduleResponse, error) {
	// Валидация рум айди
	if _, err := uuid.Parse(roomID); err != nil {
		return nil, service.ErrInvalidRequest
	}

	// Проверка времени
	if !timeRegex.MatchString(req.StartTime) || !timeRegex.MatchString(req.EndTime) {
		return nil, service.ErrInvalidRequest
	}
	if req.StartTime >= req.EndTime {
		return nil, service.ErrInvalidRequest
	}
	// проверка дней недели
	if len(req.DaysOfWeek) == 0 {
		return nil, service.ErrInvalidRequest
	}
	for _, d := range req.DaysOfWeek {
		if d < 1 || d > 7 {
			return nil, service.ErrInvalidRequest
		}
	}

	var result *dto.ScheduleResponse

	err := s.txManager.Begin(ctx, func(txCtx context.Context) error {
		// Проверяем существование комнаты
		_, err := s.roomRepo.GetById(txCtx, roomID)
		if err != nil {
			return adapters.ErrUnwrapRepoToService(err)
		}

		// Оасписание
		created, err := s.scheduleRepo.Create(txCtx, &model.Schedule{
			RoomID:     roomID,
			DaysOfWeek: req.DaysOfWeek,
			StartTime:  req.StartTime,
			EndTime:    req.EndTime,
		})
		if err != nil {
			return adapters.ErrUnwrapRepoToService(err)
		}

		result = &dto.ScheduleResponse{
			ID:         created.ID,
			RoomID:     created.RoomID,
			DaysOfWeek: created.DaysOfWeek,
			StartTime:  created.StartTime,
			EndTime:    created.EndTime,
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return result, nil
}
