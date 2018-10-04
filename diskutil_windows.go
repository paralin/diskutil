package diskutil

import (
	"strings"
	"syscall"
	"unsafe"
)

// Implicit windows only build tag.

var (
	modsetupapi = syscall.NewLazyDLL("setupapi.dll")

	procSetupDiGetClassDevsW              = modsetupapi.NewProc("SetupDiGetClassDevsW")
	procSetupDiEnumDeviceInfo             = modsetupapi.NewProc("SetupDiEnumDeviceInfo")
	procSetupDiDestroyDeviceInfoList      = modsetupapi.NewProc("SetupDiDestroyDeviceInfoList")
	procSetupDiGetDeviceRegistryPropertyA = modsetupapi.NewProc("SetupDiGetDeviceRegistryPropertyA")
)

// ListStorageDevices lists all connected storage devices.
func ListStorageDevices() ([]*DeviceDescriptor, error) {
	var res []*DeviceDescriptor
	/*
		hDeviceInfo = SetupDiGetClassDevsA(
		  &GUID_DEVICE_INTERFACE_DISK, NULL, NULL,
		  DIGCF_PRESENT | DIGCF_DEVICEINTERFACE);
	*/
	ret, _, lastErr := procSetupDiGetClassDevsW.Call(
		uintptr(unsafe.Pointer(GUID_DEVINTERFACE_DISK)),
		uintptr(0), uintptr(0),
		uintptr(DIGCF_PRESENT|DIGCF_DEVICEINTERFACE),
	)
	if ret == uintptr(0) {
		return nil, lastErr
	}
	defer syscall.Syscall(procSetupDiDestroyDeviceInfoList.Addr(), 1, ret, 0, 0)

	retHandle := syscall.Handle(ret)
	if retHandle == syscall.InvalidHandle {
		return res, nil
	}

	did := spDeviceInformationData{}
	did.CbSize = uint32(unsafe.Sizeof(did))

	for i := 0; setupDiEnumDeviceInfo(retHandle, uint32(i), &did) == nil; i++ {
		enumName, _ := getEnumeratorName(retHandle, &did)
		if enumName == "" {
			continue
		}

		descrip := &DeviceDescriptor{}
		descrip.Enumerator = enumName
		descrip.Description, _ = getDeviceRegistryPropertyString(retHandle, &did, SPDRP_FRIENDLYNAME)
		descrip.IsRemovable, _ = getIsRemovable(retHandle, &did)
		descrip.IsVirtual, _ = getIsVirtual(retHandle, &did)
		descrip.IsScsi = getIsSCSI(enumName)
		descrip.IsUSB = getIsUSB(enumName)
		descrip.IsCard = getIsSDCard(enumName)
		descrip.IsSystem = !descrip.IsRemovable && (descrip.Enumerator == "SCSI" || descrip.Enumerator == "IDE")
		descrip.IsUAS = descrip.IsScsi && descrip.IsRemovable && !descrip.IsVirtual && !descrip.IsCard

		descrip.LogicalBlockSize = 512
		descrip.BlockSize = 512

		res = append(res, descrip)
	}

	return res, nil
}

// spDeviceInformationData is SP_DEVINFO_DATA lookup structure.
type spDeviceInformationData struct {
	CbSize    uint32
	ClassGuid GUID
	DevInst   uint32
	Reserved  uintptr
}

func setupDiEnumDeviceInfo(deviceInfoSet syscall.Handle, memberIndex uint32, deviceInfoData *spDeviceInformationData) error {
	r1, _, e1 := syscall.Syscall(
		procSetupDiEnumDeviceInfo.Addr(),
		3,
		uintptr(deviceInfoSet),
		uintptr(memberIndex),
		uintptr(unsafe.Pointer(deviceInfoData)),
	)
	if r1 == 0 {
		if e1 != 0 {
			return error(e1)
		} else {
			return syscall.EINVAL
		}
	}

	return nil
}

func enumDeviceInfo(di syscall.Handle, memberIndex uint32) (*spDeviceInformationData, error) {
	did := spDeviceInformationData{}
	did.CbSize = uint32(unsafe.Sizeof(did))
	if err := setupDiEnumDeviceInfo(di, memberIndex, &did); err != nil {
		return nil, err
	}
	return &did, nil
}

// getEnumeratorName retreives the enumerator name of the device.
// Examples: SCSI, USBSTOR
func getEnumeratorName(di syscall.Handle, did *spDeviceInformationData) (string, error) {
	return getDeviceRegistryPropertyString(di, did, SPDRP_ENUMERATOR_NAME)
}

// getIsSCSI checks if a device is a scsi device.
func getIsSCSI(enumName string) bool {
	for _, usbv := range GENERIC_STORAGE_DRIVERS {
		if enumName == usbv {
			return true
		}
	}
	return false
}

// getIsUSB checks if a device is a usb device.
func getIsUSB(enumName string) bool {
	for _, usbv := range USB_STORAGE_DRIVERS {
		if enumName == usbv {
			return true
		}
	}
	return false
}

// getIsSDCard checks if a device is a sd card.
func getIsSDCard(enumName string) bool {
	return enumName == "SD"
}

// getIsRemovable checks if a device is removable.
func getIsRemovable(di syscall.Handle, did *spDeviceInformationData) (bool, error) {
	d, err := getDeviceRegistryPropertyByte(di, did, SPDRP_REMOVAL_POLICY)
	if err != nil {
		return false, err
	}
	switch d {
	case CM_REMOVAL_POLICY_EXPECT_ORDERLY_REMOVAL:
	case CM_REMOVAL_POLICY_EXPECT_SURPRISE_REMOVAL:
		return true, nil
	}
	return false, nil
}

// getIsVirtual checks if a device is virtual.
func getIsVirtual(di syscall.Handle, did *spDeviceInformationData) (bool, error) {
	hwid, err := getDeviceRegistryPropertyString(di, did, SPDRP_HARDWAREID)
	if err != nil {
		return false, err
	}

	if len(hwid) == 0 {
		return false, nil
	}

	for _, vhwid := range VHD_HARDWARE_IDS {
		if hwid == vhwid {
			return true, nil
		}
	}

	return false, nil
}

// getDeviceRegistryPropertyString retreives a device registry property value (string)
func getDeviceRegistryPropertyString(di syscall.Handle, did *spDeviceInformationData, prop SPDRP) (string, error) {
	var buf [255]byte
	_, _, _ = syscall.Syscall9(
		procSetupDiGetDeviceRegistryPropertyA.Addr(),
		7,
		uintptr(di),
		uintptr(unsafe.Pointer(did)),
		uintptr(prop),
		uintptr(0),
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(uint32(len(buf))),
		uintptr(0),
		uintptr(0),
		uintptr(0),
	)

	if buf[0] == 0x0 {
		return "", nil
	}

	bufStr := string(buf[:])
	return string(bufStr[:strings.Index(bufStr, "\x00")]), nil
}

// getDeviceRegistryPropertyByte retreives a device registry property value (byte)
func getDeviceRegistryPropertyByte(di syscall.Handle, did *spDeviceInformationData, prop SPDRP) (byte, error) {
	var dat uint32
	_, _, _ = syscall.Syscall9(
		procSetupDiGetDeviceRegistryPropertyA.Addr(),
		7,
		uintptr(di),
		uintptr(unsafe.Pointer(did)),
		uintptr(prop),
		uintptr(0),
		uintptr(unsafe.Pointer(&dat)),
		uintptr(unsafe.Sizeof(dat)),
		uintptr(0),
		uintptr(0),
		uintptr(0),
	)
	return byte(dat), nil
}
