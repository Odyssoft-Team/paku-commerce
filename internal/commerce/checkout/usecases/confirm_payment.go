package usecases

import (
	"context"
	"time"

	checkoutdomain "paku-commerce/internal/commerce/checkout/domain"
	platformbooking "paku-commerce/internal/commerce/platform/booking"
)

// ConfirmPaymentInput contiene los datos de confirmación de pago.
type ConfirmPaymentInput struct {
	OrderID    string
	PaymentRef string
	PaidAt     time.Time
}

// ConfirmPaymentOutput contiene la orden confirmada.
type ConfirmPaymentOutput struct {
	Order checkoutdomain.Order
}

// ConfirmPayment confirma el pago de una orden de forma idempotente.
type ConfirmPayment struct {
	Repo    checkoutdomain.OrderRepository
	Booking platformbooking.Client
	Now     func() time.Time
}

// Execute confirma el pago y actualiza la orden.
func (uc ConfirmPayment) Execute(ctx context.Context, input ConfirmPaymentInput) (ConfirmPaymentOutput, error) {
	// 1. Cargar la orden
	order, err := uc.Repo.GetByID(ctx, input.OrderID)
	if err != nil {
		return ConfirmPaymentOutput{}, err
	}

	// Determinar timestamp de pago
	paidAt := input.PaidAt
	if paidAt.IsZero() {
		if uc.Now != nil {
			paidAt = uc.Now()
		} else {
			paidAt = time.Now()
		}
	}

	// 2. Intentar marcar como pagada (idempotente)
	wasAlreadyPaid := order.Status == checkoutdomain.OrderStatusPaid &&
		order.PaymentRef != nil &&
		*order.PaymentRef == input.PaymentRef

	err = order.MarkPaid(input.PaymentRef, paidAt)
	if err != nil {
		return ConfirmPaymentOutput{}, err
	}

	// 3. Si ya estaba pagada con la misma ref, retornar sin side effects
	if wasAlreadyPaid {
		return ConfirmPaymentOutput{Order: order}, nil
	}

	// 4. Confirmar hold de booking si existe (solo en transición real)
	if order.BookingHoldID != nil && *order.BookingHoldID != "" {
		if err := uc.Booking.ConfirmHold(ctx, *order.BookingHoldID); err != nil {
			// No persistir cambios si booking falla
			// TODO: confirm with architect si preferimos estrategia de compensación
			return ConfirmPaymentOutput{}, err
		}
	}

	// 5. Persistir orden actualizada
	updatedOrder, err := uc.Repo.Update(ctx, order)
	if err != nil {
		return ConfirmPaymentOutput{}, err
	}

	return ConfirmPaymentOutput{Order: updatedOrder}, nil
}
