# NoHandsfree

Windows utility to automatically disable Bluetooth HFP (Hands-Free Profile).

When Bluetooth headphones connect, Windows enables two audio profiles:
- **A2DP** (Stereo) - high-quality audio
- **HFP** (Hands-Free) - low-quality, intended for calls

HFP creates an extra audio device "Headset (Hands-Free AG Audio)" in Sound Settings and may hijack audio output. This tool disables HFP, keeping only A2DP.

## Build

Requirements:

- [Go 1.25+](https://go.dev/dl/)
- [golangci-lint v2](https://golangci-lint.run/welcome/install/)


Using Make (Git Bash / MSYS2):

```bash
make build       # build binary to bin/
make test        # run tests
make lint        # run golangci-lint
make clean       # remove bin/
make install     # build + register elevated auto-start
make uninstall   # remove auto-start
```

Or directly (cmd / PowerShell):

```cmd
go build -o bin\nohandsfree.exe .
go test ./...
golangci-lint run ./...
```

## Usage

Most commands require **administrator privileges**.

```cmd
# List paired devices and HFP status
nohandsfree list

# Disable HFP on all devices
nohandsfree disable all

# Disable HFP on a specific device
nohandsfree disable AA:BB:CC:DD:EE:FF

# Re-enable HFP
nohandsfree enable AA:BB:CC:DD:EE:FF

# Configure auto-disable
nohandsfree config add AA:BB:CC:DD:EE:FF
nohandsfree config remove AA:BB:CC:DD:EE:FF
nohandsfree config interval 10      # set polling interval (default: 5)
nohandsfree config show

# Start monitor (polls every N sec)
nohandsfree watch

# Register/remove elevated auto-start at logon
nohandsfree install
nohandsfree uninstall
```

## Auto-start

`install` registers a Task Scheduler entry named `NoHandsfree` that runs
`nohandsfree watch` at logon with the highest available privileges:

```
schtasks /Create /TN NoHandsfree /TR "<exe>" watch /SC ONLOGON /RL HIGHEST /F
```

The elevated run level is the point. `watch` calls `BluetoothSetServiceState`,
which needs administrator rights, and Windows starts `Run` key entries
unelevated - so a `Run` entry would exit immediately at every logon. Both
`install` and `uninstall` must themselves run elevated, and both clear the old
`Run` value if an earlier build left one behind.

The task shows up in Task Scheduler under Task Scheduler Library.

## Configuration

File: `%APPDATA%\NoHandsfree\config.json`

```json
{
  "devices": {
    "AA:BB:CC:DD:EE:FF": {
      "auto_disable_hfp": true
    }
  },
  "poll_interval_sec": 5
}
```