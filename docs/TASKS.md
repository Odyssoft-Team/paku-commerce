# Checklist de tareas (para Copilot)

- [ ] go.mod + main + router + health
- [ ] service/domain: PetProfile + EligibilityRule + Service + Addon deps
- [ ] service/usecases: GetOfferForPet
- [ ] pricing/domain: Money + PriceRule + Quote
- [ ] pricing/usecases: QuoteItems (service/product strategies)
- [ ] promotions/domain: Coupon + Promotion (min)
- [ ] promotions/usecases: ValidateCoupon + ApplyDiscounts
- [ ] checkout/domain: Order + OrderItem + states + repository
- [ ] checkout/usecases: QuoteCheckout + CreateOrder + ConfirmPayment (idempotente)
- [ ] adapters/memory: repos (service/product/price_rules/coupons/orders)
- [ ] http: routes + handlers + dto + error mapping
- [ ] tests: eligibility, addon deps, pricing ranges, coupon apply, checkout quote
