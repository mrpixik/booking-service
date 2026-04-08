package booking

import (
	"context"
	"testing"
	"time"

	"github.com/avito-internships/test-backend-1-mrpixik/internal/domain/dto"
	"github.com/avito-internships/test-backend-1-mrpixik/internal/domain/dto/errors/repository"
	"github.com/avito-internships/test-backend-1-mrpixik/internal/domain/dto/errors/service"
	"github.com/avito-internships/test-backend-1-mrpixik/internal/domain/model"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockTxManager struct{}

func (m *mockTxManager) Begin(ctx context.Context, fn func(ctx context.Context) error) error {
	return fn(ctx)
}

func TestCreate(t *testing.T) {
	t.Parallel()

	validSlotID := uuid.New().String()
	validUserID := uuid.New().String()

	tests := []struct {
		name        string
		userID      string
		req         dto.CreateBookingRequest
		setupSlot   func(*mockSlotRepo)
		setupBook   func(*mockBookingRepo)
		setupConf   func(*mockConferenceClient)
		wantErr     error
		checkResult func(*testing.T, *dto.CreateBookingResponse)
	}{
		{
			name:   "success",
			userID: validUserID,
			req:    dto.CreateBookingRequest{SlotID: validSlotID},
			setupSlot: func(r *mockSlotRepo) {
				r.On("GetByID", mock.Anything, validSlotID).Return(&model.Slot{ID: validSlotID, Start: time.Now().Add(1 * time.Hour)}, nil)
			},
			setupBook: func(r *mockBookingRepo) {
				r.On("Create", mock.Anything, mock.Anything).Return(&model.Booking{
					ID: uuid.New().String(), SlotID: validSlotID, UserID: validUserID, Status: "active",
				}, nil)
			},
			checkResult: func(t *testing.T, res *dto.CreateBookingResponse) {
				assert.Equal(t, "active", res.Booking.Status)
				assert.Equal(t, validSlotID, res.Booking.SlotID)
			},
		},
		{
			name:   "with conference link",
			userID: uuid.New().String(),
			req:    dto.CreateBookingRequest{SlotID: validSlotID, CreateConferenceLink: true},
			setupSlot: func(r *mockSlotRepo) {
				r.On("GetByID", mock.Anything, validSlotID).Return(&model.Slot{Start: time.Now().Add(1 * time.Hour)}, nil)
			},
			setupBook: func(r *mockBookingRepo) {
				r.On("Create", mock.Anything, mock.MatchedBy(func(b *model.Booking) bool {
					return b.ConferenceLink != ""
				})).Return(&model.Booking{
					ID: uuid.New().String(), SlotID: validSlotID, Status: "active", ConferenceLink: "https://avito-ktalk.com/test",
				}, nil)
			},
			setupConf: func(c *mockConferenceClient) {
				c.On("CreateConference", mock.Anything, mock.Anything).Return("https://avito-ktalk.com/test", nil)
			},
			checkResult: func(t *testing.T, res *dto.CreateBookingResponse) {
				assert.NotEmpty(t, res.Booking.ConferenceLink)
			},
		},
		{
			name:    "invalid uuid",
			userID:  "user",
			req:     dto.CreateBookingRequest{SlotID: "bad"},
			wantErr: service.ErrInvalidRequest,
		},
		{
			name:   "slot not found",
			userID: uuid.New().String(),
			req:    dto.CreateBookingRequest{SlotID: uuid.New().String()},
			setupSlot: func(r *mockSlotRepo) {
				r.On("GetByID", mock.Anything, mock.Anything).Return(nil, repository.ErrSlotNotFound)
			},
			wantErr: service.ErrSlotNotFound,
		},
		{
			name:   "slot in past",
			userID: uuid.New().String(),
			req:    dto.CreateBookingRequest{SlotID: uuid.New().String()},
			setupSlot: func(r *mockSlotRepo) {
				r.On("GetByID", mock.Anything, mock.Anything).Return(&model.Slot{Start: time.Now().Add(-1 * time.Hour)}, nil)
			},
			wantErr: service.ErrInvalidRequest,
		},
		{
			name:   "already booked",
			userID: uuid.New().String(),
			req:    dto.CreateBookingRequest{SlotID: uuid.New().String()},
			setupSlot: func(r *mockSlotRepo) {
				r.On("GetByID", mock.Anything, mock.Anything).Return(&model.Slot{Start: time.Now().Add(1 * time.Hour)}, nil)
			},
			setupBook: func(r *mockBookingRepo) {
				r.On("Create", mock.Anything, mock.Anything).Return(nil, repository.ErrBookingExists)
			},
			wantErr: service.ErrSlotAlreadyBooked,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var slotR *mockSlotRepo
			var bookR *mockBookingRepo
			var confC *mockConferenceClient

			if tt.setupSlot != nil {
				slotR = newMockSlotRepo(t)
				tt.setupSlot(slotR)
			}
			if tt.setupBook != nil {
				bookR = newMockBookingRepo(t)
				tt.setupBook(bookR)
			}
			if tt.setupConf != nil {
				confC = newMockConferenceClient(t)
				tt.setupConf(confC)
			}

			svc := NewBookingService(bookR, slotR, &mockTxManager{}, confC)
			result, err := svc.Create(context.Background(), tt.userID, tt.req)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}
			assert.NoError(t, err)
			if tt.checkResult != nil {
				tt.checkResult(t, result)
			}
		})
	}
}

