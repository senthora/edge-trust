package fsutil

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"senthora.com/edge-trust/internal/testutil"
)

func TestWriteFileAtomic_WritesFileSuccessfully(t *testing.T) {
	_, path, content := newTestFile(t)
	log := testutil.NewLogger()

	err := WriteFileAtomic(log, path, content)

	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	assertFileContent(t, path, content)
}

func TestWriteFileAtomic_OverwritesExistingFile(t *testing.T) {
	_, path, initialContent := newTestFile(t)
	log := testutil.NewLogger()

	if err := os.WriteFile(path, initialContent, 0644); err != nil {
		t.Fatalf("failed to write initial file: %v", err)
	}
	newContent := []byte("new content")

	err := WriteFileAtomic(log, path, newContent)

	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	assertFileContent(t, path, newContent)
}

func TestWriteFileAtomic_ReturnsErrorWhenParentDirectoryDoesNotExist(t *testing.T) {
	dir := t.TempDir()
	log := testutil.NewLogger()
	path := filepath.Join(dir, "missing", "test.txt")
	content := []byte("hello world")

	err := WriteFileAtomic(log, path, content)

	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func TestWriteFileAtomic_CleansUpTmpFileAfterSuccessfulWrite(t *testing.T) {
	dir, path, content := newTestFile(t)
	log := testutil.NewLogger()

	err := WriteFileAtomic(log, path, content)

	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	entries, err := os.ReadDir(dir)

	if err != nil {
		t.Fatalf("failed to read directory: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 directory entry, got %d", len(entries))
	}
	if entries[0].Name() != "test.txt" {
		t.Fatalf("expected only test.txt to exist, got %q", entries[0].Name())
	}
}

func TestWriteFileAtomic_CleansUpTmpFileAfterWriteFailure(t *testing.T) {
	dir, path, content := newTestFile(t)
	log := testutil.NewLogger()

	if err := os.Mkdir(path, 0755); err != nil {
		t.Fatalf("failed to create blocking directory: %v", err)
	}
	err := WriteFileAtomic(log, path, content)

	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	entries, err := os.ReadDir(dir)

	if err != nil {
		t.Fatalf("failed to read directory: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 directory entry, got %d", len(entries))
	}
	if entries[0].Name() != "test.txt" {
		t.Fatalf("expected only blocking directory to exist, got %q", entries[0].Name())
	}
}

func assertFileContent(t *testing.T, path string, expected []byte) {
	t.Helper()

	actual, err := os.ReadFile(path)

	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}
	if !bytes.Equal(actual, expected) {
		t.Fatalf("expected content %q, got %q", expected, actual)
	}
}

func newTestFile(t *testing.T) (string, string, []byte) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.txt")
	content := []byte("hello world")

	return dir, path, content
}
