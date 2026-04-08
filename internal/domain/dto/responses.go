package dto

import (
	"time"

	"github.com/avito-internships/test-backend-1-mrpixik/internal/domain/model"
)

type TokenResponse struct {
	Token string `json:"token"`
}

type UserRegisterResponse struct {
	ID           string     `json:"id"`
	Email        string     `json:"email"`
	Role         model.Role `json:"role"`
	PasswordHash string     `json:"-"`
	CreatedAt    time.Time  `json:"createdAt"`
}

type UserLoginResponse struct {
	Token string `json:"token"`
}

type RoomResponse struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Description *string    `json:"description"`
	Capacity    *int       `json:"capacity"`
	CreatedAt   *time.Time `json:"createdAt"`
}

type RoomCreateResponse struct {
	Room *RoomResponse `json:"room"`
}

type RoomListResponse struct {
	Rooms []RoomResponse `json:"rooms"`
}

type ScheduleResponse struct {
	ID         string `json:"id"`
	RoomID     string `json:"roomId"`
	DaysOfWeek []int  `json:"daysOfWeek"`
	StartTime  string `json:"startTime"`
	EndTime    string `json:"endTime"`
}

type ScheduleCreateResponse struct {
	Schedule *ScheduleResponse `json:"schedule"`
}

type SlotResponse struct {
	ID     string    `json:"id"`
	RoomID string    `json:"roomId"`
	Start  time.Time `json:"start"`
	End    time.Time `json:"end"`
}

type SlotListResponse struct {
	Slots []SlotResponse `json:"slots"`
}

type BookingResponse struct {
	ID             string    `json:"id"`
	SlotID         string    `json:"slotId"`
	UserID         string    `json:"userId"`
	Status         string    `json:"status"`
	ConferenceLink string    `json:"conferenceLink,omitempty"`
	CreatedAt      time.Time `json:"createdAt"`
}

type CreateBookingResponse struct {
	Booking BookingResponse `json:"booking"`
}

type PaginationResponse struct {
	Page       int `json:"page"`
	PageSize   int `json:"pageSize"`
	TotalItems int `json:"totalItems"`
	TotalPages int `json:"totalPages"`
}

type BookingListResponse struct {
	Bookings   []BookingResponse  `json:"bookings"`
	Pagination PaginationResponse `json:"pagination"`
}

type MyBookingsResponse struct {
	Bookings []BookingResponse `json:"bookings"`
}

type CancelBookingResponse struct {
	Booking BookingResponse `json:"booking"`
}
