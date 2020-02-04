// SPDX-License-Identifier: BSD-3-Clause
// Copyright(c) 2019-2020 Intel Corporation

package jevents

import (
	"pmdt.org/perf"
)

// CPUEvent is information for each CPU
type CPUEvent struct {
	enabled bool        // CPU event is enabled
	cpuID   int         // CPU Id value
	Event   *perf.Event // Event structure created by perf.Open()
	Count   perf.Count  // Count structure returned on MeasureEnd()
	Scaled  uint64      // Event Scaled value of event Count
}

// PerfEvent - local event structure
type PerfEvent struct {
	EventName   string      // Name of this event
	Attr        *perf.Attr  // perf attribute structure for PerfOpenEvent()
	GroupLeader bool        // true when this event is the group leader
	EndGroup    bool        // true when this is the last event in a group
	Uncore      bool        // true when a event is for the uncore
	CPUEvents   []*CPUEvent // Event information per CPU
	TotalSum    uint64      // Scaled Sum of all CPUEvents
}

// PerfEventInfo - Event information
type PerfEventInfo struct {
	Event string // Event name or string
	Desc  string // Descrition of the event
	Pmu   string // PMU string
}

// eventInfoMap - Global list of events for jevents to use
var eventInfoMap map[string]*PerfEventInfo

// walkFunc - function  to call for walking a list
type walkFunc func(w interface{}, n, e, d, p string) int

func init() {
	eventInfoMap = make(map[string]*PerfEventInfo)
}
