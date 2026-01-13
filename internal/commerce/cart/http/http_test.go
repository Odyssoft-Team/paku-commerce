package http

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

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
