-- name: GetCouponByCode :one
SELECT * FROM coupons WHERE code = $1 LIMIT 1;

-- name: IncrementCouponUsedCount :exec
UPDATE coupons SET used_count = used_count + 1, updated_at = NOW() WHERE id = $1;

-- name: RecordUserCoupon :exec
INSERT INTO user_coupons (id, user_id, coupon_id, is_used, used_at)
VALUES ($1, $2, $3, $4, $5);

-- name: GetAvailableDiscountCoupon :one
SELECT c.*, uc.id as user_coupon_id 
FROM user_coupons uc
JOIN coupons c ON uc.coupon_id = c.id
WHERE uc.user_id = $1 AND uc.is_used = FALSE AND c.type = 'discount'
ORDER BY c.value DESC LIMIT 1;

-- name: MarkUserCouponAsUsed :exec
UPDATE user_coupons SET is_used = TRUE, used_at = NOW() WHERE id = $1;

-- name: GetSpecificUserCouponByCode :one
SELECT c.*, uc.id as user_coupon_id 
FROM user_coupons uc
JOIN coupons c ON uc.coupon_id = c.id
WHERE uc.user_id = $1 AND c.code = $2 AND uc.is_used = FALSE AND c.type = 'discount'
LIMIT 1;

-- name: CheckUserCouponUsed :one
SELECT EXISTS(
    SELECT 1 FROM user_coupons WHERE user_id = $1 AND coupon_id = $2
);

-- name: GetMyAvailableCoupons :many
SELECT c.*, uc.id as user_coupon_id, uc.used_at as collected_at
FROM user_coupons uc
JOIN coupons c ON uc.coupon_id = c.id
WHERE uc.user_id = $1 AND uc.is_used = FALSE AND c.expired_at > NOW()
ORDER BY uc.used_at DESC;

-- name: CreateCoupon :exec
INSERT INTO coupons (id, code, type, value, min_amount, start_at, max_uses, used_count, expired_at, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, 0, $8, NOW(), NOW());
