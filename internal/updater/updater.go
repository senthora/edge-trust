package updater

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
	"senthora.com/edge-trust/internal/cloudflare"
	"senthora.com/edge-trust/internal/nginx"
	"senthora.com/edge-trust/internal/state"
)

type ConfigPaths struct {
	ProxySourcesPath    string
	OriginAllowlistPath string
	StateJSONPath       string
	ReloadSignalPath    string
}

type Updater struct {
	client *cloudflare.Client
	paths  ConfigPaths
}

// NewUpdater creates and returns a configured updater
func NewUpdater(
	client *cloudflare.Client,
	path ConfigPaths,
) *Updater {
	return &Updater{
		client: client,
		paths:  path,
	}
}

// Run fetches the latest Cloudflare IP ranges and applies
// updates when the fetched CIDRs differ from the current state.
// Returns the resulting state, whether a change was applied,
// and an error if the update process fails.
func (u *Updater) Run(
	ctx context.Context,
	logger *zap.Logger,
	currentState state.State,
) (state.State, bool, error) {
	apiURL := u.client.APIURL()
	logger.Info("fetching cloudflare IP ranges", zap.String("url", apiURL))

	ipData, err := u.client.FetchIPs(ctx)
	if err != nil {
		return currentState, false, fmt.Errorf("fetch cloudflare IPs: %w", err)
	}
	logger.Info("fetched IPs from cloudflare",
		zap.String("etag", ipData.ETag),
		zap.Int("count", len(ipData.CIDRs)),
		zap.Strings("cidrs", ipData.CIDRs),
	)
	if currentState.ETag == ipData.ETag {
		logger.Info(
			"cloudflare etag unchanged, skipping update",
			zap.String("etag", ipData.ETag),
		)
		return currentState, false, nil
	}
	newState := state.New(apiURL, ipData.ETag, ipData.CIDRs, time.Now())

	logger.Debug("generating trusted proxy config file")
	proxySourcesConfig := nginx.GenerateTrustedProxyConfig(
		newState.CIDRs,
		newState.SourceURL,
		newState.WrittenAt,
	)
	logger.Debug("generating origin allowlist config file")
	originAllowlistConfig := nginx.GenerateOriginAllowlistConfig(
		newState.CIDRs,
		newState.SourceURL,
		newState.WrittenAt,
	)
	logger.Info("writing configuration files", zap.String("file", u.paths.ProxySourcesPath))
	if err := proxySourcesConfig.Write(logger, u.paths.ProxySourcesPath); err != nil {
		return newState, false, fmt.Errorf("write proxy sources nginx config: %w", err)
	}
	if err := originAllowlistConfig.Write(logger, u.paths.OriginAllowlistPath); err != nil {
		return newState, false, fmt.Errorf("write origin allowlist nginx config: %w", err)
	}
	logger.Info("saving new state", zap.String("path", u.paths.StateJSONPath))

	if err := state.Save(logger, u.paths.StateJSONPath, newState); err != nil {
		return newState, false, fmt.Errorf("save state: %w", err)
	}
	logger.Info("emitting nginx reload signal", zap.String("path", u.paths.ReloadSignalPath))
	if err := nginx.EmitReloadSignal(logger, u.paths.ReloadSignalPath); err != nil {
		return newState, false, fmt.Errorf("emit nginx reload signal: %w", err)
	}
	return newState, true, nil
}
