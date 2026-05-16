CREATE TABLE rides (
    id              UUID PRIMARY KEY,
    user_id         UUID NOT NULL REFERENCES users(id),
    bike_id         VARCHAR(50) NOT NULL REFERENCES bikes(id),
    start_lat       DOUBLE PRECISION NOT NULL,
    start_lon       DOUBLE PRECISION NOT NULL,
    end_lat         DOUBLE PRECISION,
    end_lon         DOUBLE PRECISION,
    distance_km     DOUBLE PRECISION NOT NULL DEFAULT 0, -- แก้จาก distance เป็น distance_km
    total_fare      DECIMAL(10,2) NOT NULL DEFAULT 0,
    status          VARCHAR(20) NOT NULL DEFAULT 'ongoing',
    type            VARCHAR(20) NOT NULL DEFAULT 'normal', -- เพิ่มกลับมา
    start_time      TIMESTAMP NOT NULL DEFAULT NOW(),
    end_time        TIMESTAMP,
    created_at      TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMP NOT NULL DEFAULT NOW()
);
