package server

type ErrorCode string

const (
	CodeInvalidRequest    ErrorCode = "INVALID_REQUEST"
	CodeUnauthorized      ErrorCode = "UNAUTHORIZED"
	CodeNotFound          ErrorCode = "NOT_FOUND"
	CodeRoomNotFound      ErrorCode = "ROOM_NOT_FOUND"
	CodeSlotNotFound      ErrorCode = "SLOT_NOT_FOUND"
	CodeSlotAlreadyBooked ErrorCode = "SLOT_ALREADY_BOOKED"
	CodeBookingNotFound   ErrorCode = "BOOKING_NOT_FOUND" // в документации нет упоминания в каких кейсах используется
	CodeForbidden         ErrorCode = "FORBIDDEN"
	CodeScheduleExists    ErrorCode = "SCHEDULE_EXISTS"
	CodeInternalError     ErrorCode = "INTERNAL_ERROR"
)

const (
	// AUTH
	InvalidAuthorisationHeaderMsg = "missing or invalid authorization header"
	InvalidTokenMsg               = "invalid token"
	AdminRoleRequiredMsg          = "admin role required"
	UserRoleRequiredMsg           = "user role required"
	EmptyCredentialsMsg           = "empty credentials"
	EmailTakenMsg                 = "email already taken"
	InvalidCredentialsMsg         = "invalid credential"

	//Room
	RoomNotFoundMsg = "room not found"
	// Schedule
	ScheduleExistsMsg = "schedule for this room already exists and cannot be changed"

	// Booking
	BookingNotFoundMsg           = "booking not found"
	CancelBookFromAnotherUserMsg = "cannot cancel book from another user"

	// Slot
	SlotNotFoundMsg      = "slot not found"
	SlotAlreadyBookedMsg = "slot already booked"
	// Default
	NotFoundMsg        = "not found"
	RequestCanceledMsg = "request canceled"
	InvalidRequestMsg  = "invalid request body"
	InternalErrorMsg   = "internal error"
)

type ErrorDetail struct {
	Code    ErrorCode `json:"code"`
	Message string    `json:"message"`
}

type ErrorResponse struct {
	Error  ErrorDetail `json:"error"`
	Status int         `json:"-"`
}

type InternalErrorResponse = ErrorResponse
