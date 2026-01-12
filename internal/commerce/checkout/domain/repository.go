package domain

import (
	"context"
	"errors"
)

var ErrOrderNotFound = errors.New("order not found")

// OrderRepository define el acceso a órdenes.
type OrderRepository interface {
	Create(ctx context.Context, order Order) (Order, error)
	GetByID(ctx context.Context, id string) (Order, error)
	Update(ctx context.Context, order Order) (Order, error)
}
