package booking

import (
	"context"
	"time"

	"github.com/avito-internships/test-backend-1-mrpixik/internal/adapters"
	"github.com/avito-internships/test-backend-1-mrpixik/internal/domain/dto"
	"github.com/avito-internships/test-backend-1-mrpixik/internal/domain/dto/errors/service"
	"github.com/avito-internships/test-backend-1-mrpixik/internal/domain/model"
	"github.com/google/uuid"
)

//go:generate mockery --name=slotRepo --inpackage --testonly
type slotRepo interface {
	GetByID(ctx context.Context, slotID string) (*model.Slot, error)
}

//go:generate mockery --name=conferenceClient --inpackage --testonly
type conferenceClient interface {
	CreateConference(ctx context.Context, bookingID string) (string, error)
}

//go:generate mockery --name=bookingRepo --inpackage --testonly
type bookingRepo interface {
	GetByID(ctx context.Context, bookingID string) (*model.Booking, error)
	GetAll(ctx context.Context, limit, offset int) ([]model.Booking, error)
	GetBySlotID(ctx context.Context, slotID string) (*model.Booking, error)
	GetAllByUserID(ctx context.Context, userID string) ([]model.Booking, error)
	Create(ctx context.Context, booking *model.Booking) (*model.Booking, error)
	Cancel(ctx context.Context, bookingID string) (*model.Booking, error)
	Count(ctx context.Context) (int, error)
}

type txManager interface {
	Begin(ctx context.Context, fn func(ctx context.Context) error) error
}

type BookingService struct {
	bookingRepo      bookingRepo
	slotRepo         slotRepo
	txManager        txManager
	conferenceClient conferenceClient
}

func NewBookingService(bookingRepo bookingRepo, slotRepo slotRepo, txManager txManager, conferenceClient conferenceClient) *BookingService {
	return &BookingService{
		bookingRepo:      bookingRepo,
		slotRepo:         slotRepo,
		txManager:        txManager,
		conferenceClient: conferenceClient,
	}
}

