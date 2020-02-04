// SPDX-License-Identifier: BSD-3-Clause
// Copyright(c) 2019-2020 Intel Corporation

package jevents

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"

	tlog "pmdt.org/ttylog"
)

// This file will translate the event strings into a set of structures we
// can use for perf events calling perf_event_open() system call.

// The JSON files are converted to Go structures and placed in a list to
// help convert the event strings into perf event attributes

// EventData jevents structure
type EventData struct {
	EventCode         string
	UMask             string
	EventName         string
	BriefDescription  string
	PublicDescription string
	Counter           string
	CounterHTOff      string
	SampleAfterValue  string
	MSRIndex          string
	MSRValue          string
	TakenAlone        string
	CounterMask       string
	Invert            string
	AnyThread         string
	EdgeDetect        string
	Pebs              string
	DataLA            string
	L1HitIndication   string
	Errata            string
	Ellc              string
	Offcore           string
	ExtSel            string
	Unit              string
	Filter            string
}

// EventUncoreData - is the structure for Uncore json files
type EventUncoreData struct {
	Unit              string
	EventCode         string
	UMask             string
	EventName         string
	BriefDescription  string
	PublicDescription string
}

func init() {
	tlog.Register("JeventsLogID")
}

// jPrintf - send message to the ttylog interface
func jPrintf(format string, a ...interface{}) {
	tlog.Log("JeventsLogID", "jevents."+format, a...)
}

// Create the JSON file name given the basic json file name
// Use the environment variable PME_SDK to help create the name
// Return the file path name string
func jsonDefaultName(file string) string {

	id, _ := cpuStringType(file)

	gopath, ok := os.LookupEnv("PME_SDK")
	if ok {
		paths := strings.Split(gopath, ":")
		for _, p := range paths {
			filepath := fmt.Sprintf("%s/go/src/pmdt.org/jevents/events/%s.json", p, id)

			jPrintf("filepath: %v\n", filepath)
			_, err := os.Stat(filepath)
			if err != nil {
				continue
			}
			return filepath
		}
	}

	return ""
}

// Update an event string to the correct format of an event.
func updateEvent(str, format string, event *string, before bool) {
	if len(str) > 0 {
		jPrintf("updateEvent: str %s %s\n", format, str)
		if str == "0x00" || str == "0x0" || str == "0" || str == "na" {
			return
		}
		if before {
			*event = format + str + "," + *event
		} else {
			*event += "," + format + str
		}
	}
	return
}

// processEvents - read the board event list
// Decode the json file(s) calling the walkFunc to process each event and file.
func processEvents(dec *json.Decoder, fn walkFunc, w interface{}) {
	var msr *msrMapping
	var eventcode uint64

	cnt := 0
	precise := ""
	pmu := ""
	desc := ""
	name := ""
	msrval := ""

	for dec.More() {
		var e EventData

		// Create the decoder for the JSON file data.
		err := dec.Decode(&e)
		if err != nil {
			log.Fatalf("jevents: %v\n", err)
		}

		jPrintf("\n")
		eventcode = 0

		// if the EventCode is not zero in length convert string in an event code value.
		if len(e.EventCode) > 0 {
			jPrintf("processEvents: EventCode %v\n", e.EventCode)
			d := strings.Split(e.EventCode, ",")
			jPrintf("processEvents: EventCode %v, d %v\n", e.EventCode, d)
			n, err := strconv.ParseUint(d[0], 0, 64)
			if err != nil {
				log.Fatalf("jevents: %v\n", err)
			}
			jPrintf("processEvents: n %#x\n", n)
			eventcode |= n
		}

		// If the ExtSel value is not zero length we convert it to a event code
		// value to be ORed to the eventcode variable
		if len(e.ExtSel) > 0 {
			jPrintf("processEvents: ExtSel %v\n", e.ExtSel)
			d := strings.Split(e.ExtSel, ",")
			n, err := strconv.ParseUint(d[0], 0, 64)
			if err != nil {
				log.Fatalf("jevents: %v\n", err)
			}
			jPrintf("processEvents: n %#x\n", n)
			eventcode |= n << 21
		}

		// Trim the description to exclude the '.'
		name = e.EventName
		desc = strings.TrimRight(e.BriefDescription, ".")

		// Process precise event values if present
		if len(desc) > 0 && strings.Contains(desc, "(Percise Event)") {
			precise = e.Pebs
		}
		msr = lookupMSR(e.MSRIndex)
		msrval = e.MSRValue

		if len(e.Errata) > 0 && e.Errata != "null" {
			desc = desc + ". Spec Update: " + e.Errata
		}

		if e.DataLA != "0" && e.DataLA != "0x00" {
			desc = desc + ". Supports address when precise"
		}

		if len(e.Unit) > 0 {
			pmu = fieldToPerf(e.Unit)
		}

		if len(precise) > 0 {
			if precise == "2" {
				desc = desc + "(Must be precise)"
			} else {
				desc = desc + "(Precise event)"
			}
		}

		// Done processing the event data values into values we can use for Events

		event := "event=0"
		if eventcode != 0 {
			event = fmt.Sprintf("event=%#x", eventcode)
		}
		jPrintf("processEvents: EventCode event(%s)\n", event)

		if len(e.Filter) > 0 && e.Filter != "na" {
			event = event + "," + e.Filter
			jPrintf("processEvents: Add Filter event(%s)\n", event)
		}

		if msr != nil {
			event = event + "," + msr.pname
			if len(msrval) > 0 {
				event = event + msrval
				jPrintf("processEvents: Add MSR event(%s)\n", event)
			}
		}

		if len(pmu) == 0 {
			pmu = "cpu"
		}

		updateEvent(e.AnyThread, "any=", &event, true)
		updateEvent(e.EdgeDetect, "edge=", &event, true)
		updateEvent(e.Invert, "inv=", &event, true)
		updateEvent(e.CounterMask, "cmask=", &event, true)
		updateEvent(e.SampleAfterValue, "period=", &event, true)
		updateEvent(e.UMask, "umask=", &event, true)

		if len(name) > 0 && len(event) > 0 {
			jPrintf("processEvents: name(%s), desc(%s), event(%s), pmu(%s)\n",
				name, desc, event, pmu)

			// if name and event are not zero length then we have a valid event
			// and it needs to be saved via the 'fn' function callback
			fn(w, name, event, desc, pmu)

			// Update the event code
			AddNewHardwareLabel(name, "", eventcode)
		}

		cnt++
	}
}

