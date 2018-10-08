package diskutil

import (
	"regexp"
	"strconv"

	"github.com/pkg/errors"
)

var diskNoRe = regexp.MustCompile(`(\\\\\.\\PHYSICALDRIVE)(\d+)`)

// FlashToDisk clears a disk and flashes a image.
// Progress callback is called with updates.
// DiskPath on windows is expected to be \\.\PHYSICALDRIVE{N}
func FlashToDisk(
	diskPath string,
	progressCb func(percent int, status string),
) error {
	diskNum, err := getDiskNumber(diskPath)
	if err != nil {
		return err
	}

	// Run diskpart script:
	//  - select disk 3
	//  - clean
	//  - rescan
	_ = diskNum

	return nil
}

// getDiskNumber matches the disk number in a PHYSICALDRIVE path.
func getDiskNumber(diskPath string) (int, error) {
	matches := diskNoRe.FindStringSubmatch(diskPath)
	if len(matches) != 3 {
		return 0, errors.Errorf(
			"not windows physicaldrive path: %s",
			diskPath,
		)
	}

	d, err := strconv.Atoi(matches[2])
	if err != nil {
		return 0, errors.Errorf(
			"windows physicaldrive path mismatch: %s %v",
			diskPath,
			err,
		)
	}
	return d, nil
}
