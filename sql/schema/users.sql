CREATE TABLE users (
    id             UUID PRIMARY KEY,
    phone_number   VARCHAR(20) UNIQUE NOT NULL,   -- เปลี่ยนจาก email เป็นเบอร์โทร
    full_name      VARCHAR(255) NOT NULL,          -- เปลี่ยนจาก name
    password       VARCHAR(255) NOT NULL,
    wallet_balance DECIMAL(10,2) NOT NULL DEFAULT 0.00,  -- ยอดเงินในกระเป๋า
    role           VARCHAR(10) NOT NULL DEFAULT 'user',   -- user, staff, admin
    status         VARCHAR(20) NOT NULL DEFAULT 'active', -- active, suspended
    created_at     TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMP NOT NULL DEFAULT NOW()
);
