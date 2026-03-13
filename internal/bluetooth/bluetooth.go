package bluetooth

import (
	"fmt"
	"unsafe"

	"golang.org/x/sys/windows"
)

// Device represents a paired Bluetooth device.
type Device struct {
	Name       string
	Address    uint64
	Connected  bool
	Remembered bool
	HFPEnabled bool
}

// AddressString returns the BT address as a colon-separated hex string.
func (d Device) AddressString() string {
	b := make([]byte, 6)
	for i := range b {
		b[i] = byte(d.Address >> (8 * i)) //nolint:gosec // intentional truncation to extract individual bytes
	}
	return fmt.Sprintf("%02X:%02X:%02X:%02X:%02X:%02X", b[5], b[4], b[3], b[2], b[1], b[0])
}

// ListPairedDevices enumerates all paired (remembered) Bluetooth devices
// and checks their HFP status.
func ListPairedDevices() ([]Device, error) {
	radios, err := enumRadios()
	if err != nil {
		return nil, fmt.Errorf("enumerate radios: %w", err)
	}
	defer closeHandles(radios)

	if len(radios) == 0 {
		return nil, nil
	}

	var devices []Device
	for _, radio := range radios {
		devs, err := enumDevices(radio)
		if err != nil {
			return nil, fmt.Errorf("enumerate devices: %w", err)
		}
		for _, di := range devs {
			hfp, _ := isHFPEnabled(radio, &di)
			devices = append(devices, Device{
				Name:       windows.UTF16ToString(di.SzName[:]),
				Address:    di.Address,
				Connected:  di.FConnected != 0,
				Remembered: di.FRemembered != 0,
				HFPEnabled: hfp,
			})
		}
	}
	return devices, nil
}

// DisableHFP disables the Handsfree service on the device with the given address.
func DisableHFP(address uint64) error {
	return setHFP(address, false)
}

// EnableHFP enables the Handsfree service on the device with the given address.
func EnableHFP(address uint64) error {
	return setHFP(address, true)
}

// IsHFPEnabled checks whether HFP is enabled on the device with the given address.
func IsHFPEnabled(address uint64) (bool, error) {
	radios, err := enumRadios()
	if err != nil {
		return false, err
	}
	defer closeHandles(radios)

	di, radio, err := findDevice(radios, address)
	if err != nil {
		return false, err
	}
	return isHFPEnabled(radio, di)
}

func setHFP(address uint64, enable bool) error {
	radios, err := enumRadios()
	if err != nil {
		return err
	}
	defer closeHandles(radios)

	di, radio, err := findDevice(radios, address)
	if err != nil {
		return err
	}

	guid := HandsfreeServiceClassGUID
	return bluetoothSetServiceState(radio, di, &guid, enable)
}

func isHFPEnabled(radioHandle windows.Handle, di *bluetoothDeviceInfo) (bool, error) {
	guids, err := bluetoothEnumerateInstalledServices(radioHandle, di)
	if err != nil {
		return false, err
	}
	for _, g := range guids {
		if g == HandsfreeServiceClassGUID {
			return true, nil
		}
	}
	return false, nil
}

func findDevice(radios []windows.Handle, address uint64) (*bluetoothDeviceInfo, windows.Handle, error) {
	for _, radio := range radios {
		devs, err := enumDevices(radio)
		if err != nil {
			continue
		}
		for i := range devs {
			if devs[i].Address == address {
				return &devs[i], radio, nil
			}
		}
	}
	return nil, 0, fmt.Errorf("device %012X not found", address)
}

func enumRadios() ([]windows.Handle, error) {
	findHandle, firstRadio, err := bluetoothFindFirstRadio()
	if err != nil {
		return nil, err
	}

	radios := []windows.Handle{firstRadio}
	for {
		radio, err := bluetoothFindNextRadio(findHandle)
		if err != nil {
			break
		}
		radios = append(radios, radio)
	}
	_ = bluetoothFindRadioClose(findHandle)
	return radios, nil
}

func enumDevices(radioHandle windows.Handle) ([]bluetoothDeviceInfo, error) {
	params := bluetoothDeviceSearchParams{
		DwSize:               uint32(unsafe.Sizeof(bluetoothDeviceSearchParams{})),
		FReturnAuthenticated: 1,
		FReturnRemembered:    1,
		FReturnConnected:     1,
		HRadio:               uintptr(radioHandle),
	}

	findHandle, first, err := bluetoothFindFirstDevice(&params)
	if err != nil {
		return nil, err
	}

	devices := []bluetoothDeviceInfo{*first}
	for {
		di, err := bluetoothFindNextDevice(findHandle)
		if err != nil {
			break
		}
		devices = append(devices, *di)
	}
	_ = bluetoothFindDeviceClose(findHandle)
	return devices, nil
}

func closeHandles(handles []windows.Handle) {
	for _, h := range handles {
		_ = windows.CloseHandle(h)
	}
}
