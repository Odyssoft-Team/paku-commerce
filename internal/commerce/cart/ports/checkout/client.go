package checkout

import "context"

// CheckoutClient define operaciones de checkout para cart.
type CheckoutClient interface {
	// CancelOrder cancela una orden.
	CancelOrder(ctx context.Context, orderID string) error
}
