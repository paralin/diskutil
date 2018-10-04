package diskutil

// http://msdn.microsoft.com/en-us/library/windows/desktop/aa373931.aspx
type GUID struct {
	Data1 uint32
	Data2 uint16
	Data3 uint16
	Data4 [8]byte
}

// https://docs.microsoft.com/en-us/windows-hardware/drivers/install/install-reference
// https://msdn.microsoft.com/en-us/library/windows/hardware/ff541389(v=vs.85).aspx

// GUID_DEVINTERFACE_DISK { 53F56307-B6BF-11D0-94F2-00A0C91EFB8B }
var (
	GUID_DEVINTERFACE_DISK = &GUID{
		Data1: 0x53F56307,
		Data2: 0xB6BF,
		Data3: 0x11D0,
		Data4: [8]byte{0x94, 0xF2, 0x00, 0xA0, 0xC9, 0x1E, 0xFB, 0x8B},
	}

	// GUID_DEVINTERFACE_CDROM { 53F56308-B6BF-11D0-94F2-00A0C91EFB8B }
	GUID_DEVINTERFACE_CDROM = &GUID{
		Data1: 0x53F56308,
		Data2: 0xB6BF,
		Data3: 0x11D0,
		Data4: [8]byte{0x94, 0xF2, 0x00, 0xA0, 0xC9, 0x1E, 0xFB, 0x8B},
	}

	// GUID_DEVINTERFACE_USB_HUB { F18A0E88-C30C-11D0-8815-00A0C906BED8 }
	GUID_DEVINTERFACE_USB_HUB = &GUID{
		Data1: 0xF18A0E88,
		Data2: 0xC30C,
		Data3: 0x11D0,
		Data4: [8]byte{0x88, 0x15, 0x00, 0xA0, 0xC9, 0x06, 0xBE, 0xD8},
	}

	// GUID_DEVINTERFACE_FLOPPY { 53F56311-B6BF-11D0-94F2-00A0C91EFB8B }
	GUID_DEVINTERFACE_FLOPPY = &GUID{
		Data1: 0x53F56311,
		Data2: 0xB6BF,
		Data3: 0x11D0,
		Data4: [8]byte{0x94, 0xF2, 0x00, 0xA0, 0xC9, 0x1E, 0xFB, 0x8B},
	}

	// GUID_DEVINTERFACE_WRITEONCEDISK { 53F5630C-B6BF-11D0-94F2-00A0C91EFB8B }
	GUID_DEVINTERFACE_WRITEONCEDISK = &GUID{
		Data1: 0x53F5630C,
		Data2: 0xB6BF,
		Data3: 0x11D0,
		Data4: [8]byte{0x94, 0xF2, 0x00, 0xA0, 0xC9, 0x1E, 0xFB, 0x8B},
	}

	// GUID_DEVINTERFACE_TAPE { 53F5630B-B6BF-11D0-94F2-00A0C91EFB8B }
	GUID_DEVINTERFACE_TAPE = &GUID{
		Data1: 0x53F5630B,
		Data2: 0xB6BF,
		Data3: 0x11D0,
		Data4: [8]byte{0x94, 0xF2, 0x00, 0xA0, 0xC9, 0x1E, 0xFB, 0x8B},
	}

	// GUID_DEVINTERFACE_USB_DEVICE { A5DCBF10-6530-11D2-901F-00C04FB951ED }
	GUID_DEVINTERFACE_USB_DEVICE = &GUID{
		Data1: 0xA5DCBF10,
		Data2: 0x6530,
		Data3: 0x11D2,
		Data4: [8]byte{0x90, 0x1F, 0x00, 0xC0, 0x4F, 0xB9, 0x51, 0xED},
	}

	// GUID_DEVINTERFACE_VOLUME { 53F5630D-B6BF-11D0-94F2-00A0C91EFB8B }
	GUID_DEVINTERFACE_VOLUME = &GUID{
		Data1: 0x53F5630D,
		Data2: 0xB6BF,
		Data3: 0x11D0,
		Data4: [8]byte{0x94, 0xF2, 0x00, 0xA0, 0xC9, 0x1E, 0xFB, 0x8B},
	}

	// GUID_DEVINTERFACE_STORAGEPORT { 2ACCFE60-C130-11D2-B082-00A0C91EFB8B }
	GUID_DEVINTERFACE_STORAGEPORT = &GUID{
		Data1: 0x2ACCFE60,
		Data2: 0xC130,
		Data3: 0x11D2,
		Data4: [8]byte{0xB0, 0x82, 0x00, 0xA0, 0xC9, 0x1E, 0xFB, 0x8B},
	}
)

