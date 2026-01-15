package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

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
// @Summary      Upsert cart
// @Description  Crear o actualizar el carrito del usuario
// @Tags         cart
// @Accept       json
// @Produce      json
// @Param        X-User-ID  header    string              true  "User ID"
// @Param        body       body      UpsertCartRequestDTO  true  "Cart data"
// @Success      200        {object}  CartResponseDTO
// @Failure      400        {object}  ErrorResponse
// @Failure      500        {object}  ErrorResponse
// @Router       /cart/me [put]
// @Security     UserID
func (h *CartHandlers) HandleUpsertCart(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		respondError(w, http.StatusBadRequest, "bad_request", "X-User-ID header is required")
		return
	}

	var req UpsertCartRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "bad_request", "invalid JSON")
		return
	}

	// Validaciones básicas
	if len(req.Items) == 0 {
		respondError(w, http.StatusBadRequest, "bad_request", "items cannot be empty")
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
		code, msg := mapErrorToCodeAndMessage(err)
		respondError(w, mapErrorToHTTPStatus(err), code, msg)
		return
	}

	resp := CartResponseDTO{Cart: toCartDTO(output.Cart)}
	respondJSON(w, http.StatusOK, resp)
}

// HandleGetCart maneja GET /cart/me.
// @Summary      Get cart
// @Description  Obtener el carrito del usuario
// @Tags         cart
// @Produce      json
// @Param        X-User-ID  header    string  true  "User ID"
// @Success      200        {object}  CartResponseDTO
// @Failure      404        {object}  ErrorResponse
// @Failure      500        {object}  ErrorResponse
// @Router       /cart/me [get]
// @Security     UserID
func (h *CartHandlers) HandleGetCart(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		respondError(w, http.StatusBadRequest, "bad_request", "X-User-ID header is required")
		return
	}

	input := cartusecases.GetCartInput{UserID: userID}
	output, err := h.GetCartUC.Execute(r.Context(), input)
	if err != nil {
		code, msg := mapErrorToCodeAndMessage(err)
		respondError(w, mapErrorToHTTPStatus(err), code, msg)
		return
	}

	resp := CartResponseDTO{Cart: toCartDTO(output.Cart)}
	respondJSON(w, http.StatusOK, resp)
}

// HandleDeleteCart maneja DELETE /cart/me.
// @Summary      Delete cart
// @Description  Eliminar el carrito del usuario
// @Tags         cart
// @Param        X-User-ID  header  string  true  "User ID"
// @Success      204        "No Content"
// @Failure      500        {object}  ErrorResponse
// @Router       /cart/me [delete]
// @Security     UserID
func (h *CartHandlers) HandleDeleteCart(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		respondError(w, http.StatusBadRequest, "bad_request", "X-User-ID header is required")
		return
	}

	input := cartusecases.DeleteCartInput{UserID: userID}
	_, err := h.DeleteCartUC.Execute(r.Context(), input)
	if err != nil {
		code, msg := mapErrorToCodeAndMessage(err)
		respondError(w, mapErrorToHTTPStatus(err), code, msg)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// HandleExpireCarts maneja POST /cart/expire.
// @Summary      Expire carts
// @Description  Expirar carritos vencidos (dev/admin)
// @Tags         cart
// @Accept       json
// @Produce      json
// @Param        body  body      ExpireRequestDTO  false  "Optional timestamp"
// @Success      200   {object}  ExpireResponseDTO
// @Failure      400   {object}  ErrorResponse
// @Failure      500   {object}  ErrorResponse
// @Router       /cart/expire [post]
func (h *CartHandlers) HandleExpireCarts(w http.ResponseWriter, r *http.Request) {
	var req ExpireRequestDTO

	// Body opcional: si no hay body o está vacío, no fallar
	if r.Body != nil {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil && err.Error() != "EOF" {
			respondError(w, http.StatusBadRequest, "bad_request", "invalid JSON")
			return
		}
	}

	// Determinar timestamp
	now := time.Now()
	if req.Now != nil && *req.Now != "" {
		parsed, err := time.Parse(time.RFC3339, *req.Now)
		if err != nil {
			respondError(w, http.StatusBadRequest, "bad_request", "invalid now format (use RFC3339)")
			return
		}
		now = parsed
	}

	// Ejecutar usecase
	input := cartusecases.ExpireCartsInput{Now: now}
	output, err := h.ExpireCartsUC.Execute(r.Context(), input)
	if err != nil {
		code, msg := mapErrorToCodeAndMessage(err)
		respondError(w, mapErrorToHTTPStatus(err), code, msg)
		return
	}

	resp := ExpireResponseDTO{ExpiredCount: output.ExpiredCount}
	respondJSON(w, http.StatusOK, resp)
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

func mapErrorToCodeAndMessage(err error) (string, string) {
	if errors.Is(err, cartdomain.ErrCartNotFound) {
		return "not_found", err.Error()
	}
	if errors.Is(err, cartdomain.ErrInvalidUserID) || errors.Is(err, cartdomain.ErrEmptyItems) {
		return "bad_request", err.Error()
	}
	return "internal", "internal server error"
}

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, status int, code string, message string) {
	respondJSON(w, status, ErrorResponse{
		Error: ErrorDTO{
			Code:    code,
			Message: message,
		},
	})
}
