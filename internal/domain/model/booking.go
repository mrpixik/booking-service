package model

import "time"

type Booking struct {
	ID             string
	SlotID         string
	UserID         string
	Status         string
	ConferenceLink string
	CreatedAt      time.Time
}
