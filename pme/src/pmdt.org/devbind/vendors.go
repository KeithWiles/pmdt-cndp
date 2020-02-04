// SPDX-License-Identifier: BSD-3-Clause
// Copyright(c) 2019-2020 Intel Corporation

package devbind

import (
	"fmt"
	"github.com/BurntSushi/toml"

	"github.com/davecgh/go-spew/spew"
	tlog "pmdt.org/ttylog"
)

// DeviceID type
type DeviceID string

// SDeviceID type
type SDeviceID string

// VendorID type
type VendorID string

// SVendorID type
type SVendorID string

// DevClassID value
type DevClassID string

// DevSubClassID value
type DevSubClassID string

// DeviceClassNames in string format
var DeviceClassNames = map[DevClassID]string{
	UnclassifiedDevice:               "UnclassifiedDevice",
	MassStorageController:            "MassStorageController",
	NetworkController:                "NetworkController",
	DisplayController:                "DisplayController",
	MultimediaController:             "MultimediaController",
	MemoryController:                 "MemoryController",
	BridgeController:                 "BridgeController",
	CommunicationController:          "CommunicationController",
	GenericSystemPeripheral:          "GenericSystemPeripheral",
	InputDeviceController:            "InputDeviceController",
	DockingStation:                   "DockingStation",
	ProcessorClass:                   "ProcessorClass",
	SerialBusController:              "SerialBusController",
	WirelessController:               "WirelessController",
	IntelligentController:            "IntelligentController",
	SatelliteCommunicationController: "SatelliteCommunicationController",
	EncryptionController:             "EncryptionController",
	SignalProcessingController:       "SignalProcessingController",
	ProcessingAccelerators:           "ProcessingAccelerators",
	NonEssentialInstrumentation:      "NonEssentialInstrumentation",
	Coprocessor:                      "Coprocessor",
}

// PCI Device class values
const (
	UnclassifiedDevice               DevClassID = "00"
	MassStorageController            DevClassID = "01"
	NetworkController                DevClassID = "02"
	DisplayController                DevClassID = "03"
	MultimediaController             DevClassID = "04"
	MemoryController                 DevClassID = "05"
	BridgeController                 DevClassID = "06"
	CommunicationController          DevClassID = "07"
	GenericSystemPeripheral          DevClassID = "08"
	InputDeviceController            DevClassID = "09"
	DockingStation                   DevClassID = "0a"
	ProcessorClass                   DevClassID = "0b"
	SerialBusController              DevClassID = "0c"
	WirelessController               DevClassID = "0d"
	IntelligentController            DevClassID = "0e"
	SatelliteCommunicationController DevClassID = "0f"
	EncryptionController             DevClassID = "10"
	SignalProcessingController       DevClassID = "11"
	ProcessingAccelerators           DevClassID = "12"
	NonEssentialInstrumentation      DevClassID = "13"
	Coprocessor                      DevClassID = "40"
)

// VendorID values
const (
	IntelVendorID  VendorID = "8086"
	CaviumVendorID VendorID = "177d"
	RedHatVendorID VendorID = "1af4"
)

// Group constant strings
const (
	NetworkGroup  string = "NetworkGroup"
	CryptoGroup   string = "CryptoGroup"
	DMAGroup      string = "DMAGroup"
	EventdevGroup string = "EventdevGroup"
	MempoolGroup  string = "MempoolGroup"
	CompressGroup string = "CompressGroup"
)

// ValidGroups currently allowed
var ValidGroups = []string{
	NetworkGroup,
	CryptoGroup,
	DMAGroup,
	EventdevGroup,
	MempoolGroup,
	CompressGroup,
}

// DeviceClass - Network Class information
type DeviceClass struct {
	Slot  string
	Class struct {
		ID  DevClassID
		Sub DevSubClassID
		Str string
	}
	Vendor struct {
		ID  VendorID
		Str string
	}
	Device struct {
		ID  DeviceID
		Str string
	}
	SVendor struct {
		ID  SVendorID
		Str string
	}
	SDevice struct {
		ID  SDeviceID
		Str string
	}
	Rev       string
	Driver    string
	Module    string
	Unused    string
	Interface string
	NumaNode  string
	SSHIf     string
	Active    bool
}

// DeviceConfig - Network Class information
type DeviceConfig struct {
	Group string `toml:"group"`
	Desc  string `toml:"desc"`
	Class struct {
		ID  DevClassID    `toml:"devclass"`
		Sub DevSubClassID `toml:"subclass"`
	} `toml:"class"`
	Vendor  VendorID  `toml:"vendor_id"`
	Device  DeviceID  `toml:"device_id"`
	SVendor SVendorID `toml:"svendor_id"`
	SDevice SDeviceID `toml:"sdevice_id"`
}

// DeviceList is a list of all devices found in system
type DeviceList map[string]*DeviceClass

// DevConfigs is list of devices in devices.toml
type DevConfigs map[string]*DeviceConfig

// DevGroups is the map of devices to a group name
type DevGroups map[string][]*DeviceConfig

func isGroup(group string) bool {

	for _, g := range ValidGroups {
		if g == group {
			return true
		}
	}
	return false
}

// ValidateDeviceFile is a valid TOML file for devbind
func ValidateDeviceFile(file string) (DevConfigs, DevGroups, error) {

	cfgs := make(DevConfigs)
	grps := make(DevGroups)

	if _, err := toml.DecodeFile(file, &cfgs); err != nil {
		return nil, nil, err
	}

	// Create the device class grouping from the device list
	for _, v := range cfgs {
		grps[v.Group] = append(grps[v.Group], v)
		if !isGroup(v.Group) {
			return nil, nil, fmt.Errorf("group (%s) is not valid %v",
				v.Group, ValidGroups)
		}
	}
	tlog.DebugPrintf("Groups:\n%s\n", spew.Sdump(grps))

	return cfgs, grps, nil
}

// LoadDeviceFile and create the config and groups
func LoadDeviceFile(files string) (DevConfigs, DevGroups, error) {

	return ValidateDeviceFile(files)
}
