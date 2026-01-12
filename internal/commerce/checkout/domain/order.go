package domain

import (
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
}
