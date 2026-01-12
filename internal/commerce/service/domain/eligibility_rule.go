package domain

// EligibilityRule evalúa si un pet cumple con criterios específicos.
type EligibilityRule struct {
	AllowedSpecies    []string
	ExcludedSpecies   []string
	MinWeightKg       *int
	MaxWeightKg       *int
	AllowedCoatTypes  []string
	ExcludedCoatTypes []string
}

// IsEligible evalúa si el pet cumple con esta regla.
func (r EligibilityRule) IsEligible(pet PetProfile) bool {
	// Check species
	if len(r.AllowedSpecies) > 0 {
		if !contains(r.AllowedSpecies, pet.Species) {
			return false
		}
	}
	if len(r.ExcludedSpecies) > 0 {
		if contains(r.ExcludedSpecies, pet.Species) {
			return false
		}
	}

	// Check weight range
	if r.MinWeightKg != nil && pet.WeightKg < *r.MinWeightKg {
		return false
	}
	if r.MaxWeightKg != nil && pet.WeightKg > *r.MaxWeightKg {
		return false
	}

	// Check coat type
	if len(r.AllowedCoatTypes) > 0 {
		if !contains(r.AllowedCoatTypes, pet.CoatType) {
			return false
		}
	}
	if len(r.ExcludedCoatTypes) > 0 {
		if contains(r.ExcludedCoatTypes, pet.CoatType) {
			return false
		}
	}

	return true
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
