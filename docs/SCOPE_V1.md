# Scope MVP v1 (Congelado)

## Alcance implementado

### Módulo: Service
- ✅ Definición de servicios base (bath)
- ✅ Addons con dependencias (deshedding, dematting requieren bath)
- ✅ Reglas de elegibilidad por pet_profile (species, weight_kg, coat_type)
- ✅ Validación de dependencias addon/parent en checkout
- ⚠️ Catálogo estático (memory repo con 3 servicios)

### Módulo: Pricing
- ✅ Motor de precios por rango de peso (services)
- ✅ Money value object (int64 minor units, sin floats)
- ✅ QuoteItems usecase (cotiza servicios + productos)
- ✅ Reglas en memoria con ejemplos (bath, deshedding)
- ❌ Productos (scaffold pendiente)
- ❌ Surcharges dinámicos

### Módulo: Promotions
- ✅ Cupones por código (ej. BANO10)
- ✅ Promociones automáticas (ej. Tuesday Grooming)
- ✅ Validación de aplicabilidad (subtotal mínimo, tipos de item)
- ✅ ApplyDiscounts usecase (calcula descuentos sobre Quote)
- ❌ Cupones con uso limitado (max_uses)
- ❌ Promos por calendario/horario

### Módulo: Cart
- ✅ Carrito persistente por user_id (memory repo)
- ✅ TTL de 90 minutos desde last_updated
- ✅ Upsert/Get/Delete endpoints
- ✅ Expiración manual con side-effects (cancela order+hold)
- ✅ Actualización automática de refs (booking_hold_id, order_id)
- ❌ Migración automática de items al cambiar pet_profile

### Módulo: Checkout
- ✅ QuoteCheckout: valida + cotiza + aplica promos
- ✅ CreateOrder: crea orden pending_payment
- ✅ ConfirmPayment: marca paid (idempotente)
- ✅ StartCheckout: crea hold + order + actualiza cart
- ✅ Validaciones:
  - Addon requiere parent
  - Service elegible para pet
  - Qty > 0
- ✅ Reemplazo de hold previo en StartCheckout
- ❌ Webhook real de pagos
- ❌ Validación de disponibilidad en booking

### Ports (interfaces)
- ✅ BookingClient: CreateHold, ConfirmHold, CancelHold (stub no-op)
- ✅ PaymentsClient: ValidatePayment (stub no-op)
- ✅ CheckoutClient (in-process): CancelOrder
- ❌ PetsClient: GetPetProfile (pendiente integración)

### HTTP API
- ✅ GET /health
- ✅ PUT /cart/me
- ✅ GET /cart/me
- ✅ DELETE /cart/me
- ✅ POST /cart/expire
- ✅ POST /checkout/quote
- ✅ POST /checkout/orders
- ✅ POST /checkout/start
- ✅ POST /checkout/orders/{id}/confirm-payment
- ❌ Admin endpoints (CRUD servicios/reglas)

### Tests
- ✅ Unit tests: service eligibility, pricing, promotions
- ✅ Unit tests: cart usecases (expire con side-effects)
- ✅ Unit tests: checkout usecases (quote, create, confirm)
- ✅ HTTP tests: endpoints completos
- ✅ E2E test: flujo cart → start → confirm-payment
- ❌ Integration tests con DB real
- ❌ Load tests

## Fuera de scope MVP v1

### Booking
- ❌ Integración real con ms-booking
- ❌ Validación de disponibilidad de slots
- ❌ Sincronización de hold states
- ✅ Stub permite desarrollo sin dependencias

### Payments
- ❌ Integración con pasarela (Stripe/MercadoPago)
- ❌ Webhooks de confirmación
- ❌ Gestión de reembolsos
- ✅ Stub permite flujo completo sin pagos reales

### Productos
- ❌ Catálogo de productos
- ❌ Pricing fijo por SKU/variante
- ❌ Inventario y stock
- ✅ Scaffold existe en pricing (ItemTypeProduct)

### Autenticación
- ❌ JWT/OAuth
- ❌ Roles y permisos
- ✅ X-User-ID header para desarrollo

### Persistencia
- ❌ PostgreSQL
- ❌ Redis (cache)
- ❌ Migraciones
- ✅ Memory repos permiten desarrollo rápido

