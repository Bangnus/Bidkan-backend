package rank

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"github.com/Bangnus/Bidkan-backend/internal/app/domain/repository"
	"github.com/google/uuid"
)

type Response struct {
	CurrentRank        string  `json:"current_rank"`
	TotalSpent4Months  float64 `json:"total_spent_4_months"`
	CurrentDiscount    float64 `json:"current_discount"`
	NextRankThreshold  float64 `json:"next_rank_threshold"`
	ProgressPercentage float64 `json:"progress_percentage"`
}

type Service interface {
	GetUserRank(ctx context.Context, userID uuid.UUID) (*Response, error)
}

type service struct {
	spendingRepo repository.UserMonthlySpendingRepository
	rankCache    repository.UserRankCacheRepository
}

func NewService(spendingRepo repository.UserMonthlySpendingRepository, rankCache repository.UserRankCacheRepository) Service {
	return &service{
		spendingRepo: spendingRepo,
		rankCache:    rankCache,
	}
}

func (s *service) GetUserRank(ctx context.Context, userID uuid.UUID) (*Response, error) {
	// 1. ลองดึงจาก Cache ก่อน
	if cachedData, err := s.rankCache.GetRankCache(ctx, userID.String()); err == nil {
		var resp Response
		if err := json.Unmarshal([]byte(cachedData), &resp); err == nil {
			return &resp, nil
		}
	}

	now := time.Now()
	// คำนวณปีเดือนย้อนหลัง 4 เดือน (เช่น 202402)
	since := now.AddDate(0, -3, 0)
	sinceYearMonth, _ := strconv.Atoi(since.Format("200601"))

	totalSpent, err := s.spendingRepo.GetTotalInWindow(ctx, userID, sinceYearMonth)
	if err != nil {
		return nil, err
	}

	rank := "Bronze"
	discount := 0.0
	nextThreshold := 800.0

	if totalSpent > 2500 {
		rank = "Gold"
		discount = 1.0
		nextThreshold = 0 // Max rank
	} else if totalSpent > 800 {
		rank = "Silver"
		discount = 0.5
		nextThreshold = 2500.0
	}

	progress := 0.0
	if nextThreshold > 0 {
		if rank == "Bronze" {
			progress = (totalSpent / 800.0) * 100
		} else if rank == "Silver" {
			progress = ((totalSpent - 800.0) / (2500.0 - 800.0)) * 100
		}
	} else {
		progress = 100
	}

	if progress > 100 {
		progress = 100
	}

	resp := &Response{
		CurrentRank:        rank,
		TotalSpent4Months:  totalSpent,
		CurrentDiscount:    discount,
		NextRankThreshold:  nextThreshold,
		ProgressPercentage: progress,
	}

	// 2. เก็บลง Cache (10 นาที เพื่อให้เลื่อนแรงค์ได้ค่อนข้างเรียลไทม์)
	_ = s.rankCache.SetRankCache(ctx, userID.String(), resp, 10*time.Minute)

	return resp, nil
}
