package telemetry

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type Overrides struct {
	DoNotTrack   string
	MapTelemetry string
	NoTelemetry  bool
}

func ReadOverrides(noTelemetry bool) Overrides {
	return Overrides{
		DoNotTrack:   os.Getenv("DO_NOT_TRACK"),
		MapTelemetry: os.Getenv("MAPT_TELEMETRY"),
		NoTelemetry:  noTelemetry,
	}
}

type Decision struct {
	Enabled   bool
	InstallID string
}

type storedState struct {
	Enabled   *bool  `json:"enabled,omitempty"`
	InstallID string `json:"install_id,omitempty"`
}

type Client struct {
	statePath  string
	stderr     io.Writer
	endpoint   string
	apiKey     string
	httpClient *http.Client
}

func New(statePath string) *Client {
	return &Client{
		statePath:  statePath,
		stderr:     os.Stderr,
		endpoint:   posthogHost + capturePath,
		apiKey:     posthogToken,
		httpClient: &http.Client{Timeout: sendTimeout},
	}
}

// Resolve applies run-time overrides before the persisted choice. counted
// reports whether the command being run is one whose execution is tracked
// (in mapt: a leaf command that actually runs, not a group showing help);
// uncounted commands never notify, never persist, and never emit.
func (c *Client) Resolve(overrides Overrides, commandPath string, counted bool) (Decision, error) {
	if !counted || excludedCommand(commandPath) {
		return Decision{}, nil
	}
	if overrides.DoNotTrack == "1" {
		return Decision{}, nil
	}
	if overrides.NoTelemetry {
		return Decision{}, nil
	}
	if overrides.MapTelemetry == "1" {
		state, err := c.readState()
		if err == nil && state.InstallID != "" {
			return Decision{Enabled: true, InstallID: state.InstallID}, nil
		}
		// This run-only opt-in must not create persistent state.
		id, err := newInstallID()
		if err != nil {
			return Decision{}, err
		}
		return Decision{Enabled: true, InstallID: id}, nil
	}
	if overrides.MapTelemetry == "0" {
		return Decision{}, nil
	}

	state, err := c.readState()
	if err != nil {
		return Decision{}, err
	}
	if state.Enabled != nil {
		if !*state.Enabled {
			return Decision{}, nil
		}
		if state.InstallID == "" {
			return Decision{}, fmt.Errorf("enabled telemetry state is missing install id")
		}
		return Decision{Enabled: true, InstallID: state.InstallID}, nil
	}

	// No persisted choice: telemetry is on by default. Announce it once and
	// record the choice, so the notice is not repeated on later runs.
	if _, err := fmt.Fprintln(c.stderr, "mapt collects anonymous usage statistics: the command run, mapt version, OS and architecture, and a random installation id. The data is used only for aggregate statistics and nothing else. No arguments, logs, personal identifiers, or cloud details are sent. If you do not want to share it, set MAPT_TELEMETRY=0 or DO_NOT_TRACK=1 — see https://github.com/redhat-developer/mapt#telemetry for permanent opt-out."); err != nil {
		return Decision{}, err
	}
	enabled := true
	state.Enabled = &enabled
	state.InstallID, err = newInstallID()
	if err != nil {
		return Decision{}, err
	}
	if err := c.writeState(state); err != nil {
		return Decision{}, err
	}
	return Decision{Enabled: true, InstallID: state.InstallID}, nil
}

func (c *Client) readState() (storedState, error) {
	data, err := os.ReadFile(c.statePath)
	if os.IsNotExist(err) {
		return storedState{}, nil
	}
	if err != nil {
		return storedState{}, err
	}
	var state storedState
	if err := json.Unmarshal(data, &state); err != nil {
		return storedState{}, err
	}
	return state, nil
}

func (c *Client) writeState(state storedState) error {
	if err := os.MkdirAll(filepath.Dir(c.statePath), 0700); err != nil {
		return err
	}
	data, err := json.Marshal(state)
	if err != nil {
		return err
	}
	// Write to a temporary file and rename it into place, so a failed or
	// partial write can never leave a truncated state file behind — a corrupt
	// file would silently disable telemetry and defeat the config-file opt-out.
	tmp, err := os.CreateTemp(filepath.Dir(c.statePath), ".telemetry-*.tmp")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	defer func() { _ = os.Remove(tmpPath) }()
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmpPath, c.statePath)
}

func excludedCommand(commandPath string) bool {
	parts := strings.Fields(commandPath)
	if len(parts) <= 1 {
		return true
	}
	for _, part := range parts {
		switch part {
		case "help", "completion", "version":
			return true
		}
	}
	return false
}

func newInstallID() (string, error) {
	var id [16]byte
	if _, err := rand.Read(id[:]); err != nil {
		return "", err
	}
	id[6] = id[6]&0x0f | 0x40
	id[8] = id[8]&0x3f | 0x80
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x", id[0:4], id[4:6], id[6:8], id[8:10], id[10:16]), nil
}
