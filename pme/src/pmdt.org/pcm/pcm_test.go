// SPDX-License-Identifier: BSD-3-Clause
// Copyright(c) 2019-2020 Intel Corporation

package pcm

import (
	"fmt"

	"testing"
)

var p *Info

func TestOpen(t *testing.T) {
	fmt.Printf("Open PCM\n")

	p = Create()

	leaf := 0x0a
	regs := p.CPUid(leaf)

	fmt.Printf("  CPUid Registers: %+v\n", regs)
}

func TestVersion(t *testing.T) {

	p.ReadCoreCounterConfig()

	fmt.Printf("  CoreCounterConfig: %+v\n", p.core)
}

func TestDetectModel(t *testing.T) {

	ok, bs := p.DetectModel()

	if !ok {
		fmt.Printf("Failed to Detect Model\n")
		return
	}
	fmt.Printf("  Model: %s\n", bs)
	fmt.Printf("     cpuFamily %d, cpuModel %d, cpuStepping %d, hypervisor: %v\n",
		p.cpuFamily, p.cpuModel, p.cpuStepping, p.hypervisorDetected)
	fmt.Printf("     IBRS and IBPD supported: %v\n", p.ib)
	fmt.Printf("     STIBP supporte: %v\n", p.stibp)
	fmt.Printf("     Spec arch caps supported: %v\n", p.archCaps)
}


func TestClose(t *testing.T) {
	fmt.Printf("Close PCM\n")
}
