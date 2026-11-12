-- Incident reports & insurance claims per campus tenant
CREATE TABLE IF NOT EXISTS incident_reports (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    ride_id UUID REFERENCES rides(id) ON DELETE SET NULL,
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    reporter_name VARCHAR(100) NOT NULL,
    reporter_phone VARCHAR(50) NOT NULL,
    incident_type VARCHAR(50) NOT NULL, -- collision, fall, pedestrian_hit, property_damage, mechanical_failure
    severity VARCHAR(20) NOT NULL DEFAULT 'minor', -- minor, moderate, severe, critical
    description TEXT NOT NULL,
    location_name VARCHAR(150) DEFAULT '',
    location_lat NUMERIC(10,6) DEFAULT 0,
    location_lon NUMERIC(10,6) DEFAULT 0,
    police_report_no VARCHAR(100) DEFAULT '',
    insurance_claim_no VARCHAR(100) DEFAULT '',
    insurance_status VARCHAR(30) NOT NULL DEFAULT 'unfiled', -- unfiled, filed, under_review, approved, rejected, settled
    claim_amount NUMERIC(10,2) NOT NULL DEFAULT 0.00,
    approved_amount NUMERIC(10,2) NOT NULL DEFAULT 0.00,
    hospital_name VARCHAR(100) DEFAULT '',
    notes TEXT DEFAULT '',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    resolved_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX IF NOT EXISTS idx_incident_reports_tenant ON incident_reports(tenant_id);
CREATE INDEX IF NOT EXISTS idx_incident_reports_status ON incident_reports(insurance_status);
CREATE INDEX IF NOT EXISTS idx_incident_reports_created ON incident_reports(created_at DESC);
