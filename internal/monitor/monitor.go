package monitor

import (
	"context"
	"log/slog"
	"time"

	"github.com/TopPal/NoHandsfree/internal/bluetooth"
	"github.com/TopPal/NoHandsfree/internal/config"
)

// Run starts the polling loop that auto-disables HFP on configured devices.
// It blocks until ctx is cancelled.
func Run(ctx context.Context, cfg *config.Config) error {
	interval := time.Duration(cfg.PollIntervalSec) * time.Second
	slog.Info("monitor started", "interval", interval)

	// Track connection state to log connect/disconnect events.
	connected := make(map[string]bool) // address -> was connected

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	check(cfg, connected)
	for {
		select {
		case <-ctx.Done():
			slog.Info("monitor stopped")
			return ctx.Err()
		case <-ticker.C:
			check(cfg, connected)
		}
	}
}

func check(cfg *config.Config, connected map[string]bool) {
	devices, err := bluetooth.ListPairedDevices()
	if err != nil {
		slog.Error("failed to list devices", "error", err)
		return
	}

	// Build set of currently connected addresses to detect disconnects.
	nowConnected := make(map[string]bool)

	for _, dev := range devices {
		addr := dev.AddressString()

		if dev.Connected {
			nowConnected[addr] = true

			if !connected[addr] {
				slog.Info("device connected", "device", dev.Name, "address", addr, "hfp", dev.HFPEnabled)
			}

			dc, ok := cfg.Devices[addr]
			if !ok || !dc.AutoDisableHFP {
				continue
			}
			if !dev.HFPEnabled {
				continue
			}

			slog.Info("disabling HFP", "device", dev.Name, "address", addr)
			if err := bluetooth.DisableHFP(dev.Address); err != nil {
				slog.Error("failed to disable HFP", "device", dev.Name, "address", addr, "error", err)
			} else {
				slog.Info("HFP disabled", "device", dev.Name, "address", addr)
			}
		}
	}

	// Detect disconnects.
	for addr, was := range connected {
		if was && !nowConnected[addr] {
			slog.Info("device disconnected", "address", addr)
		}
	}

	// Update state.
	for addr := range connected {
		delete(connected, addr)
	}
	for addr := range nowConnected {
		connected[addr] = true
	}
}
