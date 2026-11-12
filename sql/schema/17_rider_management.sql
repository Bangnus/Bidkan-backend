-- 17_rider_management.sql
-- Extension for Campus Rider & Student Management and Penalty Dispute & Refund Desk

ALTER TABLE users 
ADD COLUMN IF NOT EXISTS is_student BOOLEAN NOT NULL DEFAULT FALSE,
ADD COLUMN IF NOT EXISTS student_id VARCHAR(50),
ADD COLUMN IF NOT EXISTS student_card_url VARCHAR(255),
ADD COLUMN IF NOT EXISTS student_verified BOOLEAN NOT NULL DEFAULT FALSE;

ALTER TABLE rides
ADD COLUMN IF NOT EXISTS parking_penalty NUMERIC(10,2) NOT NULL DEFAULT 0.00,
ADD COLUMN IF NOT EXISTS penalty_status VARCHAR(20) NOT NULL DEFAULT 'none',
ADD COLUMN IF NOT EXISTS dispute_reason TEXT,
ADD COLUMN IF NOT EXISTS refund_amount NUMERIC(10,2) NOT NULL DEFAULT 0.00,
ADD COLUMN IF NOT EXISTS refunded_at TIMESTAMP;

CREATE INDEX IF NOT EXISTS idx_users_student ON users (tenant_id, is_student, student_verified);
CREATE INDEX IF NOT EXISTS idx_rides_dispute ON rides (tenant_id, penalty_status);
