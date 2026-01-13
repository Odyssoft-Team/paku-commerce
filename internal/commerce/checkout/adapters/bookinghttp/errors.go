package bookinghttp

import (
	"errors"
	"fmt"
)

var (
	// ErrHoldNotFound indica que el hold no existe o ya expiró.
	ErrHoldNotFound = errors.New("hold not found")

	// ErrHoldExpired indica que el hold expiró antes de confirmarse.
	ErrHoldExpired = errors.New("hold expired")

	// ErrSlotUnavailable indica que el slot no está disponible.
	ErrSlotUnavailable = errors.New("slot unavailable")

	// ErrBookingUnavailable indica que el servicio booking está caído/timeout.
	ErrBookingUnavailable = errors.New("booking service unavailable")

	// ErrBookingBadRequest indica un error de validación en el request.
	ErrBookingBadRequest = errors.New("booking bad request")
)

// HttpError contiene detalles de un error HTTP del servicio booking.
type HttpError struct {
	StatusCode int
	Code       string
	Message    string
	Err        error // error subyacente (ej. timeout, network)
}

func (e *HttpError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("booking http error (status=%d, code=%s): %s", e.StatusCode, e.Code, e.Message)
	}
	if e.Err != nil {
		return fmt.Sprintf("booking http error: %v", e.Err)
	}
	return fmt.Sprintf("booking http error (status=%d)", e.StatusCode)
}

func (e *HttpError) Unwrap() error {
	return e.Err
}
