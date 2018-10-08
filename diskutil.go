package diskutil

// DeviceDescriptor is an element in the storage devices list.
type DeviceDescriptor struct {
	// InterfaceType is the interface type of this device.
	// Possibly: USB
	InterfaceType string
	// BusType is the bus type of this device.
	BusType string
	// BusVersion is the bus version of this device.
	BusVersion string
	// Device is the device id.
	Device string
	// DevicePath is the path to the device.
	DevicePath string
	// Description is the description of this device.
	Description string
	// Error was any error when accessing the device.
	Error string
	// Size is the size of the device.
	Size uint64
	// Mountpoints is the list of mountpoints in use for this device.
	Mountpoints []string

	IsReadOnly  bool
	IsSystem    bool
	IsVirtual   bool
	IsRemovable bool
	IsCard      bool // IsCard indicates device is a SD Card.
	IsScsi      bool
	IsUSB       bool
	IsUAS       bool
}
