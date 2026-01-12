package usecases

import (
	"context"

	"paku-commerce/internal/commerce/service/domain"
)

// GetOfferForPetInput representa el perfil de la mascota.
type GetOfferForPetInput struct {
	PetProfile domain.PetProfile
}

// ServiceOffer agrupa un servicio base con sus addons permitidos.
type ServiceOffer struct {
	Service       domain.Service
	AllowedAddons []domain.Service
}

// GetOfferForPetOutput contiene las ofertas de servicios elegibles.
type GetOfferForPetOutput struct {
	Offers []ServiceOffer
}

// GetOfferForPet obtiene servicios base y addons elegibles para un pet.
type GetOfferForPet struct {
	Repo domain.ServiceRepository
}

// Execute filtra servicios base elegibles y sus addons permitidos.
func (uc GetOfferForPet) Execute(ctx context.Context, input GetOfferForPetInput) (GetOfferForPetOutput, error) {
	// 1. Obtener todos los servicios
	allServices, err := uc.Repo.ListServices(ctx)
	if err != nil {
		return GetOfferForPetOutput{}, err
	}

	// 2. Separar servicios base y addons
	var baseServices, addonServices []domain.Service
	for _, svc := range allServices {
		if svc.IsAddon {
			addonServices = append(addonServices, svc)
		} else {
			baseServices = append(baseServices, svc)
		}
	}

	// 3. Filtrar servicios base elegibles
	var offers []ServiceOffer
	for _, base := range baseServices {
		if !base.IsEligibleFor(input.PetProfile) {
			continue
		}

		// 4. Encontrar addons que requieren este servicio base
		var allowedAddons []domain.Service
		for _, addon := range addonServices {
			if !containsServiceID(addon.RequiresParentIDs, base.ID) {
				continue
			}
			if !addon.IsEligibleFor(input.PetProfile) {
				continue
			}
			allowedAddons = append(allowedAddons, addon)
		}

		offers = append(offers, ServiceOffer{
			Service:       base,
			AllowedAddons: allowedAddons,
		})
	}

	return GetOfferForPetOutput{Offers: offers}, nil
}

func containsServiceID(slice []string, id string) bool {
	for _, s := range slice {
		if s == id {
			return true
		}
	}
	return false
}
