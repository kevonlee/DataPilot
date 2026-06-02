package model

import "time"

type User struct {
	Username  string    `json:"username"`
	Password  string    `json:"password"` // bcrypt hash
	CreatedAt time.Time `json:"createdAt"`
}
