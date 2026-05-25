#!/bin/bash

set -e

echo "==> Criando filas SQS..."

# DLQs primeiro (precisam existir antes das filas principais)
awslocal sqs create-queue --queue-name raw-events-dlq
awslocal sqs create-queue --queue-name processed-events-dlq

# Fila principal raw-events com redrive para DLQ
awslocal sqs create-queue --queue-name raw-events \
  --attributes '{
    "RedrivePolicy": "{\"deadLetterTargetArn\":\"arn:aws:sqs:us-east-1:000000000000:raw-events-dlq\",\"maxReceiveCount\":\"3\"}"
  }'

# Fila principal processed-events com redrive para DLQ
awslocal sqs create-queue --queue-name processed-events \
  --attributes '{
    "RedrivePolicy": "{\"deadLetterTargetArn\":\"arn:aws:sqs:us-east-1:000000000000:processed-events-dlq\",\"maxReceiveCount\":\"3\"}"
  }'

echo "==> Filas SQS criadas."

echo "==> Criando tabelas DynamoDB..."

# Tabela de eventos individuais
awslocal dynamodb create-table \
  --table-name events \
  --attribute-definitions AttributeName=event_id,AttributeType=S \
  --key-schema AttributeName=event_id,KeyType=HASH \
  --billing-mode PAY_PER_REQUEST

# Tabela de resumo por desenvolvedor
awslocal dynamodb create-table \
  --table-name developer_summary \
  --attribute-definitions AttributeName=developer_id,AttributeType=S \
  --key-schema AttributeName=developer_id,KeyType=HASH \
  --billing-mode PAY_PER_REQUEST

echo "==> Tabelas DynamoDB criadas."
echo "==> LocalStack inicializado com sucesso!"