package http

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
)

func setupTestRouter() (http.Handler, *CheckoutHandlers) {
	// Wire handlers una sola vez para compartir repos
	handlers := WireCheckoutHandlers()

	// Crear router local
	r := chi.NewRouter()
	RegisterRoutes(r, handlers)

	return r, handlers
}

func TestHTTP_Quote(t *testing.T) {
	router, _ := setupTestRouter()

	reqBody := map[string]interface{}{
		"pet_profile": map[string]interface{}{
			"species":   "dog",
			"weight_kg": 15,
			"coat_type": "double",
		},
		"items": []map[string]interface{}{
			{
				"type": "service",
				"id":   "bath",
				"qty":  1,
			},
		},
		"coupon_code": "BANO10",
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/checkout/quote", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	// Assert
	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got: %d, body: %s", rec.Code, rec.Body.String())
	}

	var resp QuoteResponseDTO
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Quote.Total.Amount == 0 {
		t.Errorf("expected non-zero total")
	}

	if resp.Quote.Subtotal.Amount <= resp.Quote.Total.Amount {
		t.Errorf("expected subtotal > total due to discount")
	}
}

func TestHTTP_CreateOrder(t *testing.T) {
	router, _ := setupTestRouter()

	reqBody := map[string]interface{}{
		"pet_profile": map[string]interface{}{
			"species":   "dog",
			"weight_kg": 10,
			"coat_type": "short",
		},
		"items": []map[string]interface{}{
			{
				"type": "service",
				"id":   "bath",
				"qty":  1,
			},
		},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/checkout/orders", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	// Assert
	if rec.Code != http.StatusCreated {
		t.Errorf("expected status 201, got: %d, body: %s", rec.Code, rec.Body.String())
	}

	var resp CreateOrderResponseDTO
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Order.ID == "" {
		t.Errorf("expected non-empty order ID")
	}

	if resp.Order.Status != "pending_payment" {
		t.Errorf("expected status pending_payment, got: %s", resp.Order.Status)
	}
}

func TestHTTP_ConfirmPayment(t *testing.T) {
	router, _ := setupTestRouter()

	// Primero crear orden
	createReqBody := map[string]interface{}{
		"pet_profile": map[string]interface{}{
			"species":   "dog",
			"weight_kg": 10,
			"coat_type": "short",
		},
		"items": []map[string]interface{}{
			{
				"type": "service",
				"id":   "bath",
				"qty":  1,
			},
		},
	}

	createBody, _ := json.Marshal(createReqBody)
	createReq := httptest.NewRequest("POST", "/checkout/orders", bytes.NewReader(createBody))
	createReq.Header.Set("Content-Type", "application/json")

	createRec := httptest.NewRecorder()
	router.ServeHTTP(createRec, createReq)

	var createResp CreateOrderResponseDTO
	json.NewDecoder(createRec.Body).Decode(&createResp)
	orderID := createResp.Order.ID

	// Ahora confirmar pago
	confirmReqBody := map[string]interface{}{
		"payment_ref": "tx_123",
	}

	confirmBody, _ := json.Marshal(confirmReqBody)
	confirmReq := httptest.NewRequest("POST", "/checkout/orders/"+orderID+"/confirm-payment", bytes.NewReader(confirmBody))
	confirmReq.Header.Set("Content-Type", "application/json")

	confirmRec := httptest.NewRecorder()
	router.ServeHTTP(confirmRec, confirmReq)

	// Assert
	if confirmRec.Code != http.StatusOK {
		t.Errorf("expected status 200, got: %d, body: %s", confirmRec.Code, confirmRec.Body.String())
	}

	var confirmResp ConfirmPaymentResponseDTO
	if err := json.NewDecoder(confirmRec.Body).Decode(&confirmResp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if confirmResp.Order.Status != "paid" {
		t.Errorf("expected status paid, got: %s", confirmResp.Order.Status)
	}

	if confirmResp.Order.PaymentRef == nil || *confirmResp.Order.PaymentRef != "tx_123" {
		t.Errorf("expected payment_ref tx_123")
	}
}

func TestHTTP_AddonWithoutParent_Returns422(t *testing.T) {
	router, _ := setupTestRouter()

	reqBody := map[string]interface{}{
		"pet_profile": map[string]interface{}{
			"species":   "dog",
			"weight_kg": 15,
			"coat_type": "double",
		},
		"items": []map[string]interface{}{
			{
				"type": "service",
				"id":   "deshedding", // addon sin parent
				"qty":  1,
			},
		},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/checkout/quote", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	// Assert
	if rec.Code != http.StatusUnprocessableEntity {
		t.Errorf("expected status 422, got: %d", rec.Code)
	}
}
