package usecases

import (
	"context"
	"errors"
	"fmt"

	checkoutdomain "paku-commerce/internal/commerce/checkout/domain"
	servicedomain "paku-commerce/internal/commerce/service/domain"
	pricingdomain "paku-commerce/internal/pricing/domain"
	pricingusecases "paku-commerce/internal/pricing/usecases"
	promotionsusecases "paku-commerce/internal/promotions/usecases"
)

var (
	ErrMissingParentService = errors.New("addon requires parent service")
	ErrServiceNotEligible   = errors.New("service not eligible for pet")
	ErrInvalidQuantity      = errors.New("quantity must be greater than zero")
	ErrUnknownService       = errors.New("service not found")
)

// CheckoutQuote contiene la cotización completa del checkout.
type CheckoutQuote struct {
	OriginalSubtotal pricingdomain.Money
	Quote            pricingdomain.Quote
	Discounts        []promotionsusecases.DiscountLine
	TotalDiscount    pricingdomain.Money
	Total            pricingdomain.Money
}

// QuoteCheckoutInput contiene la intención de compra.
type QuoteCheckoutInput struct {
	Intent checkoutdomain.PurchaseIntent
}

// QuoteCheckoutOutput contiene la cotización completa.
type QuoteCheckoutOutput struct {
	Quote CheckoutQuote
}

// QuoteCheckout valida y cotiza una intención de compra.
type QuoteCheckout struct {
	ServiceRepo  servicedomain.ServiceRepository
	PriceQuoteUC *pricingusecases.QuoteItems
	PromotionsUC *promotionsusecases.ApplyDiscounts
}

// Execute ejecuta la cotización del checkout.
func (uc QuoteCheckout) Execute(ctx context.Context, input QuoteCheckoutInput) (QuoteCheckoutOutput, error) {
	intent := input.Intent

	// 1. Validar items
	if err := uc.validateItems(ctx, intent); err != nil {
		return QuoteCheckoutOutput{}, err
	}

	// 2. Construir request para pricing
	priceRequest := pricingusecases.QuoteItemsInput{
		PetProfile: intent.PetProfile,
		Items:      make([]pricingusecases.QuoteRequestItem, 0, len(intent.Items)),
	}

	for _, item := range intent.Items {
		priceRequest.Items = append(priceRequest.Items, pricingusecases.QuoteRequestItem{
			ItemType: pricingdomain.ItemType(item.ItemType),
			ItemID:   item.ItemID,
			Qty:      item.Qty,
		})
	}

	// 3. Ejecutar cotización de precios
	priceOutput, err := uc.PriceQuoteUC.Execute(ctx, priceRequest)
	if err != nil {
		return QuoteCheckoutOutput{}, err
	}

	// Guardar subtotal original (antes de descuentos)
	originalSubtotal := priceOutput.Quote.Subtotal

	// 4. Aplicar promociones/cupón
	promoInput := promotionsusecases.ApplyDiscountsInput{
		Quote:      priceOutput.Quote,
		CouponCode: intent.CouponCode,
	}

	promoOutput, err := uc.PromotionsUC.Execute(ctx, promoInput)
	if err != nil {
		return QuoteCheckoutOutput{}, err
	}

	// 5. Calcular total (subtotal post-descuento)
	total := promoOutput.AdjustedQuote.Subtotal

	return QuoteCheckoutOutput{
		Quote: CheckoutQuote{
			OriginalSubtotal: originalSubtotal,
			Quote:            promoOutput.AdjustedQuote,
			Discounts:        promoOutput.Discounts,
			TotalDiscount:    promoOutput.TotalDiscount,
			Total:            total,
		},
	}, nil
}

// validateItems valida la intención de compra.
func (uc QuoteCheckout) validateItems(ctx context.Context, intent checkoutdomain.PurchaseIntent) error {
	// Mapa para tracking de servicios presentes
	serviceItems := make(map[string]bool)

	// Primera pasada: registrar servicios
	for _, item := range intent.Items {
		if item.Qty <= 0 {
			return ErrInvalidQuantity
		}
		if item.ItemType == checkoutdomain.ItemTypeService {
			serviceItems[item.ItemID] = true
		}
	}

	// Segunda pasada: validar servicios
	for _, item := range intent.Items {
		if item.ItemType == checkoutdomain.ItemTypeService {
			service, err := uc.ServiceRepo.GetServiceByID(ctx, item.ItemID)
			if err != nil {
				return fmt.Errorf("%w: %w", ErrUnknownService, err)
			}

			// Validar elegibilidad
			if !service.IsEligibleFor(intent.PetProfile) {
				return ErrServiceNotEligible
			}

			// Validar dependencias addon/parent
			if service.IsAddon && len(service.RequiresParentIDs) > 0 {
				hasParent := false
				for _, parentID := range service.RequiresParentIDs {
					if serviceItems[parentID] {
						hasParent = true
						break
					}
				}
				if !hasParent {
					return ErrMissingParentService
				}
			}
		}
		// Products: no validar nada adicional en este prompt
	}

	return nil
}
