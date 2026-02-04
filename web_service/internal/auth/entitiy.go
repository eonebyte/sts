package auth

import "time"

type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

type User struct {
	ID             int64     `db:"AD_USER_ID" json:"ad_user_id"`
	Name           string    `db:"NAME" json:"username"`
	HashedPassword string    `db:"hashed_password" json:"-"`
	Password       string    `db:"PASSWORD" json:"-"`
	Title          *string   `db:"TITLE" json:"title"`
	Created        time.Time `db:"created" json:"created"`
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token,omitempty"`
}

type RegisterRequest struct {
	Username string `json:"username" validate:"required,min=3,max=30,alphanum"`
	Password string `json:"password" validate:"required,min=3"`
}

type LoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}
