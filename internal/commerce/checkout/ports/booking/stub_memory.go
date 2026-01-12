package booking

import "context"

// StubBookingClient es un stub no-op de BookingClient para desarrollo.
type StubBookingClient struct{}

// ValidateHold no hace nada (stub).
func (s *StubBookingClient) ValidateHold(ctx context.Context, holdID string) error {
	return nil
}

// ConfirmHold no hace nada (stub).
func (s *StubBookingClient) ConfirmHold(ctx context.Context, holdID string) error {
	return nil
}

// CancelHold no hace nada (stub).
func (s *StubBookingClient) CancelHold(ctx context.Context, holdID string) error {
	return nil
}
