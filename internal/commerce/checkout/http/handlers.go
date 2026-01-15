package http

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	cartdomain "paku-commerce/internal/commerce/cart/domain"
	checkoutdomain "paku-commerce/internal/commerce/checkout/domain"
	checkoutusecases "paku-commerce/internal/commerce/checkout/usecases"
)

// CheckoutHandlers contiene los handlers de checkout.
type CheckoutHandlers struct {
	QuoteCheckoutUC  *checkoutusecases.QuoteCheckout
	CreateOrderUC    *checkoutusecases.CreateOrder
	ConfirmPaymentUC *checkoutusecases.ConfirmPayment
	StartCheckoutUC  *checkoutusecases.StartCheckout
}

// HandleQuote maneja POST /checkout/quote.
// @Summary      Quote checkout
// @Description  Cotizar un checkout sin crear orden
// @Tags         checkout
// @Accept       json
// @Produce      json
// @Param        body  body      QuoteRequestDTO  true  "Quote request"
// @Success      200   {object}  QuoteResponseDTO
// @Failure      400   {object}  ErrorResponse
// @Failure      422   {object}  ErrorResponse
// @Failure      500   {object}  ErrorResponse
// @Router       /api/v1/commerce/checkout/quote [post]
func (h *CheckoutHandlers) HandleQuote(w http.ResponseWriter, r *http.Request) {
	var req QuoteRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	// Validaciones básicas
	if len(req.Items) == 0 {
		respondError(w, http.StatusBadRequest, "items cannot be empty")
		return
	}

	// Construir input
	input := checkoutusecases.QuoteCheckoutInput{
		Intent: checkoutdomain.PurchaseIntent{
			PetProfile:    req.PetProfile.toPetProfile(),
			Items:         toPurchaseItems(req.Items),
			CouponCode:    req.CouponCode,
			BookingHoldID: req.BookingHoldID,
		},
	}

	// Ejecutar usecase
	output, err := h.QuoteCheckoutUC.Execute(r.Context(), input)
	if err != nil {
		respondError(w, mapErrorToHTTPStatus(err), err.Error())
		return
	}

	// Construir response
	discounts := make([]DiscountLineDTO, 0, len(output.Quote.Discounts))
	for _, d := range output.Quote.Discounts {
		discounts = append(discounts, DiscountLineDTO{
			Source: d.Source,
			Name:   d.Name,
			Amount: toMoneyDTO(d.Amount),
		})
	}

	resp := QuoteResponseDTO{
		Quote: QuoteDTO{
			Subtotal:      toMoneyDTO(output.Quote.OriginalSubtotal),
			TotalDiscount: toMoneyDTO(output.Quote.TotalDiscount),
			Total:         toMoneyDTO(output.Quote.Total),
			Discounts:     discounts,
		},
	}

	respondJSON(w, http.StatusOK, resp)
}

// HandleCreateOrder maneja POST /checkout/orders.
// @Summary      Create order
// @Description  Crear una orden pending_payment
// @Tags         checkout
// @Accept       json
// @Produce      json
// @Param        body  body      QuoteRequestDTO  true  "Order request"
// @Success      201   {object}  CreateOrderResponseDTO
// @Failure      400   {object}  ErrorResponse
// @Failure      422   {object}  ErrorResponse
// @Failure      500   {object}  ErrorResponse
// @Router       /api/v1/commerce/checkout/orders [post]
func (h *CheckoutHandlers) HandleCreateOrder(w http.ResponseWriter, r *http.Request) {
	var req QuoteRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	// Validaciones básicas
	if len(req.Items) == 0 {
		respondError(w, http.StatusBadRequest, "items cannot be empty")
		return
	}

	// Construir input
	input := checkoutusecases.CreateOrderInput{
		Intent: checkoutdomain.PurchaseIntent{
			PetProfile:    req.PetProfile.toPetProfile(),
			Items:         toPurchaseItems(req.Items),
			CouponCode:    req.CouponCode,
			BookingHoldID: req.BookingHoldID,
		},
	}

	// Ejecutar usecase
	output, err := h.CreateOrderUC.Execute(r.Context(), input)
	if err != nil {
		respondError(w, mapErrorToHTTPStatus(err), err.Error())
		return
	}

	resp := CreateOrderResponseDTO{
		Order: toOrderDTO(output.Order),
	}

	respondJSON(w, http.StatusCreated, resp)
}

