// SPDX-License-Identifier: BSD-3-Clause
// Copyright(c) 2019-2020 Intel Corporation

// +build !linux

package dpdk

import (
	"fmt"
	"testing"
)

var pi *ProcessInfo

func TestOpen(t *testing.T) {

	pi = NewProcessInfo()

	if err := pi.Open(); err != nil {
		t.Errorf("Open() failed: %v", err)
	}

	fmt.Println("Process  Open: OK")
}

func TestInfo(t *testing.T) {

	for _, a := range pi.AppList() {
		info := pi.Info(a)

		fmt.Printf("DPDK Info    : %+v\n", info)
	}
}

func TestCommands(t *testing.T) {

	for _, a := range pi.AppList() {
		cmds, err := pi.Commands(a)
		if err != nil {
			t.Errorf("unable to retrive commands: %v", err)
		} else {
			fmt.Printf("Commands     : %v\n", cmds)
		}
	}
}

func TestArgs(t *testing.T) {

	for _, a := range pi.AppList() {
		d, err := pi.Args(a)
		if err != nil {
			t.Errorf("command list failed: %v", err)
		} else {
			fmt.Printf("Args         : %v\n", d)
		}
	}
}

func TestFiles(t *testing.T) {

	files := pi.Files()
	if len(files) > 0 {
		fmt.Printf("Files        : %v\n", files)
	}
}

func TestEthdevStats(t *testing.T) {

	for _, a := range pi.AppList() {
		d, err := pi.EthdevList(a)
		if err != nil {
			t.Errorf("EthdevList failed: %v", err)
		} else {
			fmt.Printf("EthdevList   : %v\n", d)

			for _, p := range d.Ports {
				stats, err := pi.EthdevStats(a, p.PortID)
				if err != nil {
					t.Errorf("EthedevStats Failed: %v", err)
				}
				fmt.Printf("EthdevStats  : %+v\n\n", stats)
			}
		}
	}
}

func TestEthdevXStats(t *testing.T) {

	for _, a := range pi.AppList() {
		d, err := pi.EthdevList(a)
		if err != nil {
			t.Errorf("EthdevList failed: %v", err)
		} else {
			fmt.Printf("EthdevList   : %v\n", d)

			for _, p := range d.Ports {
				stats, err := pi.EthdevXStats(a, p.PortID)
				if err != nil {
					t.Errorf("EthedevXStats Failed: %v", err)
				}
				fmt.Printf("EthdevXStats  : %+v\n\n", stats)
			}
		}
	}
}

func TestRawdevStats(t *testing.T) {

	for _, a := range pi.AppList() {
		d, err := pi.RawdevList(a)
		if err != nil {
			t.Errorf("RawdevList failed: %v", err)
		} else {
			fmt.Printf("RawdevList   : %v\n", d)

			for _, p := range d.Ports {
				stats, err := pi.RawdevStats(a, p.PortID)
				if err != nil {
					t.Errorf("RawdevStats Failed: %v", err)
				}
				fmt.Printf("RawdevStats   : %+v\n\n", stats)
			}
		}
	}
}

func TestRawdevXStats(t *testing.T) {

	for _, a := range pi.AppList() {
		d, err := pi.RawdevList(a)
		if err != nil {
			t.Errorf("RawdevList failed: %v", err)
		} else {
			fmt.Printf("RawdevList   : %v\n", d)

			for _, p := range d.Ports {
				stats, err := pi.RawdevXStats(a, p.PortID)
				if err != nil {
					t.Errorf("RawdevXStats Failed: %v", err)
				}
				fmt.Printf("RawdevXStats   : %+v\n\n", stats)
			}
		}
	}
}

func TestClose(t *testing.T) {

	if pi == nil {
		t.Errorf("ProcessInfo pointer is nil")
	} else {
		pi.Close()
	}
	fmt.Println("Process Close: OK")
}
