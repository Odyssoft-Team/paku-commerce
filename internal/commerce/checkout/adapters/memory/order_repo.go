package memory

import (
	"context"
	"sync"

	"paku-commerce/internal/commerce/checkout/domain"
)

// OrderRepository implementa domain.OrderRepository en memoria.
type OrderRepository struct {
	mu     sync.RWMutex
	orders map[string]domain.Order
}

// NewOrderRepository crea un repositorio de órdenes en memoria.
func NewOrderRepository() *OrderRepository {
	return &OrderRepository{
		orders: make(map[string]domain.Order),
	}
}

// Create guarda una orden y la retorna.
func (r *OrderRepository) Create(ctx context.Context, order domain.Order) (domain.Order, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.orders[order.ID] = order
	return order, nil
}

// GetByID busca una orden por ID.
func (r *OrderRepository) GetByID(ctx context.Context, id string) (domain.Order, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	order, exists := r.orders[id]
	if !exists {
		return domain.Order{}, domain.ErrOrderNotFound
	}
	return order, nil
}

// Update actualiza una orden existente.
func (r *OrderRepository) Update(ctx context.Context, order domain.Order) (domain.Order, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.orders[order.ID]; !exists {
		return domain.Order{}, domain.ErrOrderNotFound
	}

	r.orders[order.ID] = order
	return order, nil
}
