-- name: GetSystemSummary :one
SELECT 
    (SELECT COUNT(*) FROM users WHERE role = 'user') as total_users,
    (SELECT COUNT(*) FROM bikes) as total_bikes,
    (SELECT COUNT(*) FROM bikes WHERE status = 'available') as available_bikes,
    (SELECT COUNT(*) FROM rides WHERE status = 'ongoing') as active_rides,
    (SELECT COALESCE(SUM(CAST(total_fare AS DECIMAL)), 0)::TEXT FROM rides WHERE status = 'completed') as total_revenue
FROM rides LIMIT 1;

-- name: GetBikeUsageStats :many
SELECT 
    bike_id, 
    COUNT(*) as ride_count,
    COALESCE(SUM(distance_km), 0)::FLOAT as total_distance_km,
    COALESCE(SUM(CAST(total_fare AS DECIMAL)), 0)::TEXT as total_revenue
FROM rides
WHERE status = 'completed'
GROUP BY bike_id
ORDER BY ride_count DESC;

-- name: GetDailyRevenue :many
SELECT 
    DATE(start_time) as date,
    COUNT(*) as ride_count,
    COALESCE(SUM(CAST(total_fare AS DECIMAL)), 0)::TEXT as daily_revenue
FROM rides
WHERE status = 'completed'
GROUP BY DATE(start_time)
ORDER BY date DESC
LIMIT 7;
