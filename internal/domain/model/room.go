package model

import "time"

type Room struct {
	ID          string
	Name        string
	Description *string
	Capacity    *int
	CreatedAt   *time.Time
}
