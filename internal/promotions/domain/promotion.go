package domain

import "paku-commerce/internal/pricing/domain"

// Promotion representa una promoción automática (sin código).
type Promotion struct {
	Name               string
	Active             bool
	PercentOff         int      // 0-100
	AppliesToItemTypes []string // "service", "product"; vacío = aplica a todo
	Currency           domain.Currency
}

// IsApplicable valida si la promoción es aplicable al quote dado.
func (p Promotion) IsApplicable(subtotal domain.Money, quoteItems []domain.QuoteItem) bool {
	if !p.Active {
		return false
	}

	// Check currency match
	if p.Currency != subtotal.Currency {
		return false
	}

	// Check item type filter
	if len(p.AppliesToItemTypes) > 0 {
		hasApplicableItem := false
		for _, item := range quoteItems {
			if containsItemType(p.AppliesToItemTypes, string(item.ItemType)) {
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
