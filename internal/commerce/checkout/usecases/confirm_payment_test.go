package usecases

import (
	"context"
	"errors"
	"testing"
	"time"

	checkoutmemory "paku-commerce/internal/commerce/checkout/adapters/memory"
	checkoutdomain "paku-commerce/internal/commerce/checkout/domain"
	bookingstub "paku-commerce/internal/commerce/checkout/ports/booking"
	servicememory "paku-commerce/internal/commerce/service/adapters/memory"
	servicedomain "paku-commerce/internal/commerce/service/domain"
	pricingmemory "paku-commerce/internal/pricing/adapters/memory"
	pricingusecases "paku-commerce/internal/pricing/usecases"
	promotionsmemory "paku-commerce/internal/promotions/adapters/memory"
	promotionsusecases "paku-commerce/internal/promotions/usecases"
)

func createTestOrder(t *testing.T, orderRepo checkoutdomain.OrderRepository) checkoutdomain.Order {
	// Setup repos
	serviceRepo := servicememory.NewServiceRepository()
	pricingRepo := pricingmemory.NewPriceRuleRepository()
	promoRepo := promotionsmemory.NewPromotionsRepository()

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

	createOrderUC := &CreateOrder{
		QuoteCheckoutUC: quoteCheckoutUC,
		OrderRepo:       orderRepo,
		Now:             time.Now,
	}

	intent := checkoutdomain.PurchaseIntent{
		PetProfile: servicedomain.PetProfile{
			Species:  servicedomain.SpeciesDog,
			WeightKg: 10,
			CoatType: servicedomain.CoatTypeShort,
		},
		Items: []checkoutdomain.PurchaseItem{
			{
				ItemType: checkoutdomain.ItemTypeService,
				ItemID:   "bath",
				Qty:      1,
			},
		},
	}

	output, err := createOrderUC.Execute(context.Background(), CreateOrderInput{Intent: intent})
	if err != nil {
		t.Fatalf("failed to create test order: %v", err)
	}

	return output.Order
}

func TestConfirmPayment_FirstConfirmation(t *testing.T) {
	// Setup
	orderRepo := checkoutmemory.NewOrderRepository()
	order := createTestOrder(t, orderRepo)

	fixedNow := time.Date(2026, 1, 12, 12, 0, 0, 0, time.UTC)
	uc := &ConfirmPayment{
		Repo:    orderRepo,
		Booking: &bookingstub.StubBookingClient{},
		Now:     func() time.Time { return fixedNow },
	}

	// Execute
	output, err := uc.Execute(context.Background(), ConfirmPaymentInput{
		OrderID:    order.ID,
		PaymentRef: "tx_1",
		PaidAt:     fixedNow,
	})

	// Assert
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if output.Order.Status != checkoutdomain.OrderStatusPaid {
		t.Errorf("expected status paid, got: %v", output.Order.Status)
	}

	if output.Order.PaymentRef == nil || *output.Order.PaymentRef != "tx_1" {
		t.Errorf("expected payment_ref tx_1")
	}

	if output.Order.PaidAt == nil {
		t.Errorf("expected paid_at to be set")
	}
}

func TestConfirmPayment_Idempotency(t *testing.T) {
	// Setup
	orderRepo := checkoutmemory.NewOrderRepository()
	order := createTestOrder(t, orderRepo)

	fixedNow := time.Date(2026, 1, 12, 12, 0, 0, 0, time.UTC)
	uc := &ConfirmPayment{
		Repo:    orderRepo,
		Booking: &bookingstub.StubBookingClient{},
		Now:     func() time.Time { return fixedNow },
	}

	// Primera confirmación
	_, err := uc.Execute(context.Background(), ConfirmPaymentInput{
		OrderID:    order.ID,
		PaymentRef: "tx_1",
		PaidAt:     fixedNow,
	})
	if err != nil {
		t.Fatalf("unexpected error on first confirm: %v", err)
	}

	// Segunda confirmación (idempotente)
	output, err := uc.Execute(context.Background(), ConfirmPaymentInput{
		OrderID:    order.ID,
		PaymentRef: "tx_1",
		PaidAt:     fixedNow,
	})

	// Assert
	if err != nil {
		t.Fatalf("unexpected error on idempotent confirm: %v", err)
	}

	if output.Order.Status != checkoutdomain.OrderStatusPaid {
		t.Errorf("expected status paid")
	}

	if output.Order.PaymentRef == nil || *output.Order.PaymentRef != "tx_1" {
		t.Errorf("expected payment_ref tx_1")
	}
}

func TestConfirmPayment_Conflict(t *testing.T) {
	// Setup
	orderRepo := checkoutmemory.NewOrderRepository()
	order := createTestOrder(t, orderRepo)

	fixedNow := time.Date(2026, 1, 12, 12, 0, 0, 0, time.UTC)
	uc := &ConfirmPayment{
		Repo:    orderRepo,
		Booking: &bookingstub.StubBookingClient{},
		Now:     func() time.Time { return fixedNow },
	}

	// Primera confirmación con tx_1
	_, err := uc.Execute(context.Background(), ConfirmPaymentInput{
		OrderID:    order.ID,
		PaymentRef: "tx_1",
		PaidAt:     fixedNow,
	})
	if err != nil {
		t.Fatalf("unexpected error on first confirm: %v", err)
	}

	// Segunda confirmación con tx_2 (conflicto)
	_, err = uc.Execute(context.Background(), ConfirmPaymentInput{
		OrderID:    order.ID,
		PaymentRef: "tx_2",
		PaidAt:     fixedNow,
	})

	// Assert
	if !errors.Is(err, checkoutdomain.ErrPaymentConflict) {
		t.Errorf("expected ErrPaymentConflict, got: %v", err)
	}
}
