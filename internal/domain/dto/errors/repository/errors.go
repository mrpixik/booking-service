package repository

import "errors"

var (
	ErrEmailExists    = errors.New("email already exists")
	ErrScheduleExists = errors.New("schedule already exists")

	ErrRoomNotFound = errors.New("room not found")

	ErrSlotNotFound = errors.New("slot not found")

	ErrBookingNotFound = errors.New("booking not found")
	ErrBookingExists   = errors.New("booking already exists")

	ErrNotFound      = errors.New("not found")
	ErrInternalError = errors.New("internal error")
)
