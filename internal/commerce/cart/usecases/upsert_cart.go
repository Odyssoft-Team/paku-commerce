package usecases

import (
	"context"
	"time"

	cartdomain "paku-commerce/internal/commerce/cart/domain"
	checkoutdomain "paku-commerce/internal/commerce/checkout/domain"
	servicedomain "paku-commerce/internal/commerce/service/domain"
)

// UpsertCartInput contiene los datos para crear/actualizar carrito.
type UpsertCartInput struct {
	UserID        string
	PetProfile    servicedomain.PetProfile
	Items         []checkoutdomain.PurchaseItem
	BookingHoldID *string
	OrderID       *string
}

// UpsertCartOutput contiene el carrito creado/actualizado.
type UpsertCartOutput struct {
	Cart cartdomain.Cart
}

// UpsertCart crea o actualiza un carrito.
type UpsertCart struct {
	Repo cartdomain.CartRepository
	Now  func() time.Time
}

// Execute crea o actualiza el carrito del usuario.
func (uc UpsertCart) Execute(ctx context.Context, input UpsertCartInput) (UpsertCartOutput, error) {
	// Validaciones
	if input.UserID == "" {
		return UpsertCartOutput{}, cartdomain.ErrInvalidUserID
	}
	if len(input.Items) == 0 {
		return UpsertCartOutput{}, cartdomain.ErrEmptyItems
	}
	for _, item := range input.Items {
		if item.Qty <= 0 {
			return UpsertCartOutput{}, cartdomain.ErrEmptyItems
		}
	}

	now := time.Now()
	if uc.Now != nil {
		now = uc.Now()
	}

	// Intentar obtener carrito existente
	existingCart, err := uc.Repo.GetByUserID(ctx, input.UserID)

	var cart cartdomain.Cart
	if err == cartdomain.ErrCartNotFound {
		// Crear nuevo
		cart = cartdomain.NewCart(input.UserID, input.PetProfile, input.Items, now)
	} else if err != nil {
		return UpsertCartOutput{}, err
	} else {
		// Actualizar existente
		cart = existingCart
		cart.UpdateCart(input.PetProfile, input.Items, now)
	}

	// Actualizar referencias opcionales
	cart.BookingHoldID = input.BookingHoldID
	cart.OrderID = input.OrderID

	// Persistir
	updatedCart, err := uc.Repo.Upsert(ctx, cart)
	if err != nil {
		return UpsertCartOutput{}, err
	}

	return UpsertCartOutput{Cart: updatedCart}, nil
}
