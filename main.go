package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"

	"github.com/TopPal/NoHandsfree/internal/bluetooth"
	"github.com/TopPal/NoHandsfree/internal/config"
	"github.com/TopPal/NoHandsfree/internal/monitor"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "help", "-h", "--help":
		printUsage()
		return
	case "list":
		cmdList()
	case "disable":
		requireElevated()
		cmdDisable()
	case "enable":
		requireElevated()
		cmdEnable()
	case "watch":
		requireElevated()
		cmdWatch()
	case "config":
		cmdConfig()
	case "install":
		requireElevated()
		cmdInstall()
	case "uninstall":
		requireElevated()
		cmdUninstall()
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Fprintln(os.Stderr, `NoHandsfree - auto-disable Bluetooth HFP (Hands-Free) on Windows

When Bluetooth headphones connect, Windows enables both A2DP (high-quality
stereo) and HFP (low-quality hands-free) audio profiles. This tool disables
HFP so only A2DP is used.

Usage: nohandsfree <command> [args]

Commands:
  list                       List paired Bluetooth devices and HFP status
  disable [address|all]      Disable HFP on a device or all devices
  enable  <address>          Re-enable HFP on a device (for debugging)
  watch                      Start auto-disable monitor (polling, foreground)
  config add <address>       Add device to auto-disable watch list
  config remove <address>    Remove device from auto-disable watch list
  config interval <seconds>  Set polling interval (default: 5)
  config show                Show current configuration
  install                    Add 'watch' to Windows startup
  uninstall                  Remove from Windows startup
  help                       Show this help

Address format: AA:BB:CC:DD:EE:FF or AABBCCDDEEFF
Config file:    %APPDATA%\NoHandsfree\config.json

Most commands require administrator privileges.`)
}

func cmdList() {
	devices, err := bluetooth.ListPairedDevices()
	if err != nil {
		fatal("list devices: %v", err)
	}
	if len(devices) == 0 {
		fmt.Println("No paired Bluetooth devices found.")
		return
	}

	fmt.Printf("%-30s %-18s %-10s %-5s\n", "NAME", "ADDRESS", "CONNECTED", "HFP")
	fmt.Println(strings.Repeat("-", 67))
	for _, d := range devices {
		conn := "no"
		if d.Connected {
			conn = "yes"
		}
		hfp := "off"
		if d.HFPEnabled {
			hfp = "ON"
		}
		fmt.Printf("%-30s %-18s %-10s %-5s\n", d.Name, d.AddressString(), conn, hfp)
	}
}

func cmdDisable() {
	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, "Usage: nohandsfree disable [address|all]")
		os.Exit(1)
	}
	target := os.Args[2]

	if target == "all" {
		devices, err := bluetooth.ListPairedDevices()
		if err != nil {
			fatal("list devices: %v", err)
		}
		for _, d := range devices {
			if !d.HFPEnabled {
				continue
			}
			fmt.Printf("Disabling HFP on %s (%s)...\n", d.Name, d.AddressString())
			if err := bluetooth.DisableHFP(d.Address); err != nil {
				fmt.Fprintf(os.Stderr, "  error: %v\n", err)
			} else {
				fmt.Println("  done")
			}
		}
		return
	}

	addr, err := parseAddress(target)
	if err != nil {
		fatal("invalid address %q: %v", target, err)
	}
	if err := bluetooth.DisableHFP(addr); err != nil {
		fatal("disable HFP: %v", err)
	}
	fmt.Println("HFP disabled.")
}

func cmdEnable() {
	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, "Usage: nohandsfree enable <address>")
		os.Exit(1)
	}
	addr, err := parseAddress(os.Args[2])
	if err != nil {
		fatal("invalid address %q: %v", os.Args[2], err)
	}
	if err := bluetooth.EnableHFP(addr); err != nil {
		fatal("enable HFP: %v", err)
	}
	fmt.Println("HFP enabled.")
}

