//+build windows,no_wmi

package diskutil

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"syscall"
	"unsafe"

	"golang.org/x/text/encoding/unicode"
)

// Implicit windows only build tag.

var (
	kernel32            = syscall.NewLazyDLL("kernel32.dll")
	procCreateFileW     = kernel32.NewProc("CreateFileW")
	procDeviceIoControl = kernel32.NewProc("DeviceIoControl")

	modsetupapi = syscall.NewLazyDLL("setupapi.dll")

	procSetupDiGetClassDevsW              = modsetupapi.NewProc("SetupDiGetClassDevsW")
	procSetupDiEnumDeviceInfo             = modsetupapi.NewProc("SetupDiEnumDeviceInfo")
	procSetupDiDestroyDeviceInfoList      = modsetupapi.NewProc("SetupDiDestroyDeviceInfoList")
	procSetupDiGetDeviceRegistryPropertyA = modsetupapi.NewProc("SetupDiGetDeviceRegistryPropertyA")
	procSetupDiEnumDeviceInterfaces       = modsetupapi.NewProc("SetupDiEnumDeviceInterfaces")
	procSetupDiGetDeviceInterfaceDetailW  = modsetupapi.NewProc("SetupDiGetDeviceInterfaceDetailW")
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

		if err := getDriveDetail(descrip, retHandle, &did); err != nil {
			descrip.Error = err.Error()
		}

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

type spDeviceInterfaceData struct {
	CbSize             uint32
	InterfaceClassGuid GUID
	Flags              uint32
	Reserved           uint
}

func setupDiEnumDeviceInterfaces(
	deviceInfoSet syscall.Handle,
	deviceInfoData *spDeviceInformationData,
	deviceInterfaceClassGUID *GUID,
	memberIndex uint32,
) (*spDeviceInterfaceData, error) {
	var deviceInterData spDeviceInterfaceData
	deviceInterData.CbSize = uint32(unsafe.Sizeof(deviceInterData))
	r1, _, e1 := syscall.Syscall6(
		procSetupDiEnumDeviceInterfaces.Addr(),
		5,
		uintptr(deviceInfoSet),
		uintptr(unsafe.Pointer(deviceInfoData)),
		uintptr(unsafe.Pointer(deviceInterfaceClassGUID)),
		uintptr(memberIndex),
		uintptr(unsafe.Pointer(&deviceInterData)),
		uintptr(0),
	)
	if r1 == 0 {
		// ERR_NO_MORE_ITEMS
		if e1 == 259 {
			if memberIndex == 0 {
				return nil, errors.New("no device interfaces")
			} else {
				return nil, nil
			}
		} else if e1 != 0 {
			return nil, error(e1)
		} else {
			return nil, syscall.EINVAL
		}
	}

	return &deviceInterData, nil
}

// setupDiGetDeviceInterfaceDetail returns the device path if known and/or error
func setupDiGetDeviceInterfaceDetail(
	deviceInfoSet syscall.Handle,
	deviceInterfaceData *spDeviceInterfaceData,
) (string, []byte, error) {
	// Determine how large the result needs to be
	var cbSize uint
	_, _, e1 := syscall.Syscall6(
		procSetupDiGetDeviceInterfaceDetailW.Addr(),
		6,
		uintptr(deviceInfoSet),
		uintptr(unsafe.Pointer(deviceInterfaceData)),
		uintptr(0),
		uintptr(0),
		uintptr(unsafe.Pointer(&cbSize)),
		uintptr(0),
	)
	if e1 != 0 && e1 != 122 {
		return "", nil, error(e1)
	}

	// Result is 4 bytes cbSize + result string + 1 byte
	uintSize := int(unsafe.Sizeof(cbSize))
	data := make([]byte, cbSize)
	cbSizePtr := (*uint32)(unsafe.Pointer(&data[0]))
	if uintSize == 8 {
		*cbSizePtr = 8
	} else {
		*cbSizePtr = 6
	}
	// *cbSizePtr = uint32(uintSize + 1)

	_, _, e1 = syscall.Syscall6(
		procSetupDiGetDeviceInterfaceDetailW.Addr(),
		6,
		uintptr(deviceInfoSet),
		uintptr(unsafe.Pointer(deviceInterfaceData)),
		uintptr(unsafe.Pointer(&data[0])),
		uintptr(cbSize),
		uintptr(unsafe.Pointer(&cbSize)),
		// uintptr(cbSize),
		// uintptr(unsafe.Pointer(&cbSize)),
		uintptr(0),
	)
	if e1 != 0 {
		return "", nil, error(e1)
	}

	// convert from wide char
	dec := unicode.UTF16(unicode.LittleEndian, unicode.UseBOM).NewDecoder()
	dataSlice := data[uintSize-4:]
	out, err := dec.Bytes(dataSlice)
	if err != nil {
		return "", nil, err
	}

	i := bytes.IndexByte(out, 0)
	if i == -1 {
		i = len(out)
	}

	devicePath := string(out[:i])
	return devicePath, dataSlice, nil
}

// getDriveDetail returns details about the drive to the descriptor.
func getDriveDetail(descrip *DeviceDescriptor, di syscall.Handle, did *spDeviceInformationData) error {
	var hDevice, hPhysical syscall.Handle = syscall.InvalidHandle, syscall.InvalidHandle
	for i := 0; true; i++ {
		if hDevice != syscall.InvalidHandle {
			_ = syscall.CloseHandle(hDevice)
			hDevice = syscall.InvalidHandle
		}
		if hPhysical != syscall.InvalidHandle {
			_ = syscall.CloseHandle(hPhysical)
			hPhysical = syscall.InvalidHandle
		}

		deviceInterData, err := setupDiEnumDeviceInterfaces(di, did, GUID_DEVINTERFACE_DISK, uint32(i))
		if err != nil {
			return err
		}
		if deviceInterData == nil {
			break
		}

		devicePath, devicePathBin, err := setupDiGetDeviceInterfaceDetail(di, deviceInterData)
		if err != nil {
			return err
		}
		// _ = devicePathBin
		descrip.DevicePath = devicePath

		hDeviceRet, _, lastErr := syscall.Syscall9(
			procCreateFileW.Addr(),
			7,
			uintptr(unsafe.Pointer(&devicePathBin[0])),
			uintptr(0),
			uintptr(syscall.FILE_SHARE_READ|syscall.FILE_SHARE_WRITE),
			uintptr(0),
			uintptr(syscall.OPEN_EXISTING),
			uintptr(syscall.FILE_ATTRIBUTE_NORMAL),
			uintptr(0),
			uintptr(0), uintptr(0),
		)
		hDevice = syscall.Handle(hDeviceRet)
		if hDevice == syscall.InvalidHandle {
			fmt.Printf("%v\n", devicePathBin)
			descrip.Error = "Cannot open handle to device"
			if lastErr != 0 {
				descrip.Error += ": " + lastErr.Error()
			}
			break
		}

		deviceNumber, err := getDeviceNumber(hDevice)
		if err != nil {
			descrip.Error = err.Error()
			break
		}

		descrip.Raw = "\\\\.\\PhysicalDrive" + strconv.Itoa(int(deviceNumber))
		descrip.Device = descrip.Raw
	}
	return nil
}

type diskExtent struct {
	DiskNumber     uint32
	StartingOffset uint64
	ExtentLength   uint64
}

type volumeDiskExtents []byte

func (v *volumeDiskExtents) Len() uint {
	return uint(binary.LittleEndian.Uint32([]byte(*v)))
}

func (v *volumeDiskExtents) Extent(n uint) diskExtent {
	ba := []byte(*v)
	offset := 8 + 24*n
	return diskExtent{
		DiskNumber:     binary.LittleEndian.Uint32(ba[offset:]),
		StartingOffset: binary.LittleEndian.Uint64(ba[offset+8:]),
		ExtentLength:   binary.LittleEndian.Uint64(ba[offset+16:]),
	}
}

type storageDeviceNumber struct {
	DeviceType                    uint64
	DeviceNumber, PartitionNumber uint32
}

func getDeviceNumber(devHandle syscall.Handle) (uint32, error) {
	var sdn storageDeviceNumber
	var size uint32
	ret, _, errNo := syscall.Syscall9(
		procDeviceIoControl.Addr(),
		8,
		uintptr(devHandle),
		// IOCTL_STORAGE_GET_DEVICE_NUMBER
		uintptr(0x2D1080),
		uintptr(0),
		uintptr(0),
		// ptr to disk extents
		uintptr(unsafe.Pointer(&sdn)),
		uintptr(unsafe.Sizeof(sdn)),
		uintptr(unsafe.Pointer(&size)),
		uintptr(0),
		uintptr(0),
	)
	if ret == 0 {
		return 0, error(errNo)
	}

	return sdn.DeviceNumber, nil
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
