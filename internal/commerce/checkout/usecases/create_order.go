package usecases

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	checkoutdomain "paku-commerce/internal/commerce/checkout/domain"
)

// CreateOrderInput contiene la intención de compra.
type CreateOrderInput struct {
	Intent checkoutdomain.PurchaseIntent
}

// CreateOrderOutput contiene la orden creada.
type CreateOrderOutput struct {
	Order checkoutdomain.Order
}

// CreateOrder crea una orden pendiente de pago.
type CreateOrder struct {
	QuoteCheckoutUC *QuoteCheckout
	OrderRepo       checkoutdomain.OrderRepository
	Now             func() time.Time
}

// Execute crea una orden y la persiste.
func (uc CreateOrder) Execute(ctx context.Context, input CreateOrderInput) (CreateOrderOutput, error) {
	// 1. Cotizar checkout (validación + pricing + promos)
	quoteOutput, err := uc.QuoteCheckoutUC.Execute(ctx, QuoteCheckoutInput{
		Intent: input.Intent,
	})
	if err != nil {
		return CreateOrderOutput{}, err
	}

	quote := quoteOutput.Quote

	// 2. Construir items de la orden
	orderItems := make([]checkoutdomain.OrderItem, 0, len(quote.Quote.Items))
	for _, qItem := range quote.Quote.Items {
		orderItems = append(orderItems, checkoutdomain.OrderItem{
			ItemType:  checkoutdomain.ItemType(qItem.ItemType),
			ItemID:    qItem.ItemID,
			Qty:       qItem.Qty,
			UnitPrice: qItem.UnitPrice,
			LineTotal: qItem.LineTotal,
		})
	}

	// 3. Construir orden
	now := time.Now()
	if uc.Now != nil {
		now = uc.Now()
	}

	orderID, err := generateOrderID()
	if err != nil {
		return CreateOrderOutput{}, fmt.Errorf("failed to generate order ID: %w", err)
	}

	order := checkoutdomain.Order{
		ID:            orderID,
		Status:        checkoutdomain.OrderStatusPendingPayment,
		CreatedAt:     now,
		PetProfile:    input.Intent.PetProfile,
		Items:         orderItems,
		Subtotal:      quote.OriginalSubtotal, // Subtotal original antes de descuentos
		TotalDiscount: quote.TotalDiscount,
		Total:         quote.Total,
		CouponCode:    input.Intent.CouponCode,
		BookingHoldID: input.Intent.BookingHoldID,
	}

	// 4. Persistir orden
	createdOrder, err := uc.OrderRepo.Create(ctx, order)
	if err != nil {
		return CreateOrderOutput{}, err
	}

	return CreateOrderOutput{Order: createdOrder}, nil
}

// generateOrderID genera un ID único para la orden.
// TODO: usar internal/platform/id si existe
func generateOrderID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		// Fallback a timestamp-based ID si falla random
		return fmt.Sprintf("order_%d", time.Now().UnixNano()), nil
	}
	return "order_" + hex.EncodeToString(b), nil
}
