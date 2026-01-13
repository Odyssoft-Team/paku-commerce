package booking

import (
	"context"
	"crypto/rand"
	"encoding/hex"
)

// StubClient es un stub no-op de Client para desarrollo.
type StubClient struct{}

// CreateHold genera un hold ID stub.
func (s *StubClient) CreateHold(ctx context.Context, slotID string) (string, error) {
	b := make([]byte, 8)
	rand.Read(b)
	return "hold_" + hex.EncodeToString(b), nil
}

// ValidateHold no hace nada (stub).
func (s *StubClient) ValidateHold(ctx context.Context, holdID string) error {
	return nil
}

// ConfirmHold no hace nada (stub).
func (s *StubClient) ConfirmHold(ctx context.Context, holdID string) error {
	return nil
}

// CancelHold no hace nada (stub).
func (s *StubClient) CancelHold(ctx context.Context, holdID string) error {
	return nil
}
