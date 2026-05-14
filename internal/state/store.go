package state

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"senthora.com/edge-trust/internal/fsutil"
)

// Load reads and validates State from a JSON file at path.
// Returns the loaded State. However, if the file does
// not exist, it returns an empty State and no error.
func Load(path string) (State, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return State{}, nil
		}
		return State{}, fmt.Errorf("read state file %q: %w", path, err)
	}
	var s State
	if err := json.Unmarshal(data, &s); err != nil {
		return State{}, fmt.Errorf("invalid state file: %w", err)
	}
	if err := s.Validate(); err != nil {
		return State{}, fmt.Errorf("invalid state: %w", err)
	}
	return s, nil
}

// Save validates and writes State to a JSON file at path.
// Returns an error if validation, serialization,
// directory creation, or file writing fails.
func Save(logger *zap.Logger, path string, state State) error {
	if err := state.Validate(); err != nil {
		return fmt.Errorf("invalid state: %w", err)
	}
	dirPath := filepath.Dir(path)
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return fmt.Errorf("create state directory %q: %w", dirPath, err)
	}
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal state JSON: %w", err)
	}
	data = append(data, '\n')
	if err := fsutil.WriteFileAtomic(logger, path, data); err != nil {
		return fmt.Errorf("write state file %q: %w", path, err)
	}
	return nil
}
