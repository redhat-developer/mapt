package telemetry

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"runtime/debug"
	"time"

	"github.com/redhat-developer/mapt/pkg/util/logging"
)

const sendTimeout = 2 * time.Second

type Event struct {
	InstallID string
	Version   string
	Command   string
}

type posthogEvent struct {
	APIKey     string         `json:"api_key"`
	Event      string         `json:"event"`
	DistinctID string         `json:"distinct_id"`
	Properties map[string]any `json:"properties"`
	Timestamp  string         `json:"timestamp"`
}

func (c *Client) Emit(event Event) <-chan struct{} {
	done := make(chan struct{})
	go func() {
		defer close(done)
		c.send(event)
	}()
	return done
}

func (c *Client) Wait(done <-chan struct{}, timeout time.Duration) {
	if done == nil {
		return
	}
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	select {
	case <-done:
	case <-timer.C:
	}
}

// version is injected at build time by the Makefile via
// -ldflags "-X .../pkg/telemetry.version=<VERSION>", mirroring the version
// variables of the other integrations. When unset (plain `go build`), Version
// falls back to the module version from build info.
var version = ""

func Version() string {
	if version != "" {
		return version
	}
	info, ok := debug.ReadBuildInfo()
	if !ok || info.Main.Version == "" || info.Main.Version == "(devel)" {
		return "dev"
	}
	return info.Main.Version
}

func (c *Client) send(event Event) {
	payload := posthogEvent{
		APIKey:     c.apiKey,
		Event:      "mapt.execution",
		DistinctID: event.InstallID,
		Properties: map[string]any{
			"mapt_version": event.Version,
			"command":      event.Command,
			"os":           runtime.GOOS,
			"arch":         runtime.GOARCH,
			// PostHog reads the client IP from every capture request and uses
			// it for GeoIP enrichment by default; disable it so no location
			// data is derived from users' network addresses.
			"$geoip_disable": true,
		},
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}
	data, err := json.Marshal(payload)
	if err != nil {
		logging.Debugf("telemetry delivery failed: %v", err)
		return
	}
	request, err := http.NewRequest(http.MethodPost, c.endpoint, bytes.NewReader(data))
	if err != nil {
		logging.Debugf("telemetry delivery failed: %v", err)
		return
	}
	request.Header.Set("Content-Type", "application/json")
	response, err := c.httpClient.Do(request)
	if err != nil {
		logging.Debugf("telemetry delivery failed: %v", err)
		return
	}
	defer func() { _ = response.Body.Close() }()
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		logging.Debugf("telemetry delivery failed: HTTP status %s", fmt.Sprint(response.Status))
	}
}

