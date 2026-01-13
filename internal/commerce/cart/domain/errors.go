package domain

import "errors"

var (
	ErrCartNotFound  = errors.New("cart not found")
	ErrInvalidUserID = errors.New("user_id is required")
	ErrEmptyItems    = errors.New("items cannot be empty")
)
