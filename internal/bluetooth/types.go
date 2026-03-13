package bluetooth

import (
	"fmt"
	"syscall"
	"unsafe"
)

// bluetoothFindRadioParams mirrors BLUETOOTH_FIND_RADIO_PARAMS (bluetoothapis.h:98).
type bluetoothFindRadioParams struct {
	DwSize uint32
}

// bluetoothRadioInfo mirrors BLUETOOTH_RADIO_INFO (bluetoothapis.h:214).
type bluetoothRadioInfo struct {
	DwSize          uint32
	Address         uint64      // BLUETOOTH_ADDRESS (union -> ullLong)
	SzName          [248]uint16 // WCHAR[BLUETOOTH_MAX_NAME_SIZE]
	UlClassOfDevice uint32
	LmpSubversion   uint16
	Manufacturer    uint16
}

// bluetoothDeviceInfo mirrors BLUETOOTH_DEVICE_INFO (bluetoothapis.h:268).
type bluetoothDeviceInfo struct {
	DwSize          uint32
	Address         uint64 // BLUETOOTH_ADDRESS
	UlClassOfDevice uint32
	FConnected      int32 // BOOL
	FRemembered     int32
	FAuthenticated  int32
	StLastSeen      syscall.Systemtime // 16 bytes
	StLastUsed      syscall.Systemtime // 16 bytes
	SzName          [248]uint16
}

// bluetoothDeviceSearchParams mirrors BLUETOOTH_DEVICE_SEARCH_PARAMS (bluetoothapis.h:371).
type bluetoothDeviceSearchParams struct {
	DwSize               uint32
	FReturnAuthenticated int32
	FReturnRemembered    int32
	FReturnUnknown       int32
	FReturnConnected     int32
	FIssueInquiry        int32
	CTimeoutMultiplier   uint8
	_pad                 [7]byte // alignment before HANDLE (8 bytes on amd64)
	HRadio               uintptr
}

func init() {
	checks := []struct {
		name   string
		got    uintptr
		expect uintptr
	}{
		{"bluetoothFindRadioParams", unsafe.Sizeof(bluetoothFindRadioParams{}), 4},
		{"bluetoothRadioInfo", unsafe.Sizeof(bluetoothRadioInfo{}), 520},
		{"bluetoothDeviceInfo", unsafe.Sizeof(bluetoothDeviceInfo{}), 560},
		{"bluetoothDeviceSearchParams", unsafe.Sizeof(bluetoothDeviceSearchParams{}), 40},
	}
	for _, c := range checks {
		if c.got != c.expect {
			panic(fmt.Sprintf("bluetooth: sizeof(%s) = %d, want %d", c.name, c.got, c.expect))
		}
	}
}
