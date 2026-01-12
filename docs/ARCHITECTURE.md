# Arquitectura (Clean) - paku-commerce

## Capas
- Domain: entidades + invariantes + value objects + interfaces (repos/ports).
- Usecases: casos de uso (inputs/outputs), reglas de aplicación, validación.
- Adapters: implementaciones concretas (memory/postgres/clients).
- HTTP: handlers, routes, DTO (mapping).

## Módulos
### commerce/service
Responsable de:
- Definir servicios base y addons.
- Elegibilidad por pet_profile.
- Dependencias: addon requiere servicio base (ej. desmotado requiere baño).

No responsable de:
- disponibilidad (booking)
- pagos

### pricing
Motor de precios:
- Estrategia Service: usa pet_profile (peso/coat/species) + reglas por rango.
- Estrategia Product: precio fijo/SKU/cantidad (sin mascota).

### promotions
- cupones por código (restricciones)
- promos automáticas (ej. martes 10% baños)
Se aplica sobre un Quote.

### commerce/checkout
Orquesta:
- Validación (service rules + pricing + promotions + booking hold)
- Creación de Order (pending_payment)
- Confirmación de pago (paid) + confirm hold en booking

## Reglas de diseño
- Los cálculos de precios SIEMPRE se hacen server-side.
- Checkout recalcula y revalida todo en el momento de crear orden.
- Webhooks y confirmaciones deben ser idempotentes.
