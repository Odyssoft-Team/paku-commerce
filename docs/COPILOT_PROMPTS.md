# Prompts listos para Copilot (Claude)

## Prompt 1 - Skeleton + Router
Objetivo: inicializar proyecto Go (go.mod), server router (chi), health endpoint, wiring básico.
Restricciones:
- seguir árbol existente
- no meter lógica de negocio en http handlers
- usar memory repos por defecto

## Prompt 2 - Service Domain + Eligibility
Implementar en internal/commerce/service/domain:
- PetProfile + enums
- EligibilityRule (species/weight/coat_type include/exclude)
- Service entity (base/addon) + dependencia (requires parent)
Luego usecase GetOfferForPet:
- filtra por elegibilidad
- agrupa addons permitidos por servicio base

## Prompt 3 - Pricing Engine
Implementar:
- Money (int64 minor units + currency)
- PriceRule (service: weight range)
- QuoteItems usecase con estrategia por tipo (service/product)

## Prompt 4 - Promotions
Implementar:
- Coupon (code, active, constraints mínimas)
- ApplyDiscounts sobre Quote
- ValidateCoupon usecase

## Prompt 5 - Checkout Quote/Create/Confirm
Implementar:
- QuoteCheckout usecase: valida items + service deps + pricing + promotions (+ opcional hold validate)
- CreateOrder: persistir (memory) order pending_payment
- ConfirmPayment: idempotente, cambia estado + llama booking port ConfirmHold

## Prompt 6 - HTTP DTOs + Mapping
Crear DTOs por módulo y handlers:
- /service/offers:for-pet
- /checkout/quote
- /checkout/orders
- /checkout/payments/webhook (stub idempotente)
