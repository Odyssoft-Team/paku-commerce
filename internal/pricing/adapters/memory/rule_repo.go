package memory

import (
	"context"

	"paku-commerce/internal/pricing/domain"
)

// PriceRuleRepository implementa domain.PriceRuleRepository en memoria.
type PriceRuleRepository struct {
	rules []domain.PriceRule
}

// NewPriceRuleRepository crea un repositorio con reglas de ejemplo.
func NewPriceRuleRepository() *PriceRuleRepository {
	intPtr := func(v int) *int { return &v }

	rules := []domain.PriceRule{
		// Service: bath (baño) - por rango de peso
		{
			ItemType:    domain.ItemTypeService,
			ItemID:      "bath",
			MinWeightKg: intPtr(0),
			MaxWeightKg: intPtr(10),
			UnitPrice:   domain.NewMoney(3500, domain.CurrencyPEN), // S/ 35.00
		},
		{
			ItemType:    domain.ItemTypeService,
			ItemID:      "bath",
			MinWeightKg: intPtr(11),
			MaxWeightKg: intPtr(20),
			UnitPrice:   domain.NewMoney(4500, domain.CurrencyPEN), // S/ 45.00
		},
		{
			ItemType:    domain.ItemTypeService,
			ItemID:      "bath",
			MinWeightKg: intPtr(21),
			MaxWeightKg: intPtr(40),
			UnitPrice:   domain.NewMoney(6000, domain.CurrencyPEN), // S/ 60.00
		},
		// Service: deshedding (deslanado)
		{
			ItemType:    domain.ItemTypeService,
			ItemID:      "deshedding",
			MinWeightKg: intPtr(0),
			MaxWeightKg: intPtr(20),
			UnitPrice:   domain.NewMoney(2000, domain.CurrencyPEN), // S/ 20.00
		},
		{
			ItemType:    domain.ItemTypeService,
			ItemID:      "deshedding",
			MinWeightKg: intPtr(21),
			MaxWeightKg: intPtr(40),
			UnitPrice:   domain.NewMoney(3000, domain.CurrencyPEN), // S/ 30.00
		},
		// Product: shampoo_basic (precio fijo)
		{
			ItemType:  domain.ItemTypeProduct,
			ItemID:    "shampoo_basic",
			UnitPrice: domain.NewMoney(2500, domain.CurrencyPEN), // S/ 25.00
		},
	}

	return &PriceRuleRepository{rules: rules}
}

// ListRules retorna todas las reglas.
func (r *PriceRuleRepository) ListRules(ctx context.Context) ([]domain.PriceRule, error) {
	return r.rules, nil
}

// ListRulesForItem filtra reglas por tipo e ID de item.
func (r *PriceRuleRepository) ListRulesForItem(ctx context.Context, itemType domain.ItemType, itemID string) ([]domain.PriceRule, error) {
	var filtered []domain.PriceRule
	for _, rule := range r.rules {
		if rule.ItemType == itemType && rule.ItemID == itemID {
			filtered = append(filtered, rule)
		}
	}
	return filtered, nil
}
