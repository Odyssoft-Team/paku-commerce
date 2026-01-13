package domain

import (
	"crypto/rand"
	"encoding/hex"
	"time"

	checkoutdomain "paku-commerce/internal/commerce/checkout/domain"
	servicedomain "paku-commerce/internal/commerce/service/domain"
)

const CartTTL = 90 * time.Minute

// Cart representa el carrito de compras persistente por usuario.
type Cart struct {
	ID            string
	UserID        string
	PetProfile    servicedomain.PetProfile
	Items         []checkoutdomain.PurchaseItem
	BookingHoldID *string
	OrderID       *string
	UpdatedAt     time.Time
	ExpiresAt     time.Time
}

// NewCart crea un nuevo carrito para un usuario.
func NewCart(userID string, petProfile servicedomain.PetProfile, items []checkoutdomain.PurchaseItem, now time.Time) Cart {
	return Cart{
		ID:         generateCartID(),
		UserID:     userID,
		PetProfile: petProfile,
		Items:      items,
		UpdatedAt:  now,
		ExpiresAt:  now.Add(CartTTL),
	}
}

// UpdateCart actualiza el carrito y renueva TTL.
func (c *Cart) UpdateCart(petProfile servicedomain.PetProfile, items []checkoutdomain.PurchaseItem, now time.Time) {
	c.PetProfile = petProfile
	c.Items = items
	c.UpdatedAt = now
	c.ExpiresAt = now.Add(CartTTL)
}

// IsExpired verifica si el carrito está vencido.
func (c Cart) IsExpired(now time.Time) bool {
	return now.After(c.ExpiresAt)
}

func generateCartID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return "cart_" + hex.EncodeToString(b)
}
