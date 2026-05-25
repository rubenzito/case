#!/bin/bash

# Script de seed: publica mensagens de teste na fila raw-events
# Requer o LocalStack rodando via docker-compose up

set -e

ENDPOINT="http://localhost:4566"
QUEUE="http://sqs.us-east-1.localhost.localstack.cloud:4566/000000000000/raw-events"

AWS_ACCESS_KEY_ID="${AWS_ACCESS_KEY_ID:-test}"
AWS_SECRET_ACCESS_KEY="${AWS_SECRET_ACCESS_KEY:-test}"
export AWS_ACCESS_KEY_ID AWS_SECRET_ACCESS_KEY

if ! command -v aws >/dev/null 2>&1; then
  echo "Erro: AWS CLI não encontrado. Instale o AWS CLI ou execute este script em um ambiente com aws disponível."
  echo "No Windows, uma opção é usar WSL ou Git Bash com aws instalado."
  exit 1
fi

publish() {
  local msg="$1"
  local label="$2"
  echo "==> Publicando: $label"
  aws --endpoint-url=$ENDPOINT sqs send-message \
    --queue-url $QUEUE \
    --message-body "$msg" \
    --region us-east-1 \
    --no-cli-pager
}

echo "=== Mensagens VÁLIDAS ==="

publish '{"event_id":"a1b2c3d4-e5f6-7890-abcd-ef1234567801","developer_id":"dev-alice","metric_type":"commits","value":8,"repository":"org/backend","timestamp":"2026-04-10T09:00:00Z"}' "dev-alice commits"

publish '{"event_id":"a1b2c3d4-e5f6-7890-abcd-ef1234567802","developer_id":"dev-alice","metric_type":"pull_requests","value":3,"repository":"org/backend","timestamp":"2026-04-10T10:00:00Z"}' "dev-alice pull_requests"

publish '{"event_id":"a1b2c3d4-e5f6-7890-abcd-ef1234567803","developer_id":"dev-alice","metric_type":"review_time_minutes","value":45,"repository":"org/backend","timestamp":"2026-04-10T11:00:00Z"}' "dev-alice review_time"

publish '{"event_id":"a1b2c3d4-e5f6-7890-abcd-ef1234567804","developer_id":"dev-bob","metric_type":"commits","value":12,"repository":"org/frontend","timestamp":"2026-04-11T09:00:00Z"}' "dev-bob commits"

publish '{"event_id":"a1b2c3d4-e5f6-7890-abcd-ef1234567805","developer_id":"dev-bob","metric_type":"pull_requests","value":5,"repository":"org/frontend","timestamp":"2026-04-11T10:00:00Z"}' "dev-bob pull_requests"

publish '{"event_id":"a1b2c3d4-e5f6-7890-abcd-ef1234567806","developer_id":"dev-bob","metric_type":"review_time_minutes","value":90,"repository":"org/frontend","timestamp":"2026-04-11T11:00:00Z"}' "dev-bob review_time"

publish '{"event_id":"a1b2c3d4-e5f6-7890-abcd-ef1234567807","developer_id":"dev-carol","metric_type":"commits","value":20,"repository":"org/infra","timestamp":"2026-04-12T09:00:00Z"}' "dev-carol commits"

publish '{"event_id":"a1b2c3d4-e5f6-7890-abcd-ef1234567808","developer_id":"dev-carol","metric_type":"review_time_minutes","value":120,"repository":"org/infra","timestamp":"2026-04-12T10:00:00Z"}' "dev-carol review_time"

publish '{"event_id":"a1b2c3d4-e5f6-7890-abcd-ef1234567809","developer_id":"dev-alice","metric_type":"commits","value":4,"repository":"org/api","timestamp":"2026-04-13T09:00:00Z"}' "dev-alice commits 2"

publish '{"event_id":"a1b2c3d4-e5f6-7890-abcd-ef1234567810","developer_id":"dev-bob","metric_type":"commits","value":7,"repository":"org/mobile","timestamp":"2026-04-13T10:00:00Z"}' "dev-bob commits 2"

publish '{"event_id":"a1b2c3d4-e5f6-7890-abcd-ef1234567811","developer_id":"dev-carol","metric_type":"pull_requests","value":2,"repository":"org/infra","timestamp":"2026-04-13T11:00:00Z"}' "dev-carol pull_requests"

publish '{"event_id":"a1b2c3d4-e5f6-7890-abcd-ef1234567812","developer_id":"dev-alice","metric_type":"review_time_minutes","value":30,"repository":"org/backend","timestamp":"2026-04-14T09:00:00Z"}' "dev-alice review_time 2"

publish '{"event_id":"a1b2c3d4-e5f6-7890-abcd-ef1234567813","developer_id":"dev-dave","metric_type":"commits","value":15,"repository":"org/data","timestamp":"2026-04-14T10:00:00Z"}' "dev-dave commits"

publish '{"event_id":"a1b2c3d4-e5f6-7890-abcd-ef1234567814","developer_id":"dev-dave","metric_type":"pull_requests","value":1,"repository":"org/data","timestamp":"2026-04-14T11:00:00Z"}' "dev-dave pull_requests"

publish '{"event_id":"a1b2c3d4-e5f6-7890-abcd-ef1234567815","developer_id":"dev-dave","metric_type":"review_time_minutes","value":60,"repository":"org/data","timestamp":"2026-04-14T12:00:00Z"}' "dev-dave review_time"

echo ""
echo "=== Mensagens INVÁLIDAS (vão para DLQ) ==="

publish '{"event_id":"nao-e-uuid","developer_id":"dev-alice","metric_type":"commits","value":5,"repository":"org/repo","timestamp":"2026-04-10T09:00:00Z"}' "event_id inválido"

publish '{"event_id":"a1b2c3d4-e5f6-7890-abcd-ef1234567899","developer_id":"","metric_type":"commits","value":5,"repository":"org/repo","timestamp":"2026-04-10T09:00:00Z"}' "developer_id vazio"

publish '{"event_id":"a1b2c3d4-e5f6-7890-abcd-ef1234567898","developer_id":"dev-test","metric_type":"review_time_minutes","value":9999,"repository":"org/repo","timestamp":"2026-04-10T09:00:00Z"}' "review_time_minutes > 1440"

echo ""
echo "=== Mensagens DUPLICADAS (idempotência do Aggregator) ==="

publish '{"event_id":"a1b2c3d4-e5f6-7890-abcd-ef1234567801","developer_id":"dev-alice","metric_type":"commits","value":8,"repository":"org/backend","timestamp":"2026-04-10T09:00:00Z"}' "DUPLICADA - dev-alice commits"

publish '{"event_id":"a1b2c3d4-e5f6-7890-abcd-ef1234567804","developer_id":"dev-bob","metric_type":"commits","value":12,"repository":"org/frontend","timestamp":"2026-04-11T09:00:00Z"}' "DUPLICADA - dev-bob commits"

echo ""
echo "=== Seed concluído! Total: 20 mensagens publicadas ==="
echo "=== Aguarde alguns segundos e consulte: ==="
echo "  GET http://localhost:8080/metrics/dev-alice/summary"
echo "  GET http://localhost:8080/metrics/dev-bob/summary"
echo "  GET http://localhost:8080/health"