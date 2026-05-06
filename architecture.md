# Notification Engine — Architecture

## Overview

The Notification Engine is a multi-tenant, Kafka-backed notification platform. It handles both user-triggered and system-generated notifications across multiple channels using a Strategy Pattern ingestion pipeline.

---

## High-Level Architecture

```
          ┌─────────────────────┐
          │ NGINX (API Gateway) │
          │(Rate Limiting/Routes)│
          └────────┬────────────┘
                   │
                   ▼
┌─────────────────────┐        ┌──────────────────────┐
│   Frontend / API    │        │    Billing Service   │
│  (REST via Fiber)   │        │  (gRPC + Kafka)      │
└────────┬────────────┘        └──────────┬───────────┘
         │ POST /trigger                  │ Kafka: notifications.system
         ▼                                ▼
┌────────────────────────────────────────────────────┐
│                   Engine.Ingest()                  │
│                                                    │
│  payload.IsSystem?                                 │
│  ┌── YES ──► ingestSystem()                        │
│  │           └─ GetWorkspaceOwners()               │
│  │           └─ systemStrategy (skip billing/optout│
│  │           └─ ingestNormal() per owner           │
│  │                                                 │
│  └── NO  ──► ingestNormal()                        │
│              └─ normalStrategy                     │
│              └─ CheckLimit (billing gRPC)          │
│              └─ GetContact (subscriber DB)         │
│              └─ opt-out check                      │
│              └─ CreateNotificationLog              │
│              └─ Publish to Kafka                   │
└────────────────────────────────────────────────────┘
         │
         ▼ Kafka Topics
┌─────────────────────┐   ┌──────────────────┐
│  notifications.email│   │  notifications.  │
│  (consumer)         │   │  sms (consumer)  │
└────────┬────────────┘   └──────┬───────────┘
         └──────────┬────────────┘
                    ▼
         ┌─────────────────────┐
         │   Engine.Process()  │
         │  - Fetch log from DB│
         │  - Resolve provider │
         │  - Render template  │
         │  - Send via provider│
         │  - RecordUsage      │
         │  - Update log status│
         └─────────────────────┘
```

---

## Strategy Pattern

The ingestion pipeline uses a `Strategy` interface to cleanly separate behaviors:

```
Strategy interface
├── SkipBillingCheck() bool
├── SkipOptOut() bool
└── ResolveContact() (*Contact, error)

Implementations:
├── normalStrategy   → full pipeline (billing + opt-out + subscriber lookup)
└── systemStrategy   → bypasses billing + opt-out, uses pre-resolved owner email
```

---

## Directory Structure

```text
notification-engine/
├── cmd/                        # Main entrypoint
├── config/                     # Viper config (YAML + env)
├── consts/                     # Shared constants (topics, statuses, roles)
├── db/
│   ├── migration/              # SQL migrations (up/down pairs)
│   ├── query/                  # Raw SQL files (sqlc input)
│   └── sqlc/                   # Type-safe generated Go DB code
├── deployments/
│   ├── docker-compose.yml      # Full stack: Postgres, Redis, Kafka, Backend
│   └── nginx.conf              # NGINX API Gateway and Rate Limiting config
├── engine/
│   └── notification/
│       ├── core/
│       │   ├── engine.go       # Ingest, ingestSystem, ingestNormal, Process
│       │   ├── strategy.go     # Strategy interface + normalStrategy + systemStrategy + ingestContext
│       │   ├── repository.go   # Repository, Producer, Renderer interfaces
│       │   └── types.go        # All shared DTOs (Contact, Template, NotificationLog, etc.)
│       ├── models/             # Kafka event models + TriggerPayload
│       ├── provider/           # Provider interface
│       ├── queue/              # Kafka producer, consumer, topic map
│       ├── scheduler/          # Background worker for scheduled notifications
│       └── sender/
│           ├── email/          # SendGrid + SES implementations
│           └── sms/            # Twilio implementation
├── internal/
│   ├── app/                    # App bootstrap: DI, routing, consumers, scheduler
│   ├── billing/                # gRPC billing client (CheckLimit, RecordUsage)
│   ├── domain/                 # Core domain models
│   ├── handlers/               # HTTP handlers (Auth, Notifications, Billing, etc.)
│   ├── middleware/             # Auth middleware, API key middleware, RBAC
│   ├── repositories/           # Repository implementations (sqlc-backed)
│   ├── services/               # Business logic layer
│   ├── session/                # Redis-backed session store
│   └── utils/                  # Locals, UUID helpers
├── pkg/
│   ├── cache/                  # Redis client
│   ├── conversion/             # pgtype ↔ Go type helpers
│   ├── encryptor/              # AES-256 for provider credential encryption
│   ├── httpclient/             # Shared HTTP client (for providers)
│   ├── logger/                 # zerolog + file rotation
│   ├── response/               # Standardized HTTP response helpers
│   └── validator/              # go-playground validator setup
└── proto/                      # Protobuf definitions + generated gRPC stubs
```

---

## Kafka Topics

| Topic | Producer | Consumer | Purpose |
|-------|----------|----------|---------|
| `notifications.email` | Engine (Ingest) | Engine (Process) | Queued email deliveries |
| `notifications.sms` | Engine (Ingest) | Engine (Process) | Queued SMS deliveries |
| `notifications.system` | Billing Service | Engine (Ingest) | System-level billing alerts |
| `notifications.dlq` | Engine (Process) | Engine (ProcessDLQ) | Dead-letter queue for failed sends |

---

## Key Design Decisions

### 1. Strategy Pattern
Instead of `if IsSystem { ... }` scattered throughout the engine, all behavioral differences are encapsulated in the `Strategy` interface. This makes adding new notification types (e.g., a "marketing" strategy with its own rules) trivial.

### 2. System Notification Isolation
System alerts (e.g., subscription expiry, usage limits) are published from the Billing Service to `notifications.system`. The engine resolves workspace owners from the DB and routes alerts to them without any user-facing API involvement, preventing privilege escalation.

### 3. API Gateway & Security
NGINX acts as the entry point to the platform. It handles API endpoint routing (e.g., separating `/api/v1/billing` from `/api/v1/`) and applies strict rate limits (e.g., 10 req/s with bursts) to protect the underlying Go fiber services from being overwhelmed.

### 4. Idempotency
Every notification channel gets a unique idempotency key (`{payload_key}:{channel}`). For system notifications with multiple owners, the key is further scoped per owner (`{payload_key}:{owner_email}`) to prevent collision.

### 5. Provider Resolution
Provider resolution uses a priority chain:
1. Template-level `OverrideProviderID`
2. Workspace default provider for the channel
3. Mock provider (test mode)

### 6. Credential Encryption
Provider credentials (API keys, auth tokens) are encrypted at rest using AES-256. Decryption happens only at send time inside the engine.