func TestList(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		req       dto.AllBookingsRequest
		setupBook func(*mockBookingRepo)
		wantErr   error
		wantCount int
		wantPages int
		wantTotal int
	}{
		{
			name: "success",
			req:  dto.AllBookingsRequest{Page: 1, PageSize: 20},
			setupBook: func(r *mockBookingRepo) {
				r.On("GetAll", mock.Anything, 20, 0).Return([]model.Booking{{ID: "1"}, {ID: "2"}}, nil)
				r.On("Count", mock.Anything).Return(2, nil)
			},
			wantCount: 2,
			wantPages: 1,
			wantTotal: 2,
		},
		{
			name: "pagination",
			req:  dto.AllBookingsRequest{Page: 2, PageSize: 10},
			setupBook: func(r *mockBookingRepo) {
				r.On("GetAll", mock.Anything, 10, 10).Return([]model.Booking{{ID: "1"}}, nil)
				r.On("Count", mock.Anything).Return(15, nil)
			},
			wantCount: 1,
			wantPages: 2,
			wantTotal: 15,
		},
		{
			name:    "invalid page",
			req:     dto.AllBookingsRequest{Page: 0, PageSize: 20},
			wantErr: service.ErrInvalidRequest,
		},
		{
			name:    "invalid page size zero",
			req:     dto.AllBookingsRequest{Page: 1, PageSize: 0},
			wantErr: service.ErrInvalidRequest,
		},
		{
			name:    "invalid page size over 100",
			req:     dto.AllBookingsRequest{Page: 1, PageSize: 101},
			wantErr: service.ErrInvalidRequest,
		},
		{
			name: "repo error",
			req:  dto.AllBookingsRequest{Page: 1, PageSize: 20},
			setupBook: func(r *mockBookingRepo) {
				r.On("GetAll", mock.Anything, mock.Anything, mock.Anything).Return(nil, repository.ErrInternalError)
			},
			wantErr: service.ErrInternalError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var bookR *mockBookingRepo
			if tt.setupBook != nil {
				bookR = newMockBookingRepo(t)
				tt.setupBook(bookR)
			}

			svc := NewBookingService(bookR, nil, nil, nil)
			result, err := svc.List(context.Background(), tt.req)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}
			assert.NoError(t, err)
			assert.Len(t, result.Bookings, tt.wantCount)
			assert.Equal(t, tt.wantPages, result.Pagination.TotalPages)
			assert.Equal(t, tt.wantTotal, result.Pagination.TotalItems)
		})
	}
}

