package pets

import (
	"context"

	servicedomain "paku-commerce/internal/commerce/service/domain"
)

// PetsClient define la integración con el servicio de mascotas.
// TODO: implementar en fase de integración
type PetsClient interface {
	// GetPetProfile obtiene el perfil de una mascota por ID.
	GetPetProfile(ctx context.Context, petID string) (servicedomain.PetProfile, error)
}
