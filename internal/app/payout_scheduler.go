package app

import (
	"context"
	"log"
	"time"

	infraStripe "github.com/Bangnus/Bidkan-backend/internal/app/infrastructure/stripe"
	superPayoutManage "github.com/Bangnus/Bidkan-backend/internal/app/usecase/super/payout_manage"
)

func StartAutoPayoutScheduler(ctx context.Context, stripeClient infraStripe.Client, svc superPayoutManage.Service) {
	if stripeClient == nil || !stripeClient.IsConfigured() {
		return
	}
	cfg := stripeClient.Config()
	if !cfg.AutoPayoutEnabled {
		return
	}
	log.Printf("Stripe Auto-Payout scheduler enabled (interval=%s)", cfg.AutoPayoutInterval)
	go func() {
		ticker := time.NewTicker(cfg.AutoPayoutInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				res, err := svc.RunAutoPayouts(ctx)
				if err != nil {
					log.Printf("Stripe Auto-Payout batch failed: %v", err)
					continue
				}
				log.Printf("Stripe Auto-Payout batch done: attempted=%d succeeded=%d failed=%d skipped=%d",
					res.Attempted, res.Succeeded, res.Failed, res.Skipped)
			}
		}
	}()
}

