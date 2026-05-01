# AGENTS.md — Notification Engine

## Developer Commands

| Task | Command |
|------|---------|
| Full boot (infra + migrate + logs) | `task start` |
| Start containers only | `task up` |
| Stop containers | `task down` |
| Stop + wipe volumes | `task down-v` |
| Apply migrations | `task migrate-up` |
| Rollback migrations | `task migrate-down` |
| Create migration | `task migrate-create -- <name>` |
| Regenerate sqlc models | `task gen-sqlc` (runs from `db/` dir) |
| Regenerate proto stubs | `task gen-proto` |
| Regenerate mocks | `task mock` |
| Run all tests | `go test ./...` |
| Run single test | `go test ./internal/services -run TestRemoveMember` |

Order for verification: `task gen-sqlc -> task gen-proto -> go test ./...`

## Architecture

Go backend using **Fiber v3** (HTTP) + **Kafka** (async delivery) + **PostgreSQL** (`pgxpool` + `sqlc`) + **Redis** (cache/sessions).

### Entry points
- `cmd/` — main binary entrypoint
- `internal/app/` — Fiber app setup, DI, route registration, consumer/scheduler bootstrap

### Key boundaries
- `engine/notification/core/` — ingestion engine with Strategy pattern (`normalStrategy` for user-triggered, `systemStrategy` for billing/system alerts)
- `engine/notification/queue/` — Kafka producer/consumer and topic definitions
- `engine/notification/scheduler/` — background worker for future-dated notifications
- `engine/notification/sender/` — provider implementations (SendGrid, SES, Twilio)
- `internal/billing/` — gRPC client to billing service (`CheckLimit`, `RecordUsage`)
- `internal/middleware/` — auth, API key, RBAC middleware
- `db/query/` — raw SQL files (sqlc input). Regenerate with `task gen-sqlc`
- `db/sqlc/` — generated type-safe Go code (DO NOT EDIT)
- `proto/` — protobuf definitions + generated gRPC stubs (DO NOT EDIT)

### Kafka topics
| Topic | Purpose |
|-------|---------|
| `notifications.email` | Queued email deliveries |
| `notifications.sms` | Queued SMS deliveries |
| `notifications.system` | System/billing alerts from billing service |
| `notifications.dlq` | Dead-letter queue for failed sends |

### Notification flow
1. `POST /api/v1/notifications/trigger` → `Engine.Ingest()` → strategy (billing check + opt-out) → Kafka
2. Kafka consumer → `Engine.Process()` → resolve provider → render template → send → `RecordUsage`

System notifications bypass billing and opt-out checks, route to workspace owners automatically.

## Framework Quirks

- **sqlc**: Generated code lives in `db/sqlc/`. Always run `task gen-sqlc` after changing `db/query/*.sql` or `db/migration/*.sql`. Config at `db/sqlc.yaml`.
- **migrations**: Uses `golang-migrate` CLI. Migrations in `db/migration/` with up/down SQL pairs.
- **protos**: Generated stubs in `proto/`. Run `task gen-proto` after editing `.proto` files.
- **mocks**: Generated with `mockery`. Run `task mock` to regenerate workspace mocks.
- **credentials**: Provider credentials are AES-256 encrypted at rest. Decrypted at send time.
- **IsSystem flag**: Internal-only, cannot be set via public API.
- **Docker exposes Postgres on port 5433** (not 5432) for local access.

## Testing

- Uses `testify` (assert + mock). Mocks generated via `mockery` in `internal/repositories/mocks/`.
- Only one test file exists currently: `internal/services/workspace_test.go`.
- Tests use table-driven pattern with `t.Run`.
- No CI workflows configured.

## Env

`.env` required (see README for keys). Docker compose loads via `env_file`. Local dev uses `.env` values directly; Docker overrides `DB_HOST` to `database`.
