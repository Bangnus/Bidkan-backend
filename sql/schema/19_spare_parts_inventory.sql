-- Spare parts inventory & battery ledger per campus tenant
CREATE TABLE IF NOT EXISTS campus_spare_parts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    part_name VARCHAR(100) NOT NULL,
    part_code VARCHAR(50) NOT NULL,
    category VARCHAR(50) NOT NULL, -- battery, brake, tire, electrical, body, tool
    quantity INT NOT NULL DEFAULT 0,
    unit VARCHAR(20) NOT NULL DEFAULT 'ชิ้น',
    min_stock INT NOT NULL DEFAULT 5,
    cost_per_unit NUMERIC(10,2) NOT NULL DEFAULT 0.00,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_tenant_part_code UNIQUE (tenant_id, part_code)
);

CREATE INDEX IF NOT EXISTS idx_campus_spare_parts_tenant ON campus_spare_parts(tenant_id);
CREATE INDEX IF NOT EXISTS idx_campus_spare_parts_category ON campus_spare_parts(category);

-- Log of stock in/out/maintenance usages
CREATE TABLE IF NOT EXISTS parts_usage_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    part_id UUID NOT NULL REFERENCES campus_spare_parts(id) ON DELETE CASCADE,
    change_type VARCHAR(20) NOT NULL, -- in, out, maintenance, damaged
    quantity INT NOT NULL,
    balance_after INT NOT NULL,
    reference_id VARCHAR(100) DEFAULT '',
    notes TEXT DEFAULT '',
    actor_name VARCHAR(100) NOT NULL DEFAULT 'System',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_parts_usage_logs_tenant ON parts_usage_logs(tenant_id);
CREATE INDEX IF NOT EXISTS idx_parts_usage_logs_part ON parts_usage_logs(part_id);
CREATE INDEX IF NOT EXISTS idx_parts_usage_logs_created ON parts_usage_logs(created_at DESC);
