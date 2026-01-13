package http

import (
	"context"

	bookingports "paku-commerce/internal/commerce/cart/ports/booking"
	checkoutports "paku-commerce/internal/commerce/cart/ports/checkout"
	cartusecases "paku-commerce/internal/commerce/cart/usecases"
	bookingstub "paku-commerce/internal/commerce/checkout/ports/booking"
	checkoutusecases "paku-commerce/internal/commerce/checkout/usecases"
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

// InProcessBookingClient implementa bookingports.BookingClient (stub no-op).
type InProcessBookingClient struct{}

func (c *InProcessBookingClient) CancelHold(ctx context.Context, holdID string) error {
	// Stub: no-op por ahora
	return nil
}

// WireCartHandlers construye todas las dependencias de cart.
func WireCartHandlers() *CartHandlers {
	cartRepo := runtime.CartRepoSingleton
	orderRepo := runtime.OrderRepoSingleton

	// Port: booking (stub)
	var bookingClient bookingports.BookingClient = &InProcessBookingClient{}

	// Port: checkout (in-process)
	cancelOrderUC := &checkoutusecases.CancelOrder{
		Repo:    orderRepo,
		Booking: &bookingstub.StubBookingClient{},
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
