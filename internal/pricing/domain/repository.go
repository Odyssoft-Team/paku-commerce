package domain

import "context"

// PriceRuleRepository define el acceso a reglas de precio.
type PriceRuleRepository interface {
	ListRules(ctx context.Context) ([]PriceRule, error)
	ListRulesForItem(ctx context.Context, itemType ItemType, itemID string) ([]PriceRule, error)
}
