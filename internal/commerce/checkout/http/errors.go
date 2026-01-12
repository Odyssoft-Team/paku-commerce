package http

import (
	"errors"
	"net/http"

	checkoutdomain "paku-commerce/internal/commerce/checkout/domain"
	checkoutusecases "paku-commerce/internal/commerce/checkout/usecases"
	pricingusecases "paku-commerce/internal/pricing/usecases"
	promotionsdomain "paku-commerce/internal/promotions/domain"
	promotionsusecases "paku-commerce/internal/promotions/usecases"
)

// ErrorResponse representa un error HTTP.
type ErrorResponse struct {
	Error string `json:"error"`
}

// mapErrorToHTTPStatus mapea errores de dominio/usecase a status HTTP.
func mapErrorToHTTPStatus(err error) int {
	// 404 - Not Found
	if errors.Is(err, checkoutdomain.ErrOrderNotFound) {
		return http.StatusNotFound
	}

	// 409 - Conflict
	if errors.Is(err, checkoutdomain.ErrPaymentConflict) {
		return http.StatusConflict
	}

	// 422 - Unprocessable Entity (errores de negocio)
	if errors.Is(err, checkoutusecases.ErrMissingParentService) ||
		errors.Is(err, checkoutusecases.ErrServiceNotEligible) ||
		errors.Is(err, checkoutusecases.ErrInvalidQuantity) ||
		errors.Is(err, checkoutusecases.ErrUnknownService) ||
		errors.Is(err, pricingusecases.ErrNoPriceRule) ||
		errors.Is(err, promotionsusecases.ErrInvalidCoupon) ||
		errors.Is(err, promotionsdomain.ErrCouponNotFound) ||
		errors.Is(err, checkoutdomain.ErrOrderCancelled) ||
		errors.Is(err, checkoutdomain.ErrInvalidOrderState) {
		return http.StatusUnprocessableEntity
	}

	// 500 - Internal Server Error (default)
	return http.StatusInternalServerError
}
