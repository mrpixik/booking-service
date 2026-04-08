package integration

import (
	"context"
	"testing"
	"time"

	"github.com/avito-internships/test-backend-1-mrpixik/internal/auth"
	"github.com/avito-internships/test-backend-1-mrpixik/internal/domain/dto"
	"github.com/avito-internships/test-backend-1-mrpixik/internal/domain/dto/errors/service"
	"github.com/avito-internships/test-backend-1-mrpixik/internal/repository/booking"
	"github.com/avito-internships/test-backend-1-mrpixik/internal/repository/postgres"
	"github.com/avito-internships/test-backend-1-mrpixik/internal/repository/room"
	"github.com/avito-internships/test-backend-1-mrpixik/internal/repository/schedule"
	"github.com/avito-internships/test-backend-1-mrpixik/internal/repository/slot"
	"github.com/avito-internships/test-backend-1-mrpixik/internal/repository/user"
	bookingSvc "github.com/avito-internships/test-backend-1-mrpixik/internal/service/booking"
	roomSvc "github.com/avito-internships/test-backend-1-mrpixik/internal/service/room"
	scheduleSvc "github.com/avito-internships/test-backend-1-mrpixik/internal/service/schedule"
	slotSvc "github.com/avito-internships/test-backend-1-mrpixik/internal/service/slot"
	userSvc "github.com/avito-internships/test-backend-1-mrpixik/internal/service/user"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/suite"
)

// Хелперы, чтобы генерировать новые даты для каждого теста
// если будет тестироваться прошедшая дата, то тесты будут падать с ошибкой
// (тк бизнес логика построена так, что нельзя бронить слоты на прошедшие даты)
func nextWeekday() (time.Time, string) {
	d := time.Now().AddDate(0, 0, 7)
	for d.Weekday() == time.Saturday || d.Weekday() == time.Sunday {
		d = d.AddDate(0, 0, 1)
	}
	return d, d.Format("2006-01-02")
}

func nextSunday() (time.Time, string) {
	d := time.Now().AddDate(0, 0, 1)
	for d.Weekday() != time.Sunday {
		d = d.AddDate(0, 0, 1)
	}
	return d, d.Format("2006-01-02")
}

type stubClient struct{}

func (s *stubClient) CreateConference(ctx context.Context, bookingID string) (string, error) {
	return "https://avito-ktalk.com/" + bookingID, nil
}

type BookingFlowSuite struct {
	suite.Suite
	pool            *pgxpool.Pool
	ctx             context.Context
	roomService     *roomSvc.RoomService
	scheduleService *scheduleSvc.ScheduleService
	slotService     *slotSvc.SlotService
	bookingService  *bookingSvc.BookingService
	authService     *userSvc.AuthService
}

func TestBookingFlowSuite(t *testing.T) {
	suite.Run(t, new(BookingFlowSuite))
}

func (s *BookingFlowSuite) SetupSuite() {
	if testing.Short() {
		s.T().Skip("skipping integration tests: no database")
	}

	dsn := "postgres://tester:test123@localhost:5433/booking-test-db?sslmode=disable"
	pgxCfg, err := pgxpool.ParseConfig(dsn)
	s.Require().NoError(err)

	pool, err := pgxpool.NewWithConfig(context.Background(), pgxCfg)
	s.Require().NoError(err)

	s.pool = pool
	s.ctx = context.Background()

	txManager := postgres.NewTransactionManager(pool)
	jwtManager := auth.NewJWTManager("test-secret", 3600)

	roomRepo := room.NewRoomRepository(pool)
	scheduleRepo := schedule.NewScheduleRepository(pool)
	slotRepo := slot.NewSlotRepository(pool)
	bookingRepo := booking.NewBookingRepository(pool)
	userRepo := user.NewUserRepository(pool)

	s.roomService = roomSvc.NewRoomService(roomRepo)
	s.scheduleService = scheduleSvc.NewScheduleService(scheduleRepo, roomRepo, txManager)
	s.slotService = slotSvc.NewSlotService(slotRepo, scheduleRepo, roomRepo, txManager)
	s.bookingService = bookingSvc.NewBookingService(bookingRepo, slotRepo, txManager, &stubClient{})
	s.authService = userSvc.NewAuthService(userRepo, jwtManager)
}

