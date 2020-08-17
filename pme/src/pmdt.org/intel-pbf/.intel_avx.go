// SPDX-License-Identifier: BSD-3-Clause
// Copyright(c) 2019-2020 Intel Corporation

// Package intelpbf - Intel Power Base Frequency support
package intelpbf

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
