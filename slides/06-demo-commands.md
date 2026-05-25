# Demo — Comandos Rápidos

1) Subir ambiente

```bash
docker compose up --build
```

2) Executar seed (em outro terminal)

```bash
chmod +x scripts/seed.sh && ./scripts/seed.sh
```

3) Verificar endpoints

```bash
curl -s http://localhost:8080/health
curl -s http://localhost:8080/metrics/dev-alice/summary
```
