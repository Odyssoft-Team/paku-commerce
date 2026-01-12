# Copilot Instructions (paku-commerce)

You are coding in an existing Clean Architecture tree.
Rules:
- Do NOT put business rules in HTTP handlers.
- Domain must be pure (no postgres/http imports).
- Usecases orchestrate and depend on interfaces.
- Implement memory adapters first.
- Follow docs:
  - docs/ARCHITECTURE.md
  - docs/DOMAIN_RULES.md
  - docs/API_CONTRACTS.md
  - docs/STYLEGUIDE.md
