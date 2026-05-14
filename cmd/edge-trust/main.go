package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"
	"senthora.com/edge-trust/internal/cloudflare"
	"senthora.com/edge-trust/internal/config"
	"senthora.com/edge-trust/internal/health"
	"senthora.com/edge-trust/internal/state"
	"senthora.com/edge-trust/internal/updater"
)

func main() {
	f := config.LoadFlags()
	logger := config.NewLogger(f.DebugLogLevel)
	defer func() {
		_ = logger.Sync()
	}()
	ev := config.LoadEnv()
	rawCommand := flag.Arg(0)
	if rawCommand == "" {
		flag.Usage()
		os.Exit(0)
	}
	command := config.ParseCommand(rawCommand)
	if command == nil {
		_, _ = fmt.Fprintf(
			os.Stderr,
			"edge-trust: %s\n\nRun 'edge-trust --help' for more information\n",
			fmt.Errorf("unknown command: %q", rawCommand),
		)
		os.Exit(1)
	}
	if *command == config.CommandRun {
		if err := run(logger, ev, f); err != nil {
			logger.Error("run failed", zap.Error(err))
			os.Exit(1)
		}
	}
	if *command == config.CommandHealthcheck {
		hs := initHealthService(f.HBInterval, ev.HCSignalPath)
		if err := hs.HealthCheck(time.Now()); err != nil {
			_, _ = fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}
}

func run(logger *zap.Logger, ev config.EnvVars, f config.Flags) error {
	delays := []time.Duration{
		0,
		3 * time.Second,
		5 * time.Second,
		10 * time.Second,
	}
	client := cloudflare.NewClient(logger, ev.APIURL, delays, http.DefaultClient)
	configPaths := updater.ConfigPaths{
		ProxySourcesPath:    ev.ProxySourcesPath,
		OriginAllowlistPath: ev.OriginAllowlistPath,
		StateJSONPath:       ev.StateJSONPath,
		ReloadSignalPath:    ev.ReloadSignalPath,
	}
	u := updater.NewUpdater(client, configPaths)
	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	logger.Info("starting edge-trust updater")

	if f.Daemon {
		logger.Info(
			"running in daemon mode",
			zap.Duration("interval", f.Interval),
		)
		hs := initHealthService(f.Interval, ev.HCSignalPath)
		return runDaemon(ctx, logger, u, ev, hs, f.Interval)
	}
	if err := runOnce(ctx, logger, u, ev); err != nil {
		if errors.Is(err, context.Canceled) {
			logger.Info("shutdown requested, exiting edge-trust")
			return nil
		}
		return err
	}
	return nil
}

func runOnce(
	ctx context.Context,
	logger *zap.Logger,
	u *updater.Updater,
	ev config.EnvVars,
) error {
	startedAt := time.Now()

	s, err := loadStoredState(logger, ev.StateJSONPath)
	if err != nil {
		logger.Warn(
			"stored state invalid, starting with empty state",
			zap.Error(err),
		)
		s = state.State{}
	}
	if _, _, err = u.Run(ctx, logger, s); err != nil {
		return fmt.Errorf("run updater: %w", err)
	}
	endedAt := time.Since(startedAt)
	logger.Info(
		"update run completed",
		zap.String("duration", endedAt.Round(time.Millisecond).String()),
	)
	return nil
}

func runDaemon(
	ctx context.Context,
	logger *zap.Logger,
	u *updater.Updater,
	ev config.EnvVars,
	hs *health.Service,
	runInterval time.Duration,
) error {
	wait, err := waitInitialDelay(ctx, logger, ev.StateJSONPath, hs, runInterval)
	if err != nil {
		return fmt.Errorf("wait initial delay: %w", err)
	}
	if !wait {
		return nil
	}
	runCount := 1
	for {
		logger.Info("starting update run", zap.Int("count", runCount))

		if err := runOnce(ctx, logger, u, ev); err != nil {
			if errors.Is(err, context.Canceled) {
				logger.Info("shutdown requested, exiting edge-trust")
				return nil
			}
			logger.Error("update run failed", zap.Error(err))
		}
		logger.Info(
			"next update scheduled",
			zap.String("interval", runInterval.String()),
		)
		runCount++
		wait, err := waitForNextRun(ctx, runInterval, hs)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				logger.Info("shutdown requested, exiting edge-trust")
				return nil
			}
			return fmt.Errorf("wait next run: %w", err)
		}
		if !wait {
			logger.Info("shutdown requested, exiting edge-trust")
			return nil
		}
	}
}