func (s *BookingFlowSuite) SetupTest() {
	_, err := s.pool.Exec(s.ctx, `
		TRUNCATE TABLE bookings, slots, schedules, rooms, users RESTART IDENTITY CASCADE
	`)
	s.Require().NoError(err)
}

func (s *BookingFlowSuite) TearDownSuite() {
	if s.pool != nil {
		s.pool.Close()
	}
}

// TestFullBookingFlowSuccess полный набор тестов:
// создание переговорки, ее бронь, проверка корректности бронирования,
// проверка отображения в своих бронях, отмена брони, проверка доступности слота после отмены бронирования
func (s *BookingFlowSuite) TestFullBookingFlowSuccess() {
	_, futureDateStr := nextWeekday()

	// Создание переговорки
	capacity := 10
	createdRoom, err := s.roomService.Create(s.ctx, dto.RoomCreateRequest{
		Name:     "Переговорка А",
		Capacity: &capacity,
	})
	s.Require().NoError(err)
	s.Require().NotEmpty(createdRoom.ID)
	s.Require().Equal("Переговорка А", createdRoom.Name)

	// Расписание будни 9-18
	createdSchedule, err := s.scheduleService.Create(s.ctx, createdRoom.ID, dto.ScheduleCreateRequest{
		DaysOfWeek: []int{1, 2, 3, 4, 5},
		StartTime:  "09:00",
		EndTime:    "18:00",
	})
	s.Require().NoError(err)
	s.Require().NotEmpty(createdSchedule.ID)
	s.Require().Equal(createdRoom.ID, createdSchedule.RoomID)

	// Ожидаем 18 слотов на будний день
	slots, err := s.slotService.ListAvailable(s.ctx, createdRoom.ID, futureDateStr)
	s.Require().NoError(err)
	s.Require().Len(slots.Slots, 18)

	// Регистрация user
	registeredUser, err := s.authService.Register(s.ctx, dto.UserRegisterRequest{
		Email:    "user@test.com",
		Password: "pass123",
		Role:     "user",
	})
	s.Require().NoError(err)
	s.Require().NotEmpty(registeredUser.ID)

	// Успешная бронь слота
	firstSlotID := slots.Slots[0].ID
	createdBooking, err := s.bookingService.Create(s.ctx, registeredUser.ID, dto.CreateBookingRequest{
		SlotID:               firstSlotID,
		CreateConferenceLink: true,
	})
	s.Require().NoError(err)
	s.Require().Equal("active", createdBooking.Booking.Status)
	s.Require().Equal(firstSlotID, createdBooking.Booking.SlotID)
	s.Require().NotEmpty(createdBooking.Booking.ConferenceLink)

	// Повторная бронь слота. Ожидаем ошибку
	_, err = s.bookingService.Create(s.ctx, registeredUser.ID, dto.CreateBookingRequest{
		SlotID: firstSlotID,
	})
	s.Require().ErrorIs(err, service.ErrSlotAlreadyBooked)

	// Проверка, что после брони слот исчез из списка доступных
	slotsAfter, err := s.slotService.ListAvailable(s.ctx, createdRoom.ID, futureDateStr)
	s.Require().NoError(err)
	s.Require().Len(slotsAfter.Slots, 17)

	for _, sl := range slotsAfter.Slots {
		s.Require().NotEqual(firstSlotID, sl.ID)
	}

	// Проверка отображения брони в MyBooks
	myBookings, err := s.bookingService.ListByUserID(s.ctx, registeredUser.ID)
	s.Require().NoError(err)
	s.Require().Len(myBookings.Bookings, 1)
	s.Require().Equal(firstSlotID, myBookings.Bookings[0].SlotID)

	// Успешная отмена брони
	cancelled, err := s.bookingService.Cancel(s.ctx, registeredUser.ID, createdBooking.Booking.ID)
	s.Require().NoError(err)
	s.Require().Equal("cancelled", cancelled.Booking.Status)

	// Повторная успешная отмена брони (проверка идемпотентности
	cancelledAgain, err := s.bookingService.Cancel(s.ctx, registeredUser.ID, createdBooking.Booking.ID)
	s.Require().NoError(err)
	s.Require().Equal("cancelled", cancelledAgain.Booking.Status)

	// Проверка доступности слота после отмены
	slotsAfterCancel, err := s.slotService.ListAvailable(s.ctx, createdRoom.ID, futureDateStr)
	s.Require().NoError(err)
	s.Require().Len(slotsAfterCancel.Slots, 18)
}

