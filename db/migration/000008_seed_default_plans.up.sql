INSERT INTO plans (
    id, 
    name, 
    notif_limit_month, 
    members_limit, 
    api_keys_limit, 
    log_retention_days, 
    original_price_cents, 
    price_cents, 
    is_active
) VALUES 
    (gen_random_uuid(), 'Free', 1000, 3, 2, 7, 0, 0, true),
    
    (gen_random_uuid(), 'Pro', 50000, 10, 5, 30, 2900, 1500, true),
    
    (gen_random_uuid(), 'Enterprise', 1000000, 50, 20, 90, 9900, 9900, true)

ON CONFLICT (name) DO NOTHING;