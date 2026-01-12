package usecases

import (
	"context"
	"testing"
	"time"

	checkoutmemory "paku-commerce/internal/commerce/checkout/adapters/memory"
	checkoutdomain "paku-commerce/internal/commerce/checkout/domain"
	servicememory "paku-commerce/internal/commerce/service/adapters/memory"
	servicedomain "paku-commerce/internal/commerce/service/domain"
	pricingmemory "paku-commerce/internal/pricing/adapters/memory"
	pricingusecases "paku-commerce/internal/pricing/usecases"
	promotionsmemory "paku-commerce/internal/promotions/adapters/memory"
	promotionsusecases "paku-commerce/internal/promotions/usecases"
)

func TestCreateOrder_Success(t *testing.T) {
	// Setup repos
	serviceRepo := servicememory.NewServiceRepository()
	pricingRepo := pricingmemory.NewPriceRuleRepository()
	promoRepo := promotionsmemory.NewPromotionsRepository()
	orderRepo := checkoutmemory.NewOrderRepository()

	// Build usecases
	quoteItemsUC := &pricingusecases.QuoteItems{
		RuleRepo: pricingRepo,
	}

	applyDiscountsUC := &promotionsusecases.ApplyDiscounts{
		Repo: promoRepo,
	}

	quoteCheckoutUC := &QuoteCheckout{
		ServiceRepo:  serviceRepo,
		PriceQuoteUC: quoteItemsUC,
		PromotionsUC: applyDiscountsUC,
	}

	fixedNow := time.Date(2026, 1, 12, 10, 0, 0, 0, time.UTC)
	uc := &CreateOrder{
		QuoteCheckoutUC: quoteCheckoutUC,
		OrderRepo:       orderRepo,
		Now:             func() time.Time { return fixedNow },
	}

	// Intent
	couponCode := "BANO10"
	intent := checkoutdomain.PurchaseIntent{
		PetProfile: servicedomain.PetProfile{
			Species:  servicedomain.SpeciesDog,
			WeightKg: 15,
			CoatType: servicedomain.CoatTypeDouble,
		},
		Items: []checkoutdomain.PurchaseItem{
			{
				ItemType: checkoutdomain.ItemTypeService,
				ItemID:   "bath",
				Qty:      1,
			},
		},
		CouponCode: &couponCode,
	}

	// Execute
	output, err := uc.Execute(context.Background(), CreateOrderInput{Intent: intent})

	// Assert
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if output.Order.Status != checkoutdomain.OrderStatusPendingPayment {
		t.Errorf("expected status pending_payment, got: %v", output.Order.Status)
	}

	if output.Order.ID == "" {
		t.Errorf("expected non-empty order ID")
	}

	if output.Order.Total.Amount > output.Order.Subtotal.Amount {
		t.Errorf("expected total <= subtotal")
	}

	if !output.Order.CreatedAt.Equal(fixedNow) {
		t.Errorf("expected CreatedAt to match fixed time")
	}
}
