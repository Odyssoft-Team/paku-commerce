package http

import (
	bookingstub "paku-commerce/internal/commerce/checkout/ports/booking"
	checkoutusecases "paku-commerce/internal/commerce/checkout/usecases"
	"paku-commerce/internal/commerce/runtime"
	servicememory "paku-commerce/internal/commerce/service/adapters/memory"
	pricingmemory "paku-commerce/internal/pricing/adapters/memory"
	pricingusecases "paku-commerce/internal/pricing/usecases"
	promotionsmemory "paku-commerce/internal/promotions/adapters/memory"
	promotionsusecases "paku-commerce/internal/promotions/usecases"
)

// WireCheckoutHandlers construye todas las dependencias y retorna handlers.
func WireCheckoutHandlers() *CheckoutHandlers {
	// Repos (singletons)
	serviceRepo := servicememory.NewServiceRepository()
	priceRuleRepo := pricingmemory.NewPriceRuleRepository()
	promotionsRepo := promotionsmemory.NewPromotionsRepository()
	orderRepo := runtime.OrderRepoSingleton
	cartRepo := runtime.CartRepoSingleton

	// Booking stub (no-op)
	bookingClient := &bookingstub.StubBookingClient{}

	// Usecases: pricing
	quoteItemsUC := &pricingusecases.QuoteItems{
		RuleRepo: priceRuleRepo,
	}

	// Usecases: promotions
	applyDiscountsUC := &promotionsusecases.ApplyDiscounts{
		Repo: promotionsRepo,
	}

	// Usecases: checkout
	quoteCheckoutUC := &checkoutusecases.QuoteCheckout{
		ServiceRepo:  serviceRepo,
		PriceQuoteUC: quoteItemsUC,
		PromotionsUC: applyDiscountsUC,
	}

	createOrderUC := &checkoutusecases.CreateOrder{
		QuoteCheckoutUC: quoteCheckoutUC,
		OrderRepo:       orderRepo,
		Now:             nil, // usa time.Now() por defecto
	}

	confirmPaymentUC := &checkoutusecases.ConfirmPayment{
		Repo:    orderRepo,
		Booking: bookingClient,
		Now:     nil, // usa time.Now() por defecto
	}

	startCheckoutUC := &checkoutusecases.StartCheckout{
		CartRepo:      cartRepo,
		Booking:       bookingClient,
		CreateOrderUC: createOrderUC,
	}

	return &CheckoutHandlers{
		QuoteCheckoutUC:  quoteCheckoutUC,
		CreateOrderUC:    createOrderUC,
		ConfirmPaymentUC: confirmPaymentUC,
		StartCheckoutUC:  startCheckoutUC,
	}
}
