package usecases

import (
	"context"

	cartdomain "paku-commerce/internal/commerce/cart/domain"
	checkoutdomain "paku-commerce/internal/commerce/checkout/domain"
	platformbooking "paku-commerce/internal/commerce/platform/booking"
)

// StartCheckoutInput contiene user_id y slot_id.
type StartCheckoutInput struct {
	UserID string
	SlotID string
}

// StartCheckoutOutput contiene cart, order y hold_id.
type StartCheckoutOutput struct {
	Cart          cartdomain.Cart
	Order         checkoutdomain.Order
	BookingHoldID string
}

// StartCheckout inicia el checkout creando hold, order y actualizando cart.
type StartCheckout struct {
	CartRepo      cartdomain.CartRepository
	Booking       platformbooking.Client
	CreateOrderUC *CreateOrder
}

// Execute ejecuta el flujo de start checkout.
func (uc StartCheckout) Execute(ctx context.Context, input StartCheckoutInput) (StartCheckoutOutput, error) {
	// 1. Validaciones
	if input.UserID == "" {
		return StartCheckoutOutput{}, cartdomain.ErrInvalidUserID
	}
	if input.SlotID == "" {
		return StartCheckoutOutput{}, ErrInvalidQuantity // reusar o crear ErrInvalidSlotID
	}

	// 2. Cargar cart
	cart, err := uc.CartRepo.GetByUserID(ctx, input.UserID)
	if err != nil {
		return StartCheckoutOutput{}, err
	}

	if len(cart.Items) == 0 {
		return StartCheckoutOutput{}, cartdomain.ErrEmptyItems
	}

	// 3. Reemplazar hold previo (cancelar si existe)
	if cart.BookingHoldID != nil && *cart.BookingHoldID != "" {
		_ = uc.Booking.CancelHold(ctx, *cart.BookingHoldID) // best-effort
	}

	// 4. Crear nuevo hold
	holdID, err := uc.Booking.CreateHold(ctx, input.SlotID)
	if err != nil {
		return StartCheckoutOutput{}, err
	}

	// 5. Crear orden usando CreateOrder UC
	intent := checkoutdomain.PurchaseIntent{
		PetProfile:    cart.PetProfile,
		Items:         cart.Items,
		CouponCode:    nil, // TODO: tomar del cart si se agregó cupón
		BookingHoldID: &holdID,
	}

	orderOutput, err := uc.CreateOrderUC.Execute(ctx, CreateOrderInput{Intent: intent})
	if err != nil {
		// Best-effort: cancelar hold si falló crear orden
		_ = uc.Booking.CancelHold(ctx, holdID)
		return StartCheckoutOutput{}, err
	}

	// 6. Actualizar cart con hold y order refs
	cart.BookingHoldID = &holdID
	cart.OrderID = &orderOutput.Order.ID

	updatedCart, err := uc.CartRepo.Upsert(ctx, cart)
	if err != nil {
		return StartCheckoutOutput{}, err
	}

	return StartCheckoutOutput{
		Cart:          updatedCart,
		Order:         orderOutput.Order,
		BookingHoldID: holdID,
	}, nil
}
