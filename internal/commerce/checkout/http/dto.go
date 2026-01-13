package http

import (
	"time"

	cartdomain "paku-commerce/internal/commerce/cart/domain"
	checkoutdomain "paku-commerce/internal/commerce/checkout/domain"
	servicedomain "paku-commerce/internal/commerce/service/domain"
	pricingdomain "paku-commerce/internal/pricing/domain"
)

// PetProfileDTO representa el perfil de mascota en HTTP.
type PetProfileDTO struct {
	Species  string `json:"species"`
	WeightKg int    `json:"weight_kg"`
	CoatType string `json:"coat_type"`
}

// ItemDTO representa un item a comprar.
type ItemDTO struct {
	Type string `json:"type"` // "service" | "product"
	ID   string `json:"id"`
	Qty  int    `json:"qty"`
}

// QuoteRequestDTO es el request para /checkout/quote.
type QuoteRequestDTO struct {
	PetProfile    PetProfileDTO `json:"pet_profile"`
	Items         []ItemDTO     `json:"items"`
	CouponCode    *string       `json:"coupon_code"`
	BookingHoldID *string       `json:"booking_hold_id"`
}

// MoneyDTO representa dinero en HTTP.
type MoneyDTO struct {
	Amount   int64  `json:"amount"`
	Currency string `json:"currency"`
}

// DiscountLineDTO representa una línea de descuento.
type DiscountLineDTO struct {
	Source string   `json:"source"` // "coupon" | "promotion"
	Name   string   `json:"name"`
	Amount MoneyDTO `json:"amount"`
}

// QuoteDTO representa una cotización.
type QuoteDTO struct {
	Subtotal      MoneyDTO          `json:"subtotal"`
	TotalDiscount MoneyDTO          `json:"total_discount"`
	Total         MoneyDTO          `json:"total"`
	Discounts     []DiscountLineDTO `json:"discounts"`
}

// QuoteResponseDTO es el response para /checkout/quote.
type QuoteResponseDTO struct {
	Quote QuoteDTO `json:"quote"`
}

// OrderItemDTO representa un item en una orden.
type OrderItemDTO struct {
	Type      string   `json:"type"`
	ID        string   `json:"id"`
	Qty       int      `json:"qty"`
	UnitPrice MoneyDTO `json:"unit_price"`
	LineTotal MoneyDTO `json:"line_total"`
}

// OrderDTO representa una orden.
type OrderDTO struct {
	ID            string         `json:"id"`
	Status        string         `json:"status"`
	CreatedAt     string         `json:"created_at"`
	Subtotal      MoneyDTO       `json:"subtotal"`
	TotalDiscount MoneyDTO       `json:"total_discount"`
	Total         MoneyDTO       `json:"total"`
	CouponCode    *string        `json:"coupon_code,omitempty"`
	BookingHoldID *string        `json:"booking_hold_id,omitempty"`
	PaymentRef    *string        `json:"payment_ref,omitempty"`
	PaidAt        *string        `json:"paid_at,omitempty"`
	Items         []OrderItemDTO `json:"items"`
}

// CreateOrderResponseDTO es el response para POST /checkout/orders.
type CreateOrderResponseDTO struct {
	Order OrderDTO `json:"order"`
}

// ConfirmPaymentRequestDTO es el request para POST /checkout/orders/{id}/confirm-payment.
type ConfirmPaymentRequestDTO struct {
	PaymentRef string  `json:"payment_ref"`
	PaidAt     *string `json:"paid_at,omitempty"` // ISO8601
}

// ConfirmPaymentResponseDTO es el response para confirm payment.
type ConfirmPaymentResponseDTO struct {
	Order OrderDTO `json:"order"`
}

// StartCheckoutRequestDTO es el request para POST /checkout/start.
type StartCheckoutRequestDTO struct {
	SlotID string `json:"slot_id"`
}

// CartSnapshotDTO es una versión simplificada del cart para response.
type CartSnapshotDTO struct {
	ID            string  `json:"id"`
	UserID        string  `json:"user_id"`
	BookingHoldID *string `json:"booking_hold_id,omitempty"`
	OrderID       *string `json:"order_id,omitempty"`
	ExpiresAt     string  `json:"expires_at"`
}

// StartCheckoutResponseDTO es el response para POST /checkout/start.
type StartCheckoutResponseDTO struct {
	BookingHoldID string          `json:"booking_hold_id"`
	Order         OrderDTO        `json:"order"`
	Cart          CartSnapshotDTO `json:"cart"`
}

// toPetProfile convierte DTO a dominio.
func (dto PetProfileDTO) toPetProfile() servicedomain.PetProfile {
	return servicedomain.PetProfile{
		Species:  dto.Species,
		WeightKg: dto.WeightKg,
		CoatType: dto.CoatType,
	}
}

// toPurchaseItems convierte DTOs a dominio.
func toPurchaseItems(dtos []ItemDTO) []checkoutdomain.PurchaseItem {
	items := make([]checkoutdomain.PurchaseItem, 0, len(dtos))
	for _, dto := range dtos {
		items = append(items, checkoutdomain.PurchaseItem{
			ItemType: checkoutdomain.ItemType(dto.Type),
			ItemID:   dto.ID,
			Qty:      dto.Qty,
		})
	}
	return items
}

// toMoneyDTO convierte Money a DTO.
func toMoneyDTO(m pricingdomain.Money) MoneyDTO {
	return MoneyDTO{
		Amount:   m.Amount,
		Currency: string(m.Currency),
	}
}

// toOrderDTO convierte Order a DTO.
func toOrderDTO(order checkoutdomain.Order) OrderDTO {
	items := make([]OrderItemDTO, 0, len(order.Items))
	for _, item := range order.Items {
		items = append(items, OrderItemDTO{
			Type:      string(item.ItemType),
			ID:        item.ItemID,
			Qty:       item.Qty,
			UnitPrice: toMoneyDTO(item.UnitPrice),
			LineTotal: toMoneyDTO(item.LineTotal),
		})
	}

	dto := OrderDTO{
		ID:            order.ID,
		Status:        string(order.Status),
		CreatedAt:     order.CreatedAt.Format(time.RFC3339),
		Subtotal:      toMoneyDTO(order.Subtotal),
		TotalDiscount: toMoneyDTO(order.TotalDiscount),
		Total:         toMoneyDTO(order.Total),
		CouponCode:    order.CouponCode,
		BookingHoldID: order.BookingHoldID,
		PaymentRef:    order.PaymentRef,
		Items:         items,
	}

	if order.PaidAt != nil {
		paidAtStr := order.PaidAt.Format(time.RFC3339)
		dto.PaidAt = &paidAtStr
	}

	return dto
}

// toCartSnapshotDTO convierte Cart a CartSnapshotDTO.
func toCartSnapshotDTO(cart cartdomain.Cart) CartSnapshotDTO {
	return CartSnapshotDTO{
		ID:            cart.ID,
		UserID:        cart.UserID,
		BookingHoldID: cart.BookingHoldID,
		OrderID:       cart.OrderID,
		ExpiresAt:     cart.ExpiresAt.Format(time.RFC3339),
	}
}
