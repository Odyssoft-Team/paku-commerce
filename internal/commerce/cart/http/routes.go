package http

import (
	"github.com/go-chi/chi/v5"
)

// RegisterRoutes registra las rutas de cart en el router.
func RegisterRoutes(r chi.Router, handlers *CartHandlers) {
	r.Route("/cart", func(r chi.Router) {
		r.Put("/me", handlers.HandleUpsertCart)
		r.Get("/me", handlers.HandleGetCart)
		r.Delete("/me", handlers.HandleDeleteCart)
	})
}
