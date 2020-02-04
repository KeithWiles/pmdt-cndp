// SPDX-License-Identifier: BSD-3-Clause
// Copyright(c) 2019-2020 Intel Corporation

package jevents

import (
	"fmt"
	"strings"

	"pmdt.org/perf"
	tlog "pmdt.org/ttylog"
)

// JEvents is a set of routine similar to and somewhat derived from
// Andi Kleen PMU-Tools https://github.com/andikleen/pmu-tools
//
// Most likely not the best Go code as I was learning Go and just translated
// the code from C to Go.

// List of events that need to be fixed and translated.
type fixedEvents struct {
	name  string
	event string
}

var fixedEventList = []fixedEvents{
	{"inst_retired.any", "event=0xc0"},
	{"cpu_clk_unhalted.thread", "event=0x3c"},
	{"cpu_clk_unhalted.thread_any", "event=0x3c,any=1"},
}

func init() {
	tlog.Register("CacheLogID")
}

// cPrintf - send message to the ttylog interface
func cPrintf(format string, a ...interface{}) {
	tlog.Log("CacheLogID", "jevents.cache."+format, a...)
}

// CollectEvents - collect the events
// w - interface value to collect the events from
func CollectEvents(w interface{}, name, event, desc, pmu string) int {

	s := strings.ToLower(name)
	cPrintf("CollectEvents: data: %v, name %s, event: %s, desc: %s, pmu: %s\n", w, s, event, desc, pmu)
	eventInfoMap[s] = &PerfEventInfo{Desc: desc, Event: event, Pmu: pmu}

	return 0
}

// ReadEvents - read the standard event files
func ReadEvents(fn string) int {

	if len(eventInfoMap) > 0 {
		eventInfoMap = make(map[string]*PerfEventInfo)
	}

	return readPlatformEvents(fn, CollectEvents, nil)
}

// realEventName - find the real event name
func realEventName(name string, event string) string {

	cPrintf("realEventName: name %s, event %s\n", name, event)
	en := event

	// loop over the list of events that need to be fixed
	for _, s := range fixedEventList {
		if s.name == name {
			en = s.event
			cPrintf("realEventName: found  -> %s\n", en)
			break
		}
	}

	cPrintf("realEventName: -> %s\n", en)

	return en
}

// ResolveEvent - resolve the event name
func ResolveEvent(name string, attr *perf.Attr) error {

	cPrintf("ResolveEvent: name %+v\n", name)

	// No event found just get the events again
	if len(eventInfoMap) == 0 {
		if ReadEvents("") < 0 {
			return fmt.Errorf("read events failed")
		}
	}

	// Convert the resolved events into event attribute structures
	if ei, ok := eventInfoMap[name]; ok {
		cPrintf("ResolveEvent: found %+v\n", ei)
		event := realEventName(name, ei.Event)

		cPrintf("ResolveEvent: event %s\n", event)
		return eventNameToAttr(fmt.Sprintf("%s/%s/", ei.Pmu, event), attr)
	}
	cPrintf("ResolveEvent: not found in list\n")

	// Convert the event name to an attribute structure
	if err := eventNameToAttr(name, attr); err == nil {
		return err
	}
	cPrintf("ResolveEvent: Did Not Found perf style event %s\n", name)

	event := fmt.Sprintf("cpu/%s/", name)
	cPrintf("ResolveEvent: Call eventNameToAttr(%s, attr)\n", event)

	// Do the conversion again if the first did not work
	if err := eventNameToAttr(event, attr); err == nil {
		return nil
	}

	return fmt.Errorf("(%s) not found in event list", name)
}

// WalkEvents - walk all of the events
// For a given event callback process all of the events
func WalkEvents(f func(w interface{}, n, d, e, p string) int, w interface{}) int {

	cPrintf("WalkEvents: Enter\n")
	if len(eventInfoMap) == 0 {
		cPrintf("WalkEvents: eventInfoMap is empty\n")
		if ReadEvents("") < 0 {
			return -1
		}
	}

	for k, v := range eventInfoMap {
		buf := fmt.Sprintf("%s/%s/", v.Pmu, v.Event)

		// Call the callback function for each event in the MAP
		if f(w, k, buf, v.Desc, v.Pmu) < 0 {
			cPrintf("WalkEvents: Error! %v\n", w)
			return -1
		}
	}

	cPrintf("WalkEvents: Done\n")
	return 0
}