func loadStoredState(logger *zap.Logger, stateJsonPath string) (state.State, error) {
	logger.Info("loading stored state", zap.String("path", stateJsonPath))

	storedState, err := state.Load(stateJsonPath)
	if err != nil {
		return storedState, err
	}
	if storedState.HasData() {
		logger.Info(
			"loaded stored state",
			zap.String("source_url", storedState.SourceURL),
			zap.String("etag", storedState.ETag),
			zap.Int("cidr_count", len(storedState.CIDRs)),
			zap.String("hash", storedState.Hash),
			zap.Time("written_at", storedState.WrittenAt),
		)
	} else {
		logger.Info("no stored state found, or state is empty")
	}
	return storedState, nil
}

// waitInitialDelay delays the first daemon run based on
// last successful update time stored in state.
//
// Returns false if shutdown is requested while waiting.
func waitInitialDelay(
	ctx context.Context,
	logger *zap.Logger,
	statePath string,
	hs *health.Service,
	interval time.Duration,
) (bool, error) {
	logger.Debug(
		"loading stored state for daemon scheduling",
		zap.String("path", statePath),
	)
	storedState, err := loadStoredState(logger, statePath)
	if err != nil {
		logger.Warn(
			"failed to load stored state for daemon scheduling",
			zap.Error(err),
		)
		return true, nil
	}
	if !storedState.HasData() {
		logger.Debug(
			"stored state is empty, skipping initial delay",
		)
		return true, nil
	}
	nextRunAt := storedState.WrittenAt.Add(interval)
	initialDelay := time.Until(nextRunAt)

	logger.Debug(
		"calculated initial daemon delay",
		zap.Time("last_run_at", storedState.WrittenAt),
		zap.Time("next_run_at", nextRunAt),
		zap.String(
			"delay",
			initialDelay.Round(time.Second).String(),
		),
	)
	if initialDelay <= 0 {
		logger.Debug(
			"initial delay already elapsed, starting immediately",
		)
		return true, nil
	}
	logger.Info(
		"delaying first daemon run",
		zap.String(
			"delay",
			initialDelay.Round(time.Second).String(),
		),
		zap.Time("next_run_at", nextRunAt),
	)
	return waitForNextRun(ctx, initialDelay, hs)
}

func waitForNextRun(
	ctx context.Context,
	runInterval time.Duration,
	hs *health.Service,
) (bool, error) {
	if hs == nil {
		select {
		case <-ctx.Done():
			return false, ctx.Err()
		case <-time.After(runInterval):
			return true, nil
		}
	}
	ticker := time.NewTicker(hs.PulseInterval)
	defer ticker.Stop()

	deadline := time.Now().Add(runInterval)
	for {
		if time.Now().After(deadline) {
			return true, nil
		}
		select {
		case <-ctx.Done():
			return false, ctx.Err()
		case <-ticker.C:
			if err := hs.EmitPulse(time.Now()); err != nil {
				return false, fmt.Errorf("emit heartbeat: %w", err)
			}
		}
	}
}

func initHealthService(interval time.Duration, hcSignalPath string) *health.Service {
	var hs *health.Service
	var err error
	hs, err = health.NewService(interval, hcSignalPath)
	if err != nil {
		panic(fmt.Errorf("create health service: %w", err))
	}
	return hs
}
