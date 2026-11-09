package auth

import (
	"context"
	"database/sql"
	"log"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func SeedDemoData(ctx context.Context, db *sql.DB) {
	kuID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	cmuID := uuid.MustParse("22222222-2222-2222-2222-222222222222")

	// 1. Seed Tenants (KU & CMU)
	_, err := db.ExecContext(ctx, `
		INSERT INTO tenants (id, slug, name, stripe_account_id, stripe_onboarding_completed, platform_fee_percent, unlock_fee, per_minute_rate, status)
		VALUES
		  ($1, 'ku', 'มหาวิทยาลัยเกษตรศาสตร์ (กำแพงแสน)', 'acct_demo_ku_01', true, 15.00, 5.00, 1.50, 'active'),
		  ($2, 'cmu', 'มหาวิทยาลัยเชียงใหม่', 'acct_demo_cmu_01', true, 15.00, 5.00, 1.50, 'active')
		ON CONFLICT (slug) DO NOTHING
	`, kuID, cmuID)
	if err != nil {
		log.Printf("seed tenants skipped: %v", err)
		return
	}

	// 2. Seed Backoffice Accounts (password: password123)
	hash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	pw := string(hash)

	_, _ = db.ExecContext(ctx, `
		INSERT INTO users (id, tenant_id, phone_number, username, password, wallet_balance, role, status)
		VALUES
		  ($1, NULL, '0800000001', 'superadmin', $4, 0.00, 'super_admin', 'active'),
		  ($2, $5, '0800000002', 'ku_admin', $4, 0.00, 'tenant_admin', 'active'),
		  ($3, $5, '0800000003', 'ku_staff', $4, 0.00, 'tenant_staff', 'active')
		ON CONFLICT (phone_number) DO NOTHING
	`, uuid.New(), uuid.New(), uuid.New(), pw, kuID)

	// 3. Seed Sample Bikes
	_, _ = db.ExecContext(ctx, `
		INSERT INTO bikes (id, hardware_id, tenant_id, status, lat, lon, battery_level)
		VALUES
		  ('BK-001', 'ESP32-GPS-9901', $1, 'available', 14.0234, 99.9782, 92),
		  ('BK-002', 'ESP32-GPS-9902', $1, 'in_use', 14.0241, 99.9791, 85),
		  ('BK-003', 'ESP32-GPS-9903', $1, 'available', 14.0228, 99.9765, 14),
		  ('BK-004', 'ESP32-GPS-9904', $1, 'maintenance', 14.0210, 99.9745, 18),
		  ('BK-005', 'ESP32-GPS-9905', $2, 'available', 18.7961, 98.9537, 88)
		ON CONFLICT (id) DO NOTHING
	`, kuID, cmuID)

	// 4. Seed Sample Zones (deterministic IDs to stay idempotent)
	z1 := uuid.MustParse("44444444-4444-4444-4444-444444444401")
	z2 := uuid.MustParse("44444444-4444-4444-4444-444444444402")
	_, _ = db.ExecContext(ctx, `
		INSERT INTO zones (id, tenant_id, name, type, boundary, radius)
		VALUES
		  ($1, $3, 'จุดจอดหน้าโรงอาหารกลาง (P1)', 'parking', '{"lat":14.0234,"lon":99.9782}', 30),
		  ($2, $3, 'เขตจำกัดการขับขี่ภายในวิทยาเขต', 'riding_limit', '{"lat":14.0240,"lon":99.9780}', 1500)
		ON CONFLICT (id) DO NOTHING
	`, z1, z2, kuID)

	// 5. Seed Sample Maintenance Report for KU Field Ops
	reportID := uuid.MustParse("33333333-3333-3333-3333-333333333331")
	_, _ = db.ExecContext(ctx, `
		INSERT INTO reports (id, bike_id, reported_by, issue_type, status)
		SELECT $1, 'BK-004', id, 'เบรกหน้าหลวม', 'pending'
		FROM users WHERE phone_number = '0800000002' LIMIT 1
		ON CONFLICT (id) DO NOTHING
	`, reportID)
}
