# 🔔 Notification Engine

> A robust, scalable backend service developed in Go, utilizing the **Fiber v3** framework for high-performance HTTP routing.

The Notification Engine is designed to handle user authentication, workspace management, and notification dispatching across multiple environments and channels.

---

## 🛠️ Technology Stack

| Component | Technology | Description |
|-----------|------------|-------------|
| **Language** | [Go (Golang)](https://golang.org/) | Core backend language |
| **Framework** | [Fiber v3](https://docs.gofiber.io/) | High-performance HTTP web framework |
| **Database** | PostgreSQL | Relational database (using `pgxpool`) |
| **Caching/Sessions** | Redis | In-memory data store |
| **ORM / Models** | [sqlc](https://sqlc.dev/) | Type-safe SQL code generation |
| **Authentication** | OAuth2 + JWT | Google OAuth (`goth`), Stateless JWT Sessions |
| **Logging** | `zerolog` | Structured, leveled JSON logging |

---

## ✨ Features & Recent Updates

- **🔐 Multi-Strategy Authentication:** Fully functional Local Email/Password registration/login alongside Google OAuth integration. Built with BCrypt hashing and mitigations against enumerative timing-attacks.
- **🛡️ Enhanced Security:**
  - **Payload Validation** using `go-playground/validator/v10` to enforce strict constraints (e.g., email format, min/max lengths).
  - **Rate Limiting** on authentication endpoints to automatically mitigate brute-force and DoS attempts.
  - **Role-Based Access Control (RBAC):** Middleware to enforce `owner` and `admin` scopes on critical application routes.
- **🔑 Secure API Key Management:**
  - Rigorous scoping tying API keys to specific Workspaces and Environments to prevent IDOR vulnerabilities.
  - Secure generation processes with unique prefixes (`ne_test_...`), hashed cryptographical storage, and sanitized key hints.
  - Enforceable API Key expiration rules and seamless revocation toggles.
- **🌐 Workspaces & Environments:** Supports isolated configurations per workspace (`development` / `production`) and seamlessly handles tiered subscription limits.
- **⚡ Standardized Error Handling:** Centralized internal error management (`pkg/apperrors`) handling Postgres unique-constraint violations gently across the domain.

---

## 📂 Project Architecture

This project follows a clean, highly decoupled architectural pattern, separating the core domain business logic from the HTTP transport layer and external infrastructure implementations.

```text
├── pkg/                  # Shared libraries (JWT, Redis, Postgres, Logger, Errors)
├── services/backend/
│   ├── config/           # Application configuration mapping (Viper)
│   ├── internal/
│   │   ├── auth/         # Authentication domain (Handlers & Services)
│   │   ├── user/         # User profile domain
│   │   ├── workspace/    # Workspace management
│   │   ├── api_keys/     # API Key generation, routing, and management
│   │   ├── session/      # Redis-backed token blacklist & version tracking
│   │   ├── middleware/   # Request interceptors (Auth, RBAC, Rate Limits)
│   │   ├── dtos/         # Data Transfer Objects
│   │   ├── db/           # Generated sqlc database code bindings
│   │   └── utils/        # Generic utility helpers
│   ├── app/              # Fiber App Initialization & API Router configuration
│   └── cmd/              # Application entry points
├── migrations/           # PostgreSQL schema scripts (up/down)
└── deployments/          # Docker Compose and infrastructure files
```

---

## 🚀 Getting Started

### Prerequisites

You will need the following installed:
- **Go** (1.21+)
- **Docker & Docker Compose** (for spinning up Postgres & Redis)
- **sqlc** (for modifying queries)
- **golang-migrate** (for running database migrations)

### 1. Environment Setup

Copy `.env.example` to `.env` in the root of the project to define your local configurations:

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

# JWT Secrets
ACCESS_SECRET=your_super_secret_access_key
REFRESH_SECRET=your_super_secret_refresh_key

# OAuth Config
CLIENT_ID=your_google_oauth_client_id
CLIENT_SECRET=your_google_oauth_client_secret
REDIRECT_URL=http://localhost:8080/api/v1/auth/google/callback
```

### 2. Boot Infrastructure

Start the PostgreSQL and Redis containers using Docker Compose:

```bash
cd deployments
docker-compose up -d
```

### 3. Apply Migrations

Run your database migrations to set up the default schema:

```bash
migrate -path migrations -database "postgres://postgres:secret@localhost:5432/notif_db?sslmode=disable" up
```

### 4. Run the Backend Service

Execute the main Go application from the `services/backend` directory:

```bash
cd services/backend
go run cmd/main.go cmd/cmd.go
```

The application will start the Fiber web server.
- **REST API Base URL:** `http://localhost:8080/api/v1`

---

## 📝 Scripts & Tasks

If using a Taskfile, you can easily execute standard workflows like:
- `task generate` - Regenerate `sqlc` models and queries.
- `task run` - Start the local development server.
