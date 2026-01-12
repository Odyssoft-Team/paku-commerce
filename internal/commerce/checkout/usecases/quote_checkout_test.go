package usecases

import (
	"context"
	"testing"

	checkoutdomain "paku-commerce/internal/commerce/checkout/domain"
	servicememory "paku-commerce/internal/commerce/service/adapters/memory"
	servicedomain "paku-commerce/internal/commerce/service/domain"
	pricingmemory "paku-commerce/internal/pricing/adapters/memory"
	pricingusecases "paku-commerce/internal/pricing/usecases"
	promotionsmemory "paku-commerce/internal/promotions/adapters/memory"
	promotionsusecases "paku-commerce/internal/promotions/usecases"
)

func TestQuoteCheckout_AddonWithoutParent(t *testing.T) {
	// Setup
	serviceRepo := servicememory.NewServiceRepository()
	pricingRepo := pricingmemory.NewPriceRuleRepository()
	promoRepo := promotionsmemory.NewPromotionsRepository()

	quoteItemsUC := &pricingusecases.QuoteItems{
		RuleRepo: pricingRepo,
	}

	applyDiscountsUC := &promotionsusecases.ApplyDiscounts{
		Repo: promoRepo,
	}

	uc := &QuoteCheckout{
		ServiceRepo:  serviceRepo,
		PriceQuoteUC: quoteItemsUC,
		PromotionsUC: applyDiscountsUC,
	}

	// Intent: deshedding (addon) sin parent bath
	intent := checkoutdomain.PurchaseIntent{
		PetProfile: servicedomain.PetProfile{
			Species:  servicedomain.SpeciesDog,
			WeightKg: 15,
			CoatType: servicedomain.CoatTypeDouble,
		},
		Items: []checkoutdomain.PurchaseItem{
			{
				ItemType: checkoutdomain.ItemTypeService,
				ItemID:   "deshedding",
				Qty:      1,
			},
		},
	}

	// Execute
	_, err := uc.Execute(context.Background(), QuoteCheckoutInput{Intent: intent})

	// Assert
	if err != ErrMissingParentService {
		t.Errorf("expected ErrMissingParentService, got: %v", err)
	}
}

func TestQuoteCheckout_ServiceNotEligible(t *testing.T) {
	// Setup
	serviceRepo := servicememory.NewServiceRepository()
	pricingRepo := pricingmemory.NewPriceRuleRepository()
	promoRepo := promotionsmemory.NewPromotionsRepository()

	quoteItemsUC := &pricingusecases.QuoteItems{
		RuleRepo: pricingRepo,
	}

	applyDiscountsUC := &promotionsusecases.ApplyDiscounts{
		Repo: promoRepo,
	}

	uc := &QuoteCheckout{
		ServiceRepo:  serviceRepo,
		PriceQuoteUC: quoteItemsUC,
		PromotionsUC: applyDiscountsUC,
	}

	// Intent: dematting (desmotado) no permitido para hairless
	intent := checkoutdomain.PurchaseIntent{
		PetProfile: servicedomain.PetProfile{
			Species:  servicedomain.SpeciesDog,
			WeightKg: 10,
			CoatType: servicedomain.CoatTypeHairless, // dematting excluye hairless
		},
		Items: []checkoutdomain.PurchaseItem{
			{
				ItemType: checkoutdomain.ItemTypeService,
				ItemID:   "bath",
				Qty:      1,
			},
			{
				ItemType: checkoutdomain.ItemTypeService,
				ItemID:   "dematting",
				Qty:      1,
			},
		},
	}

	// Execute
	_, err := uc.Execute(context.Background(), QuoteCheckoutInput{Intent: intent})

	// Assert
	if err != ErrServiceNotEligible {
		t.Errorf("expected ErrServiceNotEligible, got: %v", err)
	}
}

func TestQuoteCheckout_ValidIntent(t *testing.T) {
	// Setup
	serviceRepo := servicememory.NewServiceRepository()
	pricingRepo := pricingmemory.NewPriceRuleRepository()
	promoRepo := promotionsmemory.NewPromotionsRepository()

	quoteItemsUC := &pricingusecases.QuoteItems{
		RuleRepo: pricingRepo,
	}

	applyDiscountsUC := &promotionsusecases.ApplyDiscounts{
		Repo: promoRepo,
	}

	uc := &QuoteCheckout{
		ServiceRepo:  serviceRepo,
		PriceQuoteUC: quoteItemsUC,
		PromotionsUC: applyDiscountsUC,
	}

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
			{
				ItemType: checkoutdomain.ItemTypeService,
				ItemID:   "deshedding",
				Qty:      1,
			},
		},
		CouponCode: &couponCode,
	}

	// Execute
	output, err := uc.Execute(context.Background(), QuoteCheckoutInput{Intent: intent})

	// Assert
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if output.Quote.Total.Amount >= output.Quote.OriginalSubtotal.Amount {
		t.Errorf("expected total < subtotal due to discount")
	}

	if len(output.Quote.Discounts) == 0 {
		t.Errorf("expected at least one discount")
	}
}
