package order

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("order not found")

// Repository is the only place that talks SQL. Handlers never see the pool.
type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) Create(ctx context.Context, userID string, amount int64) (Order, error) {
	const q = `
		INSERT INTO orders (user_id, amount, status)
		VALUES ($1, $2, $3)
		RETURNING id::text, user_id, amount, status, created_at`

	var o Order
	err := r.pool.QueryRow(ctx, q, userID, amount, StatusPending).
		Scan(&o.ID, &o.UserID, &o.Amount, &o.Status, &o.CreatedAt)
	return o, err
}

func (r *Repository) GetByID(ctx context.Context, id string) (Order, error) {
	const q = `
		SELECT id::text, user_id, amount, status, created_at
		FROM orders
		WHERE id = $1`

	var o Order
	err := r.pool.QueryRow(ctx, q, id).
		Scan(&o.ID, &o.UserID, &o.Amount, &o.Status, &o.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return Order{}, ErrNotFound
	}
	return o, err
}
