# Testes (resumo rápido)

- `processor`:
  - validações de evento
  - worker pool concorrente
  - integração DLQ (mensagens inválidas vão para DLQ após retries)
- `aggregator`:
  - idempotência (ignore duplicatas)
  - atualização incremental do summary

Comando para rodar:

```bash
cd services/processor && go test ./... -v
cd services/aggregator && go test ./... -v
```
