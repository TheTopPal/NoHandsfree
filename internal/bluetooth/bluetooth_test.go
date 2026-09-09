package bluetooth

import (
	"errors"
	"testing"
	"unsafe"

	"golang.org/x/sys/windows"
)

func TestStructSizes(t *testing.T) {
	tests := []struct {
		name   string
		got    uintptr
		expect uintptr
	}{
		{"bluetoothFindRadioParams", unsafe.Sizeof(bluetoothFindRadioParams{}), 4},
		{"bluetoothRadioInfo", unsafe.Sizeof(bluetoothRadioInfo{}), 520},
		{"bluetoothDeviceInfo", unsafe.Sizeof(bluetoothDeviceInfo{}), 560},
		{"bluetoothDeviceSearchParams", unsafe.Sizeof(bluetoothDeviceSearchParams{}), 40},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.expect {
				t.Errorf("sizeof(%s) = %d, want %d", tt.name, tt.got, tt.expect)
			}
		})
	}
}

func TestAddressString(t *testing.T) {
	// Address 0x5818621EB9B2 -> "58:18:62:1E:B9:B2"
	d := Device{Address: 0x5818621EB9B2}
	got := d.AddressString()
	want := "58:18:62:1E:B9:B2"
	if got != want {
		t.Errorf("AddressString() = %q, want %q", got, want)
	}
}

// TestEmptySearchReportsNoMoreItems pins the Windows behaviour that enumDevices
// and enumRadios rely on: a search with no matches fails with
// ERROR_NO_MORE_ITEMS rather than returning an empty list. If that ever stops
// holding, "no paired devices" would surface as a hard error again.
func TestEmptySearchReportsNoMoreItems(t *testing.T) {
	radios, err := enumRadios()
	if err != nil {
		t.Fatalf("enumRadios: %v", err)
	}
	if len(radios) == 0 {
		t.Skip("no Bluetooth radio on this machine")
	}
	defer closeHandles(radios)

	// Every filter off, so nothing can match — the same shape as a machine
	// with a radio but no paired devices.
	params := bluetoothDeviceSearchParams{
		DwSize: uint32(unsafe.Sizeof(bluetoothDeviceSearchParams{})),
		HRadio: uintptr(radios[0]),
	}

	findHandle, _, err := bluetoothFindFirstDevice(&params)
	if err == nil {
		_ = bluetoothFindDeviceClose(findHandle)
		t.Fatal("a search that cannot match returned success")
	}
	if !errors.Is(err, windows.ERROR_NO_MORE_ITEMS) {
		t.Fatalf("empty search returned %v, want ERROR_NO_MORE_ITEMS", err)
	}
}
