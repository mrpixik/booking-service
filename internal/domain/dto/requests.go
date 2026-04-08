package dto

type DummyLoginRequest struct {
	Role string `json:"role"`
}

type UserRegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

type UserLoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RoomCreateRequest struct {
	Name        string  `json:"name"`
	Description *string `json:"description"` // nullable
	Capacity    *int    `json:"capacity"`    // nullable
}

type ScheduleCreateRequest struct {
	DaysOfWeek []int  `json:"daysOfWeek"`
	StartTime  string `json:"startTime"`
	EndTime    string `json:"endTime"`
}

type CreateBookingRequest struct {
	SlotID               string `json:"slotId"`
	CreateConferenceLink bool   `json:"createConferenceLink"`
}

type AllBookingsRequest struct {
	Page     int `json:"page"`
	PageSize int `json:"pageSize"`
}
