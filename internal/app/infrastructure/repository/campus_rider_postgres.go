package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/Bangnus/Bidkan-backend/internal/app/domain/entity"
	"github.com/google/uuid"
)

type CampusRiderPostgres struct {
	db *sql.DB
}

func NewCampusRiderPostgres(db *sql.DB) *CampusRiderPostgres {
	return &CampusRiderPostgres{db: db}
}

func (r *CampusRiderPostgres) ListCampusRiders(ctx context.Context, tenantID uuid.UUID, search, filter string) ([]*entity.CampusRider, error) {
	query := `
		SELECT 
			u.id, COALESCE(u.username, '-'), COALESCE(u.phone_number, '-'), COALESCE(u.email, '-'),
			u.status, u.wallet_balance, u.is_student, COALESCE(u.student_id, '-'),
			COALESCE(u.student_card_url, ''), u.student_verified,
			COUNT(DISTINCT r.id) as total_rides,
			COUNT(DISTINCT CASE WHEN r.penalty_status IN ('applied', 'disputed') THEN r.id END) as dispute_count,
			u.created_at
		FROM users u
		LEFT JOIN rides r ON r.user_id = u.id AND r.tenant_id = $1
		WHERE (u.tenant_id = $1 OR r.id IS NOT NULL)
	`
	args := []interface{}{tenantID}

	if search != "" {
		args = append(args, "%"+search+"%")
		query += fmt.Sprintf(" AND (u.username ILIKE $%d OR u.phone_number ILIKE $%d OR u.student_id ILIKE $%d)", len(args), len(args), len(args))
	}
	if filter == "student" {
		query += " AND u.is_student = true"
	} else if filter == "pending_verification" {
		query += " AND u.is_student = true AND u.student_verified = false"
	} else if filter == "suspended" {
		query += " AND u.status = 'suspended'"
	}

	query += " GROUP BY u.id ORDER BY u.created_at DESC LIMIT 100"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var riders []*entity.CampusRider
	for rows.Next() {
		item := &entity.CampusRider{}
		if err := rows.Scan(
			&item.UserID, &item.Username, &item.PhoneNumber, &item.Email,
			&item.Status, &item.WalletBalance, &item.IsStudent, &item.StudentID,
			&item.StudentCardURL, &item.StudentVerified, &item.TotalRides,
			&item.DisputeCount, &item.CreatedAt,
		); err != nil {
			return nil, err
		}
		riders = append(riders, item)
	}
	return riders, nil
}

func (r *CampusRiderPostgres) UpdateStudentVerification(ctx context.Context, userID uuid.UUID, verified bool) error {
	query := `UPDATE users SET student_verified = $1, updated_at = NOW() WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, verified, userID)
	return err
}

func (r *CampusRiderPostgres) UpdateUserStatus(ctx context.Context, userID uuid.UUID, status string) error {
	query := `UPDATE users SET status = $1, updated_at = NOW() WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, status, userID)
	return err
}

func (r *CampusRiderPostgres) ListDisputedRides(ctx context.Context, tenantID uuid.UUID, status string) ([]*entity.DisputedRide, error) {
	query := `
		SELECT 
			r.id, r.user_id, COALESCE(u.username, '-'), COALESCE(u.phone_number, '-'),
			r.bike_id, r.total_fare, r.parking_penalty, r.penalty_status,
			COALESCE(r.dispute_reason, '-'), r.refund_amount, r.refunded_at, r.created_at
		FROM rides r
		JOIN users u ON u.id = r.user_id
		WHERE r.tenant_id = $1 AND r.penalty_status != 'none'
	`
	args := []interface{}{tenantID}
	if status != "" && status != "all" {
		args = append(args, status)
		query += fmt.Sprintf(" AND r.penalty_status = $%d", len(args))
	}
	query += " ORDER BY r.created_at DESC LIMIT 100"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*entity.DisputedRide
	for rows.Next() {
		item := &entity.DisputedRide{}
		var refundedAt sql.NullTime
		if err := rows.Scan(
			&item.RideID, &item.UserID, &item.Username, &item.PhoneNumber,
			&item.BikeID, &item.TotalFare, &item.ParkingPenalty, &item.PenaltyStatus,
			&item.DisputeReason, &item.RefundAmount, &refundedAt, &item.CreatedAt,
		); err != nil {
			return nil, err
		}
		if refundedAt.Valid {
			item.RefundedAt = &refundedAt.Time
		}
		list = append(list, item)
	}
	return list, nil
}

func (r *CampusRiderPostgres) RefundPenalty(ctx context.Context, tenantID, rideID, userID uuid.UUID, refundAmount float64, reason string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var currentStatus string
	checkQuery := `SELECT penalty_status FROM rides WHERE id = $1 AND tenant_id = $2 FOR UPDATE`
	if err := tx.QueryRowContext(ctx, checkQuery, rideID, tenantID).Scan(&currentStatus); err != nil {
		return errors.New("ride not found for this tenant")
	}
	if currentStatus == "refunded" {
		return errors.New("this penalty has already been refunded")
	}

	updateRideQuery := `
		UPDATE rides 
		SET penalty_status = 'refunded', refund_amount = $1, refunded_at = NOW(), dispute_reason = COALESCE(dispute_reason, $2), updated_at = NOW()
		WHERE id = $3 AND tenant_id = $4
	`
	if _, err := tx.ExecContext(ctx, updateRideQuery, refundAmount, reason, rideID, tenantID); err != nil {
		return err
	}

	updateWalletQuery := `UPDATE users SET wallet_balance = wallet_balance + $1, updated_at = NOW() WHERE id = $2`
	if _, err := tx.ExecContext(ctx, updateWalletQuery, refundAmount, userID); err != nil {
		return err
	}

	txQuery := `
		INSERT INTO transactions (id, user_id, amount, type, status, reference_id, description, created_at, updated_at)
		VALUES ($1, $2, $3, 'refund_dispute', 'success', $4, $5, $6, $6)
	`
	now := time.Now()
	desc := fmt.Sprintf("คืนเงินค่าปรับ/ข้อพิพาท: %s", reason)
	if _, err := tx.ExecContext(ctx, txQuery, uuid.New(), userID, refundAmount, rideID.String(), desc, now); err != nil {
		return err
	}

	return tx.Commit()
}