func TestListByUserID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		setupBook func(*mockBookingRepo)
		wantErr   error
		wantCount int
	}{
		{
			name: "success",
			setupBook: func(r *mockBookingRepo) {
				r.On("GetAllByUserID", mock.Anything, mock.Anything).Return([]model.Booking{{ID: "1", Status: "active"}}, nil)
			},
			wantCount: 1,
		},
		{
			name: "empty",
			setupBook: func(r *mockBookingRepo) {
				r.On("GetAllByUserID", mock.Anything, mock.Anything).Return([]model.Booking{}, nil)
			},
			wantCount: 0,
		},
		{
			name: "repo error",
			setupBook: func(r *mockBookingRepo) {
				r.On("GetAllByUserID", mock.Anything, mock.Anything).Return(nil, repository.ErrInternalError)
			},
			wantErr: service.ErrInternalError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			bookR := newMockBookingRepo(t)
			tt.setupBook(bookR)

			svc := NewBookingService(bookR, nil, nil, nil)
			result, err := svc.ListByUserID(context.Background(), uuid.New().String())

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}
			assert.NoError(t, err)
			assert.Len(t, result.Bookings, tt.wantCount)
		})
	}
}

func TestCancel(t *testing.T) {
	t.Parallel()

	ownerID := uuid.New().String()
	otherID := uuid.New().String()
	bookingID := uuid.New().String()

	tests := []struct {
		name        string
		userID      string
		bookingID   string
		setupBook   func(*mockBookingRepo)
		wantErr     error
		checkResult func(*testing.T, *dto.CancelBookingResponse)
		notCanceled bool
	}{
		{
			name:      "success",
			userID:    ownerID,
			bookingID: bookingID,
			setupBook: func(r *mockBookingRepo) {
				r.On("GetByID", mock.Anything, bookingID).Return(&model.Booking{ID: bookingID, UserID: ownerID, Status: "active"}, nil)
				r.On("Cancel", mock.Anything, bookingID).Return(&model.Booking{ID: bookingID, UserID: ownerID, Status: "cancelled"}, nil)
			},
			checkResult: func(t *testing.T, res *dto.CancelBookingResponse) {
				assert.Equal(t, "cancelled", res.Booking.Status)
			},
		},
		{
			name:      "already cancelled",
			userID:    ownerID,
			bookingID: bookingID,
			setupBook: func(r *mockBookingRepo) {
				r.On("GetByID", mock.Anything, bookingID).Return(&model.Booking{ID: bookingID, UserID: ownerID, Status: "cancelled"}, nil)
			},
			checkResult: func(t *testing.T, res *dto.CancelBookingResponse) {
				assert.Equal(t, "cancelled", res.Booking.Status)
			},
			notCanceled: true,
		},
		{
			name:      "invalid uuid",
			userID:    "user",
			bookingID: "bad-id",
			wantErr:   service.ErrInvalidRequest,
		},
		{
			name:      "not found",
			userID:    uuid.New().String(),
			bookingID: uuid.New().String(),
			setupBook: func(r *mockBookingRepo) {
				r.On("GetByID", mock.Anything, mock.Anything).Return(nil, repository.ErrBookingNotFound)
			},
			wantErr: service.ErrBookingNotFound,
		},
		{
			name:      "another user",
			userID:    otherID,
			bookingID: bookingID,
			setupBook: func(r *mockBookingRepo) {
				r.On("GetByID", mock.Anything, bookingID).Return(&model.Booking{ID: bookingID, UserID: ownerID, Status: "active"}, nil)
			},
			wantErr: service.ErrCancelBookFromAnotherUser,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var bookR *mockBookingRepo
			if tt.setupBook != nil {
				bookR = newMockBookingRepo(t)
				tt.setupBook(bookR)
			}

			svc := NewBookingService(bookR, nil, &mockTxManager{}, nil)
			result, err := svc.Cancel(context.Background(), tt.userID, tt.bookingID)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}
			assert.NoError(t, err)
			if tt.checkResult != nil {
				tt.checkResult(t, result)
			}
			if tt.notCanceled {
				bookR.AssertNotCalled(t, "Cancel")
			}
		})
	}
}
