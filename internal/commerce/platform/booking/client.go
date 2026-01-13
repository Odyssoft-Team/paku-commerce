package booking

import "context"

// Client define la integración con el servicio de booking.
// Este interface unifica las operaciones de booking usadas por cart y checkout.
type Client interface {
	// CreateHold crea un hold de booking para un slot.
	CreateHold(ctx context.Context, slotID string) (holdID string, err error)

	// ValidateHold verifica que un hold de booking sea válido.
	ValidateHold(ctx context.Context, holdID string) error

	// ConfirmHold confirma un hold de booking tras pago exitoso.
	ConfirmHold(ctx context.Context, holdID string) error

	// CancelHold cancela un hold de booking.
	CancelHold(ctx context.Context, holdID string) error
}
