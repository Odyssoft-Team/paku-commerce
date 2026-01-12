# Contratos de API (propuestos)

## 1) Offers (Service)
POST /service/offers:for-pet
Request:
- pet_profile (o pet_id si luego conectamos a ms-pets)
Response:
- services[]: base + addons permitidos + flags UI (available/hidden)
- (opcional) price_estimates

## 2) Checkout Quote
POST /checkout/quote
Request:
- pet_profile
- items[]: { type: service|product, id, qty, selected_addons? }
- booking_hold_id? (opcional en quote)
- coupon_code? (opcional)
Response:
- normalized_items + computed_prices
- subtotal, discounts, total
- validation_errors[] (si aplica)

## 3) Create Order
POST /checkout/orders
Request:
- pet_profile
- items[]
- booking_hold_id (requerido si es servicio con booking)
- coupon_code?
- customer info (si aplica)
Response:
- order_id
- payment_intent (si aplica)
- totals

## 4) Payment Webhook
POST /checkout/payments/webhook
Request:
- provider payload
Behavior:
- idempotente
- si pago confirmado: mark order paid + confirm booking hold

## 5) Admin (más adelante)
CRUD services / rules / coupons / promos / price_rules
