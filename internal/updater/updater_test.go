package updater

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"senthora.com/edge-trust/internal/cloudflare"
	"senthora.com/edge-trust/internal/nginx"
	"senthora.com/edge-trust/internal/state"
	"senthora.com/edge-trust/internal/testutil"
)

var ips = testutil.IPs{
	IPV4: "103.21.244.0/22",
	IPV6: "2400:cb00::/32",
}

func TestRun_ReturnsUpdatedStateWhenCIDRsChange(t *testing.T) {
	f := newUpdaterFixture(t, testutil.CloudflareResponseJson(ips))
	defer f.server.Close()

	newState, changed, err := f.updater.Run(
		context.Background(),
		testutil.NewLogger(),
		state.State{},
	)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if !changed {
		t.Fatalf("expected changed to be true")
	}
	if !newState.HasData() {
		t.Fatalf("expected new state to contain data")
	}
	assertUpdateFilesExist(t, f)
}

func TestRun_SkipsUpdateWhenETagMatches(t *testing.T) {
	f := newUpdaterFixture(t, testutil.CloudflareResponseJson(ips))
	defer f.server.Close()

	currentState := state.New(
		f.client.APIURL(),
		testutil.SampleETag(),
		[]string{
			ips.IPV4,
			ips.IPV6,
		},
		testutil.FixedTimeNowUTC(),
	)
	newState, changed, err := f.updater.Run(
		context.Background(),
		testutil.NewLogger(),
		currentState,
	)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	assertUpdateWasSkipped(t, f, currentState, newState, changed)
}

func TestRun_UpdatesWhenETagChanges(t *testing.T) {
	f := newUpdaterFixture(t, testutil.CloudflareResponseJson(ips))
	defer f.server.Close()

	currentState := state.New(
		f.client.APIURL(),
		"old-etag",
		[]string{
			ips.IPV4,
			ips.IPV6,
		},
		testutil.FixedTimeNowUTC(),
	)

	newState, changed, err := f.updater.Run(
		context.Background(),
		testutil.NewLogger(),
		currentState,
	)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if !changed {
		t.Fatalf("expected changed to be true")
	}
	if !newState.HasData() {
		t.Fatalf("expected new state to contain data")
	}
	if newState.ETag != testutil.SampleETag() {
		t.Fatalf(
			"expected etag %q, got %q",
			testutil.SampleETag(),
			newState.ETag,
		)
	}
	assertUpdateFilesExist(t, f)
}

func TestRun_ReturnsErrorWhenETagIsEmpty(t *testing.T) {
	responseJson := fmt.Sprintf(`{
      "result": {
	    "etag": "",
	    "ipv4_cidrs": ["%s"],
	    "ipv6_cidrs": ["%s"]
      }
	}`, ips.IPV4, ips.IPV6)
	f := newUpdaterFixture(t, responseJson)
	defer f.server.Close()

	currentState := state.State{}

	newState, changed, err := f.updater.Run(
		context.Background(),
		testutil.NewLogger(),
		currentState,
	)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if changed {
		t.Fatalf("expected changed to be false")
	}
	if newState.HasData() {
		t.Fatalf("expected new state to be empty")
	}
	assertFileDoesNotExist(t, f.updater.paths.ProxySourcesPath)
	assertFileDoesNotExist(t, f.updater.paths.OriginAllowlistPath)
	assertFileDoesNotExist(t, f.updater.paths.StateJSONPath)
	assertFileDoesNotExist(t, f.updater.paths.ReloadSignalPath)
}

func TestRun_ReturnsErrorWhenCloudflareFetchFails(t *testing.T) {
	f := newUpdaterFixture(t, `invalid-json`)
	defer f.server.Close()

	currentState := state.State{}
	newState, changed, err := f.updater.Run(
		context.Background(),
		testutil.NewLogger(),
		currentState,
	)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	assertUpdateWasSkipped(t, f, currentState, newState, changed)
}

