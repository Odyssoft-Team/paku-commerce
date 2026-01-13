package memory

import (
	"context"
	"sync"
	"time"

	"paku-commerce/internal/commerce/cart/domain"
)

// CartRepository implementa domain.CartRepository en memoria.
type CartRepository struct {
	mu    sync.RWMutex
	carts map[string]domain.Cart // key: userID
}

// NewCartRepository crea un repositorio de carritos en memoria.
func NewCartRepository() *CartRepository {
	return &CartRepository{
		carts: make(map[string]domain.Cart),
	}
}

// Upsert crea o actualiza un carrito.
func (r *CartRepository) Upsert(ctx context.Context, cart domain.Cart) (domain.Cart, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.carts[cart.UserID] = cart
	return cart, nil
}

// GetByUserID obtiene un carrito por user_id.
func (r *CartRepository) GetByUserID(ctx context.Context, userID string) (domain.Cart, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	cart, exists := r.carts[userID]
	if !exists {
		return domain.Cart{}, domain.ErrCartNotFound
	}
	return cart, nil
}

// DeleteByUserID elimina un carrito.
func (r *CartRepository) DeleteByUserID(ctx context.Context, userID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.carts, userID)
	return nil
}

// ListExpired retorna carritos vencidos.
func (r *CartRepository) ListExpired(ctx context.Context, now time.Time) ([]domain.Cart, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var expired []domain.Cart
	for _, cart := range r.carts {
		if cart.IsExpired(now) {
			expired = append(expired, cart)
		}
	}
	return expired, nil
}
