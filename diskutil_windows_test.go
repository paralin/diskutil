package diskutil

import (
	"testing"
)

func TestGetDiskNumberRe(t *testing.T) {
	match := diskNoRe.FindStringSubmatch(`\\.\PHYSICALDRIVE1`)
	if len(match) != 3 {
		t.Fail()
	}
	if match[2] != "1" {
		t.Fail()
	}
}

func TestGetDiskNumber(t *testing.T) {
	d, err := getDiskNumber(`\\.\PHYSICALDRIVE1`)
	if err != nil {
		t.Fatal(err.Error())
	}
	if d != 1 {
		t.Fail()
	}
}
