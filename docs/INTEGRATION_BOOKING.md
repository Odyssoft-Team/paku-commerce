# INTEGRATION_BOOKING (paku-commerce ↔ ms-booking)

## 1) Contexto

En paku-commerce, el ciclo de vida de un booking hold se gestiona en tres momentos clave:

1. **Creación del hold**: cuando el usuario elige una fecha/hora y presiona continuar (POST /checkout/start)
2. **Cancelación del hold**: cuando expira el carrito o el usuario cancela antes de pagar
3. **Confirmación del hold**: cuando el pago se confirma exitosamente

El hold es una reserva temporal del slot que impide que otros usuarios lo reserven hasta que:
- Se confirme el pago (se convierte en booking real)
- Expire el TTL del carrito (90 min)
- Se cancele explícitamente

---

## 2) Flujos de integración

### 2.1 Start checkout → CreateHold

**Trigger:** `POST /checkout/start`

**Request paku-commerce:**
```json
{
  "slot_id": "slot_123"
}
```

**Headers:**
- `X-User-ID: user_456`
- `X-Request-ID: req_abc123` (para idempotencia)

**Proceso interno:**
1. paku-commerce carga el cart del usuario (ya contiene pet_profile + items)
2. Valida que el cart tenga items de servicios (no productos)
3. Si cart tiene booking_hold_id previo → CancelHold(old_hold_id)
4. Llama `booking.CreateHold(...)` con datos del cart

**Datos enviados a booking (CreateHold):**
- `slot_id`: desde request
- `user_id`: desde X-User-ID header
- `service_items`: desde cart.Items (filtrar solo ItemTypeService)
  - `[{service_id, qty}]`
- `request_id`: X-Request-ID (para idempotencia)
- **PENDING:** `pet_profile` (species, weight_kg, coat_type) - si booking lo requiere para validación de capacidad
- **PENDING:** `tenant_id` - para multi-tenant futuro

**Response booking (CreateHold):**
```json
{
  "hold_id": "hold_xyz789",
  "expires_at": "2026-01-15T12:30:00Z",
  "slot": {
    "id": "slot_123",
    "datetime": "2026-01-20T10:00:00Z"
  }
}
```

**Uso en paku-commerce:**
- `hold_id` → se guarda en cart.BookingHoldID y order.BookingHoldID
- `expires_at` → informativo (cart tiene su propio TTL)

**En caso de error:**
- Si booking retorna error (slot no disponible, etc.) → NO crear order
- Retornar 422 al cliente con mensaje de booking
- No actualizar cart

---

### 2.2 Expire cart → CancelHold

**Trigger:** `POST /cart/expire` (manual en v1) o job/cron futuro

**Proceso interno:**
1. ExpireCarts usecase busca carts vencidos (now > cart.ExpiresAt)
2. Para cada cart vencido con booking_hold_id:
   - Llama `booking.CancelHold(hold_id)` (best-effort)
   - Si falla: loggear pero no bloquear
3. Llama `checkout.CancelOrder(order_id)` (best-effort)
4. Elimina cart de repo

**Datos enviados a booking (CancelHold):**
- `hold_id`: desde cart.BookingHoldID
- `reason`: "cart_expired" (opcional)
- `request_id`: X-Request-ID (para trazabilidad)

**Response booking (CancelHold):**
```json
{
  "status": "cancelled"
}
```

**Idempotencia:**
- Cancelar un hold ya cancelado/expirado debe retornar 200 OK
- No debe retornar error si el hold no existe (idempotente)

**Errores esperados:**
- `hold_not_found`: OK (ya expiró/cancelado)
- `hold_already_confirmed`: WARNING (inconsistencia, loggear)
- Network error: loggear, no bloquear el proceso de expiración

---

### 2.3 Confirm payment → ConfirmHold

**Trigger:** `POST /checkout/orders/{id}/confirm-payment`

**Request paku-commerce:**
```json
{
  "payment_ref": "pay_abc123",
  "paid_at": "2026-01-15T12:00:00Z"
}
```

**Proceso interno:**
1. ConfirmPayment usecase carga order
2. Marca order como paid (idempotente por payment_ref)
3. Si order.BookingHoldID != nil:
   - Llama `booking.ConfirmHold(hold_id)`
   - Si falla: NO persistir order como paid, retornar error
4. Persiste order con status=paid

