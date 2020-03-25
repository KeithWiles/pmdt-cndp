// SPDX-License-Identifier: BSD-3-Clause
// Copyright(c) 2019-2020 Intel Corporation

package pcm

import (
)

// CoreCounterConfig information
type CoreCounterConfig struct {
	perfmonVer uint32
	genCounterMax uint32
	genCounterWidth uint32

	fixedCounterMax uint32
	fixedCounterWidth uint32
}

// Info is the data structure for PCM information
type Info struct {
	core CoreCounterConfig
}

// Create a PCM data structure
func Create() (*Info) {
	return &Info{}
}

// ReadCoreCounterConfig returns the version ID
func (p *Info) ReadCoreCounterConfig() {

	arr := p.CPUid(0x0a)

	c := &p.core
	c.perfmonVer = extractBits(arr.eax, 0, 7)
	c.genCounterMax = extractBits(arr.eax, 8, 15)
	c.genCounterWidth = extractBits(arr.eax, 16, 23)

	if c.perfmonVer > 1 {
		c.fixedCounterMax = extractBits(arr.ecx, 0, 4)
		c.fixedCounterWidth = extractBits(arr.ecx, 5, 12)
	}
}
