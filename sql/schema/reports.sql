CREATE TABLE reports (
    id          UUID PRIMARY KEY,
    bike_id     VARCHAR(50) NOT NULL REFERENCES bikes(id),
    reported_by UUID NOT NULL REFERENCES users(id),
    issue_type  VARCHAR(30) NOT NULL,   -- flat_tire, battery_dead, system_error
    status      VARCHAR(20) NOT NULL DEFAULT 'pending', -- pending, fixing, resolved
    resolved_by UUID REFERENCES users(id),              -- NULL = ยังไม่มีคนซ่อม
    created_at  TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMP NOT NULL DEFAULT NOW()
);
