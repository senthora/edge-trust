package cloudflare

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"go.uber.org/zap"
	"senthora.com/edge-trust/internal/network"
)

type Client struct {
	logger     *zap.Logger
	apiURL     string
	retries    []time.Duration
	httpClient *http.Client
}

type apiResponse struct {
	Result struct {
		ETag string   `json:"etag"`
		IPv4 []string `json:"ipv4_cidrs"`
		IPv6 []string `json:"ipv6_cidrs"`
	} `json:"result"`
}

type IPData struct {
	ETag  string
	CIDRs []string
}

// NewClient creates and returns a configured Cloudflare API
// client used to fetch Cloudflare IP ranges with retry support.
func NewClient(
	logger *zap.Logger,
	apiURL string,
	retries []time.Duration,
	httpClient *http.Client,
) *Client {
	return &Client{
		logger:     logger,
		apiURL:     apiURL,
		retries:    retries,
		httpClient: httpClient,
	}
}

// FetchIPs fetches and returns Cloudflare IPv4 and IPv6 CIDR ranges.
// Returns nil and an error if the request, response validation,
// response decoding, or retry process fails.
func (c *Client) FetchIPs(ctx context.Context) (IPData, error) {
	var result IPData

	err := network.Run(ctx, c.logger, c.retries, func() error {
		req, err := http.NewRequestWithContext(
			ctx,
			http.MethodGet,
			c.apiURL,
			nil,
		)
		resp, err := c.httpClient.Do(req)
		if err != nil {
			return fmt.Errorf("request cloudflare API: %w", err)
		}
		defer func() {
			if err := resp.Body.Close(); err != nil {
				c.logger.Warn("failed to close response body", zap.Error(err))
			}
		}()
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf(
				"unexpected cloudflare API status %d",
				resp.StatusCode,
			)
		}
		var data apiResponse

		if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
			return fmt.Errorf("decode cloudflare API response: %w", err)
		}
		if strings.TrimSpace(data.Result.ETag) == "" {
			return fmt.Errorf("cloudflare API returned empty etag")
		}
		ipLen := len(data.Result.IPv4) + len(data.Result.IPv6)
		cidrs := make([]string, 0, ipLen)

		cidrs = append(cidrs, data.Result.IPv4...)
		cidrs = append(cidrs, data.Result.IPv6...)

		result = IPData{
			ETag:  data.Result.ETag,
			CIDRs: cidrs,
		}
		return nil
	})
	if err != nil {
		return IPData{}, fmt.Errorf("fetch cloudflare IPs: %w", err)
	}
	return result, nil
}

func (c *Client) APIURL() string {
	return c.apiURL
}
