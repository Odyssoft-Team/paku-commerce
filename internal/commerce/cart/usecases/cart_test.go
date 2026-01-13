package usecases

import (
	"context"
	"testing"
	"time"

	cartmemory "paku-commerce/internal/commerce/cart/adapters/memory"
	cartdomain "paku-commerce/internal/commerce/cart/domain"
	checkoutdomain "paku-commerce/internal/commerce/checkout/domain"
	servicedomain "paku-commerce/internal/commerce/service/domain"
)

func TestUpsertCart_CreatesNew(t *testing.T) {
	repo := cartmemory.NewCartRepository()
	uc := &UpsertCart{Repo: repo, Now: func() time.Time { return time.Date(2026, 1, 15, 10, 0, 0, 0, time.UTC) }}

	input := UpsertCartInput{
		UserID: "user_1",
		PetProfile: servicedomain.PetProfile{
			Species:  servicedomain.SpeciesDog,
			WeightKg: 15,
			CoatType: servicedomain.CoatTypeShort,
		},
		Items: []checkoutdomain.PurchaseItem{
			{ItemType: checkoutdomain.ItemTypeService, ItemID: "bath", Qty: 1},
		},
	}

	output, err := uc.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if output.Cart.UserID != "user_1" {
		t.Errorf("expected user_id user_1")
	}
	if len(output.Cart.Items) != 1 {
		t.Errorf("expected 1 item")
	}
}

func TestGetCart_Expired_ReturnsNotFound(t *testing.T) {
	repo := cartmemory.NewCartRepository()
	now := time.Date(2026, 1, 15, 10, 0, 0, 0, time.UTC)

	// Crear carrito
	cart := cartdomain.NewCart("user_1", servicedomain.PetProfile{}, []checkoutdomain.PurchaseItem{{ItemType: "service", ItemID: "bath", Qty: 1}}, now)
	repo.Upsert(context.Background(), cart)

	// Intentar obtener después de expiración
	uc := &GetCart{Repo: repo, Now: func() time.Time { return now.Add(100 * time.Minute) }}
	_, err := uc.Execute(context.Background(), GetCartInput{UserID: "user_1"})

	if err != cartdomain.ErrCartNotFound {
		t.Errorf("expected ErrCartNotFound for expired cart, got: %v", err)
	}
}

func TestExpireCarts_DeletesExpired(t *testing.T) {
	repo := cartmemory.NewCartRepository()
	now := time.Date(2026, 1, 15, 10, 0, 0, 0, time.UTC)

	// Crear carrito que expirará
	cart := cartdomain.NewCart("user_1", servicedomain.PetProfile{}, []checkoutdomain.PurchaseItem{{ItemType: "service", ItemID: "bath", Qty: 1}}, now)
	repo.Upsert(context.Background(), cart)

	// Stubs para ports
	bookingStub := &stubBookingClient{}
	checkoutStub := &stubCheckoutClient{}

	uc := &ExpireCarts{Repo: repo, Booking: bookingStub, Checkout: checkoutStub}
	output, err := uc.Execute(context.Background(), ExpireCartsInput{Now: now.Add(100 * time.Minute)})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if output.ExpiredCount != 1 {
		t.Errorf("expected 1 expired cart")
	}

	// Verificar que fue eliminado
	_, err = repo.GetByUserID(context.Background(), "user_1")
	if err != cartdomain.ErrCartNotFound {
		t.Errorf("expected cart to be deleted")
	}
}

// Stubs para tests
type stubBookingClient struct{}

func (s *stubBookingClient) CancelHold(ctx context.Context, holdID string) error {
	return nil
}

type stubCheckoutClient struct{}

func (s *stubCheckoutClient) CancelOrder(ctx context.Context, orderID string) error {
	return nil
}
