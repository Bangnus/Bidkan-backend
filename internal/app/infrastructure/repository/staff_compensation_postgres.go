package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/Bangnus/Bidkan-backend/internal/app/domain/entity"
	domainRepo "github.com/Bangnus/Bidkan-backend/internal/app/domain/repository"
	"github.com/google/uuid"
)

type staffCompensationPostgres struct {
	db *sql.DB
}

func NewStaffCompensationPostgres(db *sql.DB) domainRepo.StaffCompensationRepository {
	return &staffCompensationPostgres{db: db}
}

func (r *staffCompensationPostgres) GetCompensationConfig(ctx context.Context, tenantID uuid.UUID) (*entity.StaffCompensationConfig, error) {
	query := `
		SELECT id, tenant_id, pay_model, battery_swap_fee, repair_task_fee, daily_base_wage, monthly_base_wage, created_at, updated_at
		FROM staff_compensation_configs
		WHERE tenant_id = $1`
	var cfg entity.StaffCompensationConfig
	err := r.db.QueryRowContext(ctx, query, tenantID).Scan(
		&cfg.ID, &cfg.TenantID, &cfg.PayModel, &cfg.BatterySwapFee,
		&cfg.RepairTaskFee, &cfg.DailyBaseWage, &cfg.MonthlyBaseWage,
		&cfg.CreatedAt, &cfg.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return &entity.StaffCompensationConfig{
			TenantID:        tenantID,
			PayModel:        "per_task",
			BatterySwapFee:  20.00,
			RepairTaskFee:   50.00,
			DailyBaseWage:   400.00,
			MonthlyBaseWage: 10000.00,
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		}, nil
	}
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (r *staffCompensationPostgres) UpsertCompensationConfig(ctx context.Context, cfg *entity.StaffCompensationConfig) error {
	query := `
		INSERT INTO staff_compensation_configs (tenant_id, pay_model, battery_swap_fee, repair_task_fee, daily_base_wage, monthly_base_wage, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW())
		ON CONFLICT (tenant_id) DO UPDATE SET
			pay_model = EXCLUDED.pay_model,
			battery_swap_fee = EXCLUDED.battery_swap_fee,
			repair_task_fee = EXCLUDED.repair_task_fee,
			daily_base_wage = EXCLUDED.daily_base_wage,
			monthly_base_wage = EXCLUDED.monthly_base_wage,
			updated_at = NOW()`
	_, err := r.db.ExecContext(ctx, query,
		cfg.TenantID, cfg.PayModel, cfg.BatterySwapFee,
		cfg.RepairTaskFee, cfg.DailyBaseWage, cfg.MonthlyBaseWage,
	)
	return err
}

func (r *staffCompensationPostgres) LogStaffTask(ctx context.Context, log *entity.StaffTaskLog) error {
	query := `
		INSERT INTO staff_task_logs (tenant_id, staff_user_id, task_type, target_id, earned_amount, created_at)
		VALUES ($1, $2, $3, $4, $5, NOW())`
	_, err := r.db.ExecContext(ctx, query,
		log.TenantID, log.StaffUserID, log.TaskType, log.TargetID, log.EarnedAmount,
	)
	return err
}

func (r *staffCompensationPostgres) ListStaffEarnings(ctx context.Context, tenantID uuid.UUID) ([]*entity.StaffEarningSummary, error) {
	query := `
		SELECT 
			u.id, u.username, COALESCE(u.email, ''), COALESCE(u.phone_number, ''),
			COUNT(CASE WHEN l.task_type = 'battery_swap' THEN 1 END) AS total_swaps,
			COUNT(CASE WHEN l.task_type = 'maintenance_repair' THEN 1 END) AS total_repairs,
			COALESCE(SUM(l.earned_amount), 0.00) AS total_earnings
		FROM users u
		LEFT JOIN staff_task_logs l ON u.id = l.staff_user_id AND l.created_at >= date_trunc('month', CURRENT_DATE)
		WHERE u.tenant_id = $1 AND u.role = 'tenant_staff'
		GROUP BY u.id, u.username, u.email, u.phone_number
		ORDER BY total_earnings DESC, u.username ASC`
	rows, err := r.db.QueryContext(ctx, query, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := make([]*entity.StaffEarningSummary, 0)
	for rows.Next() {
		var s entity.StaffEarningSummary
		if err := rows.Scan(&s.StaffUserID, &s.Username, &s.Email, &s.PhoneNumber, &s.TotalSwaps, &s.TotalRepairs, &s.TotalEarnings); err != nil {
			return nil, err
		}
		list = append(list, &s)
	}
	return list, nil
}

func (r *staffCompensationPostgres) GetStaffEarningsByUserID(ctx context.Context, tenantID, staffUserID uuid.UUID) (*entity.MyEarningsSummary, error) {
	cfg, err := r.GetCompensationConfig(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	query := `
		SELECT
			COUNT(CASE WHEN created_at >= CURRENT_DATE AND task_type = 'battery_swap' THEN 1 END) AS today_swaps,
			COUNT(CASE WHEN created_at >= CURRENT_DATE AND task_type = 'maintenance_repair' THEN 1 END) AS today_repairs,
			COALESCE(SUM(CASE WHEN created_at >= CURRENT_DATE THEN earned_amount ELSE 0 END), 0.00) AS today_earnings,
			COUNT(CASE WHEN created_at >= date_trunc('month', CURRENT_DATE) AND task_type = 'battery_swap' THEN 1 END) AS month_swaps,
			COUNT(CASE WHEN created_at >= date_trunc('month', CURRENT_DATE) AND task_type = 'maintenance_repair' THEN 1 END) AS month_repairs,
			COALESCE(SUM(CASE WHEN created_at >= date_trunc('month', CURRENT_DATE) THEN earned_amount ELSE 0 END), 0.00) AS month_earnings
		FROM staff_task_logs
		WHERE tenant_id = $1 AND staff_user_id = $2`

	var s entity.MyEarningsSummary
	err = r.db.QueryRowContext(ctx, query, tenantID, staffUserID).Scan(
		&s.TodaySwaps, &s.TodayRepairs, &s.TodayEarnings,
		&s.MonthSwaps, &s.MonthRepairs, &s.MonthEarnings,
	)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	s.PayModel = cfg.PayModel
	s.BatterySwapFee = cfg.BatterySwapFee
	s.RepairTaskFee = cfg.RepairTaskFee
	return &s, nil
}

