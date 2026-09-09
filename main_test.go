package main

import (
	"testing"

	"github.com/TopPal/NoHandsfree/internal/bluetooth"
)

func TestParseAddress(t *testing.T) {
	valid := map[string]uint64{
		"58:18:62:1E:B9:B2": 0x5818621EB9B2,
		"58-18-62-1e-b9-b2": 0x5818621EB9B2,
		"5818621eb9b2":      0x5818621EB9B2,
		"00:00:00:00:00:00": 0,
		"FF:FF:FF:FF:FF:FF": 0xFFFFFFFFFFFF,
	}
	for in, want := range valid {
		got, err := parseAddress(in)
		if err != nil {
			t.Errorf("parseAddress(%q): %v", in, err)
			continue
		}
		if got != want {
			t.Errorf("parseAddress(%q) = %#x, want %#x", in, got, want)
		}
	}
}

func TestParseAddressRejectsJunk(t *testing.T) {
	// Every one of these used to be accepted by config add and stored
	// verbatim, producing a key no device could ever match.
	junk := []string{
		"",
		"1234",
		"58:18:62:1E:B9",       // an octet short
		"58:18:62:1E:B9:B2:C3", // an octet long
		"zz:zz:zz:zz:zz:zz",    // not hex
		"58:18:62:1E:B9:BZ",    // one bad digit
		"58 18 62 1E B9 B2",    // spaces are not separators
	}
	for _, in := range junk {
		if got, err := parseAddress(in); err == nil {
			t.Errorf("parseAddress(%q) = %#x, want an error", in, got)
		}
	}
}

func TestConfigKeyMatchesMonitorLookup(t *testing.T) {
	// configKey writes the config file; the monitor looks devices up by
	// Device.AddressString(). Auto-disable only fires if they agree, whatever
	// separator style the user typed.
	dev := bluetooth.Device{Address: 0x5818621EB9B2}

	for _, in := range []string{"58:18:62:1E:B9:B2", "58-18-62-1e-b9-b2", "5818621eb9b2"} {
		key, err := configKey(in)
		if err != nil {
			t.Fatalf("configKey(%q): %v", in, err)
		}
		if key != dev.AddressString() {
			t.Errorf("configKey(%q) = %q, but the monitor looks up %q", in, key, dev.AddressString())
		}
	}
}

func TestConfigKeyRejectsJunk(t *testing.T) {
	if key, err := configKey("nonsense"); err == nil {
		t.Errorf("configKey(%q) = %q, want an error", "nonsense", key)
	}
}
