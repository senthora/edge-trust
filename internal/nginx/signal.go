package nginx

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"go.uber.org/zap"
)

var ErrSignalPathIsDirectory = errors.New("signal path is a directory")

// EmitReloadSignal creates or updates the nginx reload signal file.
// The file is used to notify nginx container/process that it should reload.
func EmitReloadSignal(logger *zap.Logger, path string) error {
	dirPath := filepath.Dir(path)
	if !dirExists(dirPath) {
		logger.Debug("creating signal directory", zap.String("path", dirPath))
		err := os.MkdirAll(dirPath, 0755)
		if err != nil {
			return fmt.Errorf("create signal directory %q: %w", dirPath, err)
		}
	}
	info, err := os.Stat(path)
	if err == nil && info.IsDir() {
		return fmt.Errorf("%w: %s", ErrSignalPathIsDirectory, path)
	}
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("stat reload signal path %q: %w", path, err)
	}
	now := time.Now()

	// try updating timestamps first
	err = os.Chtimes(path, now, now)
	if err == nil {
		logger.Debug("updated nginx reload signal timestamp", zap.String("path", path))
		return nil
	}
	// if file doesn't exist, create it
	if os.IsNotExist(err) {
		f, err := os.Create(path)
		if err != nil {
			return fmt.Errorf("create reload signal file %q: %w", path, err)
		}
		return f.Close()
	}
	return fmt.Errorf("update reload signal timestamps %q: %w", path, err)
}
