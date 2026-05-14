CREATE TABLE rides (
    id             UUID PRIMARY KEY,
    user_id        UUID NOT NULL REFERENCES users(id),
    bike_id        VARCHAR(50) NOT NULL REFERENCES bikes(id),
    start_time     TIMESTAMP NOT NULL DEFAULT NOW(),
    end_time       TIMESTAMP,                        -- NULL = ยังขี่อยู่
    start_lat      DOUBLE PRECISION NOT NULL,
    start_lon      DOUBLE PRECISION NOT NULL,
    end_lat        DOUBLE PRECISION,
    end_lon        DOUBLE PRECISION,
    distance_km    DOUBLE PRECISION DEFAULT 0,
    total_fare     DECIMAL(10,2) DEFAULT 0.00,
    status         VARCHAR(20) NOT NULL DEFAULT 'ongoing', -- ongoing, completed, cancelled
    created_at     TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMP NOT NULL DEFAULT NOW()
);
