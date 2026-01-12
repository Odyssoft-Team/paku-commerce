package payments

import "context"

// PaymentsClient define la integración con el proveedor de pagos.
// TODO: implementar en fase de integración
type PaymentsClient interface {
	// ValidatePayment verifica que una referencia de pago sea válida.
	ValidatePayment(ctx context.Context, paymentRef string) error
}
