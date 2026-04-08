package service

import "errors"

var (
	// Authtorisation
	ErrEmptyCredentials   = errors.New("empty credentials")
	ErrEmailTaken         = errors.New("email already taken")
	ErrInvalidCredentials = errors.New("invalid credentials")
	// Room
	ErrRoomNotFound = errors.New("room not found")
	// Schedule
	ErrScheduleExists = errors.New("schedule exists")
	// Slot
	ErrSlotAlreadyBooked = errors.New("slot already booked")
	ErrSlotNotFound      = errors.New("room not found")
	//Booking
	ErrBookingNotFound           = errors.New("booking not found")
	ErrCancelBookFromAnotherUser = errors.New("can not cancel book from another user")
	// Default
	ErrNotFound       = errors.New("not found")
	ErrInvalidRequest = errors.New("invalid request")
	ErrInternalError  = errors.New("internal error")
)