**Datos enviados a booking (ConfirmHold):**
- `hold_id`: desde order.BookingHoldID
- `order_id`: order.ID (para correlación)
- `payment_ref`: input.PaymentRef (para trazabilidad)
- `request_id`: X-Request-ID (para idempotencia)

**Response booking (ConfirmHold):**
```json
{
  "booking_id": "booking_real_123",
  "status": "confirmed"
}
```

**Uso en paku-commerce:**
- `booking_id`: **PENDING** - no se persiste en v1 (puede agregarse como order.BookingID futuro)
- Status: validar que sea "confirmed"

**Estrategia de error (ACTUAL en código):**
- Si ConfirmHold falla → NO persistir order.MarkPaid()
- Retornar error al cliente (500 o 422 según tipo)
- El cliente puede reintentar confirm-payment con mismo payment_ref (idempotente)
- **PENDING:** estrategia de compensación (deshacer pago si booking ya confirmó pero persist falló)

**Idempotencia:**
- Confirmar un hold ya confirmado debe retornar 200 OK con booking_id
- No debe crear booking duplicado

**Errores esperados:**
- `hold_not_found`: ERROR crítico (inconsistencia)
- `hold_expired`: ERROR crítico (el hold se perdió antes de confirmar)
- `hold_already_confirmed`: OK idempotente (retornar booking_id existente)

---

## 3) Contrato de datos detallado

### 3.1 CreateHold

**Method signature (port):**
```go
CreateHold(ctx context.Context, slotID string) (holdID string, error)
```

**Propuesta para integración real (HTTP):**

**Request:**
```http
POST /api/v1/holds
Content-Type: application/json
X-Request-ID: {request_id}
Authorization: Bearer {api_key}

{
  "slot_id": "slot_123",
  "user_id": "user_456",
  "service_items": [
    {"service_id": "bath", "qty": 1},
    {"service_id": "deshedding", "qty": 1}
  ],
  "request_id": "req_abc123"
}
```

**Response 201 Created:**
```json
{
  "hold_id": "hold_xyz789",
  "expires_at": "2026-01-15T12:30:00Z",
  "slot": {
    "id": "slot_123",
    "datetime": "2026-01-20T10:00:00Z",
    "service_type": "grooming"
  }
}
```

**Response 422 Unprocessable Entity:**
```json
{
  "error": {
    "code": "slot_unavailable",
    "message": "Slot is no longer available"
  }
}
```

**Origen de datos en paku-commerce:**
- `slot_id`: desde StartCheckoutInput
- `user_id`: desde X-User-ID header
- `service_items`: desde cart.Items (filtrar ItemTypeService)
- `request_id`: desde X-Request-ID header o generado

**PENDING:**
- Agregar `pet_profile` si booking lo requiere para validar capacidad
- Agregar `tenant_id` para multi-tenant

---

### 3.2 CancelHold

**Method signature (port):**
```go
CancelHold(ctx context.Context, holdID string) error
```

**Propuesta para integración real (HTTP):**

**Request:**
```http
DELETE /api/v1/holds/{hold_id}
X-Request-ID: {request_id}
Authorization: Bearer {api_key}

Optional body:
{
  "reason": "cart_expired"
}
```

**Response 200 OK:**
```json
{
  "status": "cancelled"
}
```

**Response 404 Not Found (idempotente):**
```json
{
  "error": {
    "code": "hold_not_found",
    "message": "Hold does not exist or already expired"
  }
}
```
**Nota:** paku-commerce debe tratar 404 como success (idempotente).

**Response 409 Conflict:**
```json
{
  "error": {
    "code": "hold_already_confirmed",
    "message": "Cannot cancel confirmed hold"
  }
}
```
**Nota:** loggear como warning, indica inconsistencia de estado.

---

### 3.3 ConfirmHold

**Method signature (port):**
```go
ConfirmHold(ctx context.Context, holdID string) error
```

**Propuesta para integración real (HTTP):**

**Request:**
```http
POST /api/v1/holds/{hold_id}/confirm
Content-Type: application/json
X-Request-ID: {request_id}
Authorization: Bearer {api_key}

{
  "order_id": "order_abc123",
  "payment_ref": "pay_xyz",
  "request_id": "req_confirm_123"
}
```

**Response 200 OK:**
```json
{
  "booking_id": "booking_real_456",
  "status": "confirmed"
}
```

**Response 404 Not Found:**
```json
{
  "error": {
    "code": "hold_not_found",
    "message": "Hold does not exist or expired"
  }
}
```
**Nota:** ERROR crítico en paku-commerce, abortar confirm-payment.

