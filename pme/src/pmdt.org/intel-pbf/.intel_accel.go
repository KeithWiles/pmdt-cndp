// SPDX-License-Identifier: BSD-3-Clause
// Copyright(c) 2019-2020 Intel Corporation

// Intel Acceleration - useful information for accelerating support

// Package intelpbf - Intel Power Base Frequency support
package intelpbf

import (
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"

	tlog "pmdt.org/ttylog"
)

// Access the MSR for each CPU and retrive the Power Base Frequency values and
// return the values in a structure below.

// AccelInfo per CPU
type AccelInfo struct {
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

	UpFile   string = "/proc/uptime"
	StatFile string = "/proc/stat"

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
func getCPUBusyMHz(core int) int32 {

	val, err := ReadMsr(0, 0xE7)
	if err != nil {
		tlog.ErrorPrintf("Unable to read MSR: %s\n", err)
	}
	val = ((val >> 8) & 0xFF) * 100

	return int32(val)
}

// Read the MSR register for the given core and return the values read
func getCPUAvgMHz(core int) int32 {

	val, err := ReadMsr(0, 0xE8)
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

// ReadBusyness and normalize the values
// Read all of the busyness values
func ReadBusyness() []int32 {
	var busys []int32

	str := ReadString(FreqFile)

	// Split up the string of frequencies and convert to an array of int32 values
	for _, f := range strings.Split(str, " ") {
		val, _ := strconv.ParseInt(f, 0, 32)
		busys = append(busys, int32(val))
	}

	return busys
}

// AccelInfoPerCPU using the given cpu number
// Grab all of the acceleration values from the given CPU
func AccelInfoPerCPU(cpu int) *Info {
	accel := &Info{}

	accel.MaxFreq = ReadMaxFrequency(cpu)
	accel.MinFreq = ReadMinFrequency(cpu)
	accel.CurFreq = ReadCurFrequency(cpu)
	accel.Governor = ReadGovernor(cpu)
	accel.CStateNames = CStates()

	for i := range CStates() {
		file := fmt.Sprintf("/sys/devices/system/cpu/cpu%d/cpuidle/state%d/disable", cpu, i)

		val := ReadInt32(cpu, file)
		if val == 1 {
			accel.CStates = append(accel.CStates, true)
		} else {
			accel.CStates = append(accel.CStates, false)
		}
	}

	return accel
}
