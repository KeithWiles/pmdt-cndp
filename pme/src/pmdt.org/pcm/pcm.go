// SPDX-License-Identifier: BSD-3-Clause
// Copyright(c) 2019-2020 Intel Corporation

package pcm

import (
//	"encoding/binary"
//	"strings"
)

// Info is the data structure for PCM information
type Info struct {
	dummy int
}

// Create a PCM data structure
func Create() (*Info) {
	return &Info{}
}
