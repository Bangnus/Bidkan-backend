package stripe

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

const DefaultWebhookTolerance = 5 * time.Minute

var (
	ErrMissingSignature  = errors.New("missing stripe-signature header")
	ErrInvalidSignature  = errors.New("invalid stripe webhook signature")
	ErrExpiredSignature  = errors.New("stripe webhook timestamp outside tolerance")
	ErrWebhookNotArmed   = errors.New("stripe webhook secret not configured")
)

type Event struct {
	ID      string          `json:"id"`
	Type    string          `json:"type"`
	Account string          `json:"account"`
	Data    json.RawMessage `json:"data"`
}

type EventDataObject struct {
	Object json.RawMessage `json:"object"`
}

func VerifySignature(payload []byte, header string, tolerance time.Duration, secrets ...string) error {
	var activeSecrets []string
	for _, s := range secrets {
		if trimmed := strings.TrimSpace(s); trimmed != "" {
			activeSecrets = append(activeSecrets, trimmed)
		}
	}
	if len(activeSecrets) == 0 {
		return ErrWebhookNotArmed
	}
	header = strings.TrimSpace(header)
	if header == "" {
		return ErrMissingSignature
	}

	var timestampStr string
	var sigs [][]byte
	for _, part := range strings.Split(header, ",") {
		kv := strings.SplitN(strings.TrimSpace(part), "=", 2)
		if len(kv) != 2 {
			continue
		}
		switch kv[0] {
		case "t":
			timestampStr = kv[1]
		case "v1":
			if decoded, err := hex.DecodeString(kv[1]); err == nil {
				sigs = append(sigs, decoded)
			}
		}
	}
	if timestampStr == "" || len(sigs) == 0 {
		return ErrInvalidSignature
	}

	ts, err := strconv.ParseInt(timestampStr, 10, 64)
	if err != nil {
		return ErrInvalidSignature
	}
	if tolerance <= 0 {
		tolerance = DefaultWebhookTolerance
	}
	age := time.Since(time.Unix(ts, 0))
	if age < -tolerance || age > tolerance {
		return ErrExpiredSignature
	}

	signedPayload := fmt.Sprintf("%s.%s", timestampStr, payload)
	for _, secret := range activeSecrets {
		mac := hmac.New(sha256.New, []byte(secret))
		mac.Write([]byte(signedPayload))
		expected := mac.Sum(nil)
		for _, candidate := range sigs {
			if hmac.Equal(candidate, expected) {
				return nil
			}
		}
	}
	return ErrInvalidSignature
}

