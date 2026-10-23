// Package evcc is a minimal client for the subset of EVCC's REST API that
// Ladegast needs: reading charging sessions and switching a vehicle's
// charge mode. See DESIGN.md section 3 for the endpoints this was
// verified against.
package evcc

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// Mode is a charge mode Ladegast is allowed to set on a vehicle. EVCC
// itself also knows "off", but Ladegast's MVP only ever offers these two.
type Mode string

const (
	ModeSmart Mode = "smart" // PV surplus charging
	ModeNow   Mode = "now"   // fast charging regardless of PV surplus
)

func (m Mode) valid() bool {
	return m == ModeSmart || m == ModeNow
}

type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

func NewClient(baseURL, apiKey string) *Client {
	return &Client{
		baseURL:    strings.TrimSuffix(baseURL, "/"),
		apiKey:     apiKey,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

// Sessions returns charging sessions, optionally filtered by year/month
// (pass 0 to omit a filter).
func (c *Client) Sessions(ctx context.Context, year, month int) ([]Session, error) {
	q := url.Values{}
	if year != 0 {
		q.Set("year", strconv.Itoa(year))
	}
	if month != 0 {
		q.Set("month", strconv.Itoa(month))
	}
	reqURL := c.baseURL + "/api/sessions"
	if len(q) > 0 {
		reqURL += "?" + q.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var sessions []Session
	if err := json.NewDecoder(resp.Body).Decode(&sessions); err != nil {
		return nil, fmt.Errorf("decode sessions response: %w", err)
	}
	return sessions, nil
}

// SetVehicleMode sets the charge mode for the vehicle identified by its
// EVCC config name (not its display title or internal db id).
func (c *Client) SetVehicleMode(ctx context.Context, vehicleName string, mode Mode) error {
	if !mode.valid() {
		return fmt.Errorf("mode %q is not allowed", mode)
	}
	if vehicleName == "" {
		return fmt.Errorf("vehicleName must not be empty")
	}
	reqURL := fmt.Sprintf("%s/api/vehicles/%s/mode/%s", c.baseURL, url.PathEscape(vehicleName), mode)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, nil)
	if err != nil {
		return err
	}
	resp, err := c.do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

func (c *Client) do(req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("evcc request failed: %w", err)
	}
	if resp.StatusCode >= 300 {
		defer resp.Body.Close()
		return nil, fmt.Errorf("evcc returned status %d for %s", resp.StatusCode, req.URL.Path)
	}
	return resp, nil
}
