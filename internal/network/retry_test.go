package network

import (
	"context"
	"fmt"
	"testing"
	"time"

	"senthora.com/edge-trust/internal/testutil"
)

func TestRun_SucceedsAfterRetries(t *testing.T) {
	attempts := 0
	log := testutil.NewLogger()
	delays := []time.Duration{0, 0, 0}

	operation := func() error {
		attempts++

		if attempts < 3 {
			return fmt.Errorf("fail")
		}
		return nil
	}
	ctx := context.Background()
	err := Run(ctx, log, delays, operation)

	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}
	if attempts != 3 {
		t.Fatalf("expected 3 attempts, got %d", attempts)
	}
}

func TestRun_FailsAfterAllRetries(t *testing.T) {
	attempts := 0
	log := testutil.NewLogger()
	delays := []time.Duration{0, 0, 0}

	operation := func() error {
		attempts++
		return fmt.Errorf("fail")
	}
	ctx := context.Background()
	err := Run(ctx, log, delays, operation)

	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if attempts != 3 {
		t.Fatalf("expected 3 attempts, got %d", attempts)
	}
}
