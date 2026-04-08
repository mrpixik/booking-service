package model

import "time"

type Slot struct {
	ID     string
	RoomID string
	Start  time.Time
	End    time.Time
}
