CREATE TABLE user_coupons (
    id          UUID PRIMARY KEY,
    user_id     UUID NOT NULL REFERENCES users(id),
    coupon_id   UUID NOT NULL REFERENCES coupons(id),
    is_used     BOOLEAN NOT NULL DEFAULT FALSE,
    used_at     TIMESTAMP, -- ในระบบเดิมใช้คอลัมน์นี้เก็บทั้งเวลาที่เก็บและเวลาที่ใช้
    created_at  TIMESTAMP NOT NULL DEFAULT NOW()
);
