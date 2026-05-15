CREATE TABLE transactions (
    id           UUID PRIMARY KEY,
    user_id      UUID NOT NULL REFERENCES users(id),
    amount       DECIMAL(10,2) NOT NULL,  -- บวก=เติมเงิน, ลบ=จ่ายค่าเช่า
    type         VARCHAR(30) NOT NULL,     -- topup, fare_deduction, refund
    status       VARCHAR(20) NOT NULL DEFAULT 'completed', -- pending, completed, failed
    reference_id UUID,                     -- อ้างอิงไปหา rides.id
    gateway_ref  VARCHAR(255),             -- เก็บเลขอ้างอิงจากระบบภายนอก (PaySolutions)
    created_at   TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMP NOT NULL DEFAULT NOW()
);

-- สร้าง Index เพื่อให้การคำนวณ Rank ย้อนหลัง 4 เดือนทำได้รวดเร็ว
CREATE INDEX idx_transactions_user_ranking ON transactions (user_id, type, status, created_at);

