# paku-commerce

Microservicio de comercio para Paku (servicios + productos) con checkout **dentro** del mismo servicio.

## Objetivo
- Ofertar servicios (baño y addons) y productos.
- Calcular precios (por peso/atributos en servicios, normal en productos).
- Aplicar cupones/promos.
- Ejecutar checkout: validar, cotizar, crear orden, integrar pago y confirmar booking (hold).

## Bounded modules (internos)
- internal/commerce/service: catálogo de servicios + reglas de elegibilidad + addons.
- internal/commerce/product: catálogo de productos (base).
- internal/pricing: motor de precios (estrategias por tipo).
- internal/promotions: cupones y promos.
- internal/commerce/checkout: orquestación del cierre de compra (order lifecycle).

## Principios de estilo (igual a booking / historial)
- Clean Architecture por módulo: domain (puro), usecases (orquestación), adapters (infra), http (handlers+dto).
- Dominio sin dependencias de infra.
- Usecases sin HTTP.
- DTO solo en capa http.
- Idempotencia donde aplica (confirmación de pago, confirmación de hold, webhooks).

## Documentación para Copilot
Ver docs/:
- docs/ARCHITECTURE.md
- docs/DOMAIN_RULES.md
- docs/API_CONTRACTS.md
- docs/COPILOT_WORKFLOW.md
- docs/COPILOT_PROMPTS.md
- docs/STYLEGUIDE.md
- docs/TASKS.md

---

## MVP v1 (Local)

### Requisitos
- Go 1.22+
- No requiere BD ni servicios externos (memory adapters)

### Ejecutar en local
```bash
go run ./cmd/api
# Server listening on :8080
```

### Endpoints disponibles

**Health check:**
```bash
curl http://localhost:8080/health
# ok
```

**1. Crear/actualizar carrito:**
```bash
curl -X PUT http://localhost:8080/cart/me \
  -H "X-User-ID: user_123" \
  -H "Content-Type: application/json" \
  -d '{
    "pet_profile": {
      "species": "dog",
      "weight_kg": 15,
      "coat_type": "short"
    },
    "items": [
      {"type": "service", "id": "bath", "qty": 1}
    ]
  }'
```

**2. Obtener carrito:**
```bash
curl http://localhost:8080/cart/me \
  -H "X-User-ID: user_123"
```

**3. Cotizar checkout (sin crear orden):**
```bash
curl -X POST http://localhost:8080/checkout/quote \
  -H "Content-Type: application/json" \
  -d '{
    "pet_profile": {"species": "dog", "weight_kg": 15, "coat_type": "double"},
    "items": [
      {"type": "service", "id": "bath", "qty": 1},
      {"type": "service", "id": "deshedding", "qty": 1}
    ],
    "coupon_code": "BANO10"
  }'
```

**4. Iniciar checkout (crea hold + order + actualiza cart):**
```bash
curl -X POST http://localhost:8080/checkout/start \
  -H "X-User-ID: user_123" \
  -H "Content-Type: application/json" \
  -d '{"slot_id": "slot_456"}'
```

**5. Confirmar pago (marca order como paid):**
```bash
curl -X POST http://localhost:8080/checkout/orders/{order_id}/confirm-payment \
  -H "Content-Type: application/json" \
  -d '{
    "payment_ref": "pay_xyz",
    "paid_at": "2026-01-13T10:00:00-05:00"
  }'
```

**6. Expirar carritos vencidos (manual/dev):**
```bash
curl -X POST http://localhost:8080/cart/expire
# {"expired_count": 0}
```

### Tests
```bash
# Todos los tests
go test ./...

# Tests específicos
go test ./internal/commerce/checkout/http -v
go test ./internal/commerce/cart/usecases -v

# Test E2E completo
go test ./internal/commerce/checkout/http -v -run TestHTTP_E2E
```

### Servicios de ejemplo (memory)
- **bath** (baño): S/ 35.00 (0-10kg), S/ 45.00 (11-20kg), S/ 60.00 (21-40kg)
- **deshedding** (deslanado): S/ 20.00 (0-20kg), S/ 30.00 (21-40kg) - requiere `bath`
- **dematting** (desmotado): S/ 20.00 - requiere `bath`, no permitido para hairless

### Cupones de ejemplo
- **BANO10**: 10% descuento en servicios, sin mínimo

### Limitaciones MVP v1
- Booking: stub no-op (no valida disponibilidad real)
- Payments: stub no-op (no integra pasarela)
- Repos: memoria volátil (se pierde al reiniciar)
- No auth real (X-User-ID header)
- Single tenant
- Solo servicios (productos pendientes)

### Próximos pasos
Ver `docs/SCOPE_V1.md` y `docs/TASKS.md` para roadmap.

