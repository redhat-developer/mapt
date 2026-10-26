package telemetry

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestFirstRunGeneratesUUIDv4InstallID(t *testing.T) {
	client := newTestClient(t)
	decision, err := client.Resolve(Overrides{}, "mapt aws hosts create", true)
	if err != nil {
		t.Fatal(err)
	}
	if !decision.Enabled {
		t.Fatal("first run should be enabled by default")
	}
	if !regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`).MatchString(decision.InstallID) {
		t.Fatalf("install id %q is not a UUID v4", decision.InstallID)
	}
	info, err := os.Stat(client.statePath)
	if err != nil {
		t.Fatal(err)
	}
	if got := info.Mode().Perm(); got != 0600 {
		t.Fatalf("state file mode = %04o, want 0600", got)
	}
}

func TestResolvePrecedence(t *testing.T) {
	tests := []struct {
		name       string
		state      string
		overrides  Overrides
		wantEnable bool
	}{
		{name: "do not track wins", state: "enabled", overrides: Overrides{DoNotTrack: "1", MapTelemetry: "1"}, wantEnable: false},
		{name: "flag wins over environment", state: "enabled", overrides: Overrides{NoTelemetry: true, MapTelemetry: "1"}, wantEnable: false},
		{name: "environment on wins over disabled state", state: "disabled", overrides: Overrides{MapTelemetry: "1"}, wantEnable: true},
		{name: "environment off wins over enabled state", state: "enabled", overrides: Overrides{MapTelemetry: "0"}, wantEnable: false},
		{name: "no state defaults to enabled", state: "", wantEnable: true},
		{name: "enabled state", state: "enabled", wantEnable: true},
		{name: "disabled state", state: "disabled", wantEnable: false},
		{name: "other values do not override", state: "disabled", overrides: Overrides{DoNotTrack: "yes", MapTelemetry: "true"}, wantEnable: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := newTestClient(t)
			seedState(t, client, tt.state)
			decision, err := client.Resolve(tt.overrides, "mapt aws hosts create", true)
			if err != nil {
				t.Fatal(err)
			}
			if decision.Enabled != tt.wantEnable {
				t.Fatalf("enabled = %t, want %t", decision.Enabled, tt.wantEnable)
			}
			if decision.Enabled && decision.InstallID == "" {
				t.Fatal("enabled decision has no install id")
			}
		})
	}
}

func TestRunOnlyEnvironmentOptInDoesNotPersist(t *testing.T) {
	client := newTestClient(t)
	decision, err := client.Resolve(Overrides{MapTelemetry: "1"}, "mapt aws hosts create", true)
	if err != nil {
		t.Fatal(err)
	}
	if !decision.Enabled || decision.InstallID == "" {
		t.Fatalf("environment opt-in decision = %+v, want enabled with a run-only id", decision)
	}
	if _, err := os.Stat(client.statePath); !os.IsNotExist(err) {
		t.Fatalf("run-only environment opt-in persisted state: stat error = %v", err)
	}
}

func TestExcludedCommandsDoNotNoticeOrPersist(t *testing.T) {
	for _, command := range []string{"mapt", "mapt help", "mapt completion bash", "mapt version"} {
		t.Run(command, func(t *testing.T) {
			client := newTestClient(t)
			var notice strings.Builder
			client.stderr = &notice
			decision, err := client.Resolve(Overrides{}, command, true)
			if err != nil {
				t.Fatal(err)
			}
			if decision.Enabled || notice.Len() != 0 {
				t.Fatalf("decision = %+v and notice = %q; want off with no notice", decision, notice.String())
			}
			if _, err := os.Stat(client.statePath); !os.IsNotExist(err) {
				t.Fatalf("excluded run persisted state: %v", err)
			}
		})
	}
}

func TestFirstRunNoticesOnceAndEnablesByDefault(t *testing.T) {
	client := newTestClient(t)
	var notice strings.Builder
	client.stderr = &notice
	decision, err := client.Resolve(Overrides{}, "mapt aws hosts create", true)
	if err != nil {
		t.Fatal(err)
	}
	if !decision.Enabled || decision.InstallID == "" {
		t.Fatalf("first run decision = %+v, want enabled by default with an install id", decision)
	}
	if text := notice.String(); !strings.Contains(text, "anonymous usage statistics") || !strings.Contains(text, "MAPT_TELEMETRY=0") {
		t.Fatalf("first-run notice missing collection or opt-out wording: %q", text)
	}
	state, err := client.readState()
	if err != nil {
		t.Fatal(err)
	}
	if state.Enabled == nil || !*state.Enabled || state.InstallID != decision.InstallID {
		t.Fatalf("persisted state = %+v, want enabled with the decision's install id", state)
	}

	notice.Reset()
	second, err := client.Resolve(Overrides{}, "mapt aws hosts create", true)
	if err != nil {
		t.Fatal(err)
	}
	if !second.Enabled || second.InstallID != decision.InstallID {
		t.Fatalf("second run decision = %+v, want enabled with the same install id", second)
	}
	if notice.Len() != 0 {
		t.Fatalf("notice repeated on second run: %q", notice.String())
	}
}

func TestDisabledStateSuppressesNoticeAndEmission(t *testing.T) {
	client := newTestClient(t)
	seedState(t, client, "disabled")
	var notice strings.Builder
	client.stderr = &notice
	decision, err := client.Resolve(Overrides{}, "mapt aws hosts create", true)
	if err != nil {
		t.Fatal(err)
	}
	if decision.Enabled {
		t.Fatalf("decision after disable = %+v, want off", decision)
	}
	if notice.Len() != 0 {
		t.Fatalf("notice printed for a disabled state: %q", notice.String())
	}
}

func TestEmitterPayloadAndEndpoint(t *testing.T) {
	type receivedPayload struct {
		APIKey     string         `json:"api_key"`
		Event      string         `json:"event"`
		DistinctID string         `json:"distinct_id"`
		Properties map[string]any `json:"properties"`
		Timestamp  string         `json:"timestamp"`
	}
	received := make(chan receivedPayload, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %q, want POST", r.Method)
		}
		if r.URL.Path != capturePath {
			t.Errorf("path = %q, want %q", r.URL.Path, capturePath)
		}
		var payload receivedPayload
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Errorf("decode payload: %v", err)
		}
		received <- payload
		w.WriteHeader(http.StatusAccepted)
	}))
	defer server.Close()

	client := newTestClient(t)
	client.endpoint = server.URL + capturePath
	client.apiKey = "test-token"
	done := client.Emit(Event{InstallID: "test-install", Version: "v1.2.3", Command: "mapt aws hosts create"})
	client.Wait(done, time.Second)

	var payload receivedPayload
	select {
	case payload = <-received:
	case <-time.After(time.Second):
		t.Fatal("server did not receive telemetry request")
	}
	if payload.APIKey != "test-token" || payload.Event != "mapt.execution" || payload.DistinctID != "test-install" {
		t.Fatalf("payload identity fields = %+v", payload)
	}
	wantProperties := map[string]any{
		"mapt_version":   "v1.2.3",
		"command":        "mapt aws hosts create",
		"os":             runtime.GOOS,
		"arch":           runtime.GOARCH,
		"$geoip_disable": true,
	}
	if !reflect.DeepEqual(payload.Properties, wantProperties) {
		t.Fatalf("properties = %#v, want %#v", payload.Properties, wantProperties)
	}
	if _, err := time.Parse(time.RFC3339, payload.Timestamp); err != nil {
		t.Fatalf("timestamp %q is not RFC3339: %v", payload.Timestamp, err)
	}
}

func TestEmitterTimeoutAndServerErrorAreSwallowed(t *testing.T) {
	t.Run("timeout", func(t *testing.T) {
		requestEntered := make(chan struct{})
		releaseHandler := make(chan struct{})
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			close(requestEntered)
			<-releaseHandler
			w.WriteHeader(http.StatusAccepted)
		}))
		defer func() {
			close(releaseHandler)
			server.Close()
		}()
		client := newTestClient(t)
		client.endpoint = server.URL + capturePath
		client.httpClient = &http.Client{Timeout: 25 * time.Millisecond}
		done := client.Emit(Event{InstallID: "test-install"})
		select {
		case <-requestEntered:
		case <-time.After(time.Second):
			t.Fatal("HTTP server did not receive request")
		}
		select {
		case <-done:
		case <-time.After(time.Second):
			t.Fatal("HTTP client timeout did not stop the sender")
		}
	})

	t.Run("server error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()
		client := newTestClient(t)
		client.endpoint = server.URL + capturePath
		done := client.Emit(Event{InstallID: "test-install"})
		client.Wait(done, time.Second)
		select {
		case <-done:
		default:
			t.Fatal("server error escaped or left the sender running")
		}
	})
}

func TestUncountedCommandsDoNotNoticeOrPersist(t *testing.T) {
	client := newTestClient(t)
	var notice strings.Builder
	client.stderr = &notice
	decision, err := client.Resolve(Overrides{}, "mapt aws", false)
	if err != nil {
		t.Fatal(err)
	}
	if decision.Enabled {
		t.Fatalf("uncounted command enabled telemetry: %+v", decision)
	}
	if notice.Len() > 0 {
		t.Fatalf("uncounted command printed the notice: %q", notice.String())
	}
	if _, err := os.Stat(client.statePath); !os.IsNotExist(err) {
		t.Fatalf("uncounted command persisted state: %v", err)
	}
}

func TestVersionPrefersBuildTimeInjection(t *testing.T) {
	original := version
	defer func() { version = original }()
	version = "v9.9.9-test"
	if got := Version(); got != "v9.9.9-test" {
		t.Fatalf("Version() = %q, want the ldflags-injected value", got)
	}
	version = ""
	if got := Version(); got == "" {
		t.Fatal("Version() must fall back to build info or \"dev\" when not injected")
	}
}

func TestWriteStateLeavesNoTempFilesAndKeepsPermissions(t *testing.T) {
	client := newTestClient(t)
	enabled := true
	if err := client.writeState(storedState{Enabled: &enabled, InstallID: "perm-check"}); err != nil {
		t.Fatal(err)
	}
	entries, err := os.ReadDir(filepath.Dir(client.statePath))
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), ".telemetry-") {
			t.Fatalf("writeState leaked temporary file %q", entry.Name())
		}
	}
	info, err := os.Stat(client.statePath)
	if err != nil {
		t.Fatal(err)
	}
	if perm := info.Mode().Perm(); perm != 0600 {
		t.Fatalf("state file mode = %o, want 600", perm)
	}
}

func newTestClient(t *testing.T) *Client {
	t.Helper()
	client := New(filepath.Join(t.TempDir(), "nested", "telemetry.json"))
	client.stderr = io.Discard
	return client
}

func seedState(t *testing.T, client *Client, state string) {
	t.Helper()
	switch state {
	case "enabled":
		enabled := true
		if err := client.writeState(storedState{Enabled: &enabled, InstallID: "seeded-install-id"}); err != nil {
			t.Fatal(err)
		}
	case "disabled":
		disabled := false
		if err := client.writeState(storedState{Enabled: &disabled}); err != nil {
			t.Fatal(err)
		}
	case "":
	default:
		t.Fatalf("unknown test state %q", state)
	}
}