// TestSlotsOnWeekend_Empty проверяет, что при расписании комнаты на будни, не будут создаваться слоты на выходные дни
func (s *BookingFlowSuite) TestSlotsOnWeekend_Empty() {
	_, weekendDateStr := nextSunday()

	capacity := 5
	createdRoom, err := s.roomService.Create(s.ctx, dto.RoomCreateRequest{
		Name:     "Комната Б",
		Capacity: &capacity,
	})
	s.Require().NoError(err)

	_, err = s.scheduleService.Create(s.ctx, createdRoom.ID, dto.ScheduleCreateRequest{
		DaysOfWeek: []int{1, 2, 3, 4, 5},
		StartTime:  "09:00",
		EndTime:    "18:00",
	})
	s.Require().NoError(err)

	// В выходной день слотов нет
	slots, err := s.slotService.ListAvailable(s.ctx, createdRoom.ID, weekendDateStr)
	s.Require().NoError(err)
	s.Require().Empty(slots.Slots)
}

// TestSlotGeneration_Idempotent проверяет, что UUID успешно записываются в бд при создании на новую дату
func (s *BookingFlowSuite) TestSlotGeneration_Idempotent() {
	_, futureDateStr := nextWeekday()

	createdRoom, err := s.roomService.Create(s.ctx, dto.RoomCreateRequest{Name: "Комната В"})
	s.Require().NoError(err)

	_, err = s.scheduleService.Create(s.ctx, createdRoom.ID, dto.ScheduleCreateRequest{
		DaysOfWeek: []int{1, 2, 3, 4, 5},
		StartTime:  "10:00",
		EndTime:    "12:00",
	})
	s.Require().NoError(err)

	// Первичная генерация UUID
	slots1, err := s.slotService.ListAvailable(s.ctx, createdRoom.ID, futureDateStr)
	s.Require().NoError(err)

	// Повтор и проверка, те же ли слоты
	slots2, err := s.slotService.ListAvailable(s.ctx, createdRoom.ID, futureDateStr)
	s.Require().NoError(err)

	s.Require().Len(slots1.Slots, 4)
	s.Require().Len(slots2.Slots, 4)

	for i := range slots1.Slots {
		s.Require().Equal(slots1.Slots[i].ID, slots2.Slots[i].ID)
	}
}

// TestCancelAnotherUsersBooking проверяет невозможность удалить бронь другого юзера
func (s *BookingFlowSuite) TestCancelAnotherUsersBooking() {
	_, futureDateStr := nextWeekday()

	createdRoom, err := s.roomService.Create(s.ctx, dto.RoomCreateRequest{Name: "Комната Г"})
	s.Require().NoError(err)

	_, err = s.scheduleService.Create(s.ctx, createdRoom.ID, dto.ScheduleCreateRequest{
		DaysOfWeek: []int{1, 2, 3, 4, 5},
		StartTime:  "09:00",
		EndTime:    "10:00",
	})
	s.Require().NoError(err)

	slots, err := s.slotService.ListAvailable(s.ctx, createdRoom.ID, futureDateStr)
	s.Require().NoError(err)

	user1, err := s.authService.Register(s.ctx, dto.UserRegisterRequest{
		Email: "user1@test.com", Password: "pass", Role: "user",
	})
	s.Require().NoError(err)

	user2, err := s.authService.Register(s.ctx, dto.UserRegisterRequest{
		Email: "user2@test.com", Password: "pass", Role: "user",
	})
	s.Require().NoError(err)

	booked, err := s.bookingService.Create(s.ctx, user1.ID, dto.CreateBookingRequest{
		SlotID: slots.Slots[0].ID,
	})
	s.Require().NoError(err)

	_, err = s.bookingService.Cancel(s.ctx, user2.ID, booked.Booking.ID)
	s.Require().ErrorIs(err, service.ErrCancelBookFromAnotherUser)
}
