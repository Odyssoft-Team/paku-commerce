package domain

import (
	"context"
	"time"
)

// CartRepository define el acceso a carritos.
type CartRepository interface {
	Upsert(ctx context.Context, cart Cart) (Cart, error)
	GetByUserID(ctx context.Context, userID string) (Cart, error)
	DeleteByUserID(ctx context.Context, userID string) error
	ListExpired(ctx context.Context, now time.Time) ([]Cart, error)
}
