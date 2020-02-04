// SPDX-License-Identifier: BSD-3-Clause
// Copyright(c) 2019-2020 Intel Corporation

package devbind

import (
	"fmt"
	"log"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/davecgh/go-spew/spew"
	tlog "pmdt.org/ttylog"
)

// BindInfo - Device Binding information
type BindInfo struct {
	Devices    DeviceList
	CfgDevices DevConfigs
	Groups     DevGroups
}

// UioModules supported
var UioModules = []string{"igb_uio", "vfio-pci", "uio_pci_generic"}

func init() {
	tlog.Register("devBindLog")
}

// New - create a new DevBindInfo structure
func New(devFile ...string) *BindInfo {

	db := &BindInfo{}
	file := ""
	if len(devFile) == 0 {
		file = "devices.toml"
	} else {
		if len(devFile) > 1 {
			return nil
		}
		file = devFile[0]
	}

	tlog.InfoPrintf("Process TOML device file (%s)\n", file)

	db.Devices = make(DeviceList)
	db.CfgDevices = make(DevConfigs)
	db.Groups = make(DevGroups)

	db.getDetails(db.CfgDevices)

	for k, d := range db.Devices {
		tlog.DebugPrintf("Slot: %s = %+v\n", k, d)
	}

	if cfgs, grps, err := LoadDeviceFile(file); err != nil {
		log.Fatal(err)
	} else {
		db.CfgDevices = cfgs
		db.Groups = grps
	}
	tlog.DebugPrintf("===== CfgDevices:\n%s\n", spew.Sdump(db.CfgDevices))
	tlog.DebugPrintf("===== Groups:\n%s\n", spew.Sdump(db.Groups))

	db.getDetails(db.CfgDevices)

	return db
}

