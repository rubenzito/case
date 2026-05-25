# Documentação de Testes

Este arquivo descreve os testes presentes no projeto, o que cada teste faz, o resultado esperado e como gerar evidência (saída) para apresentação.

Resumo rápido
- Testes unitários: validam regras de negócio e usecases (rápidos, não dependem de infra externa).
- Testes de integração: verificam comportamento com LocalStack (SQS / DLQ / comportamento do worker pool).

Como rodar todos os testes (rápido):
```bash
go test ./... -v
```

Nota: os testes de integração (DLQ e Worker Pool) exigem o LocalStack disponível em `http://localhost:4566`.
Se você não tiver o LocalStack rodando, os testes de integração serão ignorados automaticamente.

1) Testes unitários (já adicionados)
- `services/processor/internal/usecase/process_test.go`
  - O que faz: valida as regras de `RawEvent`, testa o enriquecimento (`Enrich`) e comportamento do usecase `ProcessEvent` (sucesso, falha do publisher, evento inválido).
  - Resultado esperado: todos os casos unitários passam.
  - Evidência: saída do comando `go test ./services/processor/... -v` contendo `PASS` e os nomes dos testes:

    Exemplo de saída esperada:
    ```
    === RUN   TestProcessEvent_Execute_Success
    --- PASS: TestProcessEvent_Execute_Success (0.00s)
    === RUN   TestProcessEvent_Execute_PublishError
    --- PASS: TestProcessEvent_Execute_PublishError (0.00s)
    PASS
    ok  github.com/rubenzito/case/processor/internal/usecase 0.288s
    ```

- `services/aggregator/internal/usecase/aggregate_test.go`
  - O que faz: testa a idempotência (evento duplicado ignorado), caminho de sucesso e erro ao salvar (mock do repositório).
  - Resultado esperado: todos os testes unitários passam.
  - Evidência: saída do comando `go test ./services/aggregator/... -v` contendo `PASS`.

2) Testes de integração (LocalStack)
Observação: Estes testes falham/skipam se o LocalStack não estiver acessível em `http://localhost:4566`.

- `services/processor/internal/infra/queue/dlq_integration_test.go`
  - O que faz: cria uma fila principal (`raw`) e uma DLQ com `maxReceiveCount=2`, envia uma mensagem e inicia o `worker.Pool` com um processor falso que sempre falha. Depois verifica se a mensagem foi movida para a DLQ.
  - Resultado esperado: a mensagem é movida para a DLQ (o teste verifica diretamente a fila DLQ via SQS).
  - Como rodar (exemplo):
    ```bash
    # subir infra (LocalStack)
    docker-compose up -d

    # rodar apenas o teste de DLQ a partir do diretório do serviço processor
    cd services/processor && go test ./internal/integration -run TestDLQ_MessagesMovedAfterMaxReceive -v
    ```
  - Evidência: saída do `go test` com `PASS` para o teste `TestDLQ_MessagesMovedAfterMaxReceive`. Você também pode inspecionar a fila DLQ com AWS CLI:
    ```bash
    aws --endpoint-url=http://localhost:4566 sqs receive-message --queue-url <DLQ-URL>
    ```

- `services/processor/internal/infra/worker/pool_integration_test.go`
  - O que faz: cria uma fila SQS de teste, publica 6 mensagens, inicia um `worker.Pool` com 3 workers e um processor que dorme 500ms por mensagem; verifica que todas as mensagens são processadas e que houve concorrência (vários workers em execução ao mesmo tempo).
  - Resultado esperado: todas as mensagens processadas e tempo total compatível com execução concorrente (significativamente menor que processamento serial).
  - Como rodar:
    ```bash
    docker-compose up -d
    go test ./services/processor/internal/infra/worker -run TestWorkerPool_ProcessesConcurrently -v
    ```
  - Evidência: saída do `go test` mostrando `PASS` e tempo de execução (`elapsed`) relativamente curto. Também é útil gravar o tempo medido e o `maxRunning` (documentado no código de teste).

3) Recomendações para gerar evidência (para o vídeo/review)
- Grave (ou salve) a saída do comando `go test` em um arquivo:
  ```bash
  go test ./... -v | tee test-output.txt
  ```
- Para os testes de integração, grave também os logs do LocalStack e do `docker-compose`:
  ```bash
  docker-compose logs --no-color > docker-logs.txt
  ```
- Se preferir uma captura visual, faça uma screenshot do terminal com os `PASS` dos testes e do `docker-compose ps` mostrando os containers `localstack`, `processor` e `aggregator`.

4) Observações técnicas / limites
- Os testes de integração dependem de LocalStack e de rede local; se você preferir não rodar o Docker, os testes unitários ainda dão boa cobertura da lógica central.
- Para tornar os testes de integração reprodutíveis em CI, configure um job que inicialize o LocalStack (via docker-compose) antes de executar `go test`.

Se quiser, eu posso:
- 1) ajustar os testes para rodarem mais rápido (diminuir timeouts),
- 2) adicionar scripts `make test`/`make integration-test` no `Makefile`,
- 3) commitar e criar um PR com as mudanças.

---
Arquivo criado automaticamente pelo assistente — sinta-se à vontade para editar a redação.
