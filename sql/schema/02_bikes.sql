CREATE TABLE bikes (
    id              VARCHAR(50) PRIMARY KEY,
    hardware_id     VARCHAR(100) UNIQUE NOT NULL,
    status          VARCHAR(20) NOT NULL DEFAULT 'available',
    lat             DOUBLE PRECISION NOT NULL DEFAULT 0,
    lon             DOUBLE PRECISION NOT NULL DEFAULT 0,
    battery_level   INT NOT NULL DEFAULT 100,
    image_url       VARCHAR(255),                 -- เพิ่มกลับมา
    last_heartbeat  TIMESTAMP,                    -- เพิ่มกลับมา
    current_ride_id UUID,
    created_at      TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE bike_locations (
    id          UUID PRIMARY KEY,
    bike_id     VARCHAR(50) NOT NULL REFERENCES bikes(id),
    lat         DOUBLE PRECISION NOT NULL,
    lon         DOUBLE PRECISION NOT NULL,
    battery     INT NOT NULL DEFAULT 0,           -- แก้ชื่อคอลัมน์ให้ตรง Query
    created_at  TIMESTAMP NOT NULL DEFAULT NOW()
);