func (s *BookingService) Create(ctx context.Context, userID string, req dto.CreateBookingRequest) (*dto.CreateBookingResponse, error) {
	if _, err := uuid.Parse(req.SlotID); err != nil {
		return nil, service.ErrInvalidRequest
	}

	// Получаем ссылку до транзакции. Решил сделать именно так по следующим причинам:
	// Не держать блокировку в бд на время получения ответа от другого сервиса
	// Если сторонний сервис упадет, то транзакция откатится. Была идея просто докидывать ссылку в бд через UPDATE (например в горутине),
	// но решил отказаться по следующим причинам:
	// 1. тогда не получится вернуть ссылку клиенту
	// 2. лишний запрос в бд

	// Итого получается, что при возникновении ошибки с получением ссылки, мы вернем fallback ссылку, чтобы пользователям было понятно что ошибка не на его стороне (забыл поставить флаг или тп)
	// На мой взгляд это хороший подход, соблюдается принцип плавной деградации, клиент не блокируется и основная логика отрабатывается
	// По хороошему можно было бы создать возможность докидывать ссылку самому пользователю или добавить ручку на повторную попытку создания ссылки
	// Но чтобы не нарушать условия ТЗ я это не реализовываю
	var conferenceLink string
	if req.CreateConferenceLink {
		link, err := s.conferenceClient.CreateConference(ctx, req.SlotID)
		if err != nil {
			conferenceLink = "https://avito-ktalk.com/unavailable"
		} else {
			conferenceLink = link
		}
	}

	var result *dto.CreateBookingResponse

	err := s.txManager.Begin(ctx, func(txCtx context.Context) error {
		// Проверяем существование слота
		slot, err := s.slotRepo.GetByID(txCtx, req.SlotID)
		if err != nil {
			return adapters.ErrUnwrapRepoToService(err)
		}

		// Проверяем что слот уже не в прошлом
		if slot.Start.Before(time.Now()) {
			return service.ErrInvalidRequest
		}

		//conferenceLink := ""
		//if req.CreateConferenceLink {
		//	conferenceLink = "https://avito-ktalk.com/" + uuid.New().String()
		//}
		// Создаём бронь. Проверка на уже активную бронь не делается, чтобы избежать гонки + чтобы не делать лишний запрос

		booking := &model.Booking{
			SlotID:         req.SlotID,
			UserID:         userID,
			Status:         "active",
			ConferenceLink: conferenceLink,
		}

		created, err := s.bookingRepo.Create(txCtx, booking)
		if err != nil {
			return adapters.ErrUnwrapRepoToService(err)
		}

		result = &dto.CreateBookingResponse{
			Booking: dto.BookingResponse{
				ID:             created.ID,
				SlotID:         created.SlotID,
				UserID:         created.UserID,
				Status:         created.Status,
				ConferenceLink: created.ConferenceLink,
				CreatedAt:      created.CreatedAt,
			},
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return result, nil
}

func (s *BookingService) List(ctx context.Context, req dto.AllBookingsRequest) (*dto.BookingListResponse, error) {
	if req.Page < 1 || req.PageSize < 1 || req.PageSize > 100 {
		return nil, service.ErrInvalidRequest
	}

	offset := (req.Page - 1) * req.PageSize

	bookings, err := s.bookingRepo.GetAll(ctx, req.PageSize, offset)
	if err != nil {
		return nil, adapters.ErrUnwrapRepoToService(err)
	}

	total, err := s.bookingRepo.Count(ctx)
	if err != nil {
		return nil, adapters.ErrUnwrapRepoToService(err)
	}

	totalPages := total / req.PageSize
	if total%req.PageSize != 0 {
		totalPages++
	}

	resp := make([]dto.BookingResponse, 0, len(bookings))
	for _, b := range bookings {
		resp = append(resp, dto.BookingResponse{
			ID:             b.ID,
			SlotID:         b.SlotID,
			UserID:         b.UserID,
			Status:         b.Status,
			ConferenceLink: b.ConferenceLink,
			CreatedAt:      b.CreatedAt,
		})
	}

	return &dto.BookingListResponse{
		Bookings: resp,
		Pagination: dto.PaginationResponse{
			Page:       req.Page,
			PageSize:   req.PageSize,
			TotalItems: total,
			TotalPages: totalPages,
		},
	}, nil
}
func (s *BookingService) ListByUserID(ctx context.Context, userID string) (*dto.MyBookingsResponse, error) {
	bookings, err := s.bookingRepo.GetAllByUserID(ctx, userID)
	if err != nil {
		return nil, adapters.ErrUnwrapRepoToService(err)
	}

	resp := make([]dto.BookingResponse, 0, len(bookings))
	for _, b := range bookings {
		resp = append(resp, dto.BookingResponse{
			ID:             b.ID,
			SlotID:         b.SlotID,
			UserID:         b.UserID,
			Status:         b.Status,
			ConferenceLink: b.ConferenceLink,
			CreatedAt:      b.CreatedAt,
		})
	}

	return &dto.MyBookingsResponse{Bookings: resp}, nil
}
func (s *BookingService) Cancel(ctx context.Context, userID string, bookingID string) (*dto.CancelBookingResponse, error) {
	if _, err := uuid.Parse(bookingID); err != nil {
		return nil, service.ErrInvalidRequest
	}

	var result *dto.CancelBookingResponse

	err := s.txManager.Begin(ctx, func(txCtx context.Context) error {
		booking, err := s.bookingRepo.GetByID(txCtx, bookingID)
		if err != nil {
			return adapters.ErrUnwrapRepoToService(err)
		}

		// Проверка совпадения пользователя, создавшего бронь
		if booking.UserID != userID {
			return service.ErrCancelBookFromAnotherUser
		}

		// Если уже отменена, то не возвращаем ошибку
		if booking.Status == "cancelled" {
			result = &dto.CancelBookingResponse{
				Booking: dto.BookingResponse{
					ID:             booking.ID,
					SlotID:         booking.SlotID,
					UserID:         booking.UserID,
					Status:         booking.Status,
					ConferenceLink: booking.ConferenceLink,
					CreatedAt:      booking.CreatedAt,
				},
			}
			return nil
		}

		cancelled, err := s.bookingRepo.Cancel(txCtx, bookingID)
		if err != nil {
			return adapters.ErrUnwrapRepoToService(err)
		}

		result = &dto.CancelBookingResponse{
			Booking: dto.BookingResponse{
				ID:             cancelled.ID,
				SlotID:         cancelled.SlotID,
				UserID:         cancelled.UserID,
				Status:         cancelled.Status,
				ConferenceLink: cancelled.ConferenceLink,
				CreatedAt:      cancelled.CreatedAt,
			},
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return result, nil
}
