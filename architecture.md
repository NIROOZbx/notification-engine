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

config/                 ← Viper config + OAuth setup
internal/
    app/                ← fiber setup, routing
    auth/               ← handler + service
    user/               ← handler + service
    workspace/          ← handler + service
    middleware/         ← JWT auth middleware
    session/            ← Redis session store
    db/                 ← sqlc generated
    dtos/               ← request/response types
    utils/              ← UUID helpers, slugify
cmd/                    ← entrypoint
```
