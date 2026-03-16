package user

import (
	"errors"
	"time"
)

var ErrUserNotFound = errors.New("user not found")

type User struct {
	ID        int       `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Password  string    `json:"password"`
	CreatedAt time.Time `json:"created_at"`
}
