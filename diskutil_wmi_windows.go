//+build windows,!no_wmi

package diskutil

import (
	"github.com/StackExchange/wmi"
)

type Win32_DiskDrive struct {
	Caption       string
	Name          string
	DeviceID      string
	Model         string
	Index         int
	Partitions    int
	Size          int
	PNPDeviceID   string
	Status        string
	SerialNumber  string
	Manufacturer  string
	MediaType     string
	Description   string
	SystemName    string
	InterfaceType string
}

// ListStorageDevices lists all connected storage devices.
func ListStorageDevices() ([]*DeviceDescriptor, error) {
	var res []*DeviceDescriptor

	var diskDrives []*Win32_DiskDrive
	if err := wmi.Query("select * from win32_diskdrive", &diskDrives); err != nil {
		return nil, err
	}

	for _, drive := range diskDrives {
		res = append(res, &DeviceDescriptor{
			InterfaceType: drive.InterfaceType,
			Device:        drive.DeviceID,
			DevicePath:    drive.Name,
			Description:   drive.Caption,
		})
	}

	return res, nil
}
