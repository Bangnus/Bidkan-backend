package audit_manage

import (
	"bytes"
	"context"
	"encoding/csv"
	"errors"
	"fmt"

	"github.com/Bangnus/Bidkan-backend/internal/app/domain/entity"
	domainRepo "github.com/Bangnus/Bidkan-backend/internal/app/domain/repository"
)

type Service struct {
	tenantRepo domainRepo.TenantRepository
	campusRepo domainRepo.CampusOpsRepository
	staffRepo  domainRepo.StaffCompensationRepository
	auditRepo  domainRepo.AuditRepository
}

func NewService(
	tenantRepo domainRepo.TenantRepository,
	campusRepo domainRepo.CampusOpsRepository,
	staffRepo domainRepo.StaffCompensationRepository,
	auditRepo domainRepo.AuditRepository,
) *Service {
	return &Service{
		tenantRepo: tenantRepo,
		campusRepo: campusRepo,
		staffRepo:  staffRepo,
		auditRepo:  auditRepo,
	}
}

func (s *Service) ListAuditLogs(ctx context.Context, slug, action string) ([]*entity.AuditLog, error) {
	tenant, err := s.tenantRepo.GetBySlug(ctx, slug)
	if err != nil {
		return nil, errors.New("campus tenant not found")
	}
	return s.auditRepo.ListAuditLogs(ctx, tenant.ID, action, 100)
}

func (s *Service) ExportRidesCSV(ctx context.Context, slug string) ([]byte, error) {
	tenant, err := s.tenantRepo.GetBySlug(ctx, slug)
	if err != nil {
		return nil, errors.New("campus tenant not found")
	}

	rides, err := s.campusRepo.ListRidesByTenant(ctx, tenant.ID)
	if err != nil {
		return nil, err
	}

	buf := &bytes.Buffer{}
	buf.WriteString("\xef\xbb\xbf") // UTF-8 BOM for Excel
	writer := csv.NewWriter(buf)

	_ = writer.Write([]string{
		"รหัสทริป (Ride ID)", "รหัสจักรยาน (Bike ID)", "ผู้ใช้งาน", "เบอร์โทร",
		"ระยะทาง (กม.)", "ค่าโดยสาร (บาท)", "ค่าธรรมเนียมแพลตฟอร์ม", "รายได้สุทธิวิทยาเขต", "สถานะ", "เวลาเริ่มต้น",
	})

	for _, r := range rides {
		_ = writer.Write([]string{
			r.ID.String(),
			r.BikeID,
			r.Username,
			r.PhoneNumber,
			fmt.Sprintf("%.2f", r.DistanceKm),
			fmt.Sprintf("%.2f", r.TotalFare),
			fmt.Sprintf("%.2f", r.PlatformFee),
			fmt.Sprintf("%.2f", r.PartnerEarnings),
			r.Status,
			r.StartTime.Format("2006-01-02 15:04:05"),
		})
	}
	writer.Flush()
	return buf.Bytes(), nil
}

func (s *Service) ExportRevenueCSV(ctx context.Context, slug string) ([]byte, error) {
	tenant, err := s.tenantRepo.GetBySlug(ctx, slug)
	if err != nil {
		return nil, errors.New("campus tenant not found")
	}

	completedStats, _ := s.campusRepo.GetCompletedStats(ctx, tenant.ID)
	unsettledStats, _ := s.campusRepo.GetUnsettledStats(ctx, tenant.ID)
	if completedStats == nil {
		completedStats = &entity.RideStatsSummary{}
	}
	if unsettledStats == nil {
		unsettledStats = &entity.RideStatsSummary{}
	}

	buf := &bytes.Buffer{}
	buf.WriteString("\xef\xbb\xbf") // UTF-8 BOM for Excel
	writer := csv.NewWriter(buf)

	_ = writer.Write([]string{"หมวดหมู่รายงาน", "ยอดรวม (บาท)"})
	_ = writer.Write([]string{"ยอดรายได้รวมสะสม (Gross Volume)", fmt.Sprintf("%.2f", completedStats.GrossAmount)})
	_ = writer.Write([]string{"ค่าธรรมเนียมแพลตฟอร์ม (Platform Fee)", fmt.Sprintf("%.2f", completedStats.GPFeeAmount)})
	_ = writer.Write([]string{"รายได้สุทธิคงเหลือ (Net Partner Earnings)", fmt.Sprintf("%.2f", completedStats.NetAmount)})
	_ = writer.Write([]string{"ยอดรอโอนรอบถัดไป (Pending Payout)", fmt.Sprintf("%.2f", unsettledStats.NetAmount)})
	_ = writer.Write([]string{"จำนวนทริปเสร็จสิ้นทั้งหมด", fmt.Sprintf("%d", completedStats.TotalRides)})

	writer.Flush()
	return buf.Bytes(), nil
}

func (s *Service) ExportStaffPayoutsCSV(ctx context.Context, slug string) ([]byte, error) {
	tenant, err := s.tenantRepo.GetBySlug(ctx, slug)
	if err != nil {
		return nil, errors.New("campus tenant not found")
	}

	earnings, err := s.staffRepo.ListStaffEarnings(ctx, tenant.ID)
	if err != nil {
		return nil, err
	}

	buf := &bytes.Buffer{}
	buf.WriteString("\xef\xbb\xbf") // UTF-8 BOM for Excel
	writer := csv.NewWriter(buf)

	_ = writer.Write([]string{
		"รหัสช่าง", "ชื่อช่าง", "อีเมล", "เบอร์โทร",
		"สลับแบตเตอรี่ (ครั้ง)", "ซ่อมบำรุง (ครั้ง)", "รายได้รวมสะสม (บาท)",
	})

	for _, e := range earnings {
		_ = writer.Write([]string{
			e.StaffUserID.String(),
			e.Username,
			e.Email,
			e.PhoneNumber,
			fmt.Sprintf("%d", e.TotalSwaps),
			fmt.Sprintf("%d", e.TotalRepairs),
			fmt.Sprintf("%.2f", e.TotalEarnings),
		})
	}
	writer.Flush()
	return buf.Bytes(), nil
}

func (s *Service) RecordAuditLog(ctx context.Context, log *entity.AuditLog) {
	if s.auditRepo != nil {
		_ = s.auditRepo.LogAction(ctx, log)
	}
}
