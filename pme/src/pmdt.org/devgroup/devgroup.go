// SPDX-License-Identifier: BSD-3-Clause
// Copyright(c) 2019-2020 Intel Corporation

package main

import (
	"bytes"
	"fmt"
	"github.com/BurntSushi/toml"
	flags "github.com/jessevdk/go-flags"
	"log"
	"os"

	db "pmdt.org/devbind"
)

var networkClass = db.DeviceConfig{
	Desc: "Network Controller",
	Class: struct {
		ID  db.DevClassID    `toml:"devclass"`
		Sub db.DevSubClassID `toml:"subclass"`
	}{ID: db.NetworkController},
	Vendor:  "",
	Device:  "",
	SVendor: "",
	SDevice: "",
}
var ifpgaClass = db.DeviceConfig{
	Desc: "IFPGA Controller",
	Class: struct {
		ID  db.DevClassID    `toml:"devclass"`
		Sub db.DevSubClassID `toml:"subclass"`
	}{ID: db.ProcessingAccelerators},
	Vendor:  "",
	Device:  "",
	SVendor: "",
	SDevice: "",
}
var encryptionClass = db.DeviceConfig{
	Desc: "Encryption Controller",
	Class: struct {
		ID  db.DevClassID    `toml:"devclass"`
		Sub db.DevSubClassID `toml:"subclass"`
	}{ID: db.EncryptionController},
	Vendor:  "",
	Device:  "",
	SVendor: "",
	SDevice: "",
}
var intelProcessorClass = db.DeviceConfig{
	Desc: "Intel Processor",
	Class: struct {
		ID  db.DevClassID    `toml:"devclass"`
		Sub db.DevSubClassID `toml:"subclass"`
	}{ID: db.ProcessorClass},
	Vendor:  "",
	Device:  "",
	SVendor: "",
	SDevice: "",
}
var caviumSSOClass = db.DeviceConfig{
	Desc: "Cavium SSO",
	Class: struct {
		ID  db.DevClassID    `toml:"devclass"`
		Sub db.DevSubClassID `toml:"subclass"`
	}{ID: db.GenericSystemPeripheral},
	Vendor:  db.CaviumVendorID,
	Device:  "a04b:a04d",
	SVendor: "",
	SDevice: "",
}
var caviumFPAClass = db.DeviceConfig{
	Desc: "Cavium FGA",
	Class: struct {
		ID  db.DevClassID    `toml:"devclass"`
		Sub db.DevSubClassID `toml:"subclass"`
	}{ID: db.GenericSystemPeripheral},
	Vendor:  db.CaviumVendorID,
	Device:  "a053",
	SVendor: "",
	SDevice: "",
}
var caviumPKXClass = db.DeviceConfig{
	Desc: "Cavium PKX",
	Class: struct {
		ID  db.DevClassID    `toml:"devclass"`
		Sub db.DevSubClassID `toml:"subclass"`
	}{ID: db.GenericSystemPeripheral},
	Vendor:  db.CaviumVendorID,
	Device:  "a0dd:a049",
	SVendor: "",
	SDevice: "",
}
var caviumTIMClass = db.DeviceConfig{
	Desc: "Cavium TIM",
	Class: struct {
		ID  db.DevClassID    `toml:"devclass"`
		Sub db.DevSubClassID `toml:"subclass"`
	}{ID: db.GenericSystemPeripheral},
	Vendor:  db.CaviumVendorID,
	Device:  "a051",
	SVendor: "",
	SDevice: "",
}
var caviumZIPClass = db.DeviceConfig{
	Desc: "Cavium ZIP",
	Class: struct {
		ID  db.DevClassID    `toml:"devclass"`
		Sub db.DevSubClassID `toml:"subclass"`
	}{ID: db.ProcessingAccelerators},
	Vendor:  db.CaviumVendorID,
	Device:  "a037",
	SVendor: "",
	SDevice: "",
}
var avpVnicClass = db.DeviceConfig{
	Desc: "AVP NIC",
	Class: struct {
		ID  db.DevClassID    `toml:"devclass"`
		Sub db.DevSubClassID `toml:"subclass"`
	}{ID: db.MemoryController},
	Vendor:  db.RedHatVendorID,
	Device:  "1110",
	SVendor: "",
	SDevice: "",
}
var octeontx2SSOClass = db.DeviceConfig{
	Desc: "Octeonx2 SSO",
	Class: struct {
		ID  db.DevClassID    `toml:"devclass"`
		Sub db.DevSubClassID `toml:"subclass"`
	}{ID: db.GenericSystemPeripheral},
	Vendor:  db.CaviumVendorID,
	Device:  "a0f9:a0fa",
	SVendor: "",
	SDevice: "",
}
var octeontx2NPAClass = db.DeviceConfig{
	Desc: "Octeonx 2NPA",
	Class: struct {
		ID  db.DevClassID    `toml:"devclass"`
		Sub db.DevSubClassID `toml:"subclass"`
	}{ID: db.GenericSystemPeripheral},
	Vendor:  db.CaviumVendorID,
	Device:  "a0fb:a0fc",
	SVendor: "",
	SDevice: "",
}
var ioatBdwClass = db.DeviceConfig{
	Desc: "Intel IOAT Broadwell",
	Class: struct {
		ID  db.DevClassID    `toml:"devclass"`
		Sub db.DevSubClassID `toml:"subclass"`
	}{ID: db.GenericSystemPeripheral},
	Vendor:  "",
	Device:  "6f20:6f21:6f22:6f23:6f24:6f25:6f26:6f27:6f2e:6f2f",
	SVendor: "",
	SDevice: "",
}
var ioatSkxClass = db.DeviceConfig{
	Desc: "Intel IOAT Skylake",
	Class: struct {
		ID  db.DevClassID    `toml:"devclass"`
		Sub db.DevSubClassID `toml:"subclass"`
	}{ID: db.GenericSystemPeripheral},
	Vendor:  "",
	Device:  "2021",
	SVendor: "",
	SDevice: "",
}

