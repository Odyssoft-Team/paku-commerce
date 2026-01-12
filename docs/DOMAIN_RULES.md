# Reglas de dominio (Paku Commerce)

## Servicio: BAÑO (base)
- Elegible por: especie, peso, coat_type (y reglas adicionales futuras).
- Permite addons: deslanado, desmotado, corte uñas, etc.
- Debe mostrarse en catálogo aunque booking no tenga cupos.

## Addons (dependientes)
- Un addon no se puede vender solo.
- Regla: si existe addon en items -> debe existir su parent (ej. baño) en los items.
- Elegibilidad de addon depende del pet_profile.
  - Ejemplo: desmotado NO permitido si coat_type = hairless
  - Ejemplo: deslanado permitido si coat_type in [double,long]

## Pet Profile (mínimo canónico)
- species: dog|cat|other
- weight_kg: number
- coat_type: hairless|short|double|curly|wire|long|unknown

NOTA: evitar reglas hardcode por raza; raza se mapea a atributos (coat/size).

## Booking
- Booking administra disponibilidad y holds.
- Commerce/Checkout solo valida el hold y luego lo confirma al pagar.

## Precios
- Servicios: por regla (rango de peso) y/o surcharge por atributo.
- Productos: precio fijo por SKU/variante (sin mascota).

## Idempotencia
- confirm_payment y confirm_hold deben ser idempotentes (mismo resultado si se repite).
