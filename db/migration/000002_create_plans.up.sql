CREATE TABLE IF not exists plans (
    id                    UUID         PRIMARY KEY,
    name                  VARCHAR(255) NOT NULL,      
    notif_limit_month     INTEGER      NOT NULL DEFAULT 1000,  
    members_limit         INTEGER      NOT NULL DEFAULT 3,      
    api_keys_limit        INTEGER      NOT NULL DEFAULT 2,      
    log_retention_days    INTEGER      NOT NULL DEFAULT 7,    
    original_price_cents  INTEGER      NOT NULL DEFAULT 0,     
    price_cents           INTEGER      NOT NULL DEFAULT 0,     
    is_active             BOOLEAN      NOT NULL DEFAULT TRUE,
    created_at            TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at            TIMESTAMPTZ  NOT NULL DEFAULT NOW()
    
);