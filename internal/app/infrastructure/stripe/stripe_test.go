package stripe

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func signTestPayload(payload []byte, secret string, ts int64) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(fmt.Sprintf("%d.%s", ts, payload)))
	return fmt.Sprintf("t=%d,v1=%s", ts, hex.EncodeToString(mac.Sum(nil)))
}

func TestVerifySignature(t *testing.T) {
	payload := []byte(`{"id":"evt_1","type":"account.updated"}`)
	secret := "whsec_test_123"
	connectSecret := "whsec_connect_456"
	now := time.Now().Unix()

	sig := signTestPayload(payload, secret, now)
	if err := VerifySignature(payload, sig, 5*time.Minute, secret); err != nil {
		t.Fatalf("expected valid signature, got %v", err)
	}

	connectSig := signTestPayload(payload, connectSecret, now)
	if err := VerifySignature(payload, connectSig, 5*time.Minute, secret, connectSecret); err != nil {
		t.Fatalf("expected valid connect signature, got %v", err)
	}

	if err := VerifySignature(payload, sig, 5*time.Minute, "wrong_secret"); err == nil {
		t.Fatalf("expected error for wrong secret")
	}

	oldSig := signTestPayload(payload, secret, time.Now().Add(-10*time.Minute).Unix())
	if err := VerifySignature(payload, oldSig, 5*time.Minute, secret); err == nil {
		t.Fatalf("expected expired timestamp error")
	}

	if err := VerifySignature(payload, "", 5*time.Minute, secret); err != ErrMissingSignature {
		t.Fatalf("expected ErrMissingSignature, got %v", err)
	}
	if err := VerifySignature(payload, sig, 5*time.Minute, ""); err != ErrWebhookNotArmed {
		t.Fatalf("expected ErrWebhookNotArmed, got %v", err)
	}
}

func TestStripeHTTPClientRequests(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, _, ok := r.BasicAuth()
		if !ok || user != "sk_test_mock" {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(`{"error":{"message":"invalid key"}}`))
			return
		}
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/v1/accounts":
			_, _ = w.Write([]byte(`{"id":"acct_test_123"}`))
		case r.Method == http.MethodGet && r.URL.Path == "/v1/accounts/acct_test_123":
			_, _ = w.Write([]byte(`{"id":"acct_test_123","charges_enabled":true,"payouts_enabled":true,"details_submitted":true,"requirements":{"currently_due":[]}}`))
		case r.Method == http.MethodPost && r.URL.Path == "/v1/account_links":
			_, _ = w.Write([]byte(`{"url":"https://connect.stripe.com/setup/test"}`))
		case r.Method == http.MethodPost && r.URL.Path == "/v1/transfers":
			if r.Header.Get("Idempotency-Key") != "payout-1" {
				t.Fatalf("missing idempotency key")
			}
			_, _ = w.Write([]byte(`{"id":"tr_test_999","amount":12550,"currency":"thb","destination":"acct_test_123"}`))
		case r.Method == http.MethodPost && r.URL.Path == "/v1/checkout/sessions":
			_, _ = w.Write([]byte(`{"id":"cs_test_1","url":"https://checkout.stripe.com/c/pay/cs_test_1","payment_status":"unpaid","amount_total":10000,"currency":"thb"}`))
		default:
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"error":{"type":"invalid_request_error","code":"resource_missing","message":"missing"}}`))
		}
	}))
	defer srv.Close()

	client := NewClient(&Config{
		SecretKey:          "sk_test_mock",
		APIBaseURL:         srv.URL,
		ConnectAccountType: "express",
		ConnectCountry:     "TH",
	})
	ctx := context.Background()

	acctID, err := client.CreateConnectedAccount(ctx, "ku", "KU", "ops@ku.th")
	if err != nil || acctID != "acct_test_123" {
		t.Fatalf("CreateConnectedAccount got (%q, %v)", acctID, err)
	}
	status, err := client.GetAccount(ctx, acctID)
	if err != nil || !status.PayoutsEnabled {
		t.Fatalf("GetAccount got (%+v, %v)", status, err)
	}
	link, err := client.CreateAccountLink(ctx, acctID, "http://a", "http://b")
	if err != nil || link == "" {
		t.Fatalf("CreateAccountLink got (%q, %v)", link, err)
	}
	tr, err := client.CreateTransfer(ctx, acctID, 125.50, "payout-1", map[string]string{"tenant_slug": "ku"})
	if err != nil || tr.ID != "tr_test_999" || tr.Amount != 12550 {
		t.Fatalf("CreateTransfer got (%+v, %v)", tr, err)
	}
	cs, err := client.CreateCheckoutSession(ctx, CheckoutSessionParams{
		TransactionID: "tx-1", UserID: "u-1", AmountTHB: 100, SuccessURL: "http://ok", CancelURL: "http://no",
	})
	if err != nil || cs.ID != "cs_test_1" {
		t.Fatalf("CreateCheckoutSession got (%+v, %v)", cs, err)
	}
	if _, err := client.GetAccount(ctx, "acct_missing"); !IsMissingResource(err) {
		t.Fatalf("expected IsMissingResource for 404, got %v", err)
	}
}

