-- ตารางสำหรับเก็บสถานะปัจจุบันของรถแต่ละคัน (Latest State)
CREATE TABLE bikes (
    id               VARCHAR(50) PRIMARY KEY, -- เช่น BK-001
    hardware_id      VARCHAR(100) UNIQUE NOT NULL, -- รหัสบอร์ด หรือ MAC Address
    lat              DOUBLE PRECISION NOT NULL DEFAULT 0,
    lon              DOUBLE PRECISION NOT NULL DEFAULT 0,
    battery_level    INT NOT NULL DEFAULT 0,
    status           VARCHAR(20) NOT NULL DEFAULT 'available', -- available, in_use, maintenance, low_battery
    current_ride_id  UUID NULL, -- ID การเช่าปัจจุบัน
    last_heartbeat   TIMESTAMP NOT NULL DEFAULT NOW(),
    created_at       TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMP NOT NULL DEFAULT NOW()
);

-- ตารางสำหรับเก็บประวัติพิกัด (History Tracking)
CREATE TABLE bike_locations (
    id         UUID PRIMARY KEY,
    bike_id    VARCHAR(50) NOT NULL REFERENCES bikes(id),
    lat        DOUBLE PRECISION NOT NULL,
    lon        DOUBLE PRECISION NOT NULL,
    battery    INT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_bike_loc_history ON bike_locations(bike_id, created_at DESC);
