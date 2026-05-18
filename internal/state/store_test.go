package state

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"senthora.com/edge-trust/internal/testutil"
)

func TestLoad_ReturnsEmptyStateWhenFileDoesNotExist(t *testing.T) {
	path := filepath.Join(t.TempDir(), "missing-state.json")

	s, err := Load(path)

	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if s.HasData() {
		t.Fatalf("expected empty state, got %+v", s)
	}
}

func TestLoad_ReturnsErrorWhenFileCannotBeRead(t *testing.T) {
	path := filepath.Join(t.TempDir(), "state.json")

	if err := os.WriteFile(path, []byte("test"), 0000); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}
	_, err := Load(path)

	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func TestLoad_ReturnsErrorWhenFileContainsInvalidJSON(t *testing.T) {
	path := filepath.Join(t.TempDir(), "state.json")

	if err := os.WriteFile(path, []byte("{invalid-json"), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}
	_, err := Load(path)

	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func TestLoad_ReturnsErrorWhenFileContainsInvalidState(t *testing.T) {
	path := filepath.Join(t.TempDir(), "state.json")

	data := []byte(`{
	  "source_url": "",
	  "cidrs": [],
	  "hash": "",
	  "written_at": "0001-01-01T00:00:00Z"
	}`)
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}
	_, err := Load(path)

	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func TestLoad_ReturnsStateWhenFileContainsValidState(t *testing.T) {
	path := filepath.Join(t.TempDir(), "state.json")
	expected := validState()

	data, err := json.MarshalIndent(expected, "", "  ")
	if err != nil {
		t.Fatalf("failed to marshal test state: %v", err)
	}
	data = append(data, '\n')

	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}
	actual, err := Load(path)

	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if !expected.Equals(actual) {
		t.Fatalf("expected %v, got %v", expected, actual)
	}
}

func TestSave_ReturnsErrorWhenStateIsInvalid(t *testing.T) {
	path := filepath.Join(t.TempDir(), "state.json")
	log := testutil.NewLogger()

	err := Save(log, path, State{})

	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func TestSave_ReturnsErrorWhenParentDirectoryCannotBeCreated(t *testing.T) {
	basePath := filepath.Join(t.TempDir(), "blocked")
	log := testutil.NewLogger()

	if err := os.WriteFile(basePath, []byte("file"), 0644); err != nil {
		t.Fatalf("failed to create blocking file: %v", err)
	}
	path := filepath.Join(basePath, "state.json")
	s := validState()

	err := Save(log, path, s)

	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func TestSave_ReturnsErrorWhenFileCannotBeWritten(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "state.json")
	log := testutil.NewLogger()

	if err := os.Mkdir(path, 0755); err != nil {
		t.Fatalf("failed to create blocking directory: %v", err)
	}
	s := validState()

	err := Save(log, path, s)

	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func TestSave_WritesStateSuccessfully(t *testing.T) {
	path := filepath.Join(t.TempDir(), "state.json")
	log := testutil.NewLogger()

	expected := validState()

	if err := Save(log, path, expected); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	actual, err := Load(path)

	if err != nil {
		t.Fatalf("failed to load saved state: %v", err)
	}
	if !expected.Equals(actual) {
		t.Fatalf("expected %v, got %v", expected, actual)
	}
}
