-- 16_staff_compensation.sql
-- Staff compensation configurations and task earnings for campus field operations

CREATE TABLE IF NOT EXISTS staff_compensation_configs (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id           UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE UNIQUE,
    pay_model           VARCHAR(20) NOT NULL DEFAULT 'per_task', -- 'per_task', 'daily', 'monthly'
    battery_swap_fee    DECIMAL(10,2) NOT NULL DEFAULT 20.00,
    repair_task_fee     DECIMAL(10,2) NOT NULL DEFAULT 50.00,
    daily_base_wage     DECIMAL(10,2) NOT NULL DEFAULT 400.00,
    monthly_base_wage   DECIMAL(10,2) NOT NULL DEFAULT 10000.00,
    created_at          TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS staff_task_logs (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    staff_user_id   UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    task_type       VARCHAR(30) NOT NULL, -- 'battery_swap', 'maintenance_repair'
    target_id       VARCHAR(100) NOT NULL, -- bike_id หรือ report_id
    earned_amount   DECIMAL(10,2) NOT NULL DEFAULT 0.00,
    created_at      TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_staff_tasks_tenant_staff ON staff_task_logs(tenant_id, staff_user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_staff_tasks_created_at ON staff_task_logs(created_at DESC);

