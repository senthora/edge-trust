package health

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"senthora.com/edge-trust/internal/testutil"
)

func TestNewService_CreatesDirectoryAndSignalFile(t *testing.T) {
	t.Parallel()

	signalPath := tmpSignalPath(t)
	pulseInterval := 5 * time.Second

	_, err := NewService(pulseInterval, signalPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	info, err := os.Stat(signalPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.IsDir() {
		t.Fatal("expected signal path to be a file")
	}
}

func TestNewService_ReturnsInitializedService(t *testing.T) {
	t.Parallel()

	signalPath := tmpSignalPath(t)
	pulseInterval := 5 * time.Second

	hs, err := NewService(pulseInterval, signalPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if hs.SignalPath != signalPath {
		t.Fatalf("expected SignalPath %q, got %q", signalPath, hs.SignalPath)
	}
	if hs.PulseInterval != pulseInterval {
		t.Fatalf("expected PulseInterval %s, got %s", pulseInterval, hs.PulseInterval)
	}
	expectedMaxAge := 4 * pulseInterval

	if hs.SignalMaxAge != expectedMaxAge {
		t.Fatalf("expected SignalMaxAge %s, got %s", expectedMaxAge, hs.SignalMaxAge)
	}
}

func TestNewService_ReturnsErrorWhenPulseDirectoryCannotBeCreated(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	signalPath := filepath.Join(tmpDir, "blocked", ".alive")
	pulseInterval := 5 * time.Second

	if err := os.WriteFile(
		filepath.Join(tmpDir, "blocked"),
		[]byte("not a directory"),
		0600,
	); err != nil {
		t.Fatalf("failed to create blocking file: %v", err)
	}
	_, err := NewService(pulseInterval, signalPath)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, ErrCreatePulseDirectory) {
		t.Fatalf("expected ErrCreatePulseDirectory error, got %v", err)
	}
}

func TestEmitPulse_CreatesMissingSignalFileAndUpdatesTimestamp(t *testing.T) {
	t.Parallel()

	signalPath := tmpSignalPath(t)
	now := testutil.FixedTimeNowUTC()

	hs := &Service{
		SignalPath: signalPath,
	}
	if err := os.MkdirAll(filepath.Dir(signalPath), 0700); err != nil {
		t.Fatalf("failed to create directory: %v", err)
	}
	if err := hs.EmitPulse(now); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	info, err := os.Stat(signalPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.IsDir() {
		t.Fatal("expected signal path to be a file")
	}
	if !info.ModTime().Equal(now) {
		t.Fatalf("expected mod time %s, got %s", now, info.ModTime())
	}
}

func TestEmitPulse_UpdatesTimestampOfExistingSignalFile(t *testing.T) {
	t.Parallel()

	signalPath := tmpSignalPath(t)
	oldTime := testutil.FixedTimeNowUTC()
	newTime := oldTime.Add(1 * time.Hour)

	hs := initHealthService(t, 1*time.Second, signalPath)

	if err := os.Chtimes(signalPath, oldTime, oldTime); err != nil {
		t.Fatalf("failed to set initial mod time: %v", err)
	}
	if err := hs.EmitPulse(newTime); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	info, err := os.Stat(signalPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !info.ModTime().Equal(newTime) {
		t.Fatalf("expected mod time %s, got %s", newTime, info.ModTime())
	}
}

func TestHealthCheck_ReturnsNilWhenPulseFileIsFresh(t *testing.T) {
	t.Parallel()

	signalPath := tmpSignalPath(t)
	hs := initHealthService(t, 1*time.Second, signalPath)

	now := testutil.FixedTimeNowUTC()
	freshPulse := now.Add(-hs.SignalMaxAge + time.Second)

	if err := os.Chtimes(signalPath, freshPulse, freshPulse); err != nil {
		t.Fatalf("failed to set pulse timestamp: %v", err)
	}
	if err := hs.HealthCheck(now); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func TestHealthCheck_ReturnsErrorWhenPulseFileIsStale(t *testing.T) {
	t.Parallel()

	signalPath := tmpSignalPath(t)
	hs := initHealthService(t, 1*time.Second, signalPath)

	now := testutil.FixedTimeNowUTC()
	stalePulse := now.Add(-(hs.SignalMaxAge + time.Second))

	if err := os.Chtimes(signalPath, stalePulse, stalePulse); err != nil {
		t.Fatalf("failed to set pulse timestamp: %v", err)
	}
	err := hs.HealthCheck(now)
	if !errors.Is(err, ErrHealthPulseStale) {
		t.Fatalf("expected ErrHealthPulseStale, got %v", err)
	}
}

func TestHealthCheck_ReturnsErrorWhenPulseFileDoesNotExist(t *testing.T) {
	t.Parallel()

	hs := &Service{
		SignalPath:   tmpSignalPath(t),
		SignalMaxAge: 4 * time.Second,
	}
	now := testutil.FixedTimeNowUTC()

	err := hs.HealthCheck(now)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func tmpSignalPath(t *testing.T) string {
	return filepath.Join(t.TempDir(), "var/run/edge-trust", ".alive")
}

func initHealthService(t *testing.T, interval time.Duration, signalPath string) *Service {
	hs, err := NewService(interval, signalPath)
	if hs == nil {
		t.Fatal("expected health service to be initialized")
	}
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	return hs
}
