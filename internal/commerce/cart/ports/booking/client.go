package booking

import "context"

// BookingClient define operaciones de booking para cart.
type BookingClient interface {
	// CancelHold cancela un hold de booking.
	CancelHold(ctx context.Context, holdID string) error
}
