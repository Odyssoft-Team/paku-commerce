package bookinghttp

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestCreateHold_OK_ReturnsHoldID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/api/v1/holds" {
			t.Errorf("expected /api/v1/holds, got %s", r.URL.Path)
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{
			"hold_id": "hold_abc",
		})
	}))
	defer server.Close()

	client, err := NewClient(Config{
		BaseURL: server.URL,
		Timeout: 1 * time.Second,
	})
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	holdID, err := client.CreateHold(context.Background(), "slot_123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if holdID != "hold_abc" {
		t.Errorf("expected hold_abc, got %s", holdID)
	}
}

func TestCreateHold_422_ReturnsErrSlotUnavailable(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": map[string]string{
				"code":    "slot_unavailable",
				"message": "Slot is no longer available",
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(Config{BaseURL: server.URL})

	_, err := client.CreateHold(context.Background(), "slot_456")
	if err != ErrSlotUnavailable {
		t.Errorf("expected ErrSlotUnavailable, got: %v", err)
	}
}

func TestValidateHold_410_ReturnsErrHoldExpired(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/api/v1/holds/hold_old/validate" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.WriteHeader(http.StatusGone)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": map[string]string{
				"code":    "hold_expired",
				"message": "Hold has expired",
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(Config{BaseURL: server.URL})

	err := client.ValidateHold(context.Background(), "hold_old")
	if err != ErrHoldExpired {
		t.Errorf("expected ErrHoldExpired, got: %v", err)
	}
}

func TestConfirmHold_404_ReturnsErrHoldNotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/api/v1/holds/hold_missing/confirm" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": map[string]string{
				"code":    "hold_not_found",
				"message": "Hold does not exist",
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(Config{BaseURL: server.URL})

	err := client.ConfirmHold(context.Background(), "hold_missing")
	if err != ErrHoldNotFound {
		t.Errorf("expected ErrHoldNotFound, got: %v", err)
	}
}

func TestCancelHold_404_IsIdempotent_ReturnsNil(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/api/v1/holds/hold_gone" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": map[string]string{
				"code":    "hold_not_found",
				"message": "Hold does not exist or already expired",
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(Config{BaseURL: server.URL})

	// Según contrato: 404 en CancelHold es idempotente (debe retornar nil)
	err := client.CancelHold(context.Background(), "hold_gone")
	if err != nil {
		t.Errorf("expected nil (idempotent), got: %v", err)
	}
}

func TestAuthorizationHeader_Sent_WhenAPIKeyProvided(t *testing.T) {
	var capturedAuth string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedAuth = r.Header.Get("Authorization")

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{
			"hold_id": "hold_xyz",
		})
	}))
	defer server.Close()

	client, _ := NewClient(Config{
		BaseURL: server.URL,
		APIKey:  "testkey",
		Timeout: 1 * time.Second,
	})

	_, err := client.CreateHold(context.Background(), "slot_auth")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedAuth := "Bearer testkey"
	if capturedAuth != expectedAuth {
		t.Errorf("expected Authorization: %s, got: %s", expectedAuth, capturedAuth)
	}
}

func TestTimeout_ReturnsErrBookingUnavailable(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client, _ := NewClient(Config{
		BaseURL: server.URL,
		Timeout: 50 * time.Millisecond,
	})

	_, err := client.CreateHold(context.Background(), "slot_timeout")
	if err == nil {
		t.Errorf("expected timeout error, got nil")
	}

	// Verificar que es ErrBookingUnavailable o HttpError con Err subyacente
	var httpErr *HttpError
	if errors.As(err, &httpErr) {
		if httpErr.Code != "booking_unavailable" {
			t.Errorf("expected booking_unavailable code, got: %s", httpErr.Code)
		}
	} else {
		t.Errorf("expected HttpError, got: %T", err)
	}
}

func TestCreateHold_200_OK_AlsoAccepted(t *testing.T) {
	// Algunos servicios pueden retornar 200 en lugar de 201
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"hold_id": "hold_200",
		})
	}))
	defer server.Close()

	client, _ := NewClient(Config{BaseURL: server.URL})

	holdID, err := client.CreateHold(context.Background(), "slot_200")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if holdID != "hold_200" {
		t.Errorf("expected hold_200, got %s", holdID)
	}
}

func TestConfirmHold_200_OK_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"booking_id": "booking_real_123",
			"status":     "confirmed",
		})
	}))
	defer server.Close()

	client, _ := NewClient(Config{BaseURL: server.URL})

	err := client.ConfirmHold(context.Background(), "hold_confirm")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestCancelHold_200_OK_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"status": "cancelled",
		})
	}))
	defer server.Close()

	client, _ := NewClient(Config{BaseURL: server.URL})

	err := client.CancelHold(context.Background(), "hold_cancel")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestValidateHold_200_OK_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client, _ := NewClient(Config{BaseURL: server.URL})

	err := client.ValidateHold(context.Background(), "hold_valid")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestNewClient_RequiresBaseURL(t *testing.T) {
	_, err := NewClient(Config{BaseURL: ""})
	if err == nil {
		t.Errorf("expected error when BaseURL is empty")
	}
}

func TestNewClient_DefaultTimeout(t *testing.T) {
	client, err := NewClient(Config{BaseURL: "http://test.local"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if client.cfg.Timeout != 5*time.Second {
		t.Errorf("expected default timeout 5s, got: %v", client.cfg.Timeout)
	}
}

func TestCreateHold_400_ReturnsBadRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": map[string]string{
				"code":    "bad_request",
				"message": "Invalid slot_id format",
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(Config{BaseURL: server.URL})

	_, err := client.CreateHold(context.Background(), "invalid")
	if err != ErrBookingBadRequest {
		t.Errorf("expected ErrBookingBadRequest, got: %v", err)
	}
}

func TestCreateHold_503_ReturnsUnavailable(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": map[string]string{
				"code":    "service_unavailable",
				"message": "Booking service is down",
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(Config{BaseURL: server.URL})

	_, err := client.CreateHold(context.Background(), "slot_down")
	if err != ErrBookingUnavailable {
		t.Errorf("expected ErrBookingUnavailable, got: %v", err)
	}
}
