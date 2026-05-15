CREATE TABLE users (
    id             UUID PRIMARY KEY,
    phone_number   VARCHAR(20) UNIQUE NOT NULL,   -- เปลี่ยนจาก email เป็นเบอร์โทร
    username       VARCHAR(255) NOT NULL,
    password       VARCHAR(255) NOT NULL,
    wallet_balance DECIMAL(10,2) NOT NULL DEFAULT 0.00,
    role           VARCHAR(10) NOT NULL DEFAULT 'user',
    status         VARCHAR(20) NOT NULL DEFAULT 'pending', 
    image_url      VARCHAR(255),
    created_at     TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMP NOT NULL DEFAULT NOW()
);
