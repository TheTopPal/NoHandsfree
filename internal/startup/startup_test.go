package startup

import "testing"

// flagValue returns the argument following flag.
func flagValue(t *testing.T, args []string, flag string) string {
	t.Helper()
	for i, a := range args {
		if a == flag {
			if i+1 >= len(args) {
				t.Fatalf("flag %s has no value in %q", flag, args)
			}
			return args[i+1]
		}
	}
	t.Fatalf("flag %s missing from %q", flag, args)
	return ""
}

func hasFlag(args []string, flag string) bool {
	for _, a := range args {
		if a == flag {
			return true
		}
	}
	return false
}

func TestCreateArgs(t *testing.T) {
	exe := `C:\Program Files\NoHandsfree\nohandsfree.exe`
	args := createArgs(exe)

	// The monitor requires elevation, so the run level is the whole reason
	// this is a scheduled task rather than a Run key entry.
	if got := flagValue(t, args, "/RL"); got != "HIGHEST" {
		t.Errorf("/RL = %q, want HIGHEST", got)
	}
	if got := flagValue(t, args, "/SC"); got != "ONLOGON" {
		t.Errorf("/SC = %q, want ONLOGON", got)
	}
	if got := flagValue(t, args, "/TN"); got != TaskName {
		t.Errorf("/TN = %q, want %q", got, TaskName)
	}

	// The path must stay quoted or a directory with a space splits the command.
	want := `"` + exe + `" watch`
	if got := flagValue(t, args, "/TR"); got != want {
		t.Errorf("/TR = %q, want %q", got, want)
	}

	// Without /F a second install fails instead of overwriting.
	if !hasFlag(args, "/F") {
		t.Errorf("createArgs missing /F, re-install would fail: %q", args)
	}
}

func TestDeleteArgs(t *testing.T) {
	args := deleteArgs()
	if got := flagValue(t, args, "/TN"); got != TaskName {
		t.Errorf("/TN = %q, want %q", got, TaskName)
	}
	if !hasFlag(args, "/F") {
		t.Errorf("deleteArgs missing /F, delete would prompt: %q", args)
	}
}

func TestQueryArgs(t *testing.T) {
	args := queryArgs()
	if got := flagValue(t, args, "/TN"); got != TaskName {
		t.Errorf("/TN = %q, want %q", got, TaskName)
	}
	// A query must never modify anything.
	for _, bad := range []string{"/F", "/Create", "/Delete"} {
		if hasFlag(args, bad) {
			t.Errorf("queryArgs contains %s: %q", bad, args)
		}
	}
}

func TestInstalledReportsFalseForUnknownTask(t *testing.T) {
	// Guards the exit-code-only contract: no task by this name should exist
	// on a test machine, and Installed must not panic or report true.
	if Installed() {
		t.Skip("a NoHandsfree task is registered on this machine")
	}
}
