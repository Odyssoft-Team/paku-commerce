package domain

import (
	"context"
	"errors"
)

var ErrCouponNotFound = errors.New("coupon not found")

// PromotionsRepository define el acceso a cupones y promociones.
type PromotionsRepository interface {
	GetCouponByCode(ctx context.Context, code string) (Coupon, error)
	ListActivePromotions(ctx context.Context) ([]Promotion, error)
}
