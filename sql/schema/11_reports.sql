CREATE TABLE reports (
    id          UUID PRIMARY KEY,
    bike_id     VARCHAR(50) NOT NULL REFERENCES bikes(id),
    reported_by UUID NOT NULL REFERENCES users(id),
    issue_type  VARCHAR(30) NOT NULL,
    status      VARCHAR(20) NOT NULL DEFAULT 'pending',
    resolved_by UUID REFERENCES users(id),
    created_at  TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMP NOT NULL DEFAULT NOW()
);
