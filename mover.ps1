$src = "C:\Users\rubens\Downloads\files"
$dst = "C:\Users\rubens\git\case"

# Cria todas as pastas necessarias
$pastas = @(
    "docs",
    "infra\localstack",
    "scripts",
    "services\processor\cmd",
    "services\processor\internal\domain",
    "services\processor\internal\usecase",
    "services\processor\internal\infra\config",
    "services\processor\internal\infra\queue",
    "services\processor\internal\infra\worker",
    "services\processor\internal\infra\tracing",
    "services\aggregator\cmd",
    "services\aggregator\internal\domain",
    "services\aggregator\internal\usecase",
    "services\aggregator\internal\infra\config",
    "services\aggregator\internal\infra\queue",
    "services\aggregator\internal\infra\repository",
    "services\aggregator\internal\infra\api",
    "services\aggregator\internal\infra\tracing"
)

foreach ($pasta in $pastas) {
    New-Item -ItemType Directory -Force -Path "$dst\$pasta" | Out-Null
}

Write-Host "Pastas criadas!" -ForegroundColor Green

# Mapa: nome do arquivo em Downloads\files -> caminho destino no projeto
$arquivos = @{
    "docker-compose.yml"  = "docker-compose.yml"
    "Makefile"            = "Makefile"
    "README.md"           = "README.md"
    "openapi.yaml"        = "docs\openapi.yaml"
    "init-aws.sh"         = "infra\localstack\init-aws.sh"
    "seed.sh"             = "scripts\seed.sh"

    # processor
    "processor-go.mod"        = "services\processor\go.mod"
    "processor-Dockerfile"    = "services\processor\Dockerfile"
    "processor-main.go"       = "services\processor\cmd\main.go"
    "processor-event.go"      = "services\processor\internal\domain\event.go"
    "processor-event_test.go" = "services\processor\internal\domain\event_test.go"
    "processor-process.go"    = "services\processor\internal\usecase\process.go"
    "processor-process_test.go" = "services\processor\internal\usecase\process_test.go"
    "processor-config.go"     = "services\processor\internal\infra\config\config.go"
    "processor-consumer.go"   = "services\processor\internal\infra\queue\consumer.go"
    "processor-publisher.go"  = "services\processor\internal\infra\queue\publisher.go"
    "processor-pool.go"       = "services\processor\internal\infra\worker\pool.go"
    "processor-tracer.go"     = "services\processor\internal\infra\tracing\tracer.go"

    # aggregator
    "aggregator-go.mod"        = "services\aggregator\go.mod"
    "aggregator-Dockerfile"    = "services\aggregator\Dockerfile"
    "aggregator-main.go"       = "services\aggregator\cmd\main.go"
    "aggregator-event.go"      = "services\aggregator\internal\domain\event.go"
    "aggregator-aggregate.go"  = "services\aggregator\internal\usecase\aggregate.go"
    "aggregator-config.go"     = "services\aggregator\internal\infra\config\config.go"
    "aggregator-consumer.go"   = "services\aggregator\internal\infra\queue\consumer.go"
    "aggregator-dynamodb.go"   = "services\aggregator\internal\infra\repository\dynamodb.go"
    "aggregator-handler.go"    = "services\aggregator\internal\infra\api\handler.go"
    "aggregator-tracer.go"     = "services\aggregator\internal\infra\tracing\tracer.go"
}

foreach ($arquivo in $arquivos.Keys) {
    $origem  = "$src\$arquivo"
    $destino = "$dst\$($arquivos[$arquivo])"

    if (Test-Path $origem) {
        Copy-Item -Path $origem -Destination $destino -Force
        Write-Host "OK: $arquivo -> $($arquivos[$arquivo])" -ForegroundColor Cyan
    } else {
        Write-Host "NAO ENCONTRADO: $arquivo" -ForegroundColor Red
    }
}

Write-Host ""
Write-Host "Pronto! Agora rode:" -ForegroundColor Green
Write-Host "  cd C:\Users\rubens\git\case\services\processor && go mod tidy"
Write-Host "  cd C:\Users\rubens\git\case\services\aggregator && go mod tidy"
Write-Host "  cd C:\Users\rubens\git\case && docker-compose up --build"