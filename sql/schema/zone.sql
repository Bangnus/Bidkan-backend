CREATE TABLE zones (
    id UUID PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    -- "ลานจอด P หน้าหอพักชาย"
    type VARCHAR(30) NOT NULL,
    -- parking_zone, riding_limit_zone
    boundary JSONB,
    -- เก็บพิกัด Polygon เป็น GeoJSON
    radius DOUBLE PRECISION DEFAULT 0,
    -- รัศมี (ถ้าเป็นวงกลม)
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);