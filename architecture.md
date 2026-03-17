# Notification Engine Architecture

## Directory Structure

```text
pkg/                    ← shared across services
    jwt/                ← token generation and parsing
    cache/              ← Redis client
    database/           ← PostgreSQL client
    logger/             ← zerolog + lumberjack
    response/           ← HTTP response helpers
    apperrors/          ← sentinel errors

services/backend/
    config/             ← Viper config + OAuth setup
    internal/
        auth/           ← handler + service
        user/           ← handler + service
        workspace/      ← handler + service
        middleware/     ← JWT auth middleware
        session/        ← Redis session store
        db/             ← sqlc generated
        dtos/           ← request/response types
        utils/          ← UUID helpers, slugify
    app/                ← fiber setup, routing
    cmd/                ← entrypoint add this in the architecture.md file
```
