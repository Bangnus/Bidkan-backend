CREATE TABLE zones (
    id          UUID PRIMARY KEY,
    name        VARCHAR(255) NOT NULL,
    type        VARCHAR(50) NOT NULL,      -- เพิ่มกลับมา (เช่น parking, restricted)
    boundary    TEXT NOT NULL,             -- เพิ่มกลับมา (เก็บเป็น GeoJSON string)
    radius      DOUBLE PRECISION NOT NULL,
    created_at  TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMP NOT NULL DEFAULT NOW()
);