**Response 410 Gone:**
```json
{
  "error": {
    "code": "hold_expired",
    "message": "Hold expired before confirmation"
  }
}
```
**Nota:** ERROR crítico, el slot se perdió.

**Response 200 OK (idempotente):**
```json
{
  "booking_id": "booking_real_456",
  "status": "confirmed",
  "message": "Hold already confirmed"
}
```
**Nota:** OK, retornar booking_id existente.

---

## 4) Errores y estrategia de manejo

### 4.1 CreateHold

| Error | Código | Acción paku-commerce |
|-------|--------|---------------------|
| Slot no disponible | 422 slot_unavailable | Retornar 422 al cliente, NO crear order |
| Hold duplicado (idempotente) | 200 OK | Retornar hold_id existente |
| Validación falla | 400 bad_request | Retornar 400 al cliente |
| Servicio caído | 503 service_unavailable | Retornar 503, cliente puede reintentar |
| Timeout | network error | Retornar 500, loggear |

**Estrategia:**
- Si CreateHold falla → NO crear order
- NO actualizar cart con booking_hold_id
- Retornar error al cliente sin side-effects

---

### 4.2 CancelHold

| Error | Código | Acción paku-commerce |
|-------|--------|---------------------|
| Hold no existe | 404 hold_not_found | OK (idempotente), continuar |
| Hold ya confirmado | 409 hold_already_confirmed | WARNING, loggear inconsistencia, continuar |
| Servicio caído | 503 | Loggear, continuar (best-effort) |
| Timeout | network error | Loggear, continuar |

**Estrategia:**
- Best-effort, no bloquear expiración de cart
- Loggear todos los errores para investigación
- Continuar con CancelOrder y DeleteCart

---

### 4.3 ConfirmHold

| Error | Código | Acción paku-commerce |
|-------|--------|---------------------|
| Hold no existe | 404 hold_not_found | ERROR, abortar confirm-payment, retornar 500 |
| Hold expirado | 410 hold_expired | ERROR, abortar, retornar 422 con mensaje |
| Hold ya confirmado (idempotente) | 200 OK | OK, continuar con persist paid |
| Servicio caído | 503 | ERROR, abortar, retornar 503 |
| Timeout | network error | ERROR, abortar, retornar 500 |

**Estrategia (ACTUAL en código):**
- Si ConfirmHold falla → NO persistir order.MarkPaid()
- Retornar error al cliente
- Cliente puede reintentar confirm-payment con mismo payment_ref (idempotente en paku-commerce)
- **PENDING:** si persist falla después de ConfirmHold exitoso, queda booking confirmado sin order paid (inconsistencia). Requiere compensación o retry logic.

---

## 5) Idempotencia

### CreateHold
**Recomendación:** Idempotente por `(user_id + slot_id)` o por `request_id`.

**Comportamiento esperado:**
- Si se llama dos veces con mismo request_id → retornar hold_id existente (200 OK)
- No crear hold duplicado
- Mismo TTL del hold original

**Implementación booking:**
- Store request_id por 24h
- Si request_id existe → retornar hold_id del registro existente

---

### CancelHold
**Comportamiento:** Idempotente por naturaleza.

**Casos:**
- Hold ya cancelado → 200 OK
- Hold no existe → 200 OK (o 404 que paku-commerce trata como OK)
- Hold ya confirmado → 409 (WARNING en paku-commerce)

**No requiere request_id:** la operación es safe para repetir.

---

### ConfirmHold
**Recomendación:** Idempotente por `request_id` o por `hold_id + payment_ref`.

**Comportamiento esperado:**
- Si hold ya confirmado con mismo payment_ref → 200 OK (retornar booking_id)
- Si hold ya confirmado con diferente payment_ref → 409 conflict
- No crear booking duplicado

**Implementación booking:**
- Store payment_ref en booking record
- Validar payment_ref match en reintentos

---

## 6) Correlación y trazabilidad

### Headers obligatorios
```http
X-Request-ID: {unique_id}
X-User-ID: {user_id}
Authorization: Bearer {api_key}
```

### Campos de logging en paku-commerce

**CreateHold:**




## 7) Estado del contrato

Este documento define el contrato de integración **v1** entre paku-commerce y ms-booking.

- El detalle HTTP es **referencial**.
- El contrato real está definido por los ports en paku-commerce.
- Cualquier cambio incompatible debe versionarse como v2.

Documento congelado para MVP v1.