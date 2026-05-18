package cloudflare

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"senthora.com/edge-trust/internal/testutil"
)

var expected = testutil.IPs{
	IPV4: "1.1.1.0/24",
	IPV6: "2606:4700::/32",
}

func TestFetchIPs_Success(t *testing.T) {
	server := httptest.NewServer(newHandler(t))
	defer server.Close()

	client := NewClient(
		testutil.NewLogger(),
		server.URL,
		[]time.Duration{0},
		server.Client(),
	)
	data, err := client.FetchIPs(context.Background())

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertExpectedIPs(t, data, expected)
}

func TestFetchIPsWithRetry_SucceedsAfterRetries(t *testing.T) {
	attempts := 0

	server := httptest.NewServer(newHandlerForRetries(t, 2, &attempts))
	defer server.Close()

	client := NewClient(
		testutil.NewLogger(),
		server.URL,
		[]time.Duration{0, 0, 0},
		server.Client(),
	)
	data, err := client.FetchIPs(context.Background())

	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}
	assertExpectedIPs(t, data, expected)
	if attempts != 3 {
		t.Fatalf("expected 3 attempts, got %d", attempts)
	}
}

func TestFetchIPsWithRetry_FailsAfterAllRetries(t *testing.T) {
	attempts := 0

	// always fail
	server := httptest.NewServer(newHandlerForRetries(t, 999, &attempts))
	defer server.Close()

	client := NewClient(
		testutil.NewLogger(),
		server.URL,
		[]time.Duration{0, 0, 0},
		server.Client(),
	)
	data, err := client.FetchIPs(context.Background())

	if err == nil {
		t.Fatalf("expected error, got nil")
		return
	}
	if len(data.CIDRs) != 0 {
		t.Fatalf("expected no IPs, got %v", data.CIDRs)
	}
	if attempts != 3 {
		t.Fatalf("expected 3 attempts, got %d", attempts)
	}
}

func newHandler(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeResponse(t, w)
	}
}

func newHandlerForRetries(t *testing.T, failures int, attempts *int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		*attempts++

		if *attempts <= failures {
			http.Error(w, "temporary failure", http.StatusInternalServerError)
			return
		}
		writeResponse(t, w)
	}
}

func writeResponse(t *testing.T, w http.ResponseWriter) {
	t.Helper()

	responseJson := testutil.CloudflareResponseJson(expected)
	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write([]byte(responseJson)); err != nil {
		t.Fatalf("failed to write response: %v", err)
	}
}

func assertExpectedIPs(t *testing.T, data IPData, expected testutil.IPs) {
	if len(data.CIDRs) != 2 {
		t.Fatalf("expected 2 IPs, got %d", len(data.CIDRs))
	}
	if data.CIDRs[0] != expected.IPV4 {
		t.Fatalf("unexpected IPv4: %s", data.CIDRs[0])
	}
	if data.CIDRs[1] != expected.IPV6 {
		t.Fatalf("unexpected IPv6: %s", data.CIDRs[1])
	}
}
