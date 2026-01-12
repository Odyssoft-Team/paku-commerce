package memory

import (
	"context"

	pricingdomain "paku-commerce/internal/pricing/domain"
	"paku-commerce/internal/promotions/domain"
)

// PromotionsRepository implementa domain.PromotionsRepository en memoria.
type PromotionsRepository struct {
	coupons    map[string]domain.Coupon
	promotions []domain.Promotion
}

// NewPromotionsRepository crea un repositorio con datos de ejemplo.
func NewPromotionsRepository() *PromotionsRepository {
	coupons := map[string]domain.Coupon{
		"BANO10": {
			Code:               "BANO10",
			Active:             true,
			PercentOff:         10,
			AppliesToItemTypes: []string{"service"},
			MinSubtotalAmount:  0, // sin mínimo
			Currency:           pricingdomain.CurrencyPEN,
		},
	}

	promotions := []domain.Promotion{
		{
			Name:               "Tuesday Grooming",
			Active:             true,
			PercentOff:         5,
			AppliesToItemTypes: []string{"service"},
			Currency:           pricingdomain.CurrencyPEN,
		},
	}

	return &PromotionsRepository{
		coupons:    coupons,
		promotions: promotions,
	}
}

// GetCouponByCode busca un cupón por código normalizado.
func (r *PromotionsRepository) GetCouponByCode(ctx context.Context, code string) (domain.Coupon, error) {
	normalizedCode := domain.NormalizeCode(code)
	coupon, exists := r.coupons[normalizedCode]
	if !exists {
		return domain.Coupon{}, domain.ErrCouponNotFound
	}
	return coupon, nil
}

// ListActivePromotions retorna todas las promociones activas.
func (r *PromotionsRepository) ListActivePromotions(ctx context.Context) ([]domain.Promotion, error) {
	var active []domain.Promotion
	for _, promo := range r.promotions {
		if promo.Active {
			active = append(active, promo)
		}
	}
	return active, nil
}
