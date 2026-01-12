package domain

import (
	"strings"

	"paku-commerce/internal/pricing/domain"
)

// Coupon representa un cupón de descuento por código.
type Coupon struct {
	Code               string
	Active             bool
	PercentOff         int      // 0-100
	AppliesToItemTypes []string // "service", "product"; vacío = aplica a todo
	MinSubtotalAmount  int64    // minor units; 0 = sin mínimo
	Currency           domain.Currency
}

// NormalizeCode normaliza el código del cupón (trim + uppercase).
func NormalizeCode(code string) string {
	return strings.ToUpper(strings.TrimSpace(code))
}

// IsApplicable valida si el cupón es aplicable al quote dado.
func (c Coupon) IsApplicable(subtotal domain.Money, quoteItems []domain.QuoteItem) bool {
	if !c.Active {
		return false
	}

	// Check currency match
	if c.Currency != subtotal.Currency {
		return false
	}

	// Check minimum subtotal
	if c.MinSubtotalAmount > 0 && subtotal.Amount < c.MinSubtotalAmount {
		return false
	}

	// Check item type filter (si vacío, aplica a todo)
	if len(c.AppliesToItemTypes) > 0 {
		hasApplicableItem := false
		for _, item := range quoteItems {
			if containsItemType(c.AppliesToItemTypes, string(item.ItemType)) {
				hasApplicableItem = true
				break
			}
		}
		if !hasApplicableItem {
			return false
		}
	}

	return true
}

func containsItemType(slice []string, itemType string) bool {
	for _, s := range slice {
		if s == itemType {
			return true
		}
	}
	return false
}
