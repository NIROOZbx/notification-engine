# 🔔 Notification Engine

> A production-grade, multi-tenant notification engine built in Go. Handles user, system, and scheduled notifications across multiple channels (Email, SMS) with a Kafka-backed async delivery pipeline, billing integration, and a strategy-based ingestion system.

---

## 🛠️ Technology Stack

| Component | Technology | Description |
|-----------|------------|-------------|
| **Language** | Go (Golang) | Core backend language |
| **Framework** | Fiber v3 | High-performance HTTP web framework |
| **API Gateway** | NGINX | Endpoint routing, Rate Limiting, Load Balancing |
| **Database** | PostgreSQL | Relational database (`pgxpool` + `sqlc`) |
| **Cache / Sessions** | Redis | In-memory store for sessions and data |
| **Message Queue** | Apache Kafka (KRaft) | Async notification delivery pipeline |
| **Billing** | gRPC (Billing Service) | Usage tracking, quota enforcement, Stripe |
| **Authentication** | OAuth2 + JWT | Google OAuth (`goth`), stateless JWT sessions |
| **Logging** | `zerolog` | Structured, leveled JSON logging |
| **Code Generation** | `sqlc` | Type-safe SQL code generation |

---

## ✨ Features

### 🔔 Notification Engine
- **Multi-Channel Delivery:** Email (SendGrid, AWS SES) and SMS (Twilio) with provider fallback.
- **Strategy Pattern Ingestion:** Clean separation between `normalStrategy` (user-triggered) and `systemStrategy` (billing/system-generated alerts).
- **System Notifications:** Billing and usage alerts bypass quota checks and opt-out rules, and are automatically routed to workspace owners.
- **Scheduled Notifications:** Future-dated notifications are stored as `scheduled` and picked up by a background scheduler.
- **Retry & DLQ:** Failed deliveries are retried with exponential backoff and terminated to a Dead Letter Queue after max attempts.
- **Idempotency:** All notifications are deduplicated via idempotency keys, preventing duplicate sends.
- **Template & Layout System:** Dynamic Handlebars-style template rendering with support for shared layouts.
- **Provider Override:** Templates can override the default workspace provider per-channel.

### 💳 Billing Integration (gRPC)
- **Quota Enforcement:** Per-channel send limits checked before every notification dispatch.
- **Usage Recording:** Tracks successful/failed sends per workspace for billing purposes.
- **Subscription Lifecycle:** Expiry reminders and usage threshold alerts published to Kafka and routed to workspace owners.

### 🔐 Security & Auth
- **API Gateway (NGINX):** Routes traffic and provides strict IP-based rate limiting to prevent abuse.
- **Multi-Strategy Auth:** Email/Password + Google OAuth with BCrypt hashing.
- **Rate Limiting:** Brute-force protection on auth endpoints.
- **RBAC Middleware:** Role-based access control (`owner` / `admin`) on critical routes.
- **API Key Management:** Scoped API keys per workspace/environment with expiry and revocation.
- **System Flag Protection:** `IsSystem` is internal-only — it cannot be spoofed via the public API.

---

## 📂 Project Structure

```text
notification-engine/
├── cmd/                        # Application entrypoint
├── config/                     # Viper config + YAML
├── consts/                     # Global constants (topics, statuses, etc.)
├── db/
│   ├── migration/              # PostgreSQL migrations (up/down)
│   ├── query/                  # Raw SQL queries (sqlc input)
│   └── sqlc/                   # sqlc-generated type-safe Go code
├── deployments/
│   ├── docker-compose.yml      # Postgres, Redis, Kafka, Backend
│   └── nginx.conf              # NGINX rate limiting and routing config
├── engine/
│   └── notification/
│       ├── core/               # Engine core (ingest, process, strategy)
│       │   ├── engine.go       # Main engine: Ingest, Process, ingestSystem, ingestNormal
│       │   ├── strategy.go     # Strategy pattern: normalStrategy, systemStrategy, ingestContext
│       │   ├── repository.go   # Repository + Producer + Renderer interfaces
│       │   └── types.go        # All DTOs and structs
│       ├── models/             # Kafka event models and trigger payloads
│       ├── provider/           # Provider interface + mock
│       ├── queue/              # Kafka producer + consumer + topic definitions
│       ├── scheduler/          # Background scheduler for future-dated notifications
│       ├── sender/
│       │   ├── email/          # SendGrid + SES providers
│       │   └── sms/            # Twilio provider
│       └── template/           # Go template renderer
├── internal/
│   ├── app/                    # Fiber app setup, routing, consumer/scheduler bootstrap
│   ├── billing/                # gRPC billing client (CheckLimit, RecordUsage)
│   ├── domain/                 # Core domain models
│   ├── handlers/               # HTTP + gRPC handlers
│   ├── middleware/             # Auth, API Key, RBAC middleware
│   ├── repositories/           # Repository implementations
│   ├── services/               # Business logic layer
│   ├── session/                # Redis session store
│   └── utils/                  # UUID helpers, locals, etc.
├── pkg/
│   ├── cache/                  # Redis client
│   ├── conversion/             # pgtype/JSON helpers
│   ├── encryptor/              # AES credential encryption
│   ├── httpclient/             # Shared HTTP client
│   ├── logger/                 # zerolog + lumberjack
│   ├── response/               # HTTP response helpers
│   └── validator/              # Request validation
└── proto/                      # Protobuf definitions + generated gRPC code
```

