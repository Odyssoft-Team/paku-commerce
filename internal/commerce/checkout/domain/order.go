package domain

import (
	"errors"
	"time"

	servicedomain "paku-commerce/internal/commerce/service/domain"
	pricingdomain "paku-commerce/internal/pricing/domain"
)

// OrderStatus representa el estado de una orden.
type OrderStatus string

const (
	OrderStatusPendingPayment OrderStatus = "pending_payment"
	OrderStatusPaid           OrderStatus = "paid"
	OrderStatusCancelled      OrderStatus = "cancelled"
)

var (
	ErrPaymentConflict   = errors.New("payment reference conflict")
	ErrOrderCancelled    = errors.New("order is cancelled")
	ErrInvalidOrderState = errors.New("invalid order state for operation")
)

// Order representa una orden de compra.
type Order struct {
	ID            string
	Status        OrderStatus
	CreatedAt     time.Time
	PetProfile    servicedomain.PetProfile
	Items         []OrderItem
	Subtotal      pricingdomain.Money
	TotalDiscount pricingdomain.Money
	Total         pricingdomain.Money
	CouponCode    *string
	BookingHoldID *string
	PaymentRef    *string
	PaidAt        *time.Time
}

// MarkPaid marca la orden como pagada de forma idempotente.
func (o *Order) MarkPaid(paymentRef string, paidAt time.Time) error {
	if o.Status == OrderStatusCancelled {
		return ErrOrderCancelled
	}

	if o.Status == OrderStatusPaid {
		// Idempotencia: si ya está paid con la misma ref, OK
		if o.PaymentRef != nil && *o.PaymentRef == paymentRef {
			return nil
		}
		// Conflicto: ya pagado con distinta ref
		return ErrPaymentConflict
	}

	if o.Status == OrderStatusPendingPayment {
		o.Status = OrderStatusPaid
		o.PaymentRef = &paymentRef
		o.PaidAt = &paidAt
		return nil
	}

	return ErrInvalidOrderState
}

// MarkCancelled marca la orden como cancelada.
func (o *Order) MarkCancelled() error {
	if o.Status == OrderStatusCancelled {
		// Idempotente: ya cancelada
		return nil
	}

	if o.Status == OrderStatusPaid {
		return ErrInvalidOrderState
	}

	if o.Status == OrderStatusPendingPayment {
		o.Status = OrderStatusCancelled
		return nil
	}

	return ErrInvalidOrderState
}
