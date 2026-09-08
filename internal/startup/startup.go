// Package startup manages the Windows auto-start entry for the monitor.
//
// The monitor needs administrator privileges, so it cannot be launched from
// HKCU\...\Run: Windows starts Run entries unelevated and 'watch' would refuse
// to run. A scheduled task with an ONLOGON trigger and RunLevel HIGHEST is the
// standard way to auto-start elevated, and needs no stored password because
// schtasks gives an ONLOGON task the interactive token of its creator.
package startup

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"

	"golang.org/x/sys/windows/registry"
)

// TaskName is the Task Scheduler entry that starts the monitor at logon.
const TaskName = "NoHandsfree"

// Where auto-start lived before the scheduled task. Kept only so Install and
// Uninstall can clear it on machines that were set up with the old build.
const (
	legacyRunKeyPath  = `Software\Microsoft\Windows\CurrentVersion\Run`
	legacyRunKeyValue = "NoHandsfree"
)

// Install registers the monitor to start elevated at logon and clears any
// leftover Run key entry. Creating a HIGHEST task requires elevation.
func Install(exe string) error {
	if err := schtasks(createArgs(exe)); err != nil {
		return err
	}
	return removeLegacyRunKey()
}

// Uninstall removes the auto-start task and any leftover Run key entry.
// It reports whether a task was actually there to remove.
func Uninstall() (bool, error) {
	if err := removeLegacyRunKey(); err != nil {
		return false, err
	}
	if !Installed() {
		return false, nil
	}
	if err := schtasks(deleteArgs()); err != nil {
		return false, err
	}
	return true, nil
}

// Installed reports whether the auto-start task is registered. A failed query
// counts as "not installed": schtasks exits non-zero both when the task is
// missing and on error, and its messages are localized, so the exit code is
// the only portable signal available.
func Installed() bool {
	return schtasks(queryArgs()) == nil
}

// createArgs builds the registration command. ONLOGON + HIGHEST is the part
// that matters: it is what makes the task start elevated.
func createArgs(exe string) []string {
	return []string{
		"/Create",
		"/TN", TaskName,
		"/TR", `"` + exe + `" watch`,
		"/SC", "ONLOGON",
		"/RL", "HIGHEST",
		"/F",
	}
}

func deleteArgs() []string {
	return []string{"/Delete", "/TN", TaskName, "/F"}
}

func queryArgs() []string {
	return []string{"/Query", "/TN", TaskName}
}

// schtasks runs schtasks.exe. It reports failures on stdout rather than
// through the exit status alone, so the output is folded into the error.
func schtasks(args []string) error {
	out, err := exec.Command("schtasks.exe", args...).CombinedOutput() //nolint:gosec // args are built above; only the executable path varies
	if err == nil {
		return nil
	}
	if msg := strings.TrimSpace(string(out)); msg != "" {
		return fmt.Errorf("schtasks %s: %w: %s", args[0], err, msg)
	}
	return fmt.Errorf("schtasks %s: %w", args[0], err)
}

// removeLegacyRunKey deletes the old Run value if one is left over.
func removeLegacyRunKey() error {
	key, err := registry.OpenKey(registry.CURRENT_USER, legacyRunKeyPath, registry.SET_VALUE)
	if err != nil {
		if errors.Is(err, registry.ErrNotExist) {
			return nil
		}
		return fmt.Errorf("open Run key: %w", err)
	}
	defer func() { _ = key.Close() }()

	if err := key.DeleteValue(legacyRunKeyValue); err != nil && !errors.Is(err, registry.ErrNotExist) {
		return fmt.Errorf("delete legacy Run value: %w", err)
	}
	return nil
}
