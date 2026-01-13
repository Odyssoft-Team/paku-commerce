package http

import (
	"time"

	cartdomain "paku-commerce/internal/commerce/cart/domain"
	checkoutdomain "paku-commerce/internal/commerce/checkout/domain"
	servicedomain "paku-commerce/internal/commerce/service/domain"
)

// PetProfileDTO representa el perfil de mascota.
type PetProfileDTO struct {
	Species  string `json:"species"`
	WeightKg int    `json:"weight_kg"`
	CoatType string `json:"coat_type"`
}

// ItemDTO representa un item del carrito.
type ItemDTO struct {
	Type string `json:"type"`
	ID   string `json:"id"`
	Qty  int    `json:"qty"`
}

// UpsertCartRequestDTO es el request para PUT /cart/me.
type UpsertCartRequestDTO struct {
	PetProfile    PetProfileDTO `json:"pet_profile"`
	Items         []ItemDTO     `json:"items"`
	BookingHoldID *string       `json:"booking_hold_id,omitempty"`
	OrderID       *string       `json:"order_id,omitempty"`
}

// CartDTO representa un carrito.
type CartDTO struct {
	ID            string        `json:"id"`
	UserID        string        `json:"user_id"`
	PetProfile    PetProfileDTO `json:"pet_profile"`
	Items         []ItemDTO     `json:"items"`
	BookingHoldID *string       `json:"booking_hold_id,omitempty"`
	OrderID       *string       `json:"order_id,omitempty"`
	UpdatedAt     string        `json:"updated_at"`
	ExpiresAt     string        `json:"expires_at"`
}

// CartResponseDTO es el response para cart operations.
type CartResponseDTO struct {
	Cart CartDTO `json:"cart"`
}

// ErrorResponse representa un error HTTP.
type ErrorResponse struct {
	Error string `json:"error"`
}

func (dto PetProfileDTO) toPetProfile() servicedomain.PetProfile {
	return servicedomain.PetProfile{
		Species:  dto.Species,
		WeightKg: dto.WeightKg,
		CoatType: dto.CoatType,
	}
}

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

func toCartDTO(cart cartdomain.Cart) CartDTO {
	items := make([]ItemDTO, 0, len(cart.Items))
	for _, item := range cart.Items {
		items = append(items, ItemDTO{
			Type: string(item.ItemType),
			ID:   item.ItemID,
			Qty:  item.Qty,
		})
	}

	return CartDTO{
		ID:     cart.ID,
		UserID: cart.UserID,
		PetProfile: PetProfileDTO{
			Species:  cart.PetProfile.Species,
			WeightKg: cart.PetProfile.WeightKg,
			CoatType: cart.PetProfile.CoatType,
		},
		Items:         items,
		BookingHoldID: cart.BookingHoldID,
		OrderID:       cart.OrderID,
		UpdatedAt:     cart.UpdatedAt.Format(time.RFC3339),
		ExpiresAt:     cart.ExpiresAt.Format(time.RFC3339),
	}
}
