package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadMissingFileReturnsDefaults(t *testing.T) {
	cfg, err := Load(filepath.Join(t.TempDir(), "absent.json"))
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.PollIntervalSec != 5 {
		t.Errorf("PollIntervalSec = %d, want 5", cfg.PollIntervalSec)
	}
	if cfg.Devices == nil {
		t.Error("Devices is nil; 'config add' assigns into it without checking")
	}
}

func TestSaveLoadRoundTrip(t *testing.T) {
	// A nested path also checks that Save creates the parent directory, which
	// is what happens on a first 'config add'.
	path := filepath.Join(t.TempDir(), "NoHandsfree", "config.json")
	const addr = "58:18:62:1E:B9:B2"

	want := &Config{
		Devices:         map[string]DeviceConfig{addr: {AutoDisableHFP: true}},
		PollIntervalSec: 12,
	}
	if err := Save(path, want); err != nil {
		t.Fatalf("Save: %v", err)
	}

	got, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if got.PollIntervalSec != want.PollIntervalSec {
		t.Errorf("PollIntervalSec = %d, want %d", got.PollIntervalSec, want.PollIntervalSec)
	}
	if len(got.Devices) != 1 {
		t.Fatalf("Devices = %v, want one entry", got.Devices)
	}
	if !got.Devices[addr].AutoDisableHFP {
		t.Errorf("AutoDisableHFP for %s did not survive the round trip", addr)
	}
}

func TestLoadRepairsUnusableValues(t *testing.T) {
	// A hand-edited config must not hand the monitor a zero tick interval
	// (time.NewTicker panics) or a nil map.
	for _, body := range []string{
		`{"devices":null,"poll_interval_sec":0}`,
		`{"poll_interval_sec":-1}`,
		`{}`,
	} {
		path := filepath.Join(t.TempDir(), "config.json")
		if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
			t.Fatal(err)
		}

		cfg, err := Load(path)
		if err != nil {
			t.Fatalf("Load(%s): %v", body, err)
		}
		if cfg.PollIntervalSec != 5 {
			t.Errorf("Load(%s) PollIntervalSec = %d, want 5", body, cfg.PollIntervalSec)
		}
		if cfg.Devices == nil {
			t.Errorf("Load(%s) left Devices nil", body)
		}
	}
}

func TestLoadRejectsMalformedJSON(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(path, []byte("{not json"), 0o600); err != nil {
		t.Fatal(err)
	}
	if cfg, err := Load(path); err == nil {
		t.Errorf("Load accepted malformed JSON, returned %+v", cfg)
	}
}
