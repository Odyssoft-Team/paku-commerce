package http

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"

	carthttp "paku-commerce/internal/commerce/cart/http"
)

func setupTestRouter() http.Handler {
	// Wire handlers con repos singleton compartidos
	checkoutHandlers := WireCheckoutHandlers()
	cartHandlers := carthttp.WireCartHandlers()

	r := chi.NewRouter()
	RegisterRoutes(r, checkoutHandlers)
	carthttp.RegisterRoutes(r, cartHandlers)

	return r
}

func TestHTTP_Quote(t *testing.T) {
	router := setupTestRouter()

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
	router := setupTestRouter()

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
	router := setupTestRouter()

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
	router := setupTestRouter()

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

func TestHTTP_StartCheckout(t *testing.T) {
	router := setupTestRouter()

	// 1. Crear cart con items
	cartBody := map[string]interface{}{
		"pet_profile": map[string]interface{}{
			"species":   "dog",
			"weight_kg": 15,
			"coat_type": "short",
		},
		"items": []map[string]interface{}{
			{"type": "service", "id": "bath", "qty": 1},
		},
	}

	body, _ := json.Marshal(cartBody)
	cartReq := httptest.NewRequest("PUT", "/cart/me", bytes.NewReader(body))
	cartReq.Header.Set("Content-Type", "application/json")
	cartReq.Header.Set("X-User-ID", "user_start_1")
	router.ServeHTTP(httptest.NewRecorder(), cartReq)

	// 2. Start checkout
	startBody := map[string]interface{}{
		"slot_id": "slot_123",
	}

	startBodyBytes, _ := json.Marshal(startBody)
	startReq := httptest.NewRequest("POST", "/checkout/start", bytes.NewReader(startBodyBytes))
	startReq.Header.Set("Content-Type", "application/json")
	startReq.Header.Set("X-User-ID", "user_start_1")

	startRec := httptest.NewRecorder()
	router.ServeHTTP(startRec, startReq)

	// 3. Assert
	if startRec.Code != http.StatusOK {
		t.Errorf("expected status 200, got: %d, body: %s", startRec.Code, startRec.Body.String())
	}

	var resp StartCheckoutResponseDTO
	json.NewDecoder(startRec.Body).Decode(&resp)

	if resp.BookingHoldID == "" {
		t.Errorf("expected booking_hold_id to be set")
	}

	if resp.Order.Status != "pending_payment" {
		t.Errorf("expected order status pending_payment, got: %s", resp.Order.Status)
	}

	if resp.Cart.BookingHoldID == nil || *resp.Cart.BookingHoldID != resp.BookingHoldID {
		t.Errorf("expected cart.booking_hold_id to match response.booking_hold_id")
	}

	if resp.Cart.OrderID == nil || *resp.Cart.OrderID != resp.Order.ID {
		t.Errorf("expected cart.order_id to match response.order.id")
	}
}

func TestHTTP_StartCheckout_ReplaceHold(t *testing.T) {
	router := setupTestRouter()

	// 1. Crear cart
	cartBody := map[string]interface{}{
		"pet_profile": map[string]interface{}{"species": "dog", "weight_kg": 10, "coat_type": "short"},
		"items":       []map[string]interface{}{{"type": "service", "id": "bath", "qty": 1}},
	}

	body, _ := json.Marshal(cartBody)
	cartReq := httptest.NewRequest("PUT", "/cart/me", bytes.NewReader(body))
	cartReq.Header.Set("Content-Type", "application/json")
	cartReq.Header.Set("X-User-ID", "user_replace")
	router.ServeHTTP(httptest.NewRecorder(), cartReq)

	// 2. Primer start checkout
	start1Body := map[string]interface{}{"slot_id": "slot_111"}
	start1Bytes, _ := json.Marshal(start1Body)
	start1Req := httptest.NewRequest("POST", "/checkout/start", bytes.NewReader(start1Bytes))
	start1Req.Header.Set("Content-Type", "application/json")
	start1Req.Header.Set("X-User-ID", "user_replace")
	start1Rec := httptest.NewRecorder()
	router.ServeHTTP(start1Rec, start1Req)

	var resp1 StartCheckoutResponseDTO
	json.NewDecoder(start1Rec.Body).Decode(&resp1)
	firstHoldID := resp1.BookingHoldID

	// 3. Segundo start checkout (reemplazo)
	start2Body := map[string]interface{}{"slot_id": "slot_222"}
	start2Bytes, _ := json.Marshal(start2Body)
	start2Req := httptest.NewRequest("POST", "/checkout/start", bytes.NewReader(start2Bytes))
	start2Req.Header.Set("Content-Type", "application/json")
	start2Req.Header.Set("X-User-ID", "user_replace")
	start2Rec := httptest.NewRecorder()
	router.ServeHTTP(start2Rec, start2Req)

	var resp2 StartCheckoutResponseDTO
	json.NewDecoder(start2Rec.Body).Decode(&resp2)
	secondHoldID := resp2.BookingHoldID

	// 4. Assert
	if secondHoldID == firstHoldID {
		t.Errorf("expected new booking_hold_id, got same: %s", secondHoldID)
	}

	if resp2.Cart.BookingHoldID == nil || *resp2.Cart.BookingHoldID != secondHoldID {
		t.Errorf("expected cart to have new booking_hold_id")
	}
}

func TestHTTP_E2E_CartToStartToConfirmPayment(t *testing.T) {
	router := setupTestRouter()

	// PASO 1: PUT /cart/me - Crear cart
	cartBody := map[string]interface{}{
		"pet_profile": map[string]interface{}{
			"species":   "dog",
			"weight_kg": 15,
			"coat_type": "short",
		},
		"items": []map[string]interface{}{
			{"type": "service", "id": "bath", "qty": 1},
		},
	}

	cartBodyBytes, _ := json.Marshal(cartBody)
	cartReq := httptest.NewRequest("PUT", "/cart/me", bytes.NewReader(cartBodyBytes))
	cartReq.Header.Set("Content-Type", "application/json")
	cartReq.Header.Set("X-User-ID", "user_e2e_1")

	cartRec := httptest.NewRecorder()
	router.ServeHTTP(cartRec, cartReq)

	if cartRec.Code != http.StatusOK {
		t.Fatalf("PASO 1 failed: expected status 200, got: %d, body: %s", cartRec.Code, cartRec.Body.String())
	}

	// PASO 2: POST /checkout/start - Crear hold + order
	startBody := map[string]interface{}{
		"slot_id": "slot_e2e_1",
	}

	startBodyBytes, _ := json.Marshal(startBody)
	startReq := httptest.NewRequest("POST", "/checkout/start", bytes.NewReader(startBodyBytes))
	startReq.Header.Set("Content-Type", "application/json")
	startReq.Header.Set("X-User-ID", "user_e2e_1")

	startRec := httptest.NewRecorder()
	router.ServeHTTP(startRec, startReq)

	if startRec.Code != http.StatusOK {
		t.Fatalf("PASO 2 failed: expected status 200, got: %d, body: %s", startRec.Code, startRec.Body.String())
	}

	var startResp StartCheckoutResponseDTO
	if err := json.NewDecoder(startRec.Body).Decode(&startResp); err != nil {
		t.Fatalf("PASO 2: failed to decode response: %v", err)
	}

	// Assert PASO 2
	if startResp.BookingHoldID == "" {
		t.Errorf("PASO 2: expected non-empty booking_hold_id")
	}
	if startResp.Order.ID == "" {
		t.Errorf("PASO 2: expected non-empty order.id")
	}
	if startResp.Order.Status != "pending_payment" {
		t.Errorf("PASO 2: expected order.status pending_payment, got: %s", startResp.Order.Status)
	}
	if startResp.Cart.BookingHoldID == nil || *startResp.Cart.BookingHoldID != startResp.BookingHoldID {
		t.Errorf("PASO 2: expected cart.booking_hold_id to match booking_hold_id")
	}
	if startResp.Cart.OrderID == nil || *startResp.Cart.OrderID != startResp.Order.ID {
		t.Errorf("PASO 2: expected cart.order_id to match order.id")
	}

	// PASO 3: POST /checkout/orders/{id}/confirm-payment - Confirmar pago
	confirmBody := map[string]interface{}{
		"payment_ref": "pay_e2e_1",
		"paid_at":     "2026-01-13T10:00:00-05:00",
	}

	confirmBodyBytes, _ := json.Marshal(confirmBody)
	confirmReq := httptest.NewRequest("POST", "/checkout/orders/"+startResp.Order.ID+"/confirm-payment", bytes.NewReader(confirmBodyBytes))
	confirmReq.Header.Set("Content-Type", "application/json")

	confirmRec := httptest.NewRecorder()
	router.ServeHTTP(confirmRec, confirmReq)

	if confirmRec.Code != http.StatusOK {
		t.Fatalf("PASO 3 failed: expected status 200, got: %d, body: %s", confirmRec.Code, confirmRec.Body.String())
	}

	var confirmResp ConfirmPaymentResponseDTO
	if err := json.NewDecoder(confirmRec.Body).Decode(&confirmResp); err != nil {
		t.Fatalf("PASO 3: failed to decode response: %v", err)
	}

	// Assert PASO 3
	if confirmResp.Order.Status != "paid" {
		t.Errorf("PASO 3: expected order.status paid, got: %s", confirmResp.Order.Status)
	}
	if confirmResp.Order.PaymentRef == nil || *confirmResp.Order.PaymentRef != "pay_e2e_1" {
		t.Errorf("PASO 3: expected payment_ref pay_e2e_1")
	}

	// PASO 3b: Idempotencia - repetir confirm-payment con mismo payment_ref
	confirmReq2 := httptest.NewRequest("POST", "/checkout/orders/"+startResp.Order.ID+"/confirm-payment", bytes.NewReader(confirmBodyBytes))
	confirmReq2.Header.Set("Content-Type", "application/json")

	confirmRec2 := httptest.NewRecorder()
	router.ServeHTTP(confirmRec2, confirmReq2)

	if confirmRec2.Code != http.StatusOK {
		t.Errorf("PASO 3b (idempotencia) failed: expected status 200, got: %d", confirmRec2.Code)
	}

	var confirmResp2 ConfirmPaymentResponseDTO
	json.NewDecoder(confirmRec2.Body).Decode(&confirmResp2)

	if confirmResp2.Order.Status != "paid" {
		t.Errorf("PASO 3b: expected order.status to remain paid, got: %s", confirmResp2.Order.Status)
	}
	if confirmResp2.Order.PaymentRef == nil || *confirmResp2.Order.PaymentRef != "pay_e2e_1" {
		t.Errorf("PASO 3b: expected payment_ref to remain pay_e2e_1")
	}

	// PASO 4 (Opcional): GET /cart/me - Verificar que cart sigue existiendo
	getCartReq := httptest.NewRequest("GET", "/cart/me", nil)
	getCartReq.Header.Set("X-User-ID", "user_e2e_1")

	getCartRec := httptest.NewRecorder()
	router.ServeHTTP(getCartRec, getCartReq)

	if getCartRec.Code != http.StatusOK {
		t.Fatalf("PASO 4 failed: expected status 200, got: %d", getCartRec.Code)
	}

	// Decodificar usando CartResponseDTO del cart http package
	// Como estamos en checkout/http, necesitamos una estructura local o importar
	var getCartResp struct {
		Cart struct {
			ID            string  `json:"id"`
			UserID        string  `json:"user_id"`
			BookingHoldID *string `json:"booking_hold_id,omitempty"`
			OrderID       *string `json:"order_id,omitempty"`
		} `json:"cart"`
	}

	if err := json.NewDecoder(getCartRec.Body).Decode(&getCartResp); err != nil {
		t.Fatalf("PASO 4: failed to decode cart response: %v", err)
	}

	// Assert PASO 4
	if getCartResp.Cart.OrderID == nil || *getCartResp.Cart.OrderID != startResp.Order.ID {
		t.Errorf("PASO 4: expected cart.order_id to match order.id")
	}
	if getCartResp.Cart.BookingHoldID == nil || *getCartResp.Cart.BookingHoldID != startResp.BookingHoldID {
		t.Errorf("PASO 4: expected cart.booking_hold_id to match booking_hold_id")
	}
}