// SPDRP is a device registry property code.
type SPDRP uint32

const (
	SPDRP_DEVICEDESC                  SPDRP = 0x00000000
	SPDRP_HARDWAREID                        = 0x00000001
	SPDRP_COMPATIBLEIDS                     = 0x00000002
	SPDRP_UNUSED0                           = 0x00000003
	SPDRP_SERVICE                           = 0x00000004
	SPDRP_UNUSED1                           = 0x00000005
	SPDRP_UNUSED2                           = 0x00000006
	SPDRP_CLASS                             = 0x00000007
	SPDRP_CLASSGUID                         = 0x00000008
	SPDRP_DRIVER                            = 0x00000009
	SPDRP_CONFIGFLAGS                       = 0x0000000A
	SPDRP_MFG                               = 0x0000000B
	SPDRP_FRIENDLYNAME                      = 0x0000000C
	SPDRP_LOCATION_INFORMATION              = 0x0000000D
	SPDRP_PHYSICAL_DEVICE_OBJECT_NAME       = 0x0000000E
	SPDRP_CAPABILITIES                      = 0x0000000F
	SPDRP_UI_NUMBER                         = 0x00000010
	SPDRP_UPPERFILTERS                      = 0x00000011
	SPDRP_LOWERFILTERS                      = 0x00000012
	SPDRP_BUSTYPEGUID                       = 0x00000013
	SPDRP_LEGACYBUSTYPE                     = 0x00000014
	SPDRP_BUSNUMBER                         = 0x00000015
	SPDRP_ENUMERATOR_NAME                   = 0x00000016
	SPDRP_SECURITY                          = 0x00000017
	SPDRP_SECURITY_SDS                      = 0x00000018
	SPDRP_DEVTYPE                           = 0x00000019
	SPDRP_EXCLUSIVE                         = 0x0000001A
	SPDRP_CHARACTERISTICS                   = 0x0000001B
	SPDRP_ADDRESS                           = 0x0000001C
	SPDRP_UI_NUMBER_DESC_FORMAT             = 0X0000001D
	SPDRP_DEVICE_POWER_DATA                 = 0x0000001E
	SPDRP_REMOVAL_POLICY                    = 0x0000001F
	SPDRP_REMOVAL_POLICY_HW_DEFAULT         = 0x00000020
	SPDRP_REMOVAL_POLICY_OVERRIDE           = 0x00000021
	SPDRP_INSTALL_STATE                     = 0x00000022
	SPDRP_LOCATION_PATHS                    = 0x00000023
)

const (
	DIGCF_DEFAULT         uint = 0x00000001 // only valid with DIGCF_DEVICEINTERFACE
	DIGCF_PRESENT         uint = 0x00000002
	DIGCF_ALLCLASSES      uint = 0x00000004
	DIGCF_PROFILE         uint = 0x00000008
	DIGCF_DEVICEINTERFACE uint = 0x00000010
)

const (
	CM_REMOVAL_POLICY_EXPECT_NO_REMOVAL       byte = 0x00000001
	CM_REMOVAL_POLICY_EXPECT_ORDERLY_REMOVAL       = 0x00000002
	CM_REMOVAL_POLICY_EXPECT_SURPRISE_REMOVAL      = 0x00000003
)

var VHD_HARDWARE_IDS = []string{
	"Arsenal_________Virtual_",
	"KernSafeVirtual_________",
	"Msft____Virtual_Disk____",
	"VMware__VMware_Virtual_S",
}

var USB_STORAGE_DRIVERS = []string{
	"USBSTOR", "UASPSTOR", "VUSBSTOR",
	"RTUSER", "CMIUCR", "EUCR",
	"ETRONSTOR", "ASUSSTPT",
}

var GENERIC_STORAGE_DRIVERS = []string{
	"SCSI", "SD", "PCISTOR",
	"RTSOR", "JMCR", "JMCF", "RIMMPTSK", "RIMSPTSK", "RIXDPTSK",
	"TI21SONY", "ESD7SK", "ESM7SK", "O2MD", "O2SD", "VIACR",
}
