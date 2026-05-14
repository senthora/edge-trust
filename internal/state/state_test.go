package state

import (
	"errors"
	"slices"
	"testing"
	"time"

	"senthora.com/edge-trust/internal/testutil"
)

func TestEquals(t *testing.T) {
	base := validState()

	tests := []struct {
		name  string
		other State
		want  bool
	}{
		{
			name:  "identical states",
			other: base,
			want:  true,
		},
		{
			name: "different source url",
			other: func() State {
				s := base
				s.SourceURL = "https://different.example.com"
				return s
			}(),
			want: false,
		},
		{
			name: "different cidrs",
			other: func() State {
				s := base
				s.CIDRs = []string{"1.1.1.0/24"}
				return s
			}(),
			want: false,
		},
		{
			name: "different hash",
			other: func() State {
				s := base
				s.Hash = "sha256:different"
				return s
			}(),
			want: false,
		},
		{
			name: "different written at",
			other: func() State {
				s := base
				s.WrittenAt = s.WrittenAt.Add(time.Minute)
				return s
			}(),
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := base.Equals(tt.other)

			if got != tt.want {
				t.Fatalf("expected %v, got %v", tt.want, got)
			}
		})
	}
}

func TestNew_NormalizesCIDRs(t *testing.T) {
	cidrs := []string{
		" 173.245.48.0/20 ",
		"",
		"2400:cb00::/32",
		"173.245.48.0/20",
		"  ",
		"103.21.244.0/22",
	}
	eTag := testutil.SampleETag()
	state := New("", eTag, cidrs, time.Time{})

	expected := []string{
		"103.21.244.0/22",
		"173.245.48.0/20",
		"2400:cb00::/32",
	}
	if !slices.Equal(state.CIDRs, expected) {
		t.Fatalf("expected %v, got %v", expected, state.CIDRs)
	}
}

func TestNew_ComputesStableHash(t *testing.T) {
	cidrsA := []string{
		"173.245.48.0/20",
		"103.21.244.0/22",
		"2400:cb00::/32",
	}
	cidrsB := []string{
		"2400:cb00::/32",
		"103.21.244.0/22",
		"173.245.48.0/20",
	}
	eTag := testutil.SampleETag()
	stateA := New("", eTag, cidrsA, time.Time{})
	stateB := New("", eTag, cidrsB, time.Time{})

	if stateA.Hash == "" {
		t.Fatalf("expected hash to be set")
	}
	if stateA.Hash != stateB.Hash {
		t.Fatalf("expected hashes to match, got %s and %s", stateA.Hash, stateB.Hash)
	}
}

func TestNew_StoresWrittenAtInUTC(t *testing.T) {
	location := time.FixedZone("TEST", 2*60*60)
	now := testutil.FixedTimeNow(location)
	eTag := testutil.SampleETag()

	state := New("", eTag, []string{}, now)

	if state.WrittenAt.Location() != time.UTC {
		t.Fatalf("expected WrittenAt location to be UTC, got %v", state.WrittenAt.Location())
	}
	expected := now.UTC()
	if !state.WrittenAt.Equal(expected) {
		t.Fatalf("expected WrittenAt %v, got %v", expected, state.WrittenAt)
	}
}

func TestValidate_ReturnsErrorWhenSourceURLIsEmpty(t *testing.T) {
	state := validState()
	state.SourceURL = ""

	err := state.Validate()

	if !errors.Is(err, ErrSourceURLRequired) {
		t.Fatalf("expected ErrSourceURLRequired, got %v", err)
	}
}

func TestValidate_ReturnsErrorWhenSourceURLIsInvalid(t *testing.T) {
	state := validState()
	state.SourceURL = "://invalid-url"

	err := state.Validate()

	if !errors.Is(err, ErrSourceURLInvalid) {
		t.Fatalf("expected ErrSourceURLInvalid, got %v", err)
	}
}

func TestValidate_ReturnsErrorWhenSourceURLDoesNotContainHost(t *testing.T) {
	state := validState()
	state.SourceURL = "https://"

	err := state.Validate()

	if !errors.Is(err, ErrSourceURLMissingHost) {
		t.Fatalf("expected ErrSourceURLMissingHost, got %v", err)
	}
}

func TestValidate_ReturnsErrorWhenSourceURLUsesUnsupportedScheme(t *testing.T) {
	state := validState()
	state.SourceURL = "ftp://example.com"

	err := state.Validate()

	if !errors.Is(err, ErrSourceURLInvalidScheme) {
		t.Fatalf("expected ErrSourceURLInvalidScheme, got %v", err)
	}
}

