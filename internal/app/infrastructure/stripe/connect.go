package stripe

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

func (c *httpClient) CreateConnectedAccount(ctx context.Context, slug, name, email string) (string, error) {
	form := url.Values{}
	form.Set("type", c.cfg.ConnectAccountType)
	form.Set("country", c.cfg.ConnectCountry)
	if email != "" {
		form.Set("email", email)
	}
	if c.cfg.ConnectAccountType == "express" {
		form.Set("capabilities[transfers][requested]", "true")
	}
	form.Set("metadata[tenant_slug]", slug)
	form.Set("metadata[tenant_name]", name)

	var res struct {
		ID string `json:"id"`
	}
	if err := c.doRequest(ctx, http.MethodPost, "/v1/accounts", form, "", &res); err != nil {
		return "", err
	}
	if res.ID == "" {
		return "", fmt.Errorf("stripe returned empty connected account id")
	}
	return res.ID, nil
}

func (c *httpClient) GetAccount(ctx context.Context, accountID string) (*AccountStatus, error) {
	var raw struct {
		ID               string            `json:"id"`
		ChargesEnabled   bool              `json:"charges_enabled"`
		PayoutsEnabled   bool              `json:"payouts_enabled"`
		DetailsSubmitted bool              `json:"details_submitted"`
		Metadata         map[string]string `json:"metadata"`
		Requirements     struct {
			CurrentlyDue   []string `json:"currently_due"`
			DisabledReason string   `json:"disabled_reason"`
		} `json:"requirements"`
	}
	path := "/v1/accounts/" + url.PathEscape(accountID)
	if err := c.doRequest(ctx, http.MethodGet, path, nil, "", &raw); err != nil {
		return nil, err
	}
	due := raw.Requirements.CurrentlyDue
	if due == nil {
		due = []string{}
	}
	return &AccountStatus{
		ID:               raw.ID,
		ChargesEnabled:   raw.ChargesEnabled,
		PayoutsEnabled:   raw.PayoutsEnabled,
		DetailsSubmitted: raw.DetailsSubmitted,
		CurrentlyDue:     due,
		DisabledReason:   raw.Requirements.DisabledReason,
		Metadata:         raw.Metadata,
	}, nil
}

func (c *httpClient) CreateAccountLink(ctx context.Context, accountID, refreshURL, returnURL string) (string, error) {
	form := url.Values{}
	form.Set("account", accountID)
	form.Set("refresh_url", refreshURL)
	form.Set("return_url", returnURL)
	form.Set("type", "account_onboarding")

	var res struct {
		URL string `json:"url"`
	}
	if err := c.doRequest(ctx, http.MethodPost, "/v1/account_links", form, "", &res); err != nil {
		return "", err
	}
	if res.URL == "" {
		return "", fmt.Errorf("stripe returned empty account link url")
	}
	return res.URL, nil
}

func (c *httpClient) CreateExpressLoginLink(ctx context.Context, accountID string) (string, error) {
	path := fmt.Sprintf("/v1/accounts/%s/login_links", url.PathEscape(accountID))
	var res struct {
		URL string `json:"url"`
	}
	if err := c.doRequest(ctx, http.MethodPost, path, url.Values{}, "", &res); err != nil {
		return "", err
	}
	if res.URL == "" {
		return "", fmt.Errorf("stripe returned empty login link url")
	}
	return res.URL, nil
}

