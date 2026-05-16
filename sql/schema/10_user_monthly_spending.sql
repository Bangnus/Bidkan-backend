CREATE TABLE user_monthly_spending (
    user_id        UUID NOT NULL REFERENCES users(id),
    year_month     INT NOT NULL, -- YYYYMM
    amount         DECIMAL(10,2) NOT NULL DEFAULT 0, -- แก้จาก total_spending เป็น amount
    updated_at     TIMESTAMP NOT NULL DEFAULT NOW(),  -- เพิ่มกลับมา
    PRIMARY KEY (user_id, year_month)
);