func TestRun_ReturnsErrorWhenWriteStepFails(t *testing.T) {
	tests := []struct {
		name       string
		blockPath  func(f *updaterFixture) string
		assertions func(t *testing.T, f *updaterFixture)
	}{
		{
			name: "trusted proxy config write fails",
			blockPath: func(f *updaterFixture) string {
				return f.updater.paths.ProxySourcesPath
			},
			assertions: func(t *testing.T, f *updaterFixture) {
				t.Helper()

				assertFileDoesNotExist(t, f.updater.paths.OriginAllowlistPath)
				assertFileDoesNotExist(t, f.updater.paths.StateJSONPath)
				assertFileDoesNotExist(t, f.updater.paths.ReloadSignalPath)
			},
		},
		{
			name: "origin allowlist write fails",
			blockPath: func(f *updaterFixture) string {
				return f.updater.paths.OriginAllowlistPath
			},
			assertions: func(t *testing.T, f *updaterFixture) {
				t.Helper()

				assertFileExists(t, f.updater.paths.ProxySourcesPath)
				assertFileDoesNotExist(t, f.updater.paths.StateJSONPath)
				assertFileDoesNotExist(t, f.updater.paths.ReloadSignalPath)
			},
		},
		{
			name: "state save fails",
			blockPath: func(f *updaterFixture) string {
				return f.updater.paths.StateJSONPath
			},
			assertions: func(t *testing.T, f *updaterFixture) {
				t.Helper()

				assertFileExists(t, f.updater.paths.ProxySourcesPath)
				assertFileExists(t, f.updater.paths.OriginAllowlistPath)
				assertFileDoesNotExist(t, f.updater.paths.ReloadSignalPath)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newUpdaterFixture(t, testutil.CloudflareResponseJson(ips))
			defer f.server.Close()

			blockPath := tt.blockPath(f)

			if err := os.Mkdir(blockPath, 0755); err != nil {
				t.Fatalf("failed to create blocking directory: %v", err)
			}
			newState, changed, err := f.updater.Run(
				context.Background(),
				testutil.NewLogger(),
				state.State{},
			)
			if !errors.Is(err, fs.ErrExist) {
				t.Fatalf("expected fs.ErrExist, got %v", err)
			}
			if changed {
				t.Fatalf("expected changed to be false")
			}
			if !newState.HasData() {
				t.Fatalf("expected new state to contain data")
			}
			tt.assertions(t, f)
		})
	}
}

func TestRun_ReturnsErrorWhenReloadSignalEmitFails(t *testing.T) {
	f := newUpdaterFixture(t, testutil.CloudflareResponseJson(ips))
	defer f.server.Close()

	if err := os.Mkdir(f.updater.paths.ReloadSignalPath, 0755); err != nil {
		t.Fatalf("failed to create blocking directory: %v", err)
	}
	newState, changed, err := f.updater.Run(
		context.Background(),
		testutil.NewLogger(),
		state.State{},
	)
	if !errors.Is(err, nginx.ErrSignalPathIsDirectory) {
		t.Fatalf("expected ErrSignalPathIsDirectory, got %v", err)
	}
	if changed {
		t.Fatalf("expected changed to be false")
	}
	if !newState.HasData() {
		t.Fatalf("expected new state to contain data")
	}
	assertFileExists(t, f.updater.paths.ProxySourcesPath)
	assertFileExists(t, f.updater.paths.OriginAllowlistPath)
	assertFileExists(t, f.updater.paths.ReloadSignalPath)
}

type updaterFixture struct {
	server *httptest.Server
	dir    string

	client  *cloudflare.Client
	updater *Updater
}

func newUpdaterFixture(t *testing.T, responseBody string) *updaterFixture {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if _, err := w.Write([]byte(responseBody)); err != nil {
			t.Fatalf("failed to write response: %v", err)
		}
	}))
	dir := t.TempDir()

	client := cloudflare.NewClient(
		testutil.NewLogger(),
		server.URL,
		[]time.Duration{0},
		server.Client(),
	)
	configPaths := ConfigPaths{
		ProxySourcesPath:    filepath.Join(dir, "trusted-proxies.conf"),
		OriginAllowlistPath: filepath.Join(dir, "origin-allowlist.conf"),
		StateJSONPath:       filepath.Join(dir, "state.json"),
		ReloadSignalPath:    filepath.Join(dir, "nginx.reload"),
	}
	updater := NewUpdater(client, configPaths)
	return &updaterFixture{
		server:  server,
		dir:     dir,
		client:  client,
		updater: updater,
	}
}

func assertUpdateFilesExist(t *testing.T, f *updaterFixture) {
	t.Helper()

	assertFileExists(t, f.updater.paths.ProxySourcesPath)
	assertFileExists(t, f.updater.paths.OriginAllowlistPath)
	assertFileExists(t, f.updater.paths.StateJSONPath)
	assertFileExists(t, f.updater.paths.ReloadSignalPath)
}

func assertUpdateWasSkipped(
	t *testing.T,
	f *updaterFixture,
	currentState state.State,
	newState state.State,
	changed bool,
) {
	t.Helper()

	if changed {
		t.Fatalf("expected changed to be false")
	}
	if !currentState.Equals(newState) {
		t.Fatalf("expected states to be equal")
	}
	assertFileDoesNotExist(t, f.updater.paths.ProxySourcesPath)
	assertFileDoesNotExist(t, f.updater.paths.OriginAllowlistPath)
	assertFileDoesNotExist(t, f.updater.paths.StateJSONPath)
	assertFileDoesNotExist(t, f.updater.paths.ReloadSignalPath)
}

func assertFileExists(t *testing.T, path string) {
	t.Helper()

	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected file %q to exist: %v", path, err)
	}
}

func assertFileDoesNotExist(t *testing.T, path string) {
	t.Helper()

	if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("expected file %q to not exist", path)
	}
}
