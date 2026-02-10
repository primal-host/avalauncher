# Avalauncher

Manage Avalanche blockchain nodes and L1s across multiple Docker hosts.

## Features

- Multi-host node management via SSH or local Docker
- L1 (subnet) deployment and validator assignment
- Declarative cluster configuration (`cluster.yaml`)
- Dark-theme web dashboard
- Audit event log
- Admin API with bearer token auth

## Quick Start

### Prerequisites

Create the database on infra-postgres:

```sql
CREATE ROLE dba_avalauncher WITH LOGIN PASSWORD 'your-password' CREATEDB;
CREATE DATABASE avalauncher OWNER dba_avalauncher;
```

### Local Development

```bash
go build -o avalauncher ./cmd/avalauncher

DB_USER=dba_avalauncher DB_PASSWORD=your-password \
  DB_HOST=localhost DB_PORT=5433 \
  ADMIN_KEY=dev \
  ./avalauncher
```

- Dashboard: http://localhost:4321/
- Health: http://localhost:4321/health
- Status API: `curl -H "Authorization: Bearer dev" http://localhost:4321/api/status`

### Docker

```bash
# Create secrets
echo -n "your-password" > ~/apps/infra/postgres/secrets/dba_avalauncher_password.txt
mkdir -p secrets
openssl rand -hex 32 > secrets/admin_key.txt

# Copy and edit config
cp .env.example .env

# Launch
./.launch.sh
```

Accessible via Traefik at `avalauncher.primal.host` or `avalauncher.localhost`.

## Configuration

### Environment Variables

| Variable | Default | Description |
|---|---|---|
| `DB_HOST` | `localhost` | PostgreSQL host |
| `DB_PORT` | `5432` | PostgreSQL port |
| `DB_NAME` | `avalauncher` | Database name |
| `DB_USER` | `dba_avalauncher` | Database user |
| `DB_PASSWORD` | | Database password |
| `DB_SSLMODE` | `disable` | SSL mode |
| `LISTEN_ADDR` | `:4321` | HTTP listen address |
| `ADMIN_KEY` | | Bearer token for API auth |

All sensitive variables support `_FILE` suffix for Docker secrets (e.g., `DB_PASSWORD_FILE=/run/secrets/db_password`).

### Cluster Config

Copy `cluster.yaml.example` to `cluster.yaml` and define your hosts, nodes, and L1s. See the example file for the full schema.