// jsonEventRead - read the board event list
func jsonEventRead(file string, suffix string, fn walkFunc, w interface{}) int {
	var f string

	if file == "" {
		f = jsonDefaultName(suffix)
		if len(f) == 0 {
			jPrintf("Unable create default json filename\n")
			return -1
		}
	}

	// Read all of the file into memory to be decoded
	dat, err := ioutil.ReadFile(f)
	if err != nil {
		panic(err)
	}

	jPrintf("filename: %s, %d\n", f, len(dat))

	dec := json.NewDecoder(strings.NewReader(string(dat)))
	_, err = dec.Token()
	if err != nil {
		log.Fatalf("jevents: %v\n", err)
	}

	// Now process all of the decoded event data.
	processEvents(dec, fn, w)

	_, err = dec.Token()
	if err != nil {
		log.Fatalf("jevents: %v\n", err)
	}

	return 0
}

// jsonPlatformEvents - read the platform event list
func readPlatformEvents(file string, f walkFunc, w interface{}) int {

	if jsonEventRead(file, "-core", f, w) < 0 {
		jPrintf("jsonEventRead: failed (core)\n")
		return -1
	}
	if jsonEventRead(file, "-uncore", f, w) < 0 {
		jPrintf("jsonEventRead: failed (uncore)\n")
		return -1
	}

	jPrintf("eventInfoMap: length %d\n", len(eventInfoMap))
	return 0
}

type mapPerf struct {
	json string
	perf string
}

var unitToPmu = []mapPerf{
	{"CBO", "cbox"},
	{"QPI LL", "qpi"},
	{"SBO", "sbox"},
	{"IMPH-U", "cbox"},
	{"NCU", "cbox"},
}

// fieldToPerf - convert feild to a perf name
func fieldToPerf(mapName string) string {

	if len(mapName) == 0 {
		return ""
	}
	for _, s := range unitToPmu {
		if s.json == mapName {
			return strings.ToLower(s.perf)
		}
	}
	return strings.ToLower(mapName)
}

// msrMapping - list of mappings to MSR
type msrMapping struct {
	num   string
	pname string
}

var msrmaps = []msrMapping{
	{"0x3f6", "ldlat="},
	{"0x1a6", "offcore_rsp="},
	{"0x1a7", "offcore_rsp="},
	{"0x3f7", "frontend="},
}

// lookupMSR - lookup the MSR value
func lookupMSR(val string) *msrMapping {

	vals := strings.Split(val, ",")
	for _, m := range msrmaps {
		s := strings.ToLower(vals[0])
		if m.num == s {
			return &m
		}
	}
	return nil
}