func TestValidate_ReturnsErrorWhenETagIsEmpty(t *testing.T) {
	state := validState()
	state.ETag = ""

	err := state.Validate()

	if !errors.Is(err, ErrETagRequired) {
		t.Fatalf("expected ErrETagRequired, got %v", err)
	}
}

func TestValidate_ReturnsErrorWhenCIDRListIsEmpty(t *testing.T) {
	state := validState()
	state.CIDRs = []string{}

	err := state.Validate()

	if !errors.Is(err, ErrCIDRListRequired) {
		t.Fatalf("expected ErrCIDRListRequired, got %v", err)
	}
}

func TestValidate_ReturnsErrorWhenCIDRListContainsInvalidCIDR(t *testing.T) {
	state := validState()
	state.CIDRs = []string{"invalid-cidr"}

	err := state.Validate()

	if !errors.Is(err, ErrCIDRInvalid) {
		t.Fatalf("expected ErrCIDRInvalid, got %v", err)
	}
}

func TestValidate_RejectsNonCanonicalCIDRs(t *testing.T) {
	tests := []struct {
		name  string
		cidrs []string
	}{
		{
			name: "duplicate entries",
			cidrs: []string{
				"103.21.244.0/22",
				"103.21.244.0/22",
			},
		},
		{
			name: "unsorted entries",
			cidrs: []string{
				"173.245.48.0/20",
				"103.21.244.0/22",
			},
		},
		{
			name: "untrimmed entries",
			cidrs: []string{
				" 103.21.244.0/22 ",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := validState()
			state.CIDRs = tt.cidrs

			err := state.Validate()

			if !errors.Is(err, ErrCIDRListNotNormalized) {
				t.Fatalf("expected ErrCIDRListNotNormalized, got %v", err)
			}
		})
	}
}

func TestValidate_ReturnsErrorWhenHashIsEmpty(t *testing.T) {
	state := validState()
	state.Hash = ""

	err := state.Validate()

	if !errors.Is(err, ErrHashRequired) {
		t.Fatalf("expected ErrHashRequired, got %v", err)
	}
}

func TestValidate_ReturnsErrorWhenHashIsWhitespace(t *testing.T) {
	state := validState()
	state.Hash = "  "

	err := state.Validate()

	if !errors.Is(err, ErrHashRequired) {
		t.Fatalf("expected ErrHashRequired, got %v", err)
	}
}

func TestValidate_ReturnsErrorWhenHashDoesNotMatchCIDRList(t *testing.T) {
	state := validState()
	state.CIDRs = []string{"103.21.244.0/22"}
	state.Hash = "sha256:invalid"

	err := state.Validate()

	if !errors.Is(err, ErrHashMismatch) {
		t.Fatalf("expected ErrHashMismatch, got %v", err)
	}
}

func TestValidate_ReturnsErrorWhenWrittenAtIsZero(t *testing.T) {
	state := validState()
	state.WrittenAt = time.Time{}

	err := state.Validate()

	if !errors.Is(err, ErrWrittenAtRequired) {
		t.Fatalf("expected ErrWrittenAtRequired, got %v", err)
	}
}

func TestValidate_ReturnsErrorWhenWrittenAtIsInFuture(t *testing.T) {
	state := validState()
	state.WrittenAt = time.Now().Add(1 * time.Hour)

	err := state.Validate()

	if !errors.Is(err, ErrWrittenAtInFuture) {
		t.Fatalf("expected ErrWrittenAtInFuture, got %v", err)
	}
}

func TestValidate_ReturnsNilWhenStateIsValid(t *testing.T) {
	state := validState()

	err := state.Validate()

	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func TestHasData_ReturnsFalseWhenStateContainsNoData(t *testing.T) {
	state := State{}

	if state.HasData() {
		t.Fatalf("expected HasData to return false")
	}
}

func TestHasData_ReturnsTrueWhenStateContainsData(t *testing.T) {
	state := validState()

	if !state.HasData() {
		t.Fatalf("expected HasData to return true")
	}
}

func validState() State {
	cidrs := []string{
		"103.21.244.0/22",
		"173.245.48.0/20",
	}
	return State{
		SourceURL: "https://example.com",
		ETag:      testutil.SampleETag(),
		CIDRs:     cidrs,
		Hash:      computeHash(cidrs),
		WrittenAt: testutil.FixedTimeNowUTC(),
	}
}
