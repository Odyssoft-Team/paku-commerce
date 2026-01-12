package memory

import (
	"context"
	"errors"

	"paku-commerce/internal/commerce/service/domain"
)

var ErrServiceNotFound = errors.New("service not found")

// ServiceRepository implementa domain.ServiceRepository en memoria.
type ServiceRepository struct {
	services []domain.Service
}

// NewServiceRepository crea un repositorio con datos de ejemplo.
func NewServiceRepository() *ServiceRepository {
	// Helper para weight pointer
	intPtr := func(v int) *int { return &v }

	services := []domain.Service{
		{
			ID:      "bath",
			Name:    "Baño",
			IsAddon: false,
			EligibilityRules: []domain.EligibilityRule{
				{
					AllowedSpecies: []string{domain.SpeciesDog},
				},
			},
			RequiresParentIDs: nil,
		},
		{
			ID:      "deshedding",
			Name:    "Deslanado",
			IsAddon: true,
			EligibilityRules: []domain.EligibilityRule{
				{
					AllowedCoatTypes: []string{domain.CoatTypeDouble, domain.CoatTypeLong},
				},
			},
			RequiresParentIDs: []string{"bath"},
		},
		{
			ID:      "dematting",
			Name:    "Desmotado",
			IsAddon: true,
			EligibilityRules: []domain.EligibilityRule{
				{
					ExcludedCoatTypes: []string{domain.CoatTypeHairless},
				},
			},
			RequiresParentIDs: []string{"bath"},
		},
	}

	return &ServiceRepository{services: services}
}

// ListServices retorna todos los servicios.
func (r *ServiceRepository) ListServices(ctx context.Context) ([]domain.Service, error) {
	return r.services, nil
}

// GetServiceByID busca un servicio por ID.
func (r *ServiceRepository) GetServiceByID(ctx context.Context, id string) (domain.Service, error) {
	for _, svc := range r.services {
		if svc.ID == id {
			return svc, nil
		}
	}
	return domain.Service{}, ErrServiceNotFound
}
