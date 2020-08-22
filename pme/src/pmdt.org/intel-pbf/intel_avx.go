// SPDX-License-Identifier: BSD-3-Clause
// Copyright(c) 2019-2020 Intel Corporation

// Package intelpbf - Intel Power Base Frequency support
package intelpbf

import tlog "pmdt.org/ttylog"

// Access the MSR for each CPU and retrive the Power Base Frequency values and
// return the values in a structure below.

// AVXInfo per CPU
type AVXInfo struct {
	MPerf          int32
	APerf          int32
	ThermStatus    int32
	PkgThermStatus int32
	Turbo1         int32
	Turbo2         int32
	Turbo3         int32
}

// List of constants
const (
/*
	CPUMaxFile string = "/sys/devices/system/cpu/cpu0/cpufreq/cpuinfo_max_freq"
	CPUMinFile string = "/sys/devices/system/cpu/cpu0/cpufreq/cpuinfo_min_freq"
	MAXFile    string = "/sys/devices/system/cpu/cpu0/cpufreq/scaling_max_freq"
	MINFile    string = "/sys/devices/system/cpu/cpu0/cpufreq/scaling_min_freq"
	FreqFile   string = "/sys/devices/system/cpu/cpu0/cpufreq/scaling_available_frequencies"
	GovFile    string = "/sys/devices/system/cpu/cpu0/cpufreq/scaling_available_governors"
	DrvFile    string = "/sys/devices/system/cpu/cpu0/cpufreq/scaling_driver"
*/
// Defined in pbf
// MsrFile    string = "/dev/cpu/0/msr"
// CPUMsrFile string = "/dev/cpu/%d/msr"

)

// Global values
/*
var (
	FreqP1  int32
	FreqP1n int32
	FreqP0  int32
	Driver  string
	Freqs   []int32
)
*/

// Read the MSR register for the given core and return the values read
func getCPUTurbo1(core int) int32 {

	val, err := ReadMsr(0, 0x28) //0x28
	if err != nil {
		tlog.ErrorPrintf("Unable to read MSR: %s\n", err)
	}
	val = ((val >> 8) & 0xff) * 100

	return int32(val)
}

// Read the MSR register for the given core and return the values read
func getCPUTurbo2(core int) int32 {

	val, err := ReadMsr(0, 0x28)
	if err != nil {
		tlog.ErrorPrintf("Unable to read MSR: %s\n", err)
	}
	val = ((val >> 8) & 0xff) * 100

	return int32(val)
}

// Read the MSR register for the given core and return the values read
func getCPUTurbo3(core int) int32 {

	val, err := ReadMsr(0, 0x28)
	if err != nil {
		tlog.ErrorPrintf("Unable to read MSR: %s\n", err)
	}
	val = ((val >> 8) & 0xff) * 100

	return int32(val)
}

// ReadMPerf Read the MSR register for the given core and return the busy core freq
func ReadMPerf(core int) int32 {

	val, err := ReadMsr(0, 0xE7)
	if err != nil {
		tlog.ErrorPrintf("Unable to read MSR: %s\n", err)
	}
	val = ((val >> 8) & 0xFF) * 100

	return int32(val)
}

// ReadAPerf Read the MSR register for the given core and return the avg core freq
func ReadAPerf(core int) int32 {

	val, err := ReadMsr(0, 0xE8)
	if err != nil {
		tlog.ErrorPrintf("Unable to read MSR: %s\n", err)
	}
	val = ((val >> 8) & 0xFF) * 100

	return int32(val)
}

// ReadThermStatus Read the MSR register for the given core and return the busy core freq
func ReadThermStatus(core int) int32 {

	val, err := ReadMsr(0, 0x19C)
	if err != nil {
		tlog.ErrorPrintf("Unable to read MSR: %s\n", err)
	}
	val = ((val >> 8) & 0xFF) * 100

	return int32(val)
}

// ReadPkgThermStatus Read the MSR register for the given core and return the busy core freq
func ReadPkgThermStatus(core int) int32 {

	val, err := ReadMsr(0, 0x1B1)
	if err != nil {
		tlog.ErrorPrintf("Unable to read MSR: %s\n", err)
	}
	val = ((val >> 8) & 0xFF) * 100

	return int32(val)
}

// AVXInfoPerCPU using the given cpu number
// Grab all of the AVX values from the given CPU
func AVXInfoPerCPU(cpu int) *AVXInfo {
	avx := &AVXInfo{}

	avx.MPerf = ReadMPerf(cpu)
	avx.APerf = ReadAPerf(cpu)
	avx.ThermStatus = ReadThermStatus(cpu)
	avx.PkgThermStatus = ReadPkgThermStatus(cpu)

	avx.Turbo1 = getCPUTurbo1(cpu)
	avx.Turbo2 = getCPUTurbo2(cpu)
	avx.Turbo3 = getCPUTurbo3(cpu)

	/*
		avx.MaxFreq = ReadMaxFrequency(cpu)
		avx.MinFreq = ReadMinFrequency(cpu)
		avx.CurFreq = ReadCurFrequency(cpu)
		avx.Governor = ReadGovernor(cpu)
	*/
	/*
		SetCell(pg.avxStats, 0, 0, cz.Orange("CPU", 4))
		SetCell(pg.avxStats, 0, 1, cz.Orange("128BLight", 6))
		SetCell(pg.avxStats, 0, 2, cz.Orange("128BHeavy", 6))
		SetCell(pg.avxStats, 0, 3, cz.Orange("256BLight", 6))
		SetCell(pg.avxStats, 0, 4, cz.Orange("256BHeavy", 6))
		SetCell(pg.avxStats, 0, 3, cz.Orange("512BLight", 6))
		SetCell(pg.avxStats, 0, 4, cz.Orange("512BHeavy", 6))
	*/

	//avx.CStateNames = CStates()
	/*
		for i := range CStates() {
			file := fmt.Sprintf("/sys/devices/system/cpu/cpu%d/cpuidle/state%d/disable", cpu, i)

			val := ReadInt32(cpu, file)
			if val == 1 {
				pbf.CStates = append(pbf.CStates, true)
			} else {
				pbf.CStates = append(pbf.CStates, false)
			}
		}
	*/
	return avx
}