// List of all device names
const (
	Network        string = "Network"
	IFPGA          string = "IPPGA"
	Encryption     string = "Encryption"
	IntelProcessor string = "IntelProcessor"
	CaviumFPA      string = "CaviumFPA"
	CaviumPKX      string = "CaviumPKX"
	CaviumSSO      string = "CaviumSSO"
	CaviumTIM      string = "CaviumTIM"
	CaviumZIP      string = "CaviumZIP"
	AVPvNic        string = "AVP-vNic"
	Octeontx2NPA   string = "Octeontx2NPA"
	Octeontx2SSO   string = "Octeontx2SSO"
	IOATBdw        string = "IOAT-Bdw"
	IOATSkx        string = "IOAT-Skx"
)

// DeviceList a list of devices
var DeviceList = map[string]db.DeviceConfig{
	Network:        networkClass,
	IFPGA:          ifpgaClass,
	Encryption:     encryptionClass,
	IntelProcessor: intelProcessorClass,
	CaviumSSO:      caviumSSOClass,
	CaviumFPA:      caviumFPAClass,
	CaviumPKX:      caviumPKXClass,
	CaviumTIM:      caviumTIMClass,
	CaviumZIP:      caviumZIPClass,
	AVPvNic:        avpVnicClass,
	Octeontx2NPA:   octeontx2NPAClass,
	Octeontx2SSO:   octeontx2SSOClass,
	IOATBdw:        ioatBdwClass,
	IOATSkx:        ioatSkxClass,
}

// DeviceGroups for each class of devices
var DeviceGroups = map[string][]*db.DeviceConfig{}

// Options command line options
type Options struct {
	File    string `short:"f" long:"file" description:"validate the TOML file"`
	Devices string `short:"d" long:"devices" description:"show devices"`
	Groups  string `short:"g" long:"groups" description:"show groups"`
	Verbose bool   `short:"v" long:"verbose" description:"Dump each file in TOML"`
}

var options Options
var parser = flags.NewParser(&options, flags.Default)

