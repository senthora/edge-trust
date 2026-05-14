package health

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

var (
	ErrCreatePulseDirectory = errors.New("create pulse directory")
	ErrHealthPulseStale     = errors.New("health pulse stale")
)

type Service struct {
	SignalPath    string
	PulseInterval time.Duration
	SignalMaxAge  time.Duration
}

func NewService(pulseInterval time.Duration, signalPath string) (*Service, error) {
	dir := filepath.Dir(signalPath)
	if err := os.MkdirAll(dir, 0744); err != nil {
		return nil, ErrCreatePulseDirectory
	}
	file, err := os.Create(signalPath)
	if err != nil {
		return nil, fmt.Errorf("create pulse file: %w", err)
	}
	if err := file.Close(); err != nil {
		return nil, fmt.Errorf("close pulse file: %w", err)
	}
	return &Service{
		SignalPath:    signalPath,
		PulseInterval: pulseInterval,
		SignalMaxAge:  4 * pulseInterval,
	}, nil
}

func (s *Service) EmitPulse(now time.Time) error {
	if _, err := os.Stat(s.SignalPath); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("stat health signal file: %w", err)
		}
		file, err := os.Create(s.SignalPath)
		if err != nil {
			return fmt.Errorf("create health signal file: %w", err)
		}
		if err := file.Close(); err != nil {
			return fmt.Errorf("close health signal file: %w", err)
		}
	}
	if err := os.Chtimes(s.SignalPath, now, now); err != nil {
		return fmt.Errorf("update health signal timestamp: %w", err)
	}
	return nil
}

func (s *Service) HealthCheck(now time.Time) error {
	pulseAge, err := s.pulseAge(now)
	if err != nil {
		return fmt.Errorf("get pulse age: %w", err)
	}
	if s.isPulseStale(pulseAge) {
		return fmt.Errorf(
			"%w: age %s exceeds max age %s",
			ErrHealthPulseStale,
			pulseAge.Round(time.Second),
			s.SignalMaxAge,
		)
	}
	return nil
}

func (s *Service) pulseAge(now time.Time) (time.Duration, error) {
	info, err := os.Stat(s.SignalPath)
	if err != nil {
		return time.Duration(0), fmt.Errorf("stat health pulse file: %w", err)
	}
	lastPulse := info.ModTime()
	return now.Sub(lastPulse), nil
}

func (s *Service) isPulseStale(pulseAge time.Duration) bool {
	return pulseAge > s.SignalMaxAge
}
