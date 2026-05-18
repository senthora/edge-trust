package state

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net"
	"net/url"
	"slices"
	"sort"
	"strings"
	"time"
)

var (
	ErrSourceURLRequired      = errors.New("source URL is required")
	ErrSourceURLInvalid       = errors.New("source URL is invalid")
	ErrSourceURLNotAbsolute   = errors.New("source URL must be absolute")
	ErrSourceURLMissingHost   = errors.New("source URL must contain host")
	ErrSourceURLInvalidScheme = errors.New("source URL must use http or https")

	ErrETagRequired = errors.New("etag is required")

	ErrCIDRListRequired      = errors.New("CIDR list is required")
	ErrCIDRInvalid           = errors.New("CIDR list contains invalid CIDR")
	ErrCIDRListNotNormalized = errors.New("CIDR list is not normalized")

	ErrHashRequired = errors.New("hash is required")
	ErrHashMismatch = errors.New("hash does not match CIDR list")

	ErrWrittenAtRequired = errors.New("written at timestamp is required")
	ErrWrittenAtInFuture = errors.New("written_at timestamp cannot be in the future")
)

// State represents the persisted canonical snapshot of synchronized
// Cloudflare CIDR ranges and their synchronization metadata.
type State struct {
	SourceURL string    `json:"source_url"`
	ETag      string    `json:"etag"`
	CIDRs     []string  `json:"cidrs"`
	Hash      string    `json:"hash"`
	WrittenAt time.Time `json:"written_at"`
}

// Equals reports whether this state is structurally equal to another state.
func (s State) Equals(other State) bool {
	return s.SourceURL == other.SourceURL &&
		s.ETag == other.ETag &&
		slices.Equal(s.CIDRs, other.CIDRs) &&
		s.Hash == other.Hash &&
		s.WrittenAt.Equal(other.WrittenAt)
}

// New creates and returns a normalized State from the given source URL and CIDR ranges.
func New(sourceURL string, etag string, cidrs []string, now time.Time) State {
	normalized := normalizeCIDRs(cidrs)
	hash := computeHash(normalized)

	return State{
		SourceURL: sourceURL,
		ETag:      etag,
		CIDRs:     normalized,
		Hash:      hash,
		WrittenAt: now.UTC(),
	}
}

// Validate verifies that the state contains a valid source URL,
// normalized CIDRs, a matching hash, and a valid timestamp.
// Returns nil or an error if any validation rule is violated.
func (s State) Validate() error {
	// validate source URL
	rawURL := strings.TrimSpace(s.SourceURL)
	if rawURL == "" {
		return ErrSourceURLRequired
	}
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return ErrSourceURLInvalid
	}
	if !parsedURL.IsAbs() {
		return ErrSourceURLNotAbsolute
	}
	if parsedURL.Host == "" {
		return ErrSourceURLMissingHost
	}
	switch parsedURL.Scheme {
	case "http", "https":
	default:
		return ErrSourceURLInvalidScheme
	}
	// validate etag presence
	if strings.TrimSpace(s.ETag) == "" {
		return ErrETagRequired
	}
	// validate CIDR list presence
	if len(s.CIDRs) == 0 {
		return ErrCIDRListRequired
	}
	// validate CIDR normalization invariants
	normalizedCIDRs := normalizeCIDRs(s.CIDRs)

	if !slices.Equal(normalizedCIDRs, s.CIDRs) {
		return ErrCIDRListNotNormalized
	}
	// validate CIDR syntax
	for _, cidr := range s.CIDRs {
		if _, _, err := net.ParseCIDR(cidr); err != nil {
			return ErrCIDRInvalid
		}
	}
	// validate hash presence
	if strings.TrimSpace(s.Hash) == "" {
		return ErrHashRequired
	}
	// validate hash integrity against CIDRs
	expectedHash := computeHash(s.CIDRs)

	if s.Hash != expectedHash {
		return ErrHashMismatch
	}
	// validate timestamp presence
	if s.WrittenAt.IsZero() {
		return ErrWrittenAtRequired
	}
	// validate timestamp is not in the future
	if s.WrittenAt.After(time.Now()) {
		return ErrWrittenAtInFuture
	}
	return nil
}

// HasData reports whether the state
// contains persisted synchronization data.
func (s State) HasData() bool {
	return s.SourceURL != "" ||
		s.ETag != "" ||
		len(s.CIDRs) > 0 ||
		s.Hash != "" ||
		!s.WrittenAt.IsZero()
}

// NormalizeCIDRs returns a sorted, deduplicated copy of the given CIDR ranges.
func normalizeCIDRs(cidrs []string) []string {
	seen := make(map[string]struct{}, len(cidrs))
	out := make([]string, 0, len(cidrs))

	for _, c := range cidrs {
		c = strings.TrimSpace(c)
		if c == "" {
			continue
		}
		if _, ok := seen[c]; ok {
			continue
		}
		seen[c] = struct{}{}
		out = append(out, c)
	}
	sort.Strings(out)
	return out
}

// ComputeHash returns a stable SHA-256 hash for the given CIDR ranges.
// The input must be normalized to ensure consistent and comparable results.
func computeHash(cidrs []string) string {
	h := sha256.New()

	// join with a delimiter to avoid ambiguity
	data := strings.Join(cidrs, ",")

	h.Write([]byte(data))

	return "sha256:" + hex.EncodeToString(h.Sum(nil))
}