func cmdWatch() {
	cfgPath, err := config.DefaultPath()
	if err != nil {
		fatal("config path: %v", err)
	}
	cfg, err := config.Load(cfgPath)
	if err != nil {
		fatal("load config: %v", err)
	}

	if len(cfg.Devices) == 0 {
		fmt.Println("No devices configured for auto-disable. Use 'nohandsfree config add <address>' first.")
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	fmt.Println("Watching for Bluetooth connections... (Ctrl+C to stop)")
	if err := monitor.Run(ctx, cfg); err != nil && err != context.Canceled {
		fatal("monitor: %v", err)
	}
}

func cmdConfig() {
	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, "Usage: nohandsfree config [add|remove|show] ...")
		os.Exit(1)
	}

	cfgPath, err := config.DefaultPath()
	if err != nil {
		fatal("config path: %v", err)
	}
	cfg, err := config.Load(cfgPath)
	if err != nil {
		fatal("load config: %v", err)
	}

	switch os.Args[2] {
	case "add":
		if len(os.Args) < 4 {
			fmt.Fprintln(os.Stderr, "Usage: nohandsfree config add <address>")
			os.Exit(1)
		}
		addr := normalizeAddress(os.Args[3])
		cfg.Devices[addr] = config.DeviceConfig{AutoDisableHFP: true}
		if err := config.Save(cfgPath, cfg); err != nil {
			fatal("save config: %v", err)
		}
		fmt.Printf("Device %s added to auto-disable list.\n", addr)

	case "remove":
		if len(os.Args) < 4 {
			fmt.Fprintln(os.Stderr, "Usage: nohandsfree config remove <address>")
			os.Exit(1)
		}
		addr := normalizeAddress(os.Args[3])
		delete(cfg.Devices, addr)
		if err := config.Save(cfgPath, cfg); err != nil {
			fatal("save config: %v", err)
		}
		fmt.Printf("Device %s removed from auto-disable list.\n", addr)

	case "interval":
		if len(os.Args) < 4 {
			fmt.Fprintln(os.Stderr, "Usage: nohandsfree config interval <seconds>")
			os.Exit(1)
		}
		sec, err := strconv.Atoi(os.Args[3])
		if err != nil || sec < 1 {
			fatal("interval must be a positive integer, got %q", os.Args[3])
		}
		cfg.PollIntervalSec = sec
		if err := config.Save(cfgPath, cfg); err != nil {
			fatal("save config: %v", err)
		}
		fmt.Printf("Poll interval set to %ds.\n", sec)

	case "show":
		fmt.Printf("Config: %s\n", cfgPath)
		fmt.Printf("Poll interval: %ds\n\n", cfg.PollIntervalSec)
		if len(cfg.Devices) == 0 {
			fmt.Println("No devices configured.")
			return
		}
		fmt.Printf("%-18s %s\n", "ADDRESS", "AUTO-DISABLE HFP")
		for addr, dc := range cfg.Devices {
			fmt.Printf("%-18s %v\n", addr, dc.AutoDisableHFP)
		}

	default:
		fmt.Fprintf(os.Stderr, "unknown config command: %s\n", os.Args[2])
		os.Exit(1)
	}
}

// parseAddress parses a Bluetooth address from colon-separated hex (AA:BB:CC:DD:EE:FF)
// or plain hex (AABBCCDDEEFF).
func parseAddress(s string) (uint64, error) {
	s = strings.ReplaceAll(s, ":", "")
	s = strings.ReplaceAll(s, "-", "")
	if len(s) != 12 {
		return 0, fmt.Errorf("expected 12 hex chars, got %d", len(s))
	}
	return strconv.ParseUint(s, 16, 64)
}

// normalizeAddress converts address to colon-separated uppercase format.
func normalizeAddress(s string) string {
	s = strings.ReplaceAll(s, ":", "")
	s = strings.ReplaceAll(s, "-", "")
	s = strings.ToUpper(s)
	if len(s) == 12 {
		return s[0:2] + ":" + s[2:4] + ":" + s[4:6] + ":" + s[6:8] + ":" + s[8:10] + ":" + s[10:12]
	}
	return s
}

const (
	registryPath = `Software\Microsoft\Windows\CurrentVersion\Run`
	registryName = "NoHandsfree"
)

func cmdInstall() {
	exe, err := os.Executable()
	if err != nil {
		fatal("get executable path: %v", err)
	}

	key, _, err := registry.CreateKey(registry.CURRENT_USER, registryPath, registry.SET_VALUE)
	if err != nil {
		fatal("open registry: %v", err)
	}
	defer func() { _ = key.Close() }()

	value := `"` + exe + `" watch`
	if err := key.SetStringValue(registryName, value); err != nil {
		fatal("set registry value: %v", err)
	}
	fmt.Printf("Installed to startup: %s\n", value)
}

func cmdUninstall() {
	key, err := registry.OpenKey(registry.CURRENT_USER, registryPath, registry.SET_VALUE)
	if err != nil {
		fatal("open registry: %v", err)
	}
	defer func() { _ = key.Close() }()

	if err := key.DeleteValue(registryName); err != nil {
		fatal("delete registry value: %v", err)
	}
	fmt.Println("Removed from startup.")
}

func requireElevated() {
	token := windows.GetCurrentProcessToken()
	elevated := token.IsElevated()
	if !elevated {
		fmt.Fprintln(os.Stderr, "Error: this command requires administrator privileges.")
		fmt.Fprintln(os.Stderr, "Please run from an elevated (Administrator) command prompt.")
		os.Exit(1)
	}
}

func fatal(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "Error: "+format+"\n", args...)
	os.Exit(1)
}
