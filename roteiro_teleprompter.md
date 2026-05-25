Roteiro Teleprompter — 5 minutos (300s)

Formato: cada linha começa com [MM:SS | S] indicando o tempo de início. Leia a linha inteira quando o cronômetro atingir o segundo indicado.
Anote o avanço de slide quando indicado (SLIDE X).

[00:00 | 0s] SLIDE 1
Olá — sou [seu nome]. Neste vídeo mostro um pipeline em Go com dois serviços: `processor` e `aggregator`, rodando localmente com LocalStack. (leia em 20s)

[00:20 | 20s] SLIDE 2
Por que dois serviços: o `processor` valida e enriquece eventos; o `aggregator` persiste e expõe a API. Separando responsabilidades, podemos escalar, testar e implantar cada parte independentemente. (leia em 30s)

[00:50 | 50s] SLIDE 2 (continuação)
Eles se comunicam por filas SQS: `raw-events` → `processed-events`. Filas desacoplam produtores e consumidores, permitem retries, e DLQ para mensagens problemáticas. (leia em 30s)

[01:20 | 80s] SLIDE 3
Decisões técnicas principais: escolhi Go 1.21 e AWS SDK v2 por maturidade e desempenho. Uso LocalStack para emular SQS e DynamoDB localmente e demonstrar integração real sem AWS. (leia em 25s)

[01:45 | 105s] SLIDE 3 (continuação)
Resiliência: worker pool configurável para paralelismo, retry com backoff e DLQ (configurado no init do LocalStack). Idempotência no aggregator previne duplicações. (leia em 25s)

[02:10 | 130s] SLIDE 4
Sobre persistência: DynamoDB tem duas tabelas: `events` para registros individuais e `developer_summary` para agregados. Observação técnica: foi necessário adicionar tags `dynamodbav` em `ProcessedEvent` para o SDK serializar corretamente. (leia em 35s)

[02:45 | 165s] SLIDE 5
Explicando `domain`, `usecase` e `infra` — objetivo rápido:
- `domain`: modelos e regras puras (entidades). Ex.: `ProcessedEvent`.
- `usecase`: orquestrações e lógica de negócio, sem acoplamento a infra.
- `infra`: implementações concretas (SQS, DynamoDB, HTTP).
(leia este bloco em 25s)

[03:10 | 190s] SLIDE 6
Testes: mostramos testes unitários e de integração.
- `processor`: testes de validação, worker pool e DLQ (integração).
- `aggregator`: teste de agregação e idempotência.
Os testes provam: validação concorrente, movimento para DLQ e agregação correta. (leia em 30s)

[03:40 | 220s] SLIDE 7
Demo rápido — vou executar os comandos principais: subir o ambiente, executar o seed e consultar a API. (leia em 5s)

[03:45 | 225s] SLIDE 7 (ação)
Mostre terminal: execute

```bash
docker compose up --build
```
(espere até o ambiente subir — fale enquanto esperamos: 20s)

[04:05 | 245s] SLIDE 7 (ação)
Em outro terminal, rode o seed:

```bash
chmod +x scripts/seed.sh && ./scripts/seed.sh
```
(espere 10s — fale: mensagens publicadas, agora verifico endpoints)

[04:15 | 255s] SLIDE 7 (verificação)
Mostrar no terminal:

```bash
curl -s http://localhost:8080/health
curl -s http://localhost:8080/metrics/dev-alice/summary
```
(5s)

[04:25 | 265s] SLIDE 8
Encerramento: autocrítica e próximos passos — o que faria diferente com mais tempo: substituir `Scan` por `Query` com GSI, adicionar validações numéricas robustas, e integrar tracing distribuído (OpenTelemetry). (leia em 25s)

[04:50 | 290s] SLIDE 9
Agradecimento e call to action: Código e instruções no repositório; se quiser, posso abrir PRs com melhorias ou gravar um walkthrough mais longo. Obrigado! (leia até 300s)
