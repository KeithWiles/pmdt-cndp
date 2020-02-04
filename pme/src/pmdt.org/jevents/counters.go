// SPDX-License-Identifier: BSD-3-Clause
// Copyright(c) 2019-2020 Intel Corporation

package jevents

import (
	"pmdt.org/perf"
	tlog "pmdt.org/ttylog"
)

type hwLabel struct {
	Name, Alias string
	EventNum    uint64
}

// AddNewHardwareLabel - add a new entry to the hardwareLabels
func AddNewHardwareLabel(name, alias string, eventNum uint64) {

	tlog.DebugPrintf("AddNewHardwareLabel: name: %s, alias: %s, eventNum: %d\n", name, alias, eventNum)
	perf.AddHardwareLabel(name, alias, eventNum)
}
