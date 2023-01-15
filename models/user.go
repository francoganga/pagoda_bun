package models

import (
	"time"

	"github.com/uptrace/bun"
)

type User struct {
	bun.BaseModel `bun:"users,alias:u"`

	ID int `bun:"id,pk,autoincrement"`

	Name string

	Email string

	Password string

	Verified bool

	CreatedAt time.Time `bun:",nullzero,notnull,default:current_timestamp"`

	Tokens []*PasswordToken `bun:"rel:has-many,join:id=user_id"`
}

type PasswordToken struct {
	bun.BaseModel `bun:"password_tokens,alias:pt"`

	ID   int `bun:"id,pk,autoincrement"`
	Hash string

    CreatedAt time.Time `bun:",nullzero,notnull,default:current_timestamp"`

	UserID int
}
