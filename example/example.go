package main

import (
	"encoding/json"
	"fmt"

	"github.com/aperturerobotics/diskutil"
)

func main() {
	devs, err := diskutil.ListStorageDevices()
	if err != nil {
		panic(err)
	}
	d, _ := json.MarshalIndent(devs, "", "  ")
	fmt.Println(string(d))
}
