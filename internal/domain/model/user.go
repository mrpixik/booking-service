package model

import "time"

type Role string

const (
	RoleAdmin Role = "admin"
	RoleUser  Role = "user"
)

const (
	DummyAdminID = "00000000-0000-0000-0000-000000000001"
	DummyUserID  = "00000000-0000-0000-0000-000000000002"
)

type ctxKey string

const (
	CtxUserID ctxKey = "user_id"
	CtxRole   ctxKey = "role"
)

type User struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	Role         Role      `json:"role"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"createdAt"`
}
