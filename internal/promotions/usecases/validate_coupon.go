package usecases

import (
	"context"
	"errors"

	pricingdomain "paku-commerce/internal/pricing/domain"
	"paku-commerce/internal/promotions/domain"
)

var ErrInvalidCoupon = errors.New("coupon is not applicable")

// ValidateCouponInput contiene el código del cupón y el quote.
type ValidateCouponInput struct {
	Code  string
	Quote pricingdomain.Quote
}

// ValidateCouponOutput contiene el cupón validado.
type ValidateCouponOutput struct {
	Coupon domain.Coupon
}

// ValidateCoupon valida si un cupón es aplicable al quote.
type ValidateCoupon struct {
	Repo domain.PromotionsRepository
}

// Execute valida el cupón sin aplicar descuentos.
func (uc ValidateCoupon) Execute(ctx context.Context, input ValidateCouponInput) (ValidateCouponOutput, error) {
	// Normalizar código
	normalizedCode := domain.NormalizeCode(input.Code)

	// Buscar cupón
	coupon, err := uc.Repo.GetCouponByCode(ctx, normalizedCode)
	if err != nil {
		return ValidateCouponOutput{}, err
	}

	// Validar aplicabilidad
	if !coupon.IsApplicable(input.Quote.Subtotal, input.Quote.Items) {
		return ValidateCouponOutput{}, ErrInvalidCoupon
	}

	return ValidateCouponOutput{Coupon: coupon}, nil
}
