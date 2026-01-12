package domain

import servicedomain "paku-commerce/internal/commerce/service/domain"

// ItemType identifica el tipo de item a comprar.
type ItemType string

const (
	ItemTypeService ItemType = "service"
	ItemTypeProduct ItemType = "product"
)

// PurchaseItem representa un item individual a comprar.
type PurchaseItem struct {
	ItemType ItemType
	ItemID   string
	Qty      int
}

// PurchaseIntent representa la intención de compra del usuario.
type PurchaseIntent struct {
	PetProfile    servicedomain.PetProfile
	Items         []PurchaseItem
	CouponCode    *string
	BookingHoldID *string
}
