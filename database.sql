-- 1. Users
CREATE TABLE IF NOT EXISTS users (
    id             UUID PRIMARY KEY,
    phone_number   VARCHAR(20) UNIQUE NOT NULL,
    username       VARCHAR(255) NOT NULL,
    password       VARCHAR(255) NOT NULL,
    wallet_balance DECIMAL(10,2) NOT NULL DEFAULT 0.00,
    role           VARCHAR(10) NOT NULL DEFAULT 'user',
    status         VARCHAR(20) NOT NULL DEFAULT 'pending', 
    image_url      VARCHAR(255),
    fcm_token      TEXT,
    created_at     TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMP NOT NULL DEFAULT NOW()
);

-- 2. Bikes
CREATE TABLE IF NOT EXISTS bikes (
    id              VARCHAR(50) PRIMARY KEY,
    hardware_id     VARCHAR(100) UNIQUE NOT NULL,
    status          VARCHAR(20) NOT NULL DEFAULT 'available',
    lat             DOUBLE PRECISION NOT NULL DEFAULT 0,
    lon             DOUBLE PRECISION NOT NULL DEFAULT 0,
    battery_level   INT NOT NULL DEFAULT 100,
    image_url       VARCHAR(255),
    last_heartbeat  TIMESTAMP,
    current_ride_id UUID,
    created_at      TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMP NOT NULL DEFAULT NOW()
);

-- 2.1 Bike Locations (History)
CREATE TABLE IF NOT EXISTS bike_locations (
    id          UUID PRIMARY KEY,
    bike_id     VARCHAR(50) NOT NULL REFERENCES bikes(id),
    lat         DOUBLE PRECISION NOT NULL,
    lon         DOUBLE PRECISION NOT NULL,
    battery     INT NOT NULL DEFAULT 0,
    created_at  TIMESTAMP NOT NULL DEFAULT NOW()
);

-- 3. Zones
CREATE TABLE IF NOT EXISTS zones (
    id          UUID PRIMARY KEY,
    name        VARCHAR(255) NOT NULL,
    type        VARCHAR(50) NOT NULL,
    boundary    TEXT NOT NULL,
    radius      DOUBLE PRECISION NOT NULL,
    created_at  TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMP NOT NULL DEFAULT NOW()
);

-- 4. Configs
CREATE TABLE IF NOT EXISTS configs (
    key         VARCHAR(100) PRIMARY KEY,
    value       TEXT NOT NULL,
    description TEXT,
    updated_at  TIMESTAMP NOT NULL DEFAULT NOW()
);

-- 5. Coupons
CREATE TABLE IF NOT EXISTS coupons (
    id             UUID PRIMARY KEY,
    code           VARCHAR(50) UNIQUE NOT NULL,
    type           VARCHAR(20) NOT NULL,
    value          DECIMAL(10,2) NOT NULL,
    min_amount     DECIMAL(10,2) DEFAULT 0.00,
    start_at       TIMESTAMP NOT NULL DEFAULT NOW(),
    max_uses       INT NOT NULL DEFAULT 1,
    used_count     INT NOT NULL DEFAULT 0,
    expired_at     TIMESTAMP NOT NULL,
    created_at     TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMP NOT NULL DEFAULT NOW()
);

-- 6. Rides
CREATE TABLE IF NOT EXISTS rides (
    id              UUID PRIMARY KEY,
    user_id         UUID NOT NULL REFERENCES users(id),
    bike_id         VARCHAR(50) NOT NULL REFERENCES bikes(id),
    start_lat       DOUBLE PRECISION NOT NULL,
    start_lon       DOUBLE PRECISION NOT NULL,
    end_lat         DOUBLE PRECISION,
    end_lon         DOUBLE PRECISION,
    distance_km     DOUBLE PRECISION NOT NULL DEFAULT 0,
    total_fare      DECIMAL(10,2) NOT NULL DEFAULT 0,
    status          VARCHAR(20) NOT NULL DEFAULT 'ongoing',
    type            VARCHAR(20) NOT NULL DEFAULT 'normal',
    start_time      TIMESTAMP NOT NULL DEFAULT NOW(),
    end_time        TIMESTAMP,
    created_at      TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMP NOT NULL DEFAULT NOW()
);

-- 7. Transactions
CREATE TABLE IF NOT EXISTS transactions (
    id           UUID PRIMARY KEY,
    user_id      UUID NOT NULL REFERENCES users(id),
    amount       DECIMAL(10,2) NOT NULL,
    type         VARCHAR(20) NOT NULL,
    status       VARCHAR(20) NOT NULL DEFAULT 'pending',
    reference_id VARCHAR(100),
    gateway_ref  VARCHAR(100),
    description  TEXT,
    created_at   TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMP NOT NULL DEFAULT NOW()
);

-- 8. User Coupons
CREATE TABLE IF NOT EXISTS user_coupons (
    id          UUID PRIMARY KEY,
    user_id     UUID NOT NULL REFERENCES users(id),
    coupon_id   UUID NOT NULL REFERENCES coupons(id),
    is_used     BOOLEAN NOT NULL DEFAULT FALSE,
    used_at     TIMESTAMP,
    created_at  TIMESTAMP NOT NULL DEFAULT NOW()
);

-- 9. Notifications
CREATE TABLE IF NOT EXISTS notifications (
    id          UUID PRIMARY KEY,
    user_id     UUID REFERENCES users(id),
    title       VARCHAR(255) NOT NULL,
    message     TEXT NOT NULL,
    type        VARCHAR(50) NOT NULL,
    is_read     BOOLEAN NOT NULL DEFAULT FALSE,
    created_at  TIMESTAMP NOT NULL DEFAULT NOW()
);

-- 10. User Monthly Spending
CREATE TABLE IF NOT EXISTS user_monthly_spending (
    user_id        UUID NOT NULL REFERENCES users(id),
    year_month     INT NOT NULL,
    amount         DECIMAL(10,2) NOT NULL DEFAULT 0,
    updated_at     TIMESTAMP NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, year_month)
);

-- 11. Reports
CREATE TABLE IF NOT EXISTS reports (
    id          UUID PRIMARY KEY,
    bike_id     VARCHAR(50) NOT NULL REFERENCES bikes(id),
    reported_by UUID NOT NULL REFERENCES users(id),
    issue_type  VARCHAR(30) NOT NULL,
    status      VARCHAR(20) NOT NULL DEFAULT 'pending',
    resolved_by UUID REFERENCES users(id),
    created_at  TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMP NOT NULL DEFAULT NOW()
);
