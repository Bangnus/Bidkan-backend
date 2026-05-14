CREATE TABLE transactions (
    id           UUID PRIMARY KEY,
    user_id      UUID NOT NULL REFERENCES users(id),
    amount       DECIMAL(10,2) NOT NULL,  -- บวก=เติมเงิน, ลบ=จ่ายค่าเช่า
    type         VARCHAR(30) NOT NULL,     -- topup, fare_deduction, refund
    reference_id UUID,                     -- อ้างอิงไปหา rides.id
    created_at   TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMP NOT NULL DEFAULT NOW()
);
