package stripe

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"strconv"
)

func (c *httpClient) CreateTransfer(
	ctx context.Context,
	destinationAccount string,
	amountTHB float64,
	idempotencyKey string,
	metadata map[string]string,
) (*TransferResult, error) {
	satang := int64(math.Round(amountTHB * 100))
	if satang < 100 {
		return nil, fmt.Errorf("transfer amount must be at least 1.00 THB")
	}
	form := url.Values{}
	form.Set("amount", strconv.FormatInt(satang, 10))
	form.Set("currency", "thb")
	form.Set("destination", destinationAccount)
	for k, v := range metadata {
		if v != "" {
			form.Set(fmt.Sprintf("metadata[%s]", k), v)
		}
	}

	var res TransferResult
	if err := c.doRequest(ctx, http.MethodPost, "/v1/transfers", form, idempotencyKey, &res); err != nil {
		return nil, err
	}
	if res.ID == "" {
		return nil, fmt.Errorf("stripe returned empty transfer id")
	}
	return &res, nil
}

func (c *httpClient) CreateCheckoutSession(ctx context.Context, params CheckoutSessionParams) (*CheckoutSession, error) {
	satang := int64(math.Round(params.AmountTHB * 100))
	if satang < 1000 {
		return nil, fmt.Errorf("top-up amount must be at least 10.00 THB")
	}
	form := url.Values{}
	form.Set("mode", "payment")
	form.Set("success_url", params.SuccessURL)
	form.Set("cancel_url", params.CancelURL)
	form.Set("client_reference_id", params.TransactionID)
	form.Set("line_items[0][price_data][currency]", "thb")
	form.Set("line_items[0][price_data][product_data][name]", "BidKan Wallet Top-up")
	form.Set("line_items[0][price_data][unit_amount]", strconv.FormatInt(satang, 10))
	form.Set("line_items[0][quantity]", "1")
	form.Set("metadata[transaction_id]", params.TransactionID)
	form.Set("metadata[user_id]", params.UserID)

	if params.CustomerEmail != "" {
		form.Set("customer_email", params.CustomerEmail)
	}
	switch params.Method {
	case "card":
		form.Set("payment_method_types[0]", "card")
	case "promptpay":
		form.Set("payment_method_types[0]", "promptpay")
	default:
		form.Set("payment_method_types[0]", "card")
		form.Set("payment_method_types[1]", "promptpay")
	}

	var res CheckoutSession
	if err := c.doRequest(ctx, http.MethodPost, "/v1/checkout/sessions", form, "checkout-"+params.TransactionID, &res); err != nil {
		return nil, err
	}
	if res.ID == "" || res.URL == "" {
		return nil, fmt.Errorf("stripe returned incomplete checkout session")
	}
	return &res, nil
}

func (c *httpClient) GetCheckoutSession(ctx context.Context, sessionID string) (*CheckoutSession, error) {
	var res CheckoutSession
	path := "/v1/checkout/sessions/" + url.PathEscape(sessionID)
	if err := c.doRequest(ctx, http.MethodGet, path, nil, "", &res); err != nil {
		return nil, err
	}
	return &res, nil
}

