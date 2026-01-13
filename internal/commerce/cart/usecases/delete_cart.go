package usecases

import (
	"context"

	cartdomain "paku-commerce/internal/commerce/cart/domain"
)

// DeleteCartInput contiene el user_id.
type DeleteCartInput struct {
	UserID string
}

// DeleteCartOutput es vacío.
type DeleteCartOutput struct{}

// DeleteCart elimina el carrito de un usuario.
type DeleteCart struct {
	Repo cartdomain.CartRepository
}

// Execute elimina el carrito sin side-effects.
func (uc DeleteCart) Execute(ctx context.Context, input DeleteCartInput) (DeleteCartOutput, error) {
	if input.UserID == "" {
		return DeleteCartOutput{}, cartdomain.ErrInvalidUserID
	}

	err := uc.Repo.DeleteByUserID(ctx, input.UserID)
	if err != nil && err != cartdomain.ErrCartNotFound {
		return DeleteCartOutput{}, err
	}

	return DeleteCartOutput{}, nil
}
