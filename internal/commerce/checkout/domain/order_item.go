package domain

import pricingdomain "paku-commerce/internal/pricing/domain"

// OrderItem representa un item individual en una orden.
type OrderItem struct {
	ItemType  ItemType
	ItemID    string
	Qty       int
	UnitPrice pricingdomain.Money
	LineTotal pricingdomain.Money
}
