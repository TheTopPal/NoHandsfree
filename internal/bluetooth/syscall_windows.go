//nolint:gosec // G103: unsafe.Pointer is required for syscall interop
package bluetooth

import (
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	modBthprops = syscall.NewLazyDLL("bthprops.cpl")

	procBluetoothFindFirstRadio             = modBthprops.NewProc("BluetoothFindFirstRadio")
	procBluetoothFindNextRadio              = modBthprops.NewProc("BluetoothFindNextRadio")
	procBluetoothFindRadioClose             = modBthprops.NewProc("BluetoothFindRadioClose")
	procBluetoothFindFirstDevice            = modBthprops.NewProc("BluetoothFindFirstDevice")
	procBluetoothFindNextDevice             = modBthprops.NewProc("BluetoothFindNextDevice")
	procBluetoothFindDeviceClose            = modBthprops.NewProc("BluetoothFindDeviceClose")
	procBluetoothEnumerateInstalledServices = modBthprops.NewProc("BluetoothEnumerateInstalledServices")
	procBluetoothSetServiceState            = modBthprops.NewProc("BluetoothSetServiceState")
)

// bluetoothFindFirstRadio wraps BluetoothFindFirstRadio.
// Returns a find handle and the first radio HANDLE.
func bluetoothFindFirstRadio() (findHandle windows.Handle, radioHandle windows.Handle, err error) {
	params := bluetoothFindRadioParams{DwSize: uint32(unsafe.Sizeof(bluetoothFindRadioParams{}))}
	var hRadio windows.Handle

	r, _, e := procBluetoothFindFirstRadio.Call(
		uintptr(unsafe.Pointer(&params)),
		uintptr(unsafe.Pointer(&hRadio)),
	)
	if r == 0 {
		return 0, 0, e
	}
	return windows.Handle(r), hRadio, nil
}

// bluetoothFindNextRadio gets the next radio HANDLE.
func bluetoothFindNextRadio(findHandle windows.Handle) (radioHandle windows.Handle, err error) {
	var hRadio windows.Handle
	r, _, e := procBluetoothFindNextRadio.Call(
		uintptr(findHandle),
		uintptr(unsafe.Pointer(&hRadio)),
	)
	if r == 0 {
		return 0, e
	}
	return hRadio, nil
}

// bluetoothFindRadioClose closes a radio enumeration handle.
func bluetoothFindRadioClose(findHandle windows.Handle) error {
	r, _, e := procBluetoothFindRadioClose.Call(uintptr(findHandle))
	if r == 0 {
		return e
	}
	return nil
}

// bluetoothFindFirstDevice starts enumerating paired/connected devices.
func bluetoothFindFirstDevice(params *bluetoothDeviceSearchParams) (findHandle windows.Handle, info *bluetoothDeviceInfo, err error) {
	var di bluetoothDeviceInfo
	di.DwSize = uint32(unsafe.Sizeof(di))

	r, _, e := procBluetoothFindFirstDevice.Call(
		uintptr(unsafe.Pointer(params)),
		uintptr(unsafe.Pointer(&di)),
	)
	if r == 0 {
		return 0, nil, e
	}
	return windows.Handle(r), &di, nil
}

// bluetoothFindNextDevice gets the next device in the enumeration.
func bluetoothFindNextDevice(findHandle windows.Handle) (*bluetoothDeviceInfo, error) {
	var di bluetoothDeviceInfo
	di.DwSize = uint32(unsafe.Sizeof(di))

	r, _, e := procBluetoothFindNextDevice.Call(
		uintptr(findHandle),
		uintptr(unsafe.Pointer(&di)),
	)
	if r == 0 {
		return nil, e
	}
	return &di, nil
}

// bluetoothFindDeviceClose closes a device enumeration handle.
func bluetoothFindDeviceClose(findHandle windows.Handle) error {
	r, _, e := procBluetoothFindDeviceClose.Call(uintptr(findHandle))
	if r == 0 {
		return e
	}
	return nil
}

// bluetoothEnumerateInstalledServices returns the list of service GUIDs
// installed for a device.
func bluetoothEnumerateInstalledServices(radioHandle windows.Handle, devInfo *bluetoothDeviceInfo) ([]windows.GUID, error) {
	var count uint32

	// First call: get count.
	r, _, _ := procBluetoothEnumerateInstalledServices.Call(
		uintptr(radioHandle),
		uintptr(unsafe.Pointer(devInfo)),
		uintptr(unsafe.Pointer(&count)),
		0,
	)
	if r != 0 && syscall.Errno(r) != syscall.ERROR_MORE_DATA {
		return nil, syscall.Errno(r)
	}
	if count == 0 {
		return nil, nil
	}

	guids := make([]windows.GUID, count)
	r, _, _ = procBluetoothEnumerateInstalledServices.Call(
		uintptr(radioHandle),
		uintptr(unsafe.Pointer(devInfo)),
		uintptr(unsafe.Pointer(&count)),
		uintptr(unsafe.Pointer(&guids[0])),
	)
	if r != 0 {
		return nil, syscall.Errno(r)
	}
	return guids[:count], nil
}

// bluetoothSetServiceState enables or disables a service on a device.
func bluetoothSetServiceState(radioHandle windows.Handle, devInfo *bluetoothDeviceInfo, serviceGUID *windows.GUID, enable bool) error {
	var flags uint32
	if enable {
		flags = 0x0001 // BLUETOOTH_SERVICE_ENABLE
	} else {
		flags = 0x0000 // BLUETOOTH_SERVICE_DISABLE
	}

	r, _, _ := procBluetoothSetServiceState.Call(
		uintptr(radioHandle),
		uintptr(unsafe.Pointer(devInfo)),
		uintptr(unsafe.Pointer(serviceGUID)),
		uintptr(flags),
	)
	if r != 0 {
		return syscall.Errno(r)
	}
	return nil
}
