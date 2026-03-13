package bluetooth

import "golang.org/x/sys/windows"

// Service class GUIDs from bthdef.h:197.
// Base UUID: {00000000-0000-1000-8000-00805F9B34FB}

// HandsfreeServiceClassGUID is the Handsfree service (HFP), UUID16 = 0x111E.
var HandsfreeServiceClassGUID = windows.GUID{
	Data1: 0x0000111E,
	Data2: 0x0000,
	Data3: 0x1000,
	Data4: [8]byte{0x80, 0x00, 0x00, 0x80, 0x5F, 0x9B, 0x34, 0xFB},
}

// A2DPSinkServiceClassGUID is the Advanced Audio Distribution (A2DP Sink), UUID16 = 0x110B.
var A2DPSinkServiceClassGUID = windows.GUID{
	Data1: 0x0000110B,
	Data2: 0x0000,
	Data3: 0x1000,
	Data4: [8]byte{0x80, 0x00, 0x00, 0x80, 0x5F, 0x9B, 0x34, 0xFB},
}

// HandsfreeAudioGatewayServiceClassGUID is the HFP Audio Gateway, UUID16 = 0x111F.
var HandsfreeAudioGatewayServiceClassGUID = windows.GUID{
	Data1: 0x0000111F,
	Data2: 0x0000,
	Data3: 0x1000,
	Data4: [8]byte{0x80, 0x00, 0x00, 0x80, 0x5F, 0x9B, 0x34, 0xFB},
}
