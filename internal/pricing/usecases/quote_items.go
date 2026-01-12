package usecases

import (
	"context"
	"errors"
	"sort"

	"paku-commerce/internal/commerce/service/domain"
	pricingdomain "paku-commerce/internal/pricing/domain"
)

var ErrNoPriceRule = errors.New("no price rule found for item")

// QuoteRequestItem representa un item a cotizar.
type QuoteRequestItem struct {
	ItemType pricingdomain.ItemType
	ItemID   string
	Qty      int
}

// QuoteItemsInput contiene el pet profile y los items a cotizar.
type QuoteItemsInput struct {
	PetProfile domain.PetProfile
	Items      []QuoteRequestItem
}

// QuoteItemsOutput contiene la cotización generada.
type QuoteItemsOutput struct {
	Quote pricingdomain.Quote
}

// QuoteItems cotiza items aplicando reglas de precio.
type QuoteItems struct {
	RuleRepo pricingdomain.PriceRuleRepository
}

// Execute cotiza los items según las reglas de precio.
func (uc QuoteItems) Execute(ctx context.Context, input QuoteItemsInput) (QuoteItemsOutput, error) {
	var quoteItems []pricingdomain.QuoteItem
	subtotal := pricingdomain.Zero(pricingdomain.CurrencyPEN)

	for _, reqItem := range input.Items {
		// Obtener reglas para el item
		rules, err := uc.RuleRepo.ListRulesForItem(ctx, reqItem.ItemType, reqItem.ItemID)
		if err != nil {
			return QuoteItemsOutput{}, err
		}

		// Seleccionar regla aplicable
		var selectedRule *pricingdomain.PriceRule
		if reqItem.ItemType == pricingdomain.ItemTypeService {
			selectedRule = selectServiceRule(rules, reqItem.ItemID, input.PetProfile)
		} else if reqItem.ItemType == pricingdomain.ItemTypeProduct {
			selectedRule = selectProductRule(rules)
		}

		if selectedRule == nil {
			return QuoteItemsOutput{}, ErrNoPriceRule
		}

		// Calcular line total
		lineTotal := selectedRule.UnitPrice.MulInt(int64(reqItem.Qty))

		quoteItems = append(quoteItems, pricingdomain.QuoteItem{
			ItemType:  reqItem.ItemType,
			ItemID:    reqItem.ItemID,
			Qty:       reqItem.Qty,
			UnitPrice: selectedRule.UnitPrice,
			LineTotal: lineTotal,
		})

		// Sumar al subtotal
		newSubtotal, err := subtotal.Add(lineTotal)
		if err != nil {
			return QuoteItemsOutput{}, err
		}
		subtotal = newSubtotal
	}

	return QuoteItemsOutput{
		Quote: pricingdomain.Quote{
			Items:    quoteItems,
			Subtotal: subtotal,
		},
	}, nil
}

// selectServiceRule elige la regla más específica que matchea el pet.
func selectServiceRule(rules []pricingdomain.PriceRule, itemID string, pet domain.PetProfile) *pricingdomain.PriceRule {
	var matching []pricingdomain.PriceRule
	for _, rule := range rules {
		if rule.MatchesService(itemID, pet) {
			matching = append(matching, rule)
		}
	}

	if len(matching) == 0 {
		return nil
	}

	// Ordenar por especificidad (más específica primero)
	sort.Slice(matching, func(i, j int) bool {
		if matching[i].Specificity() != matching[j].Specificity() {
			return matching[i].Specificity() > matching[j].Specificity()
		}
		// Desempate: menor rango
		return matching[i].RangeSize() < matching[j].RangeSize()
	})

	return &matching[0]
}

// selectProductRule elige la primera regla de producto que matchea.
func selectProductRule(rules []pricingdomain.PriceRule) *pricingdomain.PriceRule {
	for _, rule := range rules {
		if rule.ItemType == pricingdomain.ItemTypeProduct {
			return &rule
		}
	}
	return nil
}
