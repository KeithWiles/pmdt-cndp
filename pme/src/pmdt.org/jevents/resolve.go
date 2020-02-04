// SPDX-License-Identifier: BSD-3-Clause
// Copyright(c) 2019-2020 Intel Corporation

package jevents

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"golang.org/x/sys/unix"
	"pmdt.org/perf"
	tlog "pmdt.org/ttylog"
)

// Resolve the event into a raw event from the JSON files.

func init() {
	tlog.Register("ResolveLogID")
}

// rPrintf - send message to the ttylog interface
func rPrintf(format string, a ...interface{}) {
	tlog.Log("ResolveLogID", "jevents.resolve."+format, a...)
}

// read the file into memory for parsing
func readFile(f string, a ...interface{}) string {

	fn := fmt.Sprintf(f, a...)

	dat, _ := ioutil.ReadFile(fn)

	return strings.TrimSpace(string(dat))
}

// Read the qualifiers into memory and set attribute values
func readQual(qual string, attr *perf.Attr, str string) int {

	for by := range qual {
		switch by {
		case 'p':
			attr.Options.PreciseIP++
		case 'k':
			attr.Options.ExcludeUser = true
		case 'u':
			attr.Options.ExcludeKernel = true
		case 'h':
			attr.Options.ExcludeGuest = true
		default:
			fmt.Printf("Unknown modifier %c at end for %s\n", by, str)
		}
	}
	return 0
}

// Handle the special attribute setting from the string.
func specialAttr(name string, val uint64, attr *perf.Attr) bool {

	switch name {
	case "period":
		attr.SetSamplePeriod(val)
		return true
	case "freq":
		attr.SetSampleFreq(val)
		attr.Options.Freq = true
		return true
	case "config":
		attr.Config = val
		return true
	case "config1":
		attr.Config1 = val
		return true
	case "config2":
		attr.Config2 = val
		return true
	case "name":
		return true
	}
	return false
}

// determine the number of bits in a value
func bits(n uint) uint64 {
	if n == 64 {
		return 0xFFFFFFFFFFFFFFFF
	}
	return (1 << uint(n)) - 1
}

// Try to parse the event string into a uint64 value
func tryParse(format, f string, val uint64, config *uint64) bool {
	var start, end uint

	n, _ := fmt.Sscanf(format, f, &start, &end)
	if n == 0 {
		return false
	}
	if n == 1 {
		end = start + 1
	}

	rPrintf("tryParse: end %d, start %d\n", end, start)
	*config |= (val & bits(end-start+1)) << start

	rPrintf("tryParse: *config %016x\n", *config)
	return true
}

// Parse the terms used to create the perf Attr structure
func parseTerms(pmu, config string, attr *perf.Attr, recur int) int {
	var term string

	config = strings.TrimSpace(config)

	rPrintf("parseTerms: config %s recur %d\n", config, recur)
	cfg := strings.Split(config, ",")
	for _, term = range cfg {
		var name string
		var val uint64
		var err error

		val = 1

		rPrintf("parseTerms: term %v\n", term)

		// Split the event string into a set of key/value pairs from the config
		toks := strings.Split(term, "=")
		if len(toks) < 1 {
			break
		}
		rPrintf("parseTerms: toks %v\n", toks)
		name = toks[0]

		// we found more then one toks after split convert value to a variable from string
		if len(toks) > 1 {
			val, err = strconv.ParseUint(toks[1], 0, 64)
			if err != nil {
				rPrintf("*** parseTerms: %s\n", err)
				return -1
			}
		}

		// Check if this is a special attribute string
		if specialAttr(name, val, attr) {
			rPrintf("*** parseTerms: specialAttr returned true\n")
			continue
		}

		format := readFile("/sys/devices/%s/format/%s", pmu, name)
		if len(format) == 0 {
			if recur == 0 {
				alias := readFile("/sys/devices/%s/events/%s", pmu, name)
				if len(alias) == 0 {
					continue
				}
				rPrintf("ParseTerms: alias %v\n", alias)

				rPrintf("ParseTerms: alias is empty\n")

				// Parse the Alias terms into an event value
				if parseTerms(pmu, alias, attr, 1) < 0 {
					rPrintf("*** parseTerms: Cannot parse kernel event alias %s for %s\n", name, term)
					break
				}
				continue
			}
			rPrintf("Cannot parse qualifier %s for %s\n", name, term)
			break
		}
		rPrintf("ParseTerms: format %v\n", format)

		ok := tryParse(format, "config:%d-%d", val, &attr.Config) ||
			tryParse(format, "config:%d", val, &attr.Config) ||
			tryParse(format, "config1:%d-%d", val, &attr.Config1) ||
			tryParse(format, "config1:%d", val, &attr.Config)

		ok2 := tryParse(format, "config2:%d=%d", val, &attr.Config2) ||
			tryParse(format, "config2:%d", val, &attr.Config2)

		rPrintf("ParseTerms: ok %v, ok2 %v, config %v, config1 %v, config2 %v\n",
			ok, ok2, attr.Config, attr.Config1, attr.Config2)
		if ok == false && ok2 == false {
			fmt.Printf("*** parseTerms: Cannot parse kernel format %s: %s for %s, ok %v, ok2 %v\n", name, format, term, ok, ok2)
			break
		}
	}

	return 0
}

