package domain

import "context"

// ServiceRepository define el acceso a servicios.
type ServiceRepository interface {
	ListServices(ctx context.Context) ([]Service, error)
	GetServiceByID(ctx context.Context, id string) (Service, error)
}
