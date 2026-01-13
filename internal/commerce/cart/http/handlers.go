package http

import (
	"encoding/json"
	"errors"
	"net/http"

	cartdomain "paku-commerce/internal/commerce/cart/domain"
	cartusecases "paku-commerce/internal/commerce/cart/usecases"
)

// CartHandlers contiene los handlers de cart.
type CartHandlers struct {
	UpsertCartUC  *cartusecases.UpsertCart
	GetCartUC     *cartusecases.GetCart
	DeleteCartUC  *cartusecases.DeleteCart
	ExpireCartsUC *cartusecases.ExpireCarts
}

// HandleUpsertCart maneja PUT /cart/me.
func (h *CartHandlers) HandleUpsertCart(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		respondError(w, http.StatusBadRequest, "X-User-ID header is required")
		return
	}

	var req UpsertCartRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	// Validaciones básicas
	if len(req.Items) == 0 {
		respondError(w, http.StatusBadRequest, "items cannot be empty")
		return
	}

	input := cartusecases.UpsertCartInput{
		UserID:        userID,
		PetProfile:    req.PetProfile.toPetProfile(),
		Items:         toPurchaseItems(req.Items),
		BookingHoldID: req.BookingHoldID,
		OrderID:       req.OrderID,
	}

	output, err := h.UpsertCartUC.Execute(r.Context(), input)
	if err != nil {
		respondError(w, mapErrorToHTTPStatus(err), err.Error())
		return
	}

	resp := CartResponseDTO{Cart: toCartDTO(output.Cart)}
	respondJSON(w, http.StatusOK, resp)
}

// HandleGetCart maneja GET /cart/me.
func (h *CartHandlers) HandleGetCart(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		respondError(w, http.StatusBadRequest, "X-User-ID header is required")
		return
	}

	input := cartusecases.GetCartInput{UserID: userID}
	output, err := h.GetCartUC.Execute(r.Context(), input)
	if err != nil {
		respondError(w, mapErrorToHTTPStatus(err), err.Error())
		return
	}

	resp := CartResponseDTO{Cart: toCartDTO(output.Cart)}
	respondJSON(w, http.StatusOK, resp)
}

// HandleDeleteCart maneja DELETE /cart/me.
func (h *CartHandlers) HandleDeleteCart(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		respondError(w, http.StatusBadRequest, "X-User-ID header is required")
		return
	}

	input := cartusecases.DeleteCartInput{UserID: userID}
	_, err := h.DeleteCartUC.Execute(r.Context(), input)
	if err != nil {
		respondError(w, mapErrorToHTTPStatus(err), err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func mapErrorToHTTPStatus(err error) int {
	if errors.Is(err, cartdomain.ErrCartNotFound) {
		return http.StatusNotFound
	}
	if errors.Is(err, cartdomain.ErrInvalidUserID) || errors.Is(err, cartdomain.ErrEmptyItems) {
		return http.StatusBadRequest
	}
	return http.StatusInternalServerError
}

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, ErrorResponse{Error: message})
}
