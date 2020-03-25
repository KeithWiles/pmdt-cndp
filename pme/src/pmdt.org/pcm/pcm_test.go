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

	fmt.Printf("leaf %d, regs %+v\n", leaf, regs)
}

func TestVersion(t *testing.T) {
	fmt.Printf("Open PCM\n")

	p.ReadCoreCounterConfig()

	fmt.Printf("CoreCounterConfig %+v\n", p.core)
}

func TestClose(t *testing.T) {
	fmt.Printf("Close PCM\n")
}