func checkOutput(args ...string) (string, error) {

	baseCmd := args[0]
	cmdArgs := args[1:]

	cmd := exec.Command(baseCmd, cmdArgs...)
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func (db *BindInfo) getDeviceDetails(dev *DeviceClass) {

	paths, err := filepath.Glob(fmt.Sprintf("/sys/bus/pci/devices/%s/net", dev.Slot))
	if err != nil {
		return
	}

	for _, path := range paths {
		files := []string{}

		ifaces, err := filepath.Glob(path + "/*")

		if err != nil {
			continue
		}

		tlog.DebugPrintf("Interfaces: %s\n", ifaces)

		for _, iface := range ifaces {
			files = append(files, filepath.Base(iface))
		}
		dev.Interface = strings.Join(files, ",")
	}
	dev.SSHIf = ""
	dev.Active = false

	out, err := checkOutput("ip", "-o", "route")
	if err != nil {
		return
	}
	devLines := strings.Split(out, "\n")

	for _, route := range devLines {
		words := strings.Split(route, " ")

		if strings.Contains(words[0], "169.254.") {
			continue
		}
		for i, word := range words {
			if word == "dev" {
				iface := words[i+1]
				for _, dev := range db.Devices {
					if dev.Interface == iface {
						dev.Active = true
					}
				}
			}
		}
	}
}

func (db *BindInfo) getDetails(devicesType DevConfigs) {

	out, err := checkOutput("lspci", "-Dvmmnnk")
	if err != nil {
		return
	}
	devLines := strings.Split(out, "\n")

	dev := &DeviceClass{}
	for _, devLine := range devLines {
		if len(devLine) == 0 {
			for _, d := range devicesType {
				if len(dev.Slot) == 0 {
					break
				}
				if compareDevices(d, dev) {
					db.Devices[dev.Slot] = dev
					tlog.DebugPrintf("Add: Slot %s, Class: %v, Vendor %s, Device %s, SVendor %s, SDevice %s\n",
						dev.Slot, dev.Class.ID, dev.Vendor.ID, dev.Device.ID, dev.SVendor.ID, dev.SDevice.ID)
					break
				}
			}
			dev = &DeviceClass{}
		} else {
			vals := strings.Split(devLine, "\t")
			name := strings.TrimRight(vals[0], ":")
			value := vals[1]

			switch name {
			case "Slot":
				dev.Slot = value
				tlog.DebugPrintf("\nSlot   : %s\n", dev.Slot)
			case "Class":
				id, _, desc := devClassID(value)
				dev.Class.ID = id
				dev.Class.Str = desc
				tlog.DebugPrintf("Class  : [%02x] %s\n", dev.Class.ID, dev.Class.Str)
			case "Vendor":
				id, desc := getDeviceID(value)
				dev.Vendor.ID = VendorID(id)
				dev.Vendor.Str = desc
				tlog.DebugPrintf("Vendor : [%s] %s\n", dev.Vendor.ID, dev.Vendor.Str)
			case "Device":
				id, desc := getDeviceID(value)
				dev.Device.ID = DeviceID(id)
				dev.Device.Str = desc
				tlog.DebugPrintf("Device : [%s] %s\n", dev.Device.ID, dev.Device.Str)
			case "SVendor":
				id, desc := getDeviceID(value)
				dev.SVendor.ID = SVendorID(id)
				dev.SVendor.Str = desc
				tlog.DebugPrintf("SVendor: [%s] %s\n", dev.SVendor.ID, dev.SVendor.Str)
			case "SDevice":
				id, desc := getDeviceID(value)
				dev.SDevice.ID = SDeviceID(id)
				dev.SDevice.Str = desc
				tlog.DebugPrintf("SDevice: [%s] %s\n", dev.SDevice.ID, dev.SDevice.Str)
			case "Rev":
				dev.Rev = value
			case "Driver":
				dev.Driver = value
			case "Module":
				dev.Module = value
			case "NUMANode":
				dev.NumaNode = value
			}
		}
	}
	delete(db.Devices, "")

	for _, dev := range db.Devices {
		db.getDeviceDetails(dev)
	}
}

func devClassID(str string) (DevClassID, DevSubClassID, string) {

	start := strings.Index(str, "[")

	if start == -1 {
		return "FF", "FF", ""
	}
	devClass := str[start+1 : start+3]
	subClass := str[start+2 : start+5]

	return DevClassID(devClass), DevSubClassID(subClass), str[:start-1]
}

func getDeviceID(str string) (string, string) {
	start := strings.Index(str, "[")
	end := strings.Index(str, "]")

	if start == -1 || end == -1 {
		return "", ""
	}
	return str[start+1 : end], str[:start-1]
}

// FindDevicesByDeviceClass all devices matching the given device class
func (db *BindInfo) FindDevicesByDeviceClass(name string, devClasses []*DeviceConfig) map[string]*DeviceClass {

	tlog.DebugPrintf("Find devices for %s\n", name)

	devices := make(map[string]*DeviceClass)

	if len(db.Devices) == 0 {
		tlog.WarnPrintf("Devices list is empty\n")
		return nil
	}
	tlog.DebugPrintf("devClasses: %+v\n", devClasses)
	for _, d := range devClasses {
		tlog.DebugPrintf("Search for: Class %v, Vendor %s, Device %s\n",
			d.Class.ID, d.Vendor, d.Device)
	}
	for k, dev := range db.Devices {
		for _, dc := range devClasses {
			if compareDevices(dc, dev) {
				tlog.DebugPrintf("dev: %v, %v, %v  **Found**\n",
					dev.Class.ID, dev.Vendor.ID, dev.Device.ID)
				devices[k] = dev
			}
		}
	}
	return devices
}

func compareDevices(d1 *DeviceConfig, d2 *DeviceClass) bool {

	cmpSDevice := func(d1 *DeviceConfig, d2 *DeviceClass) bool {
		if len(d1.SDevice) >= 0 && len(d2.SDevice.ID) > 0 {
			if len(d1.SDevice) == 0 || strings.Contains(string(d1.Device), string(d2.SDevice.ID)) {
				return true
			}
		}
		return false
	}

	if (len(d1.Class.ID) == 0 || d1.Class.ID == d2.Class.ID) &&
		(len(d1.Vendor) == 0 || d2.Vendor.ID == d1.Vendor) &&
		(len(d1.Device) == 0 || strings.Contains(string(d1.Device), string(d2.Device.ID))) &&
		cmpSDevice(d1, d2) {
		return true
	}
	return false
}
