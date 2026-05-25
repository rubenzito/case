# Domain / Usecase / Infra

- `domain`: entidades e modelos puros (ex.: `ProcessedEvent`).
- `usecase`: regras de negócio e orquestração (ex.: `AggregateEvent`).
- `infra`: adaptação para SQS, DynamoDB e HTTP (repositório + handlers).

Referências: `services/aggregator/internal/domain`, `.../usecase`, `.../infra`.
