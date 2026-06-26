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
| **Cache / Sessions** | Redis / ElastiCache | In-memory store (local dev) / managed Redis (production) |
| **Message Queue** | Apache Kafka (KRaft) | Async notification delivery pipeline |
| **Container Registry** | Amazon ECR | Docker image storage for envoy-backend and billing-service |
| **Container Orchestration** | Amazon EKS (K8s 1.35) | Production Kubernetes cluster (`envoy-cluster`) |
| **Infrastructure as Code** | Terraform (AWS) | VPC, EKS, ECR, ElastiCache, IAM OIDC roles |
| **GitOps** | Argo CD | Declarative Kubernetes deployment (App of Apps pattern) |
| **Billing** | gRPC (Billing Service) | Usage tracking, quota enforcement, Stripe |
| **Authentication** | OAuth2 + JWT | Google OAuth (`goth`), stateless JWT sessions |
| **Monitoring** | Prometheus + Grafana Alloy | Metrics collection, service monitors, alerting rules |
| **Logging** | `zerolog` | Structured, leveled JSON logging |
| **Code Generation** | `sqlc` + `protoc` | Type-safe SQL and protobuf/gRPC code generation |
| **Load Testing** | K6 | Performance and stress testing |

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
├── .github/
│   └── workflows/
│       └── deploy.yml              # CI/CD: test, build ECR, GitOps writeback
├── cmd/                            # Application entrypoint
├── config/                         # Viper config + YAML
├── consts/                         # Global constants (topics, statuses, etc.)
├── db/
│   ├── migration/                  # PostgreSQL migrations (up/down)
│   ├── query/                      # Raw SQL queries (sqlc input)
│   └── sqlc/                       # sqlc-generated type-safe Go code
├── deployments/
│   ├── argocd/
│   │   ├── root-app.yaml           # ArgoCD App of Apps root
│   │   └── apps/                   # Child app manifests (envoy-backend, billing-service)
│   ├── docker-compose.yml          # Local dev: Postgres, Redis, Kafka, Backend
│   ├── helm/
│   │   ├── envoy-backend/          # Backend Helm chart (deployment, ingress, ESO, monitoring)
│   │   └── warpstream-infra/       # WarpStream agent Helm chart
│   ├── nginx.conf                  # NGINX API Gateway (rate limiting + routing)
│   └── terraform/
│       └── prod/                   # AWS IaC: VPC, EKS, ECR, ElastiCache, IAM OIDC, IRSA
├── docs/                           # Architecture & design documentation
├── engine/
│   └── notification/
│       ├── core/                   # Engine core (ingest, process, strategy)
│       │   ├── engine.go           # Ingest, Process, ingestSystem, ingestNormal
│       │   ├── strategy.go         # Strategy pattern: normalStrategy, systemStrategy
│       │   ├── repository.go       # Repository + Producer + Renderer interfaces
│       │   └── types.go            # All DTOs and structs
│       ├── models/                 # Kafka event models and trigger payloads
│       ├── provider/               # Provider interface + mock
│       ├── queue/                  # Kafka producer + consumer + topic definitions
│       ├── scheduler/              # Background scheduler for future-dated notifications
│       ├── sender/
│       │   ├── email/              # SendGrid + SES providers
│       │   └── sms/                # Twilio provider
│       └── template/               # Go template renderer
├── internal/
│   ├── app/                        # Fiber app setup, routing, consumer/scheduler bootstrap
│   ├── billing/                    # gRPC billing client (CheckLimit, RecordUsage)
│   ├── domain/                     # Core domain models
│   ├── handlers/                   # HTTP + gRPC handlers
│   ├── middleware/                  # Auth, API Key, RBAC middleware
│   ├── repositories/               # Repository implementations
│   ├── services/                   # Business logic layer
│   ├── session/                    # Redis session store
│   └── utils/                      # UUID helpers, locals, etc.
├── k6/                             # Load testing scripts
├── k8s-compiled/                   # Pre-rendered Kubernetes manifests
├── logs/                           # Runtime log output
├── pkg/
│   ├── cache/                      # Redis client
│   ├── conversion/                 # pgtype/JSON helpers
│   ├── encryptor/                  # AES-256 credential encryption
│   ├── httpclient/                 # Shared HTTP client
│   ├── logger/                     # zerolog + lumberjack
│   ├── response/                   # HTTP response helpers
│   └── validator/                  # Request validation
├── proto/                          # Protobuf definitions + generated gRPC stubs
├── sdk/                            # Go client SDK for external consumers
├── .env.example                    # Environment variable template
├── alloy-config.alloy              # Grafana Alloy pipeline config
├── alloy-values.yaml               # Alloy Helm values
├── deploy-infra.ps1                # Local infra bootstrap script
├── Dockerfile                      # Production multi-stage Docker image
├── go.mod / go.sum                 # Go module dependencies
├── prometheus.yml                  # Prometheus scrape configuration
├── skaffold.yaml                   # Skaffold dev workflow
└── Taskfile.yml                    # Task runner (start, migrate, gen, tf, k8s)
```

---

## 🚀 Getting Started

### Prerequisites

- **Go** 1.25.1+
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
| `task mock` | Regenerate mockery mocks |
| `task build-image` | Build production Docker image |
| `task tf-init` | Initialize Terraform |
| `task tf-plan` | Preview infrastructure changes |
| `task tf-apply` | Apply infrastructure changes |
| `task tf-destroy` | Destroy Terraform-managed resources |
| `task tf-validate` | Validate Terraform configuration |
| `task tf-fmt` | Format Terraform files |
| `task k8s-deploy` | Deploy K8s manifests |
| `task k8s-delete` | Remove K8s manifests |
| `task k8s-status` | Check pod/service/ingress status |
| `task k8s-logs-backend` | Tail envoy-backend pod logs |
| `task k8s-logs-billing` | Tail billing-service pod logs |

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

## 🌐 GitOps & EKS Deployment

The production environment is deployed on an Amazon EKS cluster (`envoy-cluster`, K8s 1.35) using a **GitOps** pipeline powered by **Argo CD** (App of Apps pattern) and **GitHub Actions** with keyless OIDC authentication.

### Infrastructure as Code (Terraform)

All AWS infrastructure is defined in `deployments/terraform/prod/`:

| Resource | File | Description |
|----------|------|-------------|
| **VPC** | `vpc.tf` | 3-tier VPC (public/private/intra subnets), NAT gateway, DNS |
| **EKS Cluster** | `eks.tf` | `envoy-cluster`, managed node groups (t3.small, 1-3 nodes), K8s 1.35 |
| **ECR Repositories** | `ecr.tf` | `envoy-backend` and `billing-service` image repos with scan-on-push |
| **ElastiCache** | `elasticache.tf` | Redis 7.1 cluster (`cache.m7g.large`), subnet group, security group |
| **GitHub OIDC Provider** | `irsa-github.tf` | `aws_iam_openid_connect_provider.github` for keyless GitHub Actions auth |
| **Deployer Roles** | `irsa-github.tf` | `github-actions-deploy-role` and `github-actions-billing-deploy-role` scoped by repo |
| **External Secrets Operator** | `irsa-eso.tf` | IRSA role for ESO to read Secrets Manager via `KubernetesSecretsReaderPolicy` |
| **WarpStream IRSA** | `irsa-warpstream.tf` | IRSA role + S3 access policy + Secrets Manager policy for WarpStream agents |

### CI/CD Pipeline (GitHub Actions)

On pushes to `main`, `.github/workflows/deploy.yml`:
1. Runs Go unit tests.
2. Assumes the AWS IAM Deployer role via keyless OIDC (no static secrets).
3. Builds and pushes the Docker image to Amazon ECR.
4. Overrides the image tag via git writeback to `deployments/helm/envoy-backend/values-production.yaml`.

### GitOps Sync (Argo CD)

- **App of Apps Pattern:** The cluster is bootstrapped with a root app (`deployments/argocd/root-app.yaml`) that monitors `deployments/argocd/apps/`.
- **Child Apps:** Automatically deploys and synchronizes:
  - `envoy-backend-app` (`deployments/helm/envoy-backend/` with `values-production.yaml`)
  - `billing-service-app` (billing service Helm chart)

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
