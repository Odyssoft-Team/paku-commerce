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