func outputHeader() {
	fmt.Printf("# SPDX-License-Identifier: BSD-3-Clause\n")
	fmt.Printf("# Copyright(c) 2019-2020 Intel Corporation\n\n")

	fmt.Printf("# TOML configuration of Devices and Groups of devices\n")
	fmt.Printf("#\n")
	fmt.Printf("# Each device maps to a GO structure called devbind.DeviceConfig\n#\n")
	fmt.Printf("# Basic layout of each TOML section to this structure\n#\n")
	fmt.Printf("# [IOAT-Bdw]                        # Device Name\n")
	fmt.Printf("#   group = \"DMAGroup\"            # Group for this device\n")
	fmt.Printf("#   desc = \"Intel IOAT Broadwell\" # Device Description\n")
	fmt.Printf("#   vendor_id = \"8086\"            # Device ID\n")
	fmt.Printf("#   device_id = \"6f20:6f21:6f22:6f23:6f24:6f25:6f26:6f27:6f2e:6f2f\" # List of device IDs\n")
	fmt.Printf("#   svendor_id = \"\"               # Sub Vendor ID\n")
	fmt.Printf("#   sdevice_id = \"\"               # Sub Device ID\n")
	fmt.Printf("#   [IOAT-Bdw.class]                # Device Class and SubClass\n")
	fmt.Printf("#     devclass = \"08\"             # Device Class\n")
	fmt.Printf("#     subclass = \"00\"             # Device SubClass\n")
	fmt.Printf("#\n")
	fmt.Printf("# Note: the indented places use tabs or spaces and must be present\n")
	fmt.Printf("# All IDs are in hex values without leading '0x' prefix\n#\n")
	fmt.Printf("# Groups: NetworkGroup, CryptoGroup, DMAGroup, EventdevGroup\n")
	fmt.Printf("#         MempoolGroup, CompressGroup\n\n")
}

func validateDevBind(files ...string) error {

	for _, f := range files {
		cfgs, grps, err := db.ValidateDeviceFile(f)
		if err != nil {
			return err
		}
		if options.Verbose {
			outputHeader()

			buf := new(bytes.Buffer)
			if err := toml.NewEncoder(buf).Encode(cfgs); err != nil {
				log.Fatalf("devgroup: %v\n", err)
			}
			fmt.Println(buf.String())

			buf = new(bytes.Buffer)
			if err := toml.NewEncoder(buf).Encode(grps); err != nil {
				log.Fatalf("devgroup: %v\n", err)
			}
			fmt.Println("\n====================================")
			fmt.Printf("Group List =\n%s\n", buf.String())
		}
		fmt.Printf("File %s is valid\n", f)
	}
	return nil
}

func main() {

	args, err := parser.Parse()
	if err != nil {
		fmt.Printf("*** invalid arguments %v\n", err)
		os.Exit(1)
	}

	if len(options.File) > 0 {
		if err := validateDevBind(options.File); err != nil {
			log.Fatalf("devgroup: %v\n", err)
		}
		fmt.Printf("File %s is valid\n", options.File)
		os.Exit(0)
	}

	if err := validateDevBind(args...); err != nil {
		log.Fatalf("devgroup: %v\n", err)
	}

	/*
	   	var allDevices map[string]db.DeviceConfig

	   	if _, err := toml.DecodeFile("devices.toml", &allDevices); err != nil {
	   		fmt.Printf("devgroup: failed to decode file, %+v\n", allDevices)
	   		log.Fatalf("devgroup: %v\n", err)
	   	}

	   	fmt.Printf("allDevices: %+v\n", allDevices)

	   	fmt.Printf("\n# ****** TOML decode of device.toml file\n")
	   	buf := new(bytes.Buffer)
	   	if err := toml.NewEncoder(buf).Encode(allDevices); err != nil {
	   		log.Fatalf("devgroup: %v\n", err)
	   	}
	   	fmt.Println(buf.String())

	   	for _, v := range allDevices {
	   		DeviceGroups[v.Group] = append(DeviceGroups[v.Group], &v)
	   	}
	   	fmt.Printf("DeviceGroups: %+v\n", DeviceGroups)

	   	outputHeader()

	   	buf = new(bytes.Buffer)
	   	for k, g := range DeviceGroups {
	   		fmt.Printf("[%s]\n", k)
	   //		for _, d := range g {
	   			if err := toml.NewEncoder(buf).Encode(g); err != nil {
	   				log.Fatalf("devgroup: %v", err)
	   			}
	   			fmt.Printf("%s\n", buf.String())
	   //		}
	   	}

	   	fmt.Printf("\n# ***** Decode allDevices\n")

	   	var foobar map[string]db.DeviceConfig

	   	if _, err := toml.Decode(buf.String(), &foobar); err != nil {
	   		log.Fatalf("devgroup: %v\n", err)
	   	}

	   	fmt.Printf("foobar: %+v\n", foobar)
	*/
}
