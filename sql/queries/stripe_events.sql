-- name: TryRecordStripeEvent :execrows
INSERT INTO stripe_webhook_events (event_id, event_type, processed_at)
VALUES ($1, $2, NOW())
ON CONFLICT (event_id) DO NOTHING;

-- name: HasStripeEvent :one
SELECT EXISTS (
    SELECT 1 FROM stripe_webhook_events WHERE event_id = $1
)::BOOLEAN;

