package nginx

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"senthora.com/edge-trust/internal/testutil"
)

func TestEmitReloadSignal_CreatesFileWhenNotExists(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nginx.reload")
	log := testutil.NewLogger()

	err := EmitReloadSignal(log, path)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("expected file to exist, got error: %v", err)
	}
	if info.IsDir() {
		t.Fatalf("expected a file, got directory")
	}
}

func TestEmitReloadSignal_CreatesParentDirectoryWhenMissing(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nginx.reload")
	log := testutil.NewLogger()

	err := EmitReloadSignal(log, path)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	parentDir := filepath.Dir(path)
	info, err := os.Stat(parentDir)
	if err != nil {
		t.Fatalf("expected parent directory to exist, got error: %v", err)
	}
	if !info.IsDir() {
		t.Fatalf("expected parent path to be a directory")
	}
	fileInfo, err := os.Stat(path)
	if err != nil {
		t.Fatalf("expected signal file to exist, got error: %v", err)
	}
	if fileInfo.IsDir() {
		t.Fatalf("expected a file, got directory")
	}
}

func TestEmitReloadSignal_UpdatesTimestampWhenFileExists(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nginx.reload")
	log := testutil.NewLogger()

	// create the file
	if err := os.WriteFile(path, []byte(""), 0644); err != nil {
		t.Fatalf("failed to create initial file: %v", err)
	}
	// set old timestamp
	oldTime := testutil.FixedTimeNowUTC().Add(-1 * time.Hour)
	if err := os.Chtimes(path, oldTime, oldTime); err != nil {
		t.Fatalf("failed to set old timestamp: %v", err)
	}
	// capture before
	beforeInfo, err := os.Stat(path)
	if err != nil {
		t.Fatalf("failed to stat file: %v", err)
	}
	beforeModTime := beforeInfo.ModTime()

	err = EmitReloadSignal(log, path)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	afterInfo, err := os.Stat(path)
	if err != nil {
		t.Fatalf("failed to stat file after: %v", err)
	}
	afterModTime := afterInfo.ModTime()
	if !afterModTime.After(beforeModTime) {
		t.Fatalf("expected mod time to be updated, before=%v after=%v", beforeModTime, afterModTime)
	}
}

func TestEmitReloadSignal_ReturnsErrorWhenPathIsDirectory(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nginx.reload")
	log := testutil.NewLogger()

	if err := os.Mkdir(path, 0755); err != nil {
		t.Fatalf("failed to create directory: %v", err)
	}
	err := EmitReloadSignal(log, path)

	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}
