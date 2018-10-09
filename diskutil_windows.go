package diskutil

import (
	"io"
	"io/ioutil"
	"os/exec"
	"regexp"
	"strconv"
	"syscall"
	"time"

	"github.com/pkg/errors"
)

var diskNoRe = regexp.MustCompile(`(\\\\\.\\PHYSICALDRIVE)(\d+)`)

// CleanDisk uses diskutil to clean a drive by number.
func CleanDisk(diskNum int) error {
	// Run diskpart script:
	//  - select disk 3
	//  - clean
	//  - rescan
	scrFile, err := ioutil.TempFile("", "*.txt")
	if err != nil {
		return errors.Wrap(err, "open temporary file")
	}
	scrFilePath := scrFile.Name()
	scrFile.WriteString("select disk ")
	scrFile.WriteString(strconv.Itoa(diskNum))
	scrFile.WriteString("\nclean\nrescan\n")
	if err := scrFile.Close(); err != nil {
		return errors.Wrap(err, "close temporary file")
	}

	// requires elevation
	diskPart := exec.Command("diskpart", "/s", scrFilePath)
	if err := diskPart.Run(); err != nil {
		return errors.Wrap(err, "diskpart clean disk")
	}

	return nil
}

// FlashToDisk clears a disk and flashes a image.
// Progress callback is called with updates.
// DiskPath on windows is expected to be \\.\PHYSICALDRIVE{N}
func FlashToDisk(
	image io.Reader,
	imageSize uint64, // imageSize in bytes for progress, if zero, disables progress.
	diskPath string,
	progressCb func(percent int, status string),
) error {
	imageSizeF := float64(imageSize)
	// Read the first chunk, we want to defer writing this.
	const chunkSize = 65536

	var firstChunk [chunkSize]byte
	if _, err := io.ReadAtLeast(image, firstChunk[:], chunkSize); err != nil {
		return errors.Wrap(err, "read first chunk")
	}

	diskNum, err := getDiskNumber(diskPath)
	if err != nil {
		return err
	}

	progressCb(5, "Cleaning disk...")
	if err := CleanDisk(diskNum); err != nil {
		return err
	}

	// wait 2 seconds for partitions to update
	for i := 0; i < 10; i++ {
		progressCb(10+i, "Windows rescanning disk...")
		<-time.After(time.Millisecond * 200)
	}

	progressCb(23, "Opening and locking disk...")
	f, err := OpenDiskRaw(diskPath)
	if err != nil {
		return err
	}
	defer syscall.CloseHandle(f)

	progressCb(25, "Initializing transfer...")

	// Pipe the rest.
	chunkBuf := make([]byte, chunkSize)
	writeOffset := int64(chunkSize)
	lastProgressUpdate := 0
	writeProgress := func() {
		lastProgressUpdate = 0
		// compute percent written
		// maximum offset is
		percent := uint32(25) + uint32((float64(writeOffset)/imageSizeF)*float64(75))
		progressCb(int(percent), "Flashing image to disk...")
	}

	if _, err := syscall.SetFilePointer(f, chunkSize, nil, syscall.FILE_BEGIN); err != nil {
		return err
	}

	for {
		nr, er := image.Read(chunkBuf)
		if nr > 0 && nr < chunkSize {
			zeroSeg := chunkBuf[nr:]
			for i := range zeroSeg {
				zeroSeg[i] = 0
			}
		}
		if nr > 0 {
			var nw uint32
			ew := syscall.WriteFile(f, chunkBuf, &nw, nil)
			if ew != nil {
				err = ew
				break
			}
			if chunkSize != int(nw) {
				err = io.ErrShortWrite
				break
			}

			writeOffset += chunkSize
			if imageSize != 0 {
				lastProgressUpdate++
				if lastProgressUpdate == 15 {
					writeProgress()
				}
			}
		}
		if er != nil {
			if er != io.EOF {
				err = er
			}
			break
		}
	}
	if err != nil {
		return errors.Wrap(err, "copy image to disk")
	}

	if _, err := syscall.SetFilePointer(f, 0, nil, syscall.FILE_BEGIN); err != nil {
		return err
	}

	{
		var nw uint32
		ew := syscall.WriteFile(f, firstChunk[:], &nw, nil)
		if ew != nil {
			return errors.Wrap(ew, "write header chunk to disk")
		}
	}

	progressCb(100, "Done flashing.")
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

	if d == 0 {
		return 0, errors.Errorf("physicaldrive 0 selected, this is probably in error: %v", matches)
	}

	return d, nil
}
