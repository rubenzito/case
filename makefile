# ── Dev ──────────────────────────────────────────────────────────────────────

buildar:
	docker-compose up --build

subir:
	docker-compose up

down:
	docker-compose down -v


testes:
	cd services/processor && go test ./... -v
	cd services/aggregator && go test ./... -v

# comandos de teste específicos:
# cd services/processor && go test ./internal/integration -run TestDLQ_MessagesMovedAfterMaxReceive -v
# cd services/processor && go test ./internal/infra/worker -run TestWorkerPool_ProcessesConcurrently -v

seed:
	chmod +x scripts/seed.sh && ./scripts/seed.sh

# ── Inspecionar filas e tabelas ───────────────────────────────────────────────

dlq:
	curl -s "http://localhost:4566/000000000000/raw-events-dlq?Action=GetQueueAttributes&AttributeName.1=ApproximateNumberOfMessages"

events:
	curl -s http://localhost:8080/metrics/dev-alice
	curl -s http://localhost:8080/metrics/dev-bob

summary:
	curl -s http://localhost:8080/metrics/dev-alice/summary 
	curl -s http://localhost:8080/metrics/dev-bob/summary   

# ── API ───────────────────────────────────────────────────────────────────────

health:
	curl -s http://localhost:8080/health


