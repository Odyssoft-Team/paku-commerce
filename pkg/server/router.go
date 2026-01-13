package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"

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

	// Registrar rutas de checkout
	checkoutHandlers := checkouthttp.WireCheckoutHandlers()
	checkouthttp.RegisterRoutes(r, checkoutHandlers)

	// Registrar rutas de cart
	cartHandlers := carthttp.WireCartHandlers()
	carthttp.RegisterRoutes(r, cartHandlers)

	return r
}
