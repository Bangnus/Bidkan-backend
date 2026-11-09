-- ============================================================================
-- BidKan Performance Indexes
-- ============================================================================

-- 1. Users Indexes
CREATE INDEX IF NOT EXISTS idx_users_tenant_id ON users (tenant_id);
CREATE INDEX IF NOT EXISTS idx_users_tenant_role ON users (tenant_id, role);

-- 2. Bikes & IoT Tracking Indexes
CREATE INDEX IF NOT EXISTS idx_bikes_status_battery ON bikes (status, battery_level);
CREATE INDEX IF NOT EXISTS idx_bikes_tenant_status ON bikes (tenant_id, status);
CREATE INDEX IF NOT EXISTS idx_bike_locations_bike_created ON bike_locations (bike_id, created_at DESC);

-- 3. Zones Indexes
CREATE INDEX IF NOT EXISTS idx_zones_tenant_type ON zones (tenant_id, type);

-- 4. Coupons & User Coupons Indexes
CREATE INDEX IF NOT EXISTS idx_coupons_tenant_expired ON coupons (tenant_id, expired_at);
CREATE INDEX IF NOT EXISTS idx_user_coupons_user_unused ON user_coupons (user_id, is_used);
CREATE INDEX IF NOT EXISTS idx_user_coupons_user_coupon ON user_coupons (user_id, coupon_id);

-- 5. Rides Indexes (Critical for real-time ride tracking, cutoff, & history)
CREATE INDEX IF NOT EXISTS idx_rides_user_status ON rides (user_id, status);
CREATE INDEX IF NOT EXISTS idx_rides_bike_status ON rides (bike_id, status);
CREATE INDEX IF NOT EXISTS idx_rides_user_created ON rides (user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_rides_tenant_created ON rides (tenant_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_rides_status ON rides (status);

-- 6. Transactions & Payment Webhooks Indexes
CREATE INDEX IF NOT EXISTS idx_transactions_user_created ON transactions (user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_transactions_user_type_status ON transactions (user_id, type, status, created_at);
CREATE INDEX IF NOT EXISTS idx_transactions_reference_id ON transactions (reference_id);
CREATE INDEX IF NOT EXISTS idx_transactions_gateway_ref ON transactions (gateway_ref);

-- 7. Notifications Indexes
CREATE INDEX IF NOT EXISTS idx_notifications_user_created ON notifications (user_id, created_at DESC);

-- 8. Reports & Maintenance Indexes
CREATE INDEX IF NOT EXISTS idx_reports_bike_id ON reports (bike_id);
CREATE INDEX IF NOT EXISTS idx_reports_status ON reports (status);

-- 9. Tenant Payouts Indexes
CREATE INDEX IF NOT EXISTS idx_tenant_payouts_tenant_period ON tenant_payouts (tenant_id, period_end DESC);

-- 10. Organizations & Hardware Orders Indexes
CREATE INDEX IF NOT EXISTS idx_tenants_organization_id ON tenants (organization_id);
CREATE INDEX IF NOT EXISTS idx_users_organization_id ON users (organization_id);
CREATE INDEX IF NOT EXISTS idx_rides_tenant_status_end ON rides (tenant_id, status, end_time DESC);
CREATE INDEX IF NOT EXISTS idx_hardware_orders_tenant ON hardware_orders (tenant_id, created_at DESC);

