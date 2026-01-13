package http

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"

	checkoutmemory "paku-commerce/internal/commerce/checkout/adapters/memory"
)

func setupTestCartRouter() http.Handler {
	orderRepo := checkoutmemory.NewOrderRepository()
	handlers := WireCartHandlers(orderRepo)

	r := chi.NewRouter()
	RegisterRoutes(r, handlers)
	return r
}

func TestHTTP_UpsertCart(t *testing.T) {
	router := setupTestCartRouter()

	reqBody := map[string]interface{}{
		"pet_profile": map[string]interface{}{
			"species":   "dog",
			"weight_kg": 15,
			"coat_type": "short",
		},
		"items": []map[string]interface{}{
			{"type": "service", "id": "bath", "qty": 1},
		},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("PUT", "/cart/me", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", "user_123")

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got: %d, body: %s", rec.Code, rec.Body.String())
	}

	var resp CartResponseDTO
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp.Cart.UserID != "user_123" {
		t.Errorf("expected user_id user_123")
	}
}

func TestHTTP_UpsertCart_MissingUserID(t *testing.T) {
	router := setupTestCartRouter()

	reqBody := map[string]interface{}{
		"pet_profile": map[string]interface{}{"species": "dog", "weight_kg": 15, "coat_type": "short"},
		"items":       []map[string]interface{}{{"type": "service", "id": "bath", "qty": 1}},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("PUT", "/cart/me", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	// Sin X-User-ID

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got: %d", rec.Code)
	}

	var resp ErrorResponse
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp.Error.Code != "bad_request" {
		t.Errorf("expected error code bad_request, got: %s", resp.Error.Code)
	}
}

func TestHTTP_GetCart(t *testing.T) {
	router := setupTestCartRouter()

	// Primero crear cart
	createBody := map[string]interface{}{
		"pet_profile": map[string]interface{}{"species": "dog", "weight_kg": 10, "coat_type": "short"},
		"items":       []map[string]interface{}{{"type": "service", "id": "bath", "qty": 1}},
	}
	body, _ := json.Marshal(createBody)
	createReq := httptest.NewRequest("PUT", "/cart/me", bytes.NewReader(body))
	createReq.Header.Set("Content-Type", "application/json")
	createReq.Header.Set("X-User-ID", "user_456")
	router.ServeHTTP(httptest.NewRecorder(), createReq)

	// Ahora obtener
	getReq := httptest.NewRequest("GET", "/cart/me", nil)
	getReq.Header.Set("X-User-ID", "user_456")
	getRec := httptest.NewRecorder()
	router.ServeHTTP(getRec, getReq)

	if getRec.Code != http.StatusOK {
		t.Errorf("expected status 200, got: %d", getRec.Code)
	}
}

func TestHTTP_GetCart_NotFound(t *testing.T) {
	router := setupTestCartRouter()

	getReq := httptest.NewRequest("GET", "/cart/me", nil)
	getReq.Header.Set("X-User-ID", "nonexistent_user")
	getRec := httptest.NewRecorder()
	router.ServeHTTP(getRec, getReq)

	if getRec.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got: %d", getRec.Code)
	}

	var resp ErrorResponse
	json.NewDecoder(getRec.Body).Decode(&resp)
	if resp.Error.Code != "not_found" {
		t.Errorf("expected error code not_found, got: %s", resp.Error.Code)
	}
}

func TestHTTP_DeleteCart(t *testing.T) {
	router := setupTestCartRouter()

	// Crear cart
	createBody := map[string]interface{}{
		"pet_profile": map[string]interface{}{"species": "dog", "weight_kg": 10, "coat_type": "short"},
		"items":       []map[string]interface{}{{"type": "service", "id": "bath", "qty": 1}},
	}
	body, _ := json.Marshal(createBody)
	createReq := httptest.NewRequest("PUT", "/cart/me", bytes.NewReader(body))
	createReq.Header.Set("Content-Type", "application/json")
	createReq.Header.Set("X-User-ID", "user_789")
	router.ServeHTTP(httptest.NewRecorder(), createReq)

	// Eliminar
	delReq := httptest.NewRequest("DELETE", "/cart/me", nil)
	delReq.Header.Set("X-User-ID", "user_789")
	delRec := httptest.NewRecorder()
	router.ServeHTTP(delRec, delReq)

	if delRec.Code != http.StatusNoContent {
		t.Errorf("expected status 204, got: %d", delRec.Code)
	}

	// Verificar que no existe
	getReq := httptest.NewRequest("GET", "/cart/me", nil)
	getReq.Header.Set("X-User-ID", "user_789")
	getRec := httptest.NewRecorder()
	router.ServeHTTP(getRec, getReq)

	if getRec.Code != http.StatusNotFound {
		t.Errorf("expected status 404 after delete, got: %d", getRec.Code)
	}
}

func TestHTTP_ExpireCarts(t *testing.T) {
	router := setupTestCartRouter()

	// Crear cart con timestamp antiguo (vencido)
	// Usar upsert y luego manipular directamente el repo si es posible,
	// o simplemente crear y esperar que expire endpoint funcione
	createBody := map[string]interface{}{
		"pet_profile": map[string]interface{}{"species": "dog", "weight_kg": 10, "coat_type": "short"},
		"items":       []map[string]interface{}{{"type": "service", "id": "bath", "qty": 1}},
	}
	body, _ := json.Marshal(createBody)
	createReq := httptest.NewRequest("PUT", "/cart/me", bytes.NewReader(body))
	createReq.Header.Set("Content-Type", "application/json")
	createReq.Header.Set("X-User-ID", "user_expire")
	router.ServeHTTP(httptest.NewRecorder(), createReq)

	// Ejecutar expire con timestamp futuro (91 minutos después)
	futureTime := time.Now().Add(91 * time.Minute).Format(time.RFC3339)
	expireBody := map[string]interface{}{
		"now": futureTime,
	}
	expireBodyBytes, _ := json.Marshal(expireBody)
	expireReq := httptest.NewRequest("POST", "/cart/expire", bytes.NewReader(expireBodyBytes))
	expireReq.Header.Set("Content-Type", "application/json")

	expireRec := httptest.NewRecorder()
	router.ServeHTTP(expireRec, expireReq)

	if expireRec.Code != http.StatusOK {
		t.Errorf("expected status 200, got: %d, body: %s", expireRec.Code, expireRec.Body.String())
	}

	var expireResp ExpireResponseDTO
	json.NewDecoder(expireRec.Body).Decode(&expireResp)
	if expireResp.ExpiredCount < 1 {
		t.Errorf("expected at least 1 expired cart, got: %d", expireResp.ExpiredCount)
	}

	// Verificar que el cart fue eliminado
	getReq := httptest.NewRequest("GET", "/cart/me", nil)
	getReq.Header.Set("X-User-ID", "user_expire")
	getRec := httptest.NewRecorder()
	router.ServeHTTP(getRec, getReq)

	if getRec.Code != http.StatusNotFound {
		t.Errorf("expected cart to be deleted (404), got: %d", getRec.Code)
	}
}

func TestHTTP_ExpireCarts_EmptyBody(t *testing.T) {
	router := setupTestCartRouter()

	// Sin body
	expireReq := httptest.NewRequest("POST", "/cart/expire", nil)
	expireRec := httptest.NewRecorder()
	router.ServeHTTP(expireRec, expireReq)

	if expireRec.Code != http.StatusOK {
		t.Errorf("expected status 200 even with empty body, got: %d", expireRec.Code)
	}

	var expireResp ExpireResponseDTO
	json.NewDecoder(expireRec.Body).Decode(&expireResp)
	// Debería retornar 0 o más, pero no fallar
	if expireResp.ExpiredCount < 0 {
		t.Errorf("invalid expired count: %d", expireResp.ExpiredCount)
	}
}
