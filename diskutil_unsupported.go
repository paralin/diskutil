//+build !windows

package diskutil

import (
	"io"
	"runtime"

	"github.com/pkg/errors"
)

var errUnsupported = errors.Errorf("unsupported on platform: %s @ %s", runtime.GOOS, runtime.GOARCH)

// FlashToDisk clears a disk and flashes a image.
// Progress callback is called with updates.
func FlashToDisk(
	reader io.Reader,
	imageSize uint64,
	diskPath string,
	progressCb func(percent int, status string),
) error {
	return errUnsupported
}

// ListStorageDevices lists all connected storage devices.
func ListStorageDevices() ([]*DeviceDescriptor, error) {
	return nil, errUnsupported
}
