package usecases

import (
	"context"

	checkoutdomain "paku-commerce/internal/commerce/checkout/domain"
	platformbooking "paku-commerce/internal/commerce/platform/booking"
)

// CancelOrderInput contiene el ID de la orden a cancelar.
type CancelOrderInput struct {
	OrderID string
}

// CancelOrderOutput contiene la orden cancelada.
type CancelOrderOutput struct {
	Order checkoutdomain.Order
}

// CancelOrder cancela una orden de forma idempotente.
type CancelOrder struct {
	Repo    checkoutdomain.OrderRepository
	Booking platformbooking.Client
}

// Execute cancela la orden y libera el hold de booking si existe.
func (uc CancelOrder) Execute(ctx context.Context, input CancelOrderInput) (CancelOrderOutput, error) {
	// 1. Cargar la orden
	order, err := uc.Repo.GetByID(ctx, input.OrderID)
	if err != nil {
		return CancelOrderOutput{}, err
	}

	// 2. Verificar si ya estaba cancelada (idempotente)
	wasAlreadyCancelled := order.Status == checkoutdomain.OrderStatusCancelled

	// 3. Marcar como cancelada
	err = order.MarkCancelled()
	if err != nil {
		return CancelOrderOutput{}, err
	}

	// 4. Si ya estaba cancelada, retornar sin side effects
	if wasAlreadyCancelled {
		return CancelOrderOutput{Order: order}, nil
	}

	// 5. Cancelar hold de booking si existe (solo en transición real)
	if order.BookingHoldID != nil && *order.BookingHoldID != "" {
		// No fallar si la cancelación de hold falla (best-effort)
		// TODO: confirm with architect si queremos retry o compensación
		_ = uc.Booking.CancelHold(ctx, *order.BookingHoldID)
	}

	// 6. Persistir orden cancelada
	updatedOrder, err := uc.Repo.Update(ctx, order)
	if err != nil {
		return CancelOrderOutput{}, err
	}

	return CancelOrderOutput{Order: updatedOrder}, nil
}
