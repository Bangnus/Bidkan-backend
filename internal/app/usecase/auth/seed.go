package auth

import (
	"context"
	"database/sql"
	"log"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func SeedDemoData(ctx context.Context, db *sql.DB) {
	kuOrgID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	kuID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	kuBkkID := uuid.MustParse("11111111-1111-1111-1111-111111111112")
	cmuID := uuid.MustParse("22222222-2222-2222-2222-222222222222")

	// 1. Seed Organization (KU Multi-Campus University)
	_, err := db.ExecContext(ctx, `
		INSERT INTO organizations (id, slug, name, contact_email, status)
		VALUES ($1, 'ku-org', 'มหาวิทยาลัยเกษตรศาสตร์ (Kasetsart University)', 'mobility@ku.th', 'active')
		ON CONFLICT (slug) DO NOTHING
	`, kuOrgID)
	if err != nil {
		log.Printf("seed organization skipped: %v", err)
	}

	// 2. Seed Tenants (KU Kamphaeng Saen, KU Bangkhen, CMU)
	_, err = db.ExecContext(ctx, `
		INSERT INTO tenants (id, organization_id, slug, name, stripe_account_id, stripe_onboarding_completed, platform_fee_percent, auto_tiered_gp, unlock_fee, per_minute_rate, parking_penalty_fee, status)
		VALUES
		  ($1, $4, 'ku', 'ม.เกษตรศาสตร์ (วิทยาเขตกำแพงแสน)', 'acct_demo_ku_01', true, 18.00, true, 5.00, 1.50, 50.00, 'active'),
		  ($2, $4, 'ku-bkk', 'ม.เกษตรศาสตร์ (วิทยาเขตบางเขน)', NULL, false, 18.00, true, 5.00, 1.50, 50.00, 'active'),
		  ($3, NULL, 'cmu', 'มหาวิทยาลัยเชียงใหม่', 'acct_demo_cmu_01', true, 18.00, true, 5.00, 1.50, 50.00, 'active')
		ON CONFLICT (slug) DO NOTHING
	`, kuID, kuBkkID, cmuID, kuOrgID)
	if err != nil {
		log.Printf("seed tenants skipped: %v", err)
		return
	}
	_, _ = db.ExecContext(ctx, `UPDATE tenants SET organization_id = $1 WHERE slug = 'ku' AND organization_id IS NULL`, kuOrgID)

	// 3. Seed Backoffice Accounts (password: password123)
	hash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	pw := string(hash)

	_, _ = db.ExecContext(ctx, `
		INSERT INTO users (id, tenant_id, organization_id, phone_number, email, username, password, wallet_balance, role, status)
		VALUES
		  ($1, NULL, NULL, '0800000001', 'superadmin@bidkan.com', 'superadmin', $5, 0.00, 'super_admin', 'active'),
		  ($2, $6, $7, '0800000002', 'ku_admin@bidkan.com', 'ku_admin', $5, 0.00, 'tenant_admin', 'active'),
		  ($3, $6, $7, '0800000003', 'ku_staff@bidkan.com', 'ku_staff', $5, 0.00, 'tenant_staff', 'active'),
		  ($4, $6, $7, '0800000004', 'ku_org@bidkan.com', 'ku_org_admin', $5, 0.00, 'org_admin', 'active')
		ON CONFLICT (phone_number) DO UPDATE SET
		  email = COALESCE(users.email, EXCLUDED.email),
		  organization_id = COALESCE(users.organization_id, EXCLUDED.organization_id)
	`, uuid.New(), uuid.New(), uuid.New(), uuid.New(), pw, kuID, kuOrgID)

	// 4. Seed Sample Bikes across campuses
	_, _ = db.ExecContext(ctx, `
		INSERT INTO bikes (id, hardware_id, tenant_id, status, lat, lon, battery_level)
		VALUES
		  ('BK-001', 'ESP32-GPS-9901', $1, 'available', 14.0234, 99.9782, 92),
		  ('BK-002', 'ESP32-GPS-9902', $1, 'in_use', 14.0241, 99.9791, 85),
		  ('BK-003', 'ESP32-GPS-9903', $1, 'available', 14.0228, 99.9765, 14),
		  ('BK-004', 'ESP32-GPS-9904', $1, 'maintenance', 14.0210, 99.9745, 18),
		  ('BK-005', 'ESP32-GPS-9905', $3, 'available', 18.7961, 98.9537, 88),
		  ('BK-006', 'ESP32-GPS-9906', $2, 'available', 13.8476, 100.5696, 95)
		ON CONFLICT (id) DO NOTHING
	`, kuID, kuBkkID, cmuID)

	// 5. Seed Sample Zones & Initial Maintenance Report
	z1 := uuid.MustParse("44444444-4444-4444-4444-444444444401")
	z2 := uuid.MustParse("44444444-4444-4444-4444-444444444402")
	z3 := uuid.MustParse("44444444-4444-4444-4444-444444444403")
	_, _ = db.ExecContext(ctx, `
		INSERT INTO zones (id, tenant_id, name, type, boundary, radius)
		VALUES
		  ($1, $4, 'จุดจอดหน้าโรงอาหารกลาง (P1)', 'parking', '{"lat":14.0234,"lon":99.9782}', 30),
		  ($2, $4, 'เขตจำกัดการขับขี่ภายในวิทยาเขต', 'riding_limit', '{"lat":14.0240,"lon":99.9780}', 1500),
		  ($3, $5, 'จุดจอดอาคารเทพศาสตร์สถิตย์ (BKK-P1)', 'parking', '{"lat":13.8476,"lon":100.5696}', 40)
		ON CONFLICT (id) DO NOTHING
	`, z1, z2, z3, kuID, kuBkkID)

	reportID := uuid.MustParse("33333333-3333-3333-3333-333333333331")
	_, _ = db.ExecContext(ctx, `
		INSERT INTO reports (id, bike_id, reported_by, issue_type, status)
		SELECT $1, 'BK-004', id, 'เบรกหน้าหลวม', 'pending'
		FROM users WHERE phone_number = '0800000002' LIMIT 1
		ON CONFLICT (id) DO NOTHING
	`, reportID)
}
