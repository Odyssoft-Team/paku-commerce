package usecases

import (
	"context"

	pricingdomain "paku-commerce/internal/pricing/domain"
	"paku-commerce/internal/promotions/domain"
)

// DiscountLine representa una línea de descuento aplicada.
type DiscountLine struct {
	Source string // "coupon" o "promotion"
	Name   string // código o nombre de promo
	Amount pricingdomain.Money
}

// ApplyDiscountsInput contiene el quote y opcionalmente un cupón.
type ApplyDiscountsInput struct {
	Quote      pricingdomain.Quote
	CouponCode *string
}

// ApplyDiscountsOutput contiene el quote ajustado y desglose de descuentos.
type ApplyDiscountsOutput struct {
	AdjustedQuote pricingdomain.Quote
	Discounts     []DiscountLine
	TotalDiscount pricingdomain.Money
}

// ApplyDiscounts aplica cupón y promociones al quote.
type ApplyDiscounts struct {
	Repo domain.PromotionsRepository
}

// Execute calcula descuentos y retorna quote ajustado.
func (uc ApplyDiscounts) Execute(ctx context.Context, input ApplyDiscountsInput) (ApplyDiscountsOutput, error) {
	adjustedQuote := input.Quote
	var discounts []DiscountLine
	totalDiscount := pricingdomain.Zero(input.Quote.Subtotal.Currency)
	currentSubtotal := input.Quote.Subtotal

	// 1. Aplicar cupón si existe
	if input.CouponCode != nil && *input.CouponCode != "" {
		normalizedCode := domain.NormalizeCode(*input.CouponCode)
		coupon, err := uc.Repo.GetCouponByCode(ctx, normalizedCode)
		if err != nil {
			return ApplyDiscountsOutput{}, err
		}

		if !coupon.IsApplicable(currentSubtotal, input.Quote.Items) {
			return ApplyDiscountsOutput{}, ErrInvalidCoupon
		}

		discountAmount := calculateDiscount(currentSubtotal, coupon.PercentOff)

		newSubtotal, err := currentSubtotal.Sub(discountAmount)
		if err != nil {
			return ApplyDiscountsOutput{}, err
		}
		currentSubtotal = newSubtotal

		newTotalDiscount, err := totalDiscount.Add(discountAmount)
		if err != nil {
			return ApplyDiscountsOutput{}, err
		}
		totalDiscount = newTotalDiscount

		discounts = append(discounts, DiscountLine{
			Source: "coupon",
			Name:   coupon.Code,
			Amount: discountAmount,
		})
	}

	// 2. Aplicar promociones activas
	promos, err := uc.Repo.ListActivePromotions(ctx)
	if err != nil {
		return ApplyDiscountsOutput{}, err
	}

	for _, promo := range promos {
		if promo.IsApplicable(currentSubtotal, input.Quote.Items) {
			discountAmount := calculateDiscount(currentSubtotal, promo.PercentOff)

			newSubtotal, err := currentSubtotal.Sub(discountAmount)
			if err != nil {
				return ApplyDiscountsOutput{}, err
			}
			currentSubtotal = newSubtotal

			newTotalDiscount, err := totalDiscount.Add(discountAmount)
			if err != nil {
				return ApplyDiscountsOutput{}, err
			}
			totalDiscount = newTotalDiscount

			discounts = append(discounts, DiscountLine{
				Source: "promotion",
				Name:   promo.Name,
				Amount: discountAmount,
			})
		}
	}

	// Actualizar quote con nuevo subtotal
	adjustedQuote.Subtotal = currentSubtotal

	return ApplyDiscountsOutput{
		AdjustedQuote: adjustedQuote,
		Discounts:     discounts,
		TotalDiscount: totalDiscount,
	}, nil
}

// calculateDiscount calcula el descuento porcentual (redondeo hacia abajo).
func calculateDiscount(amount pricingdomain.Money, percentOff int) pricingdomain.Money {
	if percentOff < 0 {
		percentOff = 0
	}
	if percentOff > 100 {
		percentOff = 100
	}

	discountAmount := (amount.Amount * int64(percentOff)) / 100

	// Clamp para evitar descuento mayor al monto
	if discountAmount > amount.Amount {
		discountAmount = amount.Amount
	}

	return pricingdomain.Money{
		Amount:   discountAmount,
		Currency: amount.Currency,
	}
}
