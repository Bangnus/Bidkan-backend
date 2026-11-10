package stripe_webhook

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"math"
	"strconv"

	"github.com/Bangnus/Bidkan-backend/internal/app/domain/entity"
	domainRepo "github.com/Bangnus/Bidkan-backend/internal/app/domain/repository"
	infraStripe "github.com/Bangnus/Bidkan-backend/internal/app/infrastructure/stripe"
	"github.com/google/uuid"
)

type Service interface {
	HandleEvent(ctx context.Context, evt infraStripe.Event) error
}

type service struct {
	tenantRepo domainRepo.TenantRepository
	superRepo  domainRepo.SuperAdminRepository
	txRepo     domainRepo.TransactionRepository
	userRepo   domainRepo.UserRepository
}

func NewService(
	tenantRepo domainRepo.TenantRepository,
	superRepo domainRepo.SuperAdminRepository,
	txRepo domainRepo.TransactionRepository,
	userRepo domainRepo.UserRepository,
) Service {
	return &service{
		tenantRepo: tenantRepo,
		superRepo:  superRepo,
		txRepo:     txRepo,
		userRepo:   userRepo,
	}
}

func (s *service) HandleEvent(ctx context.Context, evt infraStripe.Event) error {
	if evt.ID == "" || evt.Type == "" {
		return errors.New("invalid stripe event envelope")
	}
	isNew, err := s.superRepo.TryRecordStripeEvent(ctx, evt.ID, evt.Type)
	if err != nil {
		return err
	}
	if !isNew {
		return nil
	}

	var envelope infraStripe.EventDataObject
	if err := json.Unmarshal(evt.Data, &envelope); err != nil {
		return err
	}

	switch evt.Type {
	case "account.updated":
		return s.onAccountUpdated(ctx, evt.Account, envelope.Object)
	case "transfer.reversed", "transfer.updated":
		return s.onTransferUpdated(ctx, envelope.Object)
	case "checkout.session.completed", "checkout.session.async_payment_succeeded":
		return s.onCheckoutPaid(ctx, envelope.Object)
	case "checkout.session.expired", "checkout.session.async_payment_failed":
		return s.onCheckoutFailed(ctx, envelope.Object)
	default:
		return nil
	}
}

func (s *service) onAccountUpdated(ctx context.Context, eventAcct string, raw json.RawMessage) error {
	var acct infraStripe.AccountStatus
	if err := json.Unmarshal(raw, &acct); err != nil {
		return err
	}
	acctID := acct.ID
	if acctID == "" {
		acctID = eventAcct
	}
	if acctID == "" {
		return nil
	}
	tenant, err := s.tenantRepo.GetByStripeAccountID(ctx, acctID)
	if err != nil {
		log.Printf("stripe webhook account.updated skipped unknown account=%s: %v", acctID, err)
		return nil
	}
	return s.tenantRepo.UpdateStripeStatus(ctx, tenant.ID, acctID, acct.PayoutsEnabled)
}

func (s *service) onTransferUpdated(ctx context.Context, raw json.RawMessage) error {
	var tr struct {
		ID             string `json:"id"`
		AmountReversed int64  `json:"amount_reversed"`
		Reversed       bool   `json:"reversed"`
	}
	if err := json.Unmarshal(raw, &tr); err != nil {
		return err
	}
	if tr.ID == "" || (!tr.Reversed && tr.AmountReversed <= 0) {
		return nil
	}
	payout, err := s.superRepo.GetPayoutByTransferID(ctx, tr.ID)
	if err != nil {
		log.Printf("stripe webhook transfer=%s not linked to local payout", tr.ID)
		return nil
	}
	msg := "Stripe transfer was reversed"
	return s.superRepo.UpdatePayoutStatus(ctx, payout.ID, "failed", &tr.ID, &msg)
}

func (s *service) onCheckoutPaid(ctx context.Context, raw json.RawMessage) error {
	var session infraStripe.CheckoutSession
	if err := json.Unmarshal(raw, &session); err != nil {
		return err
	}
	if session.PaymentStatus != "paid" {
		return nil
	}
	tx, err := s.resolveTransaction(ctx, session)
	if err != nil {
		log.Printf("stripe webhook checkout.session=%s unmatched: %v", session.ID, err)
		return nil
	}
	expectedTHB, err := strconv.ParseFloat(tx.Amount, 64)
	if err != nil || expectedTHB <= 0 {
		return errors.New("invalid local transaction amount")
	}
	expectedSatang := int64(math.Round(expectedTHB * 100))
	if session.AmountTotal > 0 && absDiff(expectedSatang, session.AmountTotal) > 1 {
		return errors.New("stripe paid amount mismatch")
	}

	completedNow, err := s.txRepo.CompleteIfPending(ctx, tx.ID, &session.ID)
	if err != nil {
		return err
	}
	if completedNow {
		return s.userRepo.AddBalance(ctx, tx.UserID, tx.Amount)
	}
	return nil
}

func (s *service) onCheckoutFailed(ctx context.Context, raw json.RawMessage) error {
	var session infraStripe.CheckoutSession
	if err := json.Unmarshal(raw, &session); err != nil {
		return err
	}
	tx, err := s.resolveTransaction(ctx, session)
	if err != nil || tx.Status != "pending" {
		return nil
	}
	return s.txRepo.UpdateStatus(ctx, tx.ID, "failed", &session.ID)
}

func (s *service) resolveTransaction(ctx context.Context, session infraStripe.CheckoutSession) (*entity.Transaction, error) {
	txIDStr := session.Metadata["transaction_id"]
	if txIDStr == "" {
		txIDStr = session.ClientReferenceID
	}
	if txIDStr != "" {
		if txID, err := uuid.Parse(txIDStr); err == nil {
			return s.txRepo.GetByID(ctx, txID)
		}
	}
	return s.txRepo.GetByGatewayRef(ctx, session.ID)
}

func absDiff(a, b int64) int64 {
	if a > b {
		return a - b
	}
	return b - a
}

