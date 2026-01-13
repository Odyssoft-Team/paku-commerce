package usecases

import (
	"context"
	"time"

	cartdomain "paku-commerce/internal/commerce/cart/domain"
)

// GetCartInput contiene el user_id.
type GetCartInput struct {
	UserID string
}

// GetCartOutput contiene el carrito.
type GetCartOutput struct {
	Cart cartdomain.Cart
}

// GetCart obtiene el carrito de un usuario.
type GetCart struct {
	Repo cartdomain.CartRepository
	Now  func() time.Time
}

// Execute obtiene el carrito del usuario, considerando expiración.
func (uc GetCart) Execute(ctx context.Context, input GetCartInput) (GetCartOutput, error) {
	if input.UserID == "" {
		return GetCartOutput{}, cartdomain.ErrInvalidUserID
	}

	cart, err := uc.Repo.GetByUserID(ctx, input.UserID)
	if err != nil {
		return GetCartOutput{}, err
	}

	now := time.Now()
	if uc.Now != nil {
		now = uc.Now()
	}

	// Si está expirado, tratar como not found y borrarlo
	if cart.IsExpired(now) {
		_ = uc.Repo.DeleteByUserID(ctx, input.UserID) // best-effort
		return GetCartOutput{}, cartdomain.ErrCartNotFound
	}

	return GetCartOutput{Cart: cart}, nil
}
