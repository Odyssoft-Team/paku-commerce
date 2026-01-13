package usecases

import (
	"context"
	"time"

	cartdomain "paku-commerce/internal/commerce/cart/domain"
	bookingports "paku-commerce/internal/commerce/cart/ports/booking"
	checkoutports "paku-commerce/internal/commerce/cart/ports/checkout"
)

// ExpireCartsInput contiene el timestamp de referencia.
type ExpireCartsInput struct {
	Now time.Time
}

// ExpireCartsOutput contiene estadísticas de expiración.
type ExpireCartsOutput struct {
	ExpiredCount int
}

// ExpireCarts limpia carritos vencidos y ejecuta side-effects.
type ExpireCarts struct {
	Repo     cartdomain.CartRepository
	Booking  bookingports.BookingClient
	Checkout checkoutports.CheckoutClient
}

// Execute expira carritos vencidos.
func (uc ExpireCarts) Execute(ctx context.Context, input ExpireCartsInput) (ExpireCartsOutput, error) {
	expiredCarts, err := uc.Repo.ListExpired(ctx, input.Now)
	if err != nil {
		return ExpireCartsOutput{}, err
	}

	count := 0
	for _, cart := range expiredCarts {
		// Cancelar hold si existe
		if cart.BookingHoldID != nil && *cart.BookingHoldID != "" {
			// Best-effort: ignorar errores (ej. hold ya cancelado/no existe)
			_ = uc.Booking.CancelHold(ctx, *cart.BookingHoldID)
		}

		// Cancelar orden si existe
		if cart.OrderID != nil && *cart.OrderID != "" {
			// Best-effort: ignorar errores (ej. orden ya cancelada/pagada)
			_ = uc.Checkout.CancelOrder(ctx, *cart.OrderID)
		}

		// Eliminar carrito
		if err := uc.Repo.DeleteByUserID(ctx, cart.UserID); err != nil {
			// Log pero no fallar el proceso completo
			continue
		}

		count++
	}

	return ExpireCartsOutput{ExpiredCount: count}, nil
}
