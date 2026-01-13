package bookinghttp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Config contiene la configuración del cliente booking HTTP.
type Config struct {
	BaseURL string
	APIKey  string
	Timeout time.Duration
}

// Client implementa bookingports.BookingClient usando HTTP.
type Client struct {
	httpClient *http.Client
	cfg        Config
}

// NewClient crea un nuevo cliente HTTP para booking.
func NewClient(cfg Config) (*Client, error) {
	if cfg.BaseURL == "" {
		return nil, fmt.Errorf("BaseURL is required")
	}
	if cfg.Timeout == 0 {
		cfg.Timeout = 5 * time.Second
	}

	return &Client{
		httpClient: &http.Client{Timeout: cfg.Timeout},
		cfg:        cfg,
	}, nil
}

// CreateHold crea un hold de booking para un slot.
func (c *Client) CreateHold(ctx context.Context, slotID string) (string, error) {
	// TODO: confirm endpoint with booking service
	endpoint := c.cfg.BaseURL + "/api/v1/holds"

	reqBody := map[string]interface{}{
		"slot_id": slotID,
		// TODO: agregar user_id, service_items, request_id cuando esté definido cómo pasarlos
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewReader(bodyBytes))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	c.setHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", &HttpError{
			Code:       "booking_unavailable",
			Message:    "booking service unavailable",
			StatusCode: 0,
			Err:        err,
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusCreated || resp.StatusCode == http.StatusOK {
		var result struct {
			HoldID string `json:"hold_id"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return "", fmt.Errorf("failed to decode response: %w", err)
		}
		return result.HoldID, nil
	}

	return "", c.parseError(resp)
}

// ValidateHold verifica que un hold de booking sea válido.
func (c *Client) ValidateHold(ctx context.Context, holdID string) error {
	// TODO: confirm endpoint with booking service
	endpoint := c.cfg.BaseURL + "/api/v1/holds/" + holdID + "/validate"

	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	c.setHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return &HttpError{
			Code:       "booking_unavailable",
			Message:    "booking service unavailable",
			StatusCode: 0,
			Err:        err,
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return nil
	}

	return c.parseError(resp)
}

// ConfirmHold confirma un hold de booking tras pago exitoso.
func (c *Client) ConfirmHold(ctx context.Context, holdID string) error {
	// Path según docs/INTEGRATION_BOOKING.md
	endpoint := c.cfg.BaseURL + "/api/v1/holds/" + holdID + "/confirm"

	reqBody := map[string]interface{}{
		// TODO: agregar order_id, payment_ref, request_id cuando esté definido cómo pasarlos
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewReader(bodyBytes))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	c.setHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return &HttpError{
			Code:       "booking_unavailable",
			Message:    "booking service unavailable",
			StatusCode: 0,
			Err:        err,
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return nil
	}

	return c.parseError(resp)
}

// CancelHold cancela un hold de booking.
func (c *Client) CancelHold(ctx context.Context, holdID string) error {
	// Path según docs/INTEGRATION_BOOKING.md
	endpoint := c.cfg.BaseURL + "/api/v1/holds/" + holdID

	req, err := http.NewRequestWithContext(ctx, "DELETE", endpoint, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	c.setHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return &HttpError{
			Code:       "booking_unavailable",
			Message:    "booking service unavailable",
			StatusCode: 0,
			Err:        err,
		}
	}
	defer resp.Body.Close()

	// Según contrato: 404 es idempotente (hold ya cancelado/expirado)
	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusNotFound {
		return nil
	}

	return c.parseError(resp)
}

// setHeaders configura headers comunes para requests.
func (c *Client) setHeaders(req *http.Request) {
	req.Header.Set("Content-Type", "application/json")

	if c.cfg.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.cfg.APIKey)
	}

	// TODO: obtener X-Request-ID del context si existe
	// Necesitaría acceso a la key usada por RequestIDMiddleware
	// Por ahora dejamos como pending
}

// parseError parsea errores HTTP del servicio booking.
func (c *Client) parseError(resp *http.Response) error {
	bodyBytes, _ := io.ReadAll(resp.Body)

	var errorResponse struct {
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}

	_ = json.Unmarshal(bodyBytes, &errorResponse)

	httpErr := &HttpError{
		StatusCode: resp.StatusCode,
		Code:       errorResponse.Error.Code,
		Message:    errorResponse.Error.Message,
	}

	// Mapear códigos específicos según docs/INTEGRATION_BOOKING.md
	switch resp.StatusCode {
	case http.StatusNotFound:
		if errorResponse.Error.Code == "hold_not_found" {
			return ErrHoldNotFound
		}
		return httpErr
	case http.StatusGone:
		if errorResponse.Error.Code == "hold_expired" {
			return ErrHoldExpired
		}
		return httpErr
	case http.StatusUnprocessableEntity:
		if errorResponse.Error.Code == "slot_unavailable" {
			return ErrSlotUnavailable
		}
		return httpErr
	case http.StatusBadRequest:
		return ErrBookingBadRequest
	case http.StatusServiceUnavailable:
		return ErrBookingUnavailable
	default:
		return httpErr
	}
}
