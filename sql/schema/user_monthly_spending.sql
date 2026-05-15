CREATE TABLE user_monthly_spending (
    user_id     UUID NOT NULL REFERENCES users(id),
    year_month  INT NOT NULL, -- เก็บในรูปแบบ YYYYMM เช่น 202405
    amount      DECIMAL(10,2) NOT NULL DEFAULT 0.00,
    updated_at  TIMESTAMP NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, year_month)
);
