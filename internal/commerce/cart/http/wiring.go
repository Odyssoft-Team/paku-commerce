package http

import (
	"context"

	checkoutports "paku-commerce/internal/commerce/cart/ports/checkout"
	cartusecases "paku-commerce/internal/commerce/cart/usecases"
	checkoutusecases "paku-commerce/internal/commerce/checkout/usecases"
	platformbooking "paku-commerce/internal/commerce/platform/booking"
	"paku-commerce/internal/commerce/runtime"
)

// InProcessCheckoutClient implementa checkoutports.CheckoutClient usando checkout domain directamente.
type InProcessCheckoutClient struct {
	CancelOrderUC *checkoutusecases.CancelOrder
}

func (c *InProcessCheckoutClient) CancelOrder(ctx context.Context, orderID string) error {
	input := checkoutusecases.CancelOrderInput{OrderID: orderID}
	_, err := c.CancelOrderUC.Execute(ctx, input)
	return err
}

// InProcessBookingClient implementa platform booking.Client (stub no-op).
type InProcessBookingClient struct{}

func (c *InProcessBookingClient) CreateHold(ctx context.Context, slotID string) (string, error) {
	return "", nil // stub
}

func (c *InProcessBookingClient) ValidateHold(ctx context.Context, holdID string) error {
	return nil
}

func (c *InProcessBookingClient) ConfirmHold(ctx context.Context, holdID string) error {
	return nil
}

func (c *InProcessBookingClient) CancelHold(ctx context.Context, holdID string) error {
	return nil
}

// WireCartHandlers construye todas las dependencias de cart.
func WireCartHandlers() *CartHandlers {
	cartRepo := runtime.CartRepoSingleton
	orderRepo := runtime.OrderRepoSingleton

	// Port: booking (stub)
	var bookingClient platformbooking.Client = &InProcessBookingClient{}

	// Port: checkout (in-process)
	cancelOrderUC := &checkoutusecases.CancelOrder{
		Repo:    orderRepo,
		Booking: &platformbooking.StubClient{},
	}
	var checkoutClient checkoutports.CheckoutClient = &InProcessCheckoutClient{CancelOrderUC: cancelOrderUC}

	// Usecases
	upsertCartUC := &cartusecases.UpsertCart{
		Repo: cartRepo,
		Now:  nil,
	}

	getCartUC := &cartusecases.GetCart{
		Repo: cartRepo,
		Now:  nil,
	}

	deleteCartUC := &cartusecases.DeleteCart{
		Repo: cartRepo,
	}

	expireCartsUC := &cartusecases.ExpireCarts{
		Repo:     cartRepo,
		Booking:  bookingClient,
		Checkout: checkoutClient,
	}

	return &CartHandlers{
		UpsertCartUC:  upsertCartUC,
		GetCartUC:     getCartUC,
		DeleteCartUC:  deleteCartUC,
		ExpireCartsUC: expireCartsUC,
	}
}
