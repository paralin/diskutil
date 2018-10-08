package diskutil

import (
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"time"

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
	scrFile, err := ioutil.TempFile("", "*.txt")
	if err != nil {
		return errors.Wrap(err, "open temporary file")
	}
	scrFilePath := scrFile.Name()
	// defer os.Remove(scrFilePath)

	progressCb(5, "Cleaning disk (may request elevation)...")
	scrFile.WriteString("select disk ")
	scrFile.WriteString(strconv.Itoa(diskNum))
	scrFile.WriteString("\nclean\nrescan\n")
	if err := scrFile.Close(); err != nil {
		return errors.Wrap(err, "close temporary file")
	}

	// requires elevation
	// diskPart := exec.Command("diskpart", "/s", scrFilePath)
	// powershell -ExecutionPolicy ByPass -Command "Start-Process cmd -ArgumentList '/c','diskpart','/s','C:\Users\kidov\Desktop\diskutil-script.txt'" -Verb RunAs
	diskPart := exec.Command(
		"powershell",
		"-ExecutionPolicy", "ByPass",
		"-Command", "Start-Process cmd -ArgumentList '/c','diskpart','/s','"+scrFilePath+"'",
		"-Verb", "RunAs",
	)
	if err := diskPart.Run(); err != nil {
		return errors.Wrap(err, "diskpart clean disk")
	}

	// wait 2 seconds for partitions to update
	for i := 0; i < 10; i++ {
		progressCb(10+i, "Waiting for Windows to rescan the disk...")
		<-time.After(time.Millisecond * 200)
	}

	progressCb(23, "Opening disk...")
	f, err := os.OpenFile(diskPath, os.O_SYNC|os.O_RDWR, 0777)
	if err != nil {
		return err
	}
	f.Close()

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
