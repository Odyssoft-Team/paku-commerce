package domain

import "errors"

// Currency representa la moneda.
type Currency string

const (
	CurrencyPEN Currency = "PEN"
)

// Money representa una cantidad monetaria en minor units (centavos).
// Amount es int64 para evitar floats (ej: 3590 = S/ 35.90).
type Money struct {
	Amount   int64
	Currency Currency
}

var ErrCurrencyMismatch = errors.New("currency mismatch")

// NewMoney crea un Money con la cantidad y moneda especificadas.
func NewMoney(amount int64, currency Currency) Money {
	return Money{Amount: amount, Currency: currency}
}

// Zero retorna Money con valor cero en la moneda especificada.
func Zero(currency Currency) Money {
	return Money{Amount: 0, Currency: currency}
}

// Add suma dos Money. Retorna error si las monedas no coinciden.
func (m Money) Add(other Money) (Money, error) {
	if m.Currency != other.Currency {
		return Money{}, ErrCurrencyMismatch
	}
	return Money{Amount: m.Amount + other.Amount, Currency: m.Currency}, nil
}

// Sub resta otro Money. Retorna error si las monedas no coinciden.
func (m Money) Sub(other Money) (Money, error) {
	if m.Currency != other.Currency {
		return Money{}, ErrCurrencyMismatch
	}
	return Money{Amount: m.Amount - other.Amount, Currency: m.Currency}, nil
}

// MulInt multiplica el Money por un entero (para cantidades).
func (m Money) MulInt(n int64) Money {
	return Money{Amount: m.Amount * n, Currency: m.Currency}
}
