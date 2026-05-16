CREATE TABLE notifications (
    id          UUID PRIMARY KEY,
    user_id     UUID REFERENCES users(id), -- NULL = Broadcast
    title       VARCHAR(255) NOT NULL,
    message     TEXT NOT NULL,
    type        VARCHAR(50) NOT NULL,      -- promo, news, system, wallet
    is_read     BOOLEAN NOT NULL DEFAULT FALSE,
    created_at  TIMESTAMP NOT NULL DEFAULT NOW()
);