// HandleConfirmPayment maneja POST /checkout/orders/{id}/confirm-payment.
// @Summary      Confirm payment
// @Description  Confirmar el pago de una orden
// @Tags         checkout
// @Accept       json
// @Produce      json
// @Param        id    path      string                    true  "Order ID"
// @Param        body  body      ConfirmPaymentRequestDTO  true  "Payment confirmation"
// @Success      200   {object}  ConfirmPaymentResponseDTO
// @Failure      400   {object}  ErrorResponse
// @Failure      404   {object}  ErrorResponse
// @Failure      500   {object}  ErrorResponse
// @Router       /api/v1/commerce/checkout/orders/{id}/confirm-payment [post]
func (h *CheckoutHandlers) HandleConfirmPayment(w http.ResponseWriter, r *http.Request) {
	orderID := chi.URLParam(r, "id")
	if orderID == "" {
		respondError(w, http.StatusBadRequest, "order ID is required")
		return
	}

	var req ConfirmPaymentRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	// Validaciones básicas
	if req.PaymentRef == "" {
		respondError(w, http.StatusBadRequest, "payment_ref is required")
		return
	}

	// Parse paidAt
	var paidAt time.Time
	if req.PaidAt != nil && *req.PaidAt != "" {
		parsed, err := time.Parse(time.RFC3339, *req.PaidAt)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid paid_at format (use RFC3339)")
			return
		}
		paidAt = parsed
	}

	// Construir input
	input := checkoutusecases.ConfirmPaymentInput{
		OrderID:    orderID,
		PaymentRef: req.PaymentRef,
		PaidAt:     paidAt,
	}

	// Ejecutar usecase
	output, err := h.ConfirmPaymentUC.Execute(r.Context(), input)
	if err != nil {
		respondError(w, mapErrorToHTTPStatus(err), err.Error())
		return
	}

	resp := ConfirmPaymentResponseDTO{
		Order: toOrderDTO(output.Order),
	}

	respondJSON(w, http.StatusOK, resp)
}

// HandleStartCheckout maneja POST /checkout/start.
// @Summary      Start checkout
// @Description  Iniciar checkout (crear hold + order + actualizar cart)
// @Tags         checkout
// @Accept       json
// @Produce      json
// @Param        X-User-ID  header    string                   true  "User ID"
// @Param        body       body      StartCheckoutRequestDTO  true  "Start checkout request"
// @Success      200        {object}  StartCheckoutResponseDTO
// @Failure      400        {object}  ErrorResponse
// @Failure      500        {object}  ErrorResponse
// @Router       /api/v1/commerce/checkout/start [post]
// @Security     UserID
func (h *CheckoutHandlers) HandleStartCheckout(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		respondError(w, http.StatusBadRequest, "X-User-ID header is required")
		return
	}

	var req StartCheckoutRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	if req.SlotID == "" {
		respondError(w, http.StatusBadRequest, "slot_id is required")
		return
	}

	input := checkoutusecases.StartCheckoutInput{
		UserID: userID,
		SlotID: req.SlotID,
	}

	output, err := h.StartCheckoutUC.Execute(r.Context(), input)
	if err != nil {
		status := mapStartCheckoutErrorToHTTPStatus(err)
		respondError(w, status, err.Error())
		return
	}

	resp := StartCheckoutResponseDTO{
		BookingHoldID: output.BookingHoldID,
		Order:         toOrderDTO(output.Order),
		Cart:          toCartSnapshotDTO(output.Cart),
	}

	respondJSON(w, http.StatusOK, resp)
}

func mapStartCheckoutErrorToHTTPStatus(err error) int {
	if err == cartdomain.ErrCartNotFound || err == cartdomain.ErrInvalidUserID || err == cartdomain.ErrEmptyItems {
		return http.StatusBadRequest
	}
	return http.StatusInternalServerError
}

// respondJSON escribe una respuesta JSON.
func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// respondError escribe una respuesta de error JSON.
func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, ErrorResponse{Error: message})
}