// Try a PMU type and return the string type and pmu string if found
func tryPmuType(f string, pmu string) (typ string, npmu string) {

	rPrintf("tryPmuType: format: %s pmu %s\n", f, pmu)
	newpmu := fmt.Sprintf(f, pmu)

	rPrintf("tryPmuType: newpmu %s\n", newpmu)
	dat := strings.TrimSpace(readFile("/sys/devices/%s/type", newpmu))
	if len(dat) > 0 {
		pmu = newpmu
	}

	return dat, pmu
}

// Try all PMU types if we have more then one
func tryPmuTypeAll(lst []string, pmu string) (typ string, npmu string) {

	for _, f := range lst {
		d, p := tryPmuType(f, pmu)
		if len(d) > 0 {
			return d, p
		}
	}

	return "", ""
}

// jeventPmuUncore - read the PMU uncore cpumask value
func jeventPmuUncore(str string) bool {
	var pmu string
	var cpus int

	if strings.Contains(str, "/") {
		return false
	}
	n, _ := fmt.Sscanf(str, "%30s", pmu)
	if n < 1 {
		return false
	}
	cpumask := readFile("/sys/devices/%s/cpumask", pmu)
	if len(cpumask) == 0 {
		return false
	}
	n, _ = fmt.Sscanf(cpumask, "%d", cpus)

	return (n == 1 && cpus == 0)
}

// eventNameToAttr - convert a name to perf attribute
// Given the event Name string convert it to a perf attribute structure
func eventNameToAttr(str string, attr *perf.Attr) error {
	var pmu, config, s string

	attr.Type = unix.PERF_TYPE_RAW

	rPrintf("eventNameToAttr: str %v\n", str)
	n, _ := fmt.Sscanf(str, "r%x", &attr.Config)
	if n == 1 {
		k := strings.Index(str, ":")
		if k >= 0 {
			k++
			readQual(str[k:], attr, str)
		}
		return nil
	}

	if s = str; str[0] == '/' {
		s = str[1:]
	}
	toks := strings.Split(s, "/")
	if len(toks) < 2 {
		return fmt.Errorf("wrong number of tokens in %s", s)
	}
	rPrintf("eventNameToAttr: toks %v\n", toks)
	pmu = toks[0]
	config = toks[1]

	t, p := tryPmuTypeAll([]string{"%s", "uncore_%s", "uncore_%s_0", "uncore_%s_1"}, pmu)
	rPrintf("JEvetnNameToAttr: t %v, p %v\n", t, p)

	v, err := strconv.ParseUint(t, 0, 32)
	if err != nil {
		rPrintf("eventNameToAttr: error %s\n", err)
	}
	attr.Type = perf.EventType(v)

	rPrintf("eventNameToAttr: pmu %s, Type %v\n", pmu, attr.Type)
	rPrintf("eventNameToAttr: config %v\n", config)

	parseTerms(p, config, attr, 0)

	k := strings.Index(str, ":")
	if k >= 0 {
		readQual(str[k:], attr, str)
	}

	return nil
}

// WalkPerfEvents - walk the perf events and add to collection of events
func WalkPerfEvents(f walkFunc, w interface{}) bool {

	lst, err := filepath.Glob("/sys/devices/*/events/*")
	if err != nil {
		rPrintf("WalkPerfEvent: %s\n", err)
		return false
	}

	rPrintf("lst %+v\n", lst)
	for _, path := range lst {
		var pmu, event string

		toks := strings.Split(path[1:], "/")
		pmu = toks[2]
		event = toks[4]

		if strings.Contains(event, ".") {
			continue
		}

		dat := readFile(path)
		if len(dat) <= 0 {
			continue
		}

		val2 := fmt.Sprintf("%s/%s/", pmu, dat)

		buf := fmt.Sprintf("%s/%s/", pmu, event)

		rPrintf("WalkPerfEvent: buf %s, val2 %s\n", buf, val2)
		f(w, buf, val2, "", pmu)
	}

	return true
}

// resolvePMU - resolve the PMU state
func resolvePMU(typ int) string {

	lst, err := filepath.Glob("/sys/devices/*/type")
	if err != nil {
		rPrintf("resolvePMU: %s\n", err)
		return ""
	}

	for _, path := range lst {
		var pmu string

		toks := strings.Split(path[1:], "/")
		pmu = toks[2]

		file, err := os.Open(path)
		if err != nil {
			rPrintf("WalkPerfEvent: ReadFile(%s)\n", err)
			return ""
		}
		defer file.Close()

		s := bufio.NewScanner(file)
		for s.Scan() {
			var num int

			val := s.Text()
			val = strings.TrimSpace(val)

			fmt.Sscanf(val, "%d", &num)
			if num == typ {
				return pmu
			}
		}
	}
	return ""
}
