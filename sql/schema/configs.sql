CREATE TABLE configs (
    key VARCHAR(255) PRIMARY KEY,
    value VARCHAR(255) NOT NULL,
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- ค่าตั้งค่าเริ่มต้นของระบบ
INSERT INTO configs (key, value) VALUES 
('parking_penalty_fee', '50.00'),    -- ค่าปรับจอดนอกโซน P
('base_fare_rate', '2.00'),          -- อัตราค่าเช่ารถปกติ (บาท/นาที)
('rank_silver_threshold', '800.00'), -- เกณฑ์เงินสะสมสำหรับ Silver (4 เดือน)
('rank_gold_threshold', '2500.00'),  -- เกณฑ์เงินสะสมสำหรับ Gold (4 เดือน)
('rank_silver_discount', '0.50'),    -- ส่วนลดสำหรับ Silver (บาท/นาที)
('rank_gold_discount', '1.00');      -- ส่วนลดสำหรับ Gold (บาท/นาที)

