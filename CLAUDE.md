# Avalauncher

Avalanche L1 chain management dashboard.

## Project Structure

- `cmd/avalauncher/` — Entry point
- `internal/config/` — Environment + cluster.yaml config
- `internal/database/` — pgx pool, schema bootstrap
- `internal/server/` — Echo HTTP server, routes, dashboard

## Build & Run

```bash
go build -o avalauncher ./cmd/avalauncher
go vet ./...

# Local run (needs postgres)
DB_USER=dba_avalauncher DB_PASSWORD=xxx DB_HOST=localhost DB_PORT=5433 ADMIN_KEY=dev ./avalauncher

# Docker
./.launch.sh
```

## Conventions

- Go module: `github.com/primal-host/avalauncher`
- HTTP framework: Echo v4
- Database: pgx v5 on infra-postgres, database `avalauncher`
- Container name: `crypto-avalauncher`
- Schema auto-bootstraps via `CREATE TABLE IF NOT EXISTS`
- Config uses env vars with `_FILE` suffix support for Docker secrets

## Database

Postgres on `infra-postgres:5432` (host port 5433), database `avalauncher`, user `dba_avalauncher`.

Tables: `hosts`, `nodes`, `l1s`, `l1_validators`, `events`.

## Docker

- Image/container: `crypto-avalauncher`
- Network: `infra`
- Port: 4321
- Traefik: `avalauncher.primal.host` / `avalauncher.localhost`
- DNS: `192.168.147.53` (infra CoreDNS)
