# Referencia de estilo (booking / historial)

- Usecases con struct { Repo ...; Now func() time.Time }
- Inputs/Outputs explícitos por caso de uso
- Ports para servicios externos (booking, payments, pets)
- Memory adapters como default para build rápido
- Postgres adapters como scaffold sin romper build
- Idempotencia en operaciones de confirmación/cancelación