---

## 🚀 Getting Started

### Prerequisites

- **Go** 1.21+
- **Docker & Docker Compose**
- **`sqlc`** — `go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest`
- **`golang-migrate`** — `go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest`
- **Task** — `go install github.com/go-task/task/v3/cmd/task@latest`

### 1. Environment Setup

Copy `.env.example` to `.env` and fill in your values:

```env
# Database
DB_USER=postgres
DB_PASSWORD=secret
DB_HOST=localhost
DB_PORT=5432
DB_NAME=notif_db

# Redis
REDIS_PASSWORD=secret
REDIS_ADDR=localhost:6379

# Kafka
KAFKA_BROKER=localhost:9092

# JWT
ACCESS_SECRET=your_access_secret
REFRESH_SECRET=your_refresh_secret

# OAuth
CLIENT_ID=your_google_oauth_client_id
CLIENT_SECRET=your_google_oauth_client_secret
REDIRECT_URL=http://localhost:8080/api/v1/auth/google/callback

# Billing gRPC
GRPC_PORT=50051

# Encryption
SECRET_KEY=your_32_byte_aes_key
```

### 2. Start Infrastructure

```bash
task start
```

This will:
1. Start Postgres, Redis, Kafka, and the backend via Docker Compose
2. Run all pending database migrations
3. Tail the backend logs

### 3. Individual Task Commands

| Command | Description |
|---------|-------------|
| `task start` | Full boot: infra + migrate + logs |
| `task up` | Start containers (no rebuild) |
| `task build` | Rebuild and start containers |
| `task down` | Stop all containers |
| `task down-v` | Stop and wipe all volumes |
| `task migrate-up` | Apply all pending migrations |
| `task migrate-down` | Roll back migrations |
| `task migrate-create -- <name>` | Create a new migration |
| `task gen-sqlc` | Regenerate sqlc models |
| `task gen-proto` | Regenerate gRPC proto files |

---

## 🔄 Notification Flow

### User-triggered (API)
```
POST /api/v1/notifications/trigger
    → Engine.Ingest()
    → normalStrategy (billing check + opt-out check)
    → CreateNotificationLog (status: queued)
    → Publish to Kafka (email/sms topic)
    → Engine.Process() (consumer picks up)
    → Provider.Send() (SendGrid / Twilio)
    → RecordUsage (billing)
```

### System-triggered (Billing Service → Kafka)
```
Billing Service (cron/usage threshold)
    → Publish to Kafka (notifications.system topic)
    → Engine.Ingest() with IsSystem=true
    → ingestSystem() → GetWorkspaceOwners()
    → systemStrategy (skip billing + skip opt-out)
    → CreateNotificationLog per owner
    → Publish to Kafka (email topic)
    → Engine.Process() → Provider.Send()
```

---

## 📡 API Overview

| Method | Endpoint | Description |
|--------|----------|-------------|
| `POST` | `/api/v1/auth/register` | Register a new user |
| `POST` | `/api/v1/auth/login` | Login with email/password |
| `GET` | `/api/v1/auth/google` | Google OAuth login |
| `POST` | `/api/v1/notifications/trigger` | Trigger a notification |
| `GET` | `/api/v1/notifications/:id` | Get notification log |
| `POST` | `/api/v1/templates` | Create a template |
| `POST` | `/api/v1/subscribers` | Create a subscriber |
| `GET` | `/api/v1/subscribers/:id/preferences` | Get subscriber preferences |
| `POST` | `/api/v1/channel-configs` | Configure a channel provider |
| `POST` | `/api/v1/billing/checkout` | Create Stripe checkout session |

---

## 📝 License

MIT
