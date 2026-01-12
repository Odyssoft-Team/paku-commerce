package domain

import servicedomain "paku-commerce/internal/commerce/service/domain"

// ItemType identifica el tipo de item a cotizar.
type ItemType string

const (
	ItemTypeService ItemType = "service"
	ItemTypeProduct ItemType = "product"
)

// PriceRule define una regla de precio para un item.
type PriceRule struct {
	ItemType    ItemType
	ItemID      string
	MinWeightKg *int // Solo para services
	MaxWeightKg *int // Solo para services
	UnitPrice   Money
}

// MatchesService evalúa si la regla aplica al servicio dado con el pet profile.
func (r PriceRule) MatchesService(itemID string, pet servicedomain.PetProfile) bool {
	if r.ItemType != ItemTypeService {
		return false
	}
	if r.ItemID != itemID {
		return false
	}

	// Check weight range
	if r.MinWeightKg != nil && pet.WeightKg < *r.MinWeightKg {
		return false
	}
	if r.MaxWeightKg != nil && pet.WeightKg > *r.MaxWeightKg {
		return false
	}

	return true
}

// MatchesProduct evalúa si la regla aplica al producto dado.
func (r PriceRule) MatchesProduct(itemID string) bool {
	return r.ItemType == ItemTypeProduct && r.ItemID == itemID
}

// Specificity retorna un puntaje de especificidad para ordenar reglas.
// Mayor puntaje = más específica (ambos min+max definidos).
func (r PriceRule) Specificity() int {
	score := 0
	if r.MinWeightKg != nil {
		score++
	}
	if r.MaxWeightKg != nil {
		score++
	}
	return score
}

// RangeSize retorna el tamaño del rango de peso (para desempate).
// Menor rango = más específico.
func (r PriceRule) RangeSize() int {
	if r.MinWeightKg == nil || r.MaxWeightKg == nil {
		return 999999 // Rango infinito si falta alguno
	}
	return *r.MaxWeightKg - *r.MinWeightKg
}
