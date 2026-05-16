CREATE TABLE transactions (
    id          UUID PRIMARY KEY,
    user_id     UUID NOT NULL REFERENCES users(id),
    amount      DECIMAL(10,2) NOT NULL,
    type        VARCHAR(20) NOT NULL, -- topup, transfer_in, transfer_out, ride_fare, refund
    status      VARCHAR(20) NOT NULL DEFAULT 'pending', -- เพิ่มกลับมา
    reference_id VARCHAR(100), -- เพิ่มกลับมา
    gateway_ref  VARCHAR(100), -- เพิ่มกลับมา
    description TEXT,
    created_at  TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMP NOT NULL DEFAULT NOW() -- เพิ่มกลับมา
);
