package fsutil

import (
	"fmt"
	"os"
	"path/filepath"

	"go.uber.org/zap"
)

// WriteFileAtomic writes content to path using a temporary file and atomic rename.
// Note that the parent directory must already exist or an error will be thrown.
func WriteFileAtomic(logger *zap.Logger, path string, content []byte) error {
	dirPath := filepath.Dir(path)
	tmpFilename := fmt.Sprintf(".%s-*.tmp", filepath.Base(path))
	tmpFile, err := os.CreateTemp(dirPath, tmpFilename)
	if err != nil {
		return fmt.Errorf(
			"create temp file in %q for %q: %w",
			dirPath,
			path,
			err,
		)
	}
	tmpPath := tmpFile.Name()
	defer func() {
		if tmpPath == "" {
			return
		}
		if err := os.Remove(tmpPath); err != nil {
			logger.Error("failed to remove tmp file",
				zap.String("path", tmpPath),
				zap.Error(err))
		}
	}()
	if _, err := tmpFile.Write(content); err != nil {
		if cerr := tmpFile.Close(); cerr != nil {
			logger.Error("failed to close tmp file after write error",
				zap.String("path", tmpPath),
				zap.Error(cerr))
		}
		return fmt.Errorf("write temp file %q: %w", tmpPath, err)
	}
	if err := tmpFile.Sync(); err != nil {
		if cerr := tmpFile.Close(); cerr != nil {
			logger.Error("failed to close tmp file after sync error",
				zap.String("path", tmpPath),
				zap.Error(cerr))
		}
		return fmt.Errorf("sync temp file %q: %w", tmpPath, err)
	}
	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("close temp file %q: %w", tmpPath, err)
	}
	if err := os.Rename(tmpPath, path); err != nil {
		return fmt.Errorf(
			"rename temp file %q to %q: %w",
			tmpPath,
			path,
			err,
		)
	}
	// avoid errors on deferred file removal
	tmpPath = ""
	return nil
}
