// SPDX-License-Identifier: BSD-3-Clause
// Copyright(c) 2019-2020 Intel Corporation

// Package intelpbf - Intel Power Base Frequency support
package intelpbf

import (
	"fmt"
	"io/ioutil"
	tlog "pmdt.org/ttylog"
	"sort"
	"strconv"
	"strings"
)

// Access the MSR for each CPU and retrive the Power Base Frequency values and
// return the values in a structure below.

// Info per CPU
type Info struct {
	MaxFreq     int32
	MinFreq     int32
	CurFreq     int32
	Governor    string
	CStates     []bool
	CStateNames []string
}

// List of constants
const (
	CPUMaxFile string = "/sys/devices/system/cpu/cpu0/cpufreq/cpuinfo_max_freq"
	CPUMinFile string = "/sys/devices/system/cpu/cpu0/cpufreq/cpuinfo_min_freq"
	MAXFile    string = "/sys/devices/system/cpu/cpu0/cpufreq/scaling_max_freq"
	MINFile    string = "/sys/devices/system/cpu/cpu0/cpufreq/scaling_min_freq"
	FreqFile   string = "/sys/devices/system/cpu/cpu0/cpufreq/scaling_available_frequencies"
	GovFile    string = "/sys/devices/system/cpu/cpu0/cpufreq/scaling_available_governors"
	DrvFile    string = "/sys/devices/system/cpu/cpu0/cpufreq/scaling_driver"
	MsrFile    string = "/dev/cpu/0/msr"
	CPUMsrFile string = "/dev/cpu/%d/msr"
)

// Global values
var (
	FreqP1  int32
	FreqP1n int32
	FreqP0  int32
	Driver  string
	Freqs   []int32
)

// Read the MSR register for the given core and return the values read
func getCPUBaseFrequency(core int) int32 {

	val, err := ReadMsr(0, 0xCE)
	if err != nil {
		tlog.ErrorPrintf("Unable to read MSR: %s\n", err)
	}
	val = ((val >> 8) & 0xFF) * 100

	return int32(val)
}

// CheckForDrivers are installed
func CheckForDrivers() bool {

	Driver := ReadString(DrvFile)

	if Driver == "acpi-cpufreq" {
		return true
	} else if Driver == "intel_pstate" {
		if !fileExists(MsrFile) {
			tlog.ErrorPrintf("Unable to read (%s) does not exist\n", MsrFile)
			return false
		}

		FreqP1 = getCPUBaseFrequency(11)
	}
	return true
}

// Read the different pstates from the CPU and return in a slice
func getPStates() []int32 {

	freqs := []int32{}

	if Driver == "acpi-cpufreq" {
		freqs = ReadFrequencies()
	} else {
		FreqP1n = ReadMinFrequency(0)

		FreqP0 = ReadMaxFrequency(0)

		for i := FreqP1n; i < (FreqP0 + 1); i += 100 {
			freqs = append(freqs, int32(i))
		}

		sort.Slice(freqs, func(i, j int) bool {
			return freqs[j] < freqs[i]
		})
	}

	return freqs
}

// PStates of the system, return an array of int32 values.
func PStates() []int32 {

	Freqs = getPStates()

	return Freqs
}

// Get the CStates from the /sys/devices/ files for each CPU in the system
func getCStates() []string {

	stateList := []string{}

	states, err := ioutil.ReadDir("/sys/devices/system/cpu/cpu0/cpuidle")
	if err != nil {
		tlog.ErrorPrintf("Unable to read states %s", err)
		return nil
	}

	for _, state := range states {
		stateFile := fmt.Sprintf("/sys/devices/system/cpu/cpu0/cpuidle/%s/name", state.Name())

		stateList = append(stateList, ReadString(stateFile))
	}

	return stateList
}

// CStates for the system
func CStates() []string {
	return getCStates()
}

// Governors values, return a slice of strings for each Governor found
func Governors() []string {
	govs := ReadString(GovFile)

	return strings.Split(govs, " ")
}

// ReadString from file
// A helper routine to read a string and trim it of spaces
func ReadString(file string) string {

	dat, err := ioutil.ReadFile(file)
	if err != nil {
		tlog.ErrorPrintf("Unable to read (%s) %s\n", file, err)
		return "*Error*"
	}
	return strings.TrimSpace(string(dat))
}

// ReadInt32 value
// Read a string from a file and convert to a int32 value
func ReadInt32(cpu int, file string) int32 {

	str := ReadString(file)

	val, _ := strconv.ParseInt(str, 0, 32)

	return int32(val)
}

// ReadFrequency and normalize the value
// Read the frequency
func ReadFrequency(cpu int, freq string) int32 {
	file := fmt.Sprintf("/sys/devices/system/cpu/cpu%d/cpufreq/%s", cpu, freq)

	return (ReadInt32(cpu, file) / 1000)
}

// ReadGovernor for give cpu as a string
func ReadGovernor(cpu int) string {
	file := fmt.Sprintf("/sys/devices/system/cpu/cpu%d/cpufreq/scaling_governor", cpu)

	return ReadString(file)
}

// ReadMaxFrequency and normalize the value
// Read the max frequency value for the given CPU
func ReadMaxFrequency(cpu int) int32 {
	return ReadFrequency(cpu, "scaling_max_freq")
}

// ReadMinFrequency and normalize the value
// Read the min frequency value for the given CPU
func ReadMinFrequency(cpu int) int32 {
	return ReadFrequency(cpu, "scaling_min_freq")
}

// ReadCurFrequency and normalize the value
// Read the current frequency value for the given CPU
func ReadCurFrequency(cpu int) int32 {
	return ReadFrequency(cpu, "scaling_cur_freq")
}

// ReadFrequencies and normalize the values
// Read all of the frequency values
func ReadFrequencies() []int32 {
	var freqs []int32

	str := ReadString(FreqFile)

	// Split up the string of frequencies and convert to an array of int32 values
	for _, f := range strings.Split(str, " ") {
		val, _ := strconv.ParseInt(f, 0, 32)
		freqs = append(freqs, int32(val))
	}

	return freqs
}

// InfoPerCPU using the given cpu number
// Grab all of the PBF values from the given CPU
func InfoPerCPU(cpu int) *Info {
	pbf := &Info{}

	pbf.MaxFreq = ReadMaxFrequency(cpu)
	pbf.MinFreq = ReadMinFrequency(cpu)
	pbf.CurFreq = ReadCurFrequency(cpu)
	pbf.Governor = ReadGovernor(cpu)
	pbf.CStateNames = CStates()

	for i := range CStates() {
		file := fmt.Sprintf("/sys/devices/system/cpu/cpu%d/cpuidle/state%d/disable", cpu, i)

		val := ReadInt32(cpu, file)
		if val == 1 {
			pbf.CStates = append(pbf.CStates, true)
		} else {
			pbf.CStates = append(pbf.CStates, false)
		}
	}

	return pbf
}
