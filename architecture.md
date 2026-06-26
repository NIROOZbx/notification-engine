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
├── .github/
│   └── workflows/
│       └── deploy.yml              # CI/CD: test, build ECR, GitOps writeback
├── cmd/                            # Main entrypoint
├── config/                         # Viper config (YAML + env)
├── consts/                         # Shared constants (topics, statuses, roles)
├── db/
│   ├── migration/                  # SQL migrations (up/down pairs)
│   ├── query/                      # Raw SQL files (sqlc input)
│   └── sqlc/                       # Type-safe generated Go DB code
├── deployments/
│   ├── argocd/
│   │   ├── root-app.yaml           # ArgoCD App of Apps root
│   │   └── apps/                   # Child app manifests
│   ├── docker-compose.yml          # Local dev: Postgres, Redis, Kafka, Backend
│   ├── helm/
│   │   ├── envoy-backend/          # Backend Helm chart (deployment, ingress, ESO, monitoring)
│   │   └── warpstream-infra/       # WarpStream agent Helm chart
│   ├── nginx.conf                  # NGINX API Gateway and Rate Limiting
│   └── terraform/
│       └── prod/                   # AWS IaC: VPC, EKS, ECR, ElastiCache, IAM OIDC, IRSA
├── docs/                           # Architecture & design documentation
├── engine/
│   └── notification/
│       ├── core/
│       │   ├── engine.go           # Ingest, ingestSystem, ingestNormal, Process
│       │   ├── strategy.go         # Strategy interface + normalStrategy + systemStrategy
│       │   ├── repository.go       # Repository, Producer, Renderer interfaces
│       │   └── types.go            # All shared DTOs
│       ├── models/                 # Kafka event models + TriggerPayload
│       ├── provider/               # Provider interface
│       ├── queue/                  # Kafka producer, consumer, topic map
│       ├── scheduler/              # Background worker for scheduled notifications
│       └── sender/
│           ├── email/              # SendGrid + SES implementations
│           └── sms/                # Twilio implementation
├── internal/
│   ├── app/                        # App bootstrap: DI, routing, consumers, scheduler
│   ├── billing/                    # gRPC billing client (CheckLimit, RecordUsage)
│   ├── domain/                     # Core domain models
│   ├── handlers/                   # HTTP handlers (Auth, Notifications, Billing, etc.)
│   ├── middleware/                  # Auth middleware, API key middleware, RBAC
│   ├── repositories/               # Repository implementations (sqlc-backed)
│   ├── services/                   # Business logic layer
│   ├── session/                    # Redis-backed session store
│   └── utils/                      # Locals, UUID helpers
├── k6/                             # Load testing scripts
├── k8s-compiled/                   # Pre-rendered Kubernetes manifests
├── logs/                           # Runtime log output
├── pkg/
│   ├── cache/                      # Redis client
│   ├── conversion/                 # pgtype ↔ Go type helpers
│   ├── encryptor/                  # AES-256 for provider credential encryption
│   ├── httpclient/                 # Shared HTTP client (for providers)
│   ├── logger/                     # zerolog + file rotation
│   ├── response/                   # Standardized HTTP response helpers
│   └── validator/                  # go-playground validator setup
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

## Kafka Topics

| Topic | Producer | Consumer | Purpose |
|-------|----------|----------|---------|
| `notifications.email` | Engine (Ingest) | Engine (Process) | Queued email deliveries |
| `notifications.sms` | Engine (Ingest) | Engine (Process) | Queued SMS deliveries |
| `notifications.system` | Billing Service | Engine (Ingest) | System-level billing alerts |
| `notifications.dlq` | Engine (Process) | Engine (ProcessDLQ) | Dead-letter queue for failed sends |

---

## AWS Infrastructure

All production infrastructure is provisioned via Terraform (`deployments/terraform/prod/`).

### Network (VPC)
A 3-tier VPC (`10.0.0.0/16`) across two AZs (`ap-south-1a/b`) with public subnets (load balancers), private subnets (EKS nodes), and intra subnets (ElastiCache, RDS). A single NAT gateway enables outbound connectivity for private resources.

### Compute (EKS)
An Amazon EKS cluster (`envoy-cluster`, K8s 1.35) with a managed node group (`t3.small`, 1-3 nodes, ON_DEMAND). Cluster addons include VPC CNI (prefix delegation), CoreDNS, and kube-proxy.

### Storage & Cache
- **ECR:** Two repositories (`envoy-backend`, `billing-service`) with image scanning on push.
- **ElastiCache:** Redis 7.1 (`cache.m7g.large`) in a dedicated subnet group, secured to accept traffic only from the EKS node security group.

### IAM & Security
- **GitHub OIDC Provider:** `aws_iam_openid_connect_provider.github` enables keyless GitHub Actions authentication.
- **Deployer Roles:** `github-actions-deploy-role` and `github-actions-billing-deploy-role`, each scoped to its repository via `token.actions.githubusercontent.com:sub`.
- **IRSA Roles:** `prod-external-secrets-operator-irsa-role` (ESO → Secrets Manager) and `prod-warpstream-irsa-role` (WarpStream → S3 + Secrets Manager).

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

### 7. Keyless OIDC Authentication
Rather than storing persistent AWS credentials inside GitHub Secrets, IAM OIDC federation is configured via a Terraform-managed `aws_iam_openid_connect_provider.github` resource. GitHub Actions workflows assume scoped roles (`github-actions-deploy-role`, `github-actions-billing-deploy-role`) restricted by repo subject claims (`repo:NIROOZbx/notification-engine:*` and `repo:NIROOZbx/billing-service:*`).

### 8. Declarative GitOps (App of Apps)
Image building is decoupled from cluster management. GitHub Actions CI/CD builds and pushes images to ECR, then writes the new image tag back to Helm values. Argo CD running in-cluster monitors the repository declarations via the App of Apps pattern: the `root-app` detects the changes and synchronizes child Helm charts (`envoy-backend`, `billing-service`) without exposing the EKS API to the GitHub runner.

### 9. IRSA (IAM Roles for Service Accounts)
Each Kubernetes service account that needs AWS access is mapped to a dedicated IAM role via OIDC federation from the EKS cluster's built-in OIDC provider:
- **External Secrets Operator** → `prod-external-secrets-operator-irsa-role` (reads Secrets Manager)
- **WarpStream Agents** → `prod-warpstream-irsa-role` (reads/writes S3, reads Secrets Manager)

### 10. Observability
The platform exposes Prometheus metrics via ServiceMonitor resources. Grafana Alloy collects and routes telemetry data, with pre-configured alerting rules for critical conditions (deployment failures, high error rates).