### Observabilidad
- ❌ Métricas (Prometheus)
- ❌ Tracing (Jaeger/OpenTelemetry)
- ❌ Logs estructurados (Logrus/Zap)
- ✅ Request-ID middleware básico

## Decisiones de diseño MVP v1

### Arquitectura
- **Clean Architecture** por módulo (domain/usecases/adapters/http)
- **Dominio puro**: sin imports de infra/http
- **Usecases orquestan**: sin lógica en handlers
- **DTOs solo en HTTP**: mapeo explícito domain ↔ DTO

### Money
- **int64 minor units** (centavos) para evitar floats
- **Currency enum** (PEN por ahora)
- **Operaciones seguras** (Add/Sub retornan error si currency mismatch)

### Idempotencia
- **ConfirmPayment**: mismo payment_ref retorna OK sin side-effects
- **CancelOrder**: ya cancelada retorna OK
- **StartCheckout**: reemplaza hold previo automáticamente

### Cart TTL
- **90 minutos** desde last_updated
- **Expiración con side-effects**: cancela order + hold vía ports
- **Manual trigger** vía POST /cart/expire (dev/cron)

### Reglas de negocio
- **Addon requiere parent**: validado en QuoteCheckout
- **Elegibilidad por atributos**: no hardcode de razas
- **Pricing server-side**: cliente nunca envía precios
- **Recalculo en checkout**: CreateOrder recotiza siempre

### Repos singleton
- **runtime.CartRepoSingleton**: compartido entre cart y checkout
- **runtime.OrderRepoSingleton**: compartido entre cart y checkout
- **Razón**: cart.expire necesita cancelar orders

## Criterios de éxito MVP v1

- ✅ `go test ./...` pasa sin errores
- ✅ `go build ./cmd/api` compila
- ✅ Servidor arranca en localhost:8080
- ✅ Flujo E2E completo: cart → start → confirm funciona
- ✅ Validaciones de negocio funcionan (addon sin parent = 422)
- ✅ Idempotencia confirmada (repetir confirm-payment = OK)
- ✅ Tests HTTP cubren happy path y errores principales
- ❌ No requiere DB ni servicios externos para desarrollo

## Roadmap post-MVP v1

### Fase 2: Persistencia
- PostgreSQL adapters para cart/order/service/pricing
- Migraciones con golang-migrate
- Connection pooling y health checks

### Fase 3: Integración Booking
- Implementar BookingClient real (HTTP/gRPC)
- Validar disponibilidad en StartCheckout
- Sincronizar estados de hold
- Retry logic y circuit breaker

### Fase 4: Integración Payments
- Webhook handler para provider
- Idempotencia por event_id
- Manejo de estados intermedios (processing/failed)
- Reconciliación de pagos

### Fase 5: Productos
- Catálogo de productos (shampoos, accesorios)
- Pricing fijo por SKU
- Inventario básico (count available)
- Agregar productos a cart

### Fase 6: Admin
- CRUD servicios (crear/editar/desactivar)
- CRUD reglas de pricing
- CRUD cupones/promociones
- Dashboard de órdenes

### Fase 7: Observabilidad
- Structured logging (Zap)
- Metrics (Prometheus)
- Tracing (OpenTelemetry)
- Health checks avanzados

## Notas técnicas

### Limitaciones conocidas
- **Cart memory**: volátil, se pierde al reiniciar
- **No concurrencia**: memory repos no thread-safe para writes masivos
- **Single currency**: solo PEN
- **No timezone handling**: timestamps en UTC
- **Stubs sin logs**: BookingClient/PaymentsClient no registran llamadas

### Deuda técnica aceptada
- Repos memory sin locks sofisticados (OK para MVP)
- No validation de pet_profile contra ms-pets (OK sin integración)
- Error messages no i18n (OK para MVP local)
- No rate limiting (OK sin producción)
- No graceful degradation si booking falla (OK con stubs)

### Dependencias externas
- **github.com/go-chi/chi/v5**: router HTTP
- **stdlib**: todo lo demás (net/http, encoding/json, context, time)

## Conclusión

MVP v1 cumple con:
- ✅ Demostrar arquitectura Clean
- ✅ Flujo completo cart → checkout → payment
- ✅ Validaciones de negocio críticas
- ✅ Idempotencia donde aplica
- ✅ Tests automatizados sin Postman
- ✅ Desarrollo local sin dependencias externas

**Estado:** ✅ COMPLETO y funcionando en local.
