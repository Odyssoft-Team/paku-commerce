# Styleguide (igual que booking/historial)

## Go
- Paquetes pequeños, nombres claros.
- Dominio sin imports de infra/HTTP.
- Usecases reciben interfaces (repos/ports) por struct.
- Validación en dominio para invariantes; en usecase para reglas de aplicación.

## Errores
- Errores tipados por dominio (ej. ErrNotEligible, ErrMissingParent, ErrInvalidMoneyCurrency).
- HTTP mapea errores a status codes (400/404/409/422).

## DTO
- DTO solo en /http.
- Mapear Domain <-> DTO en handlers.

## Idempotencia
- ConfirmPaymentUseCase y Booking confirm deben aceptar reintentos sin duplicar.
- Webhook debe reconocer event_id/transaction_id.

## Testing (mínimo esperado)
- Unit tests para:
  - eligibility rules
  - addon parent validation
  - pricing ranges
  - promotions apply
  - checkout quote invariants
