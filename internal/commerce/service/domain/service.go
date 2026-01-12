package domain

// Service representa un servicio de grooming (base o addon).
type Service struct {
	ID                string
	Name              string
	IsAddon           bool
	EligibilityRules  []EligibilityRule
	RequiresParentIDs []string // IDs de servicios base requeridos (solo para addons)
}

// IsEligibleFor evalúa si el servicio es elegible para el pet dado.
// Todas las reglas deben cumplirse (AND).
func (s Service) IsEligibleFor(pet PetProfile) bool {
	for _, rule := range s.EligibilityRules {
		if !rule.IsEligible(pet) {
			return false
		}
	}
	return true
}

// RequiresParent indica si este servicio requiere un servicio base.
func (s Service) RequiresParent() bool {
	return len(s.RequiresParentIDs) > 0
}
