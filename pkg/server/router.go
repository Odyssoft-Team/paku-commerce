package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	httpSwagger "github.com/swaggo/http-swagger"

	carthttp "paku-commerce/internal/commerce/cart/http"
	checkouthttp "paku-commerce/internal/commerce/checkout/http"
)

func NewRouter() http.Handler {
	r := chi.NewRouter()

	r.Use(RequestIDMiddleware)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	// API routes bajo /api/v1/commerce
	r.Route("/api/v1/commerce", func(r chi.Router) {
		// Swagger UI - DENTRO del route group
		r.Get("/swagger/*", httpSwagger.Handler(
			httpSwagger.URL("/api/v1/commerce/swagger/doc.json"),
		))

		// Registrar rutas de checkout
		checkoutHandlers := checkouthttp.WireCheckoutHandlers()
		checkouthttp.RegisterRoutes(r, checkoutHandlers)

		// Registrar rutas de cart
		cartHandlers := carthttp.WireCartHandlers()
		carthttp.RegisterRoutes(r, cartHandlers)
	})

	return r
}
