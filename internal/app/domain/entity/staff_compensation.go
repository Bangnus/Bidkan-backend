package entity

import (
	"time"

	"github.com/google/uuid"
)

type StaffCompensationConfig struct {
	ID              uuid.UUID `json:"id"`
	TenantID        uuid.UUID `json:"tenant_id"`
	PayModel        string    `json:"pay_model"` // per_task, daily, monthly
	BatterySwapFee  float64   `json:"battery_swap_fee"`
	RepairTaskFee   float64   `json:"repair_task_fee"`
	DailyBaseWage   float64   `json:"daily_base_wage"`
	MonthlyBaseWage float64   `json:"monthly_base_wage"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type StaffTaskLog struct {
	ID           uuid.UUID `json:"id"`
	TenantID     uuid.UUID `json:"tenant_id"`
	StaffUserID  uuid.UUID `json:"staff_user_id"`
	TaskType     string    `json:"task_type"` // battery_swap, maintenance_repair
	TargetID     string    `json:"target_id"` // bike_id or report_id
	EarnedAmount float64   `json:"earned_amount"`
	CreatedAt    time.Time `json:"created_at"`
}

type StaffEarningSummary struct {
	StaffUserID   uuid.UUID `json:"staff_user_id"`
	Username      string    `json:"username"`
	Email         string    `json:"email"`
	PhoneNumber   string    `json:"phone_number"`
	TotalSwaps    int       `json:"total_swaps"`
	TotalRepairs  int       `json:"total_repairs"`
	TotalEarnings float64   `json:"total_earnings"`
}

type MyEarningsSummary struct {
	TodaySwaps     int     `json:"today_swaps"`
	TodayRepairs   int     `json:"today_repairs"`
	TodayEarnings  float64 `json:"today_earnings"`
	MonthSwaps     int     `json:"month_swaps"`
	MonthRepairs   int     `json:"month_repairs"`
	MonthEarnings  float64 `json:"month_earnings"`
	PayModel       string  `json:"pay_model"`
	BatterySwapFee float64 `json:"battery_swap_fee"`
	RepairTaskFee  float64 `json:"repair_task_fee"`
}

