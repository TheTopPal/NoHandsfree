package bluetooth

import (
	"testing"
	"unsafe"
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
