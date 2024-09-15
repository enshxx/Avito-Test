package entity

import (
	"github.com/google/uuid"
	"time"
)

type Employee struct {
	Id        uuid.UUID `db:"id"`
	Username  string    `db:"username"`
	FirstName *string   `db:"first_name"`
	LastName  *string   `db:"last_name"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}
