CREATE TABLE if not exists users (
    id            UUID PRIMARY KEY ,
    email         VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) ,  
    full_name     VARCHAR(255) NOT NULL,
    auth_provider  VARCHAR(50) NOT NULL DEFAULT 'local',
    provider_id  VARCHAR(255),
    avatar_url VARCHAR(255),
    is_verified   BOOLEAN      NOT NULL DEFAULT FALSE, 
    is_active     BOOLEAN      NOT NULL DEFAULT TRUE,
    last_login_at TIMESTAMPTZ  NULL,                 
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX users_email_index ON users (email);
CREATE INDEX users_provider_index ON users (auth_provider, provider_id);