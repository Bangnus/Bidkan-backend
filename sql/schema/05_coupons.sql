CREATE TABLE coupons (
    id             UUID PRIMARY KEY,
    code           VARCHAR(50) UNIQUE NOT NULL,
    type           VARCHAR(20) NOT NULL,          -- 'credit' (เงินฟรี), 'discount' (ส่วนลด)
    value          DECIMAL(10,2) NOT NULL,
    min_amount     DECIMAL(10,2) DEFAULT 0.00,
    start_at       TIMESTAMP NOT NULL DEFAULT NOW(),
    max_uses       INT NOT NULL DEFAULT 1,
    used_count     INT NOT NULL DEFAULT 0,
    expired_at     TIMESTAMP NOT NULL,
    created_at     TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMP NOT NULL DEFAULT NOW()
);
