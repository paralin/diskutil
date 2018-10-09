package diskutil

import (
	"errors"
	"syscall"
	"unsafe"
)

var (
	kernel32            = syscall.NewLazyDLL("kernel32.dll")
	procCreateFileA     = kernel32.NewProc("CreateFileA")
	procDeviceIoControl = kernel32.NewProc("DeviceIoControl")
	procWriteFile       = kernel32.NewProc("WriteFile")
)

// OpenDiskRaw acquires a file handle to a physicaldisk path using windows API calls.
func OpenDiskRaw(diskPath string) (syscall.Handle, error) {
	// move diskpath to a buffer
	diskPathBin := []byte(diskPath)
	diskPathBuf := make([]byte, len(diskPathBin)+1)
	copy(diskPathBuf, diskPathBin)
	diskPathBuf[len(diskPathBuf)-1] = 0

	/* CreateFileA(
		diskPath,
		GENERIC_READ|GENERIC_WRITE,
		FILE_SHARE_READ|FILE_SHARE_WRITE,
		0,
		OPEN_EXISTING,
		FILE_ATTRIBUTE_NORMAL,
	) */
	ret, _, lastErr := syscall.Syscall9(
		procCreateFileA.Addr(),
		7,
		uintptr(unsafe.Pointer(&diskPathBuf[0])),
		uintptr(syscall.GENERIC_READ|syscall.GENERIC_WRITE),
		// uintptr(syscall.FILE_SHARE_READ|syscall.FILE_SHARE_WRITE),
		uintptr(0),
		uintptr(0),
		uintptr(syscall.OPEN_EXISTING),
		uintptr(syscall.FILE_ATTRIBUTE_NORMAL),
		uintptr(0),
		uintptr(0), uintptr(0), // unused args
	)
	retHandle := syscall.Handle(ret)
	if retHandle == syscall.InvalidHandle {
		if lastErr != 0 {
			return syscall.InvalidHandle, lastErr
		} else {
			return syscall.InvalidHandle, errors.New("unknown error opening disk")
		}
	}

	var dismountStatus uint32
	/* DeviceIoControl(retHandle, FSCTL_DISMOUNT_VOLUME, 0, 0, 0, 0, &status, 0) */
	ret, _, lastErr = syscall.Syscall9(
		procDeviceIoControl.Addr(),
		8,
		uintptr(retHandle),
		uintptr(0x00090020), // FSCTL_DISMOUNT_VOLUME
		uintptr(0), uintptr(0), uintptr(0), uintptr(0),
		uintptr(unsafe.Pointer(&dismountStatus)),
		uintptr(0),
		uintptr(0), // unused arg
	)
	// ignore the result
	_ = ret

	/* DeviceIoControl(retHandle, FSCTL_LOCK_VOLUME, 0, 0, 0, 0, &status, 0) */
	ret, _, lastErr = syscall.Syscall9(
		procDeviceIoControl.Addr(),
		8,
		uintptr(retHandle),
		uintptr(0x00090018), // FSCTL_LOCK_VOLUME
		uintptr(0), uintptr(0), uintptr(0), uintptr(0),
		uintptr(unsafe.Pointer(&dismountStatus)),
		uintptr(0),
		uintptr(0), // unused argument
	)
	// ignore the result
	_ = ret

	// DefineDosDevice

	return retHandle, nil
}
