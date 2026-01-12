package http

import (
	"github.com/go-chi/chi/v5"
)

// RegisterRoutes registra las rutas de checkout en el router.
func RegisterRoutes(r chi.Router, handlers *CheckoutHandlers) {
	r.Route("/checkout", func(r chi.Router) {
		r.Post("/quote", handlers.HandleQuote)
		r.Post("/orders", handlers.HandleCreateOrder)
		r.Post("/orders/{id}/confirm-payment", handlers.HandleConfirmPayment)
	})
}
