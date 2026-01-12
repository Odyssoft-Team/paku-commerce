# Workflow Copilot (Claude) - Vibe coding profesional

## Reglas para Copilot
1) NO inventar endpoints extra; seguir docs/API_CONTRACTS.md.
2) NO meter lógica en handlers. Usecases deben contener la orquestación.
3) Dominio puro: no deps a postgres/http.
4) Implementar primero adapters/memory para que el build pase.
5) Postgres: dejar scaffolding con TODOs, no bloquear el build.
6) Escribir tests mínimos para reglas críticas.

## Orden de implementación (paso a paso)
Fase 1 (Service rules):
- pet_profile + eligibility_rule evaluation
- service entity + addon dependency
- get_offer_for_pet usecase

Fase 2 (Pricing):
- money value object
- service pricing (por rango de peso)
- product pricing (fijo)
- quote_items usecase

Fase 3 (Promotions):
- coupon validation (simple)
- apply_discounts sobre quote

Fase 4 (Checkout):
- quote_checkout (compone service+pricing+promotions)
- create_order (pending_payment)
- confirm_payment (idempotente) + booking confirm via port

Fase 5 (HTTP wiring):
- router + routes + handlers + DTOs
- health endpoint

## Definition of done
- go test ./... pasa
- Build pasa
- handlers no contienen reglas
- quote y create_order recalculan precios server-side
