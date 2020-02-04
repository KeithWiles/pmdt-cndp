// SPDX-License-Identifier: BSD-3-Clause
// Copyright(c) 2019-2020 Intel Corporation

package jevents

import (
	"fmt"
	"io/ioutil"
	"runtime"
	"strings"

	"pmdt.org/perf"
	tlog "pmdt.org/ttylog"
)

// Create a jevents session and verify the core is enabled, then collect the
// event data.

func init() {
	tlog.Register("SessionLogID")
}

// sPrintf - send message to the ttylog interface
func sPrintf(format string, a ...interface{}) {
	tlog.Log("SessionLogID", "jevents.session."+format, a...)
}

// newPerfEvent - create a new PerfEvent structure with default values
// s - is the name of the event to label the attribute
func newPerfEvent(s string) *PerfEvent {

	e := new(PerfEvent)

	e.EventName = s
	e.Attr = new(perf.Attr)
	e.Attr.Label = s

	e.CPUEvents = make([]*CPUEvent, numCPUs())

	return e
}

func cpuOnLine(lid int) bool {

	file := fmt.Sprintf("/sys/devices/system/cpu/cpu%d/online", lid)

	dat, err := ioutil.ReadFile(file)
	if err != nil {
		return false
	}

	if string(dat) == "1" {
		return true
	}
	return true
}

func jeventPMUUncore(s string) bool {
	var pmu string
	var cpus int

	if !strings.Contains(s, "/") {
		return false
	}
	_, err := fmt.Sscanf(s, "%30[^/]", pmu)
	if err != nil {
		return false
	}

	dat, err := ioutil.ReadFile(fmt.Sprintf("/sys/devices/%s/cpumask", pmu))
	if err != nil {
		return false
	}

	n, err := fmt.Sscanf(string(dat), "%d", &cpus)

	if n == 1 && cpus == 0 {
		return true
	}

	return false
}

// Parse - parse the events
// events - a list of comma seperated event names
func Parse(events string) ([]*PerfEvent, error) {

	perfEvents := []*PerfEvent{}

	elist := strings.Split(events, ";")
	for _, s := range elist {
		sPrintf("\n")

		grpLeader := false
		endGrp := false

		if strings.Contains(s, "{") {
			s = s[1:]
			grpLeader = true
		} else if strings.Contains(s, "}") {
			endGrp = true
			s = strings.TrimSuffix(s, "}")
		}

		pe := newPerfEvent(s)

		pe.Uncore = jeventPMUUncore(s)
		pe.GroupLeader = grpLeader
		pe.EndGroup = endGrp

		perfEvents = append(perfEvents, pe)

		ResolveEvent(s, pe.Attr)
	}

	return perfEvents, nil
}

// openEvent
func openEvent(e *PerfEvent, cpu int, leader *PerfEvent, measureAll bool, measurePid int) error {

	a := e.Attr

	a.Options.Inherit = true
	if !measureAll {
		a.Options.Disabled = true
		a.Options.EnableOnExec = true
	}

	a.CountFormat.Enabled = true
	a.CountFormat.Running = true
	a.Label = e.EventName

	pid := measurePid
	if measureAll {
		pid = -1
	}

	var ldr *perf.Event

	if ldr = nil; leader != nil {
		ldr = leader.CPUEvents[cpu].Event
	}
	ev, err := perf.Open(a, pid, cpu, ldr)
	if err != nil {
		return fmt.Errorf("perf.Open failed %v", err)
	}
	if e.CPUEvents[cpu] == nil {
		e.CPUEvents[cpu] = &CPUEvent{enabled: true, cpuID: cpu}
	}
	e.CPUEvents[cpu].Event = ev

	return nil
}

// Open - setup all events
func Open(perfEvents []*PerfEvent, all bool, pid int) error {
	var leader *PerfEvent

	if perfEvents == nil {
		return fmt.Errorf("eventlist is nil")
	}

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	leader = nil
	for _, e := range perfEvents {

		if e.Uncore {
			if err := openEvent(e, 0, leader, all, pid); err != nil {
				return err
			}
			for cpu := 1; cpu < numCPUs(); cpu++ {
				if err := openEvent(e, cpu, leader, all, pid); err != nil {
					return err
				}
			}
		} else {
			for cpu := 0; cpu < numCPUs(); cpu++ {
				if err := openEvent(e, cpu, leader, all, pid); err != nil {
					return err
				}
			}
		}
		if e.GroupLeader {
			leader = e
			sPrintf("OpenEvents: GroupLeader: %s\n", e.EventName)
		}
		if e.EndGroup {
			sPrintf("OpenEvents: EndGroup: %s\n", e.EventName)
			leader = nil
		}
	}

	return nil
}

// Start events is the primary entry point for handling perf events
func Start(perfEvents []*PerfEvent) error {

	if perfEvents == nil {
		return fmt.Errorf("event list is empty")
	}
	for _, pe := range perfEvents {
		for _, ce := range pe.CPUEvents {
			if err := ce.Event.MeasureStart(); err != nil {
				return err
			}
		}
	}
	return nil
}

// Stop event collecting and count the values in the Count structure
func Stop(perfEvents []*PerfEvent) error {

	if perfEvents == nil {
		return fmt.Errorf("event list is empty")
	}

	for _, pe := range perfEvents {
		totalSum := uint64(0)
		if pe.CPUEvents == nil {
			continue
		}
		for _, ce := range pe.CPUEvents {
			if ce == nil {
				break
			}
			cnt, err := ce.Event.MeasureEnd()
			if err != nil {
				return err
			}
			ce.Count = cnt
			ce.Scaled = eventScaledValue(&cnt)
			totalSum += ce.Scaled
		}
		pe.TotalSum = totalSum
	}

	return nil
}

// Close by doing a clone on the event
func Close(perfEvents []*PerfEvent) error {

	if perfEvents == nil {
		return fmt.Errorf("event list is empty")
	}
	for _, pe := range perfEvents {
		for _, ce := range pe.CPUEvents {
			ce.Event.Close()
		}
	}
	return nil
}

// eventScaledValue - Scale the current count values
func eventScaledValue(c *perf.Count) uint64 {

	v := c.Value
	if c.Enabled != c.Running && c.Running > 0 {
		return (v * uint64(c.Enabled)) / uint64(c.Running)
	}
	return v
}
