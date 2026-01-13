package booking

import (
	"context"
	"crypto/rand"
	"encoding/hex"
)

// StubBookingClient es un stub no-op de BookingClient para desarrollo.
type StubBookingClient struct{}

// CreateHold genera un hold ID stub.
func (s *StubBookingClient) CreateHold(ctx context.Context, slotID string) (string, error) {
	b := make([]byte, 8)
	rand.Read(b)
	return "hold_" + hex.EncodeToString(b), nil
}

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
