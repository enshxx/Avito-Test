package entity

import (
	"github.com/google/uuid"
	"time"
)

type Bid struct {
	Id          uuid.UUID `db:"id"`
	Name        string    `db:"name"`
	Description string    `db:"description"`
	TenderId    uuid.UUID `db:"tender_id"`
	Status      string    `db:"status"`
	Decision    *string   `db:"decision"`
	AuthorType  string    `db:"author_type"`
	AuthorId    uuid.UUID `db:"author_id"`
	Version     int       `db:"version"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}
