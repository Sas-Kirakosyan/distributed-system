// Package order contains the Order model, its Postgres repository and HTTP handlers.
package order

import "time"

type Status string

const (
	StatusPending Status = "PENDING"
)

type Order struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Amount    int64     `json:"amount"` // in cents: never use floats for money
	Status    Status    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}
