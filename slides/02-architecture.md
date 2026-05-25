# Arquitetura

- `raw-events` (SQS) → `processor` → `processed-events` (SQS) → `aggregator` → DynamoDB
- Benefícios: desacoplamento, retries, DLQ, escalabilidade independente
- Componentes principais: LocalStack, Processor, Aggregator
