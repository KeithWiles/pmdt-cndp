// SPDX-License-Identifier: BSD-3-Clause
// Copyright(c) 2019-2020 Intel Corporation

package main

import (
	"fmt"
	"os"
	"sync"

	"github.com/shirou/gopsutil/cpu"
	cz "pmdt.org/colorize"
	tlog "pmdt.org/ttylog"
)

var numCPUs int

// PerfmonInfo returning the basic information string
func PerfmonInfo(color bool) string {
	if !color {
		return fmt.Sprintf("%s, Version: %s Pid: %d %s",
			"DPDK Performance Monitor", Version(), os.Getpid(),
			"Copyright © 2019-2020 Intel Corporation")
	}

	return fmt.Sprintf("[%s, Version: %s Pid: %s %s]",
		cz.Yellow("DPDK Performance Monitor"), cz.Green(Version()),
		cz.Red(os.Getpid()),
		cz.SkyBlue("Copyright © 2019-2020 Intel Corporation"))
}

// NumCPUs is the number of CPUs in the system (logical cores)
func NumCPUs() int {
	var once sync.Once

	once.Do(func() {
		num, err := cpu.Counts(true)
		if err != nil {
			tlog.FatalPrintf("Unable to get number of CPUs: %v", err)
			os.Exit(1)
		}
		numCPUs = num
	})

	return numCPUs
}

func sprintf(msg string, w ...interface{}) string {
	if len(w) > 1 {
		return fmt.Sprintf("%-36s: %6d, %6d\n", msg, w[0].(uintptr), w[1].(uintptr))
	} else if len(w) == 1 {
		return fmt.Sprintf("%-36s: %6d\n", msg, w[0].(uintptr))
	} else {
		return fmt.Sprintf("%s args is zero\n", msg)
	}
}

func dprintf(msg string, w ...interface{}) {

	tlog.DoPrintf(sprintf(msg, w...))
}

// Format the bytes into human readable format
func Format(units []string, v uint64, w ...interface{}) string {
	var index int

	bytes := float64(v)
	for index = 0; index < len(units); index++ {
		if bytes < 1024.0 {
			break
		}
		bytes = bytes / 1024.0
	}

	percision := uint64(0)
	for _, v := range w {
		percision = v.(uint64)
	}

	return fmt.Sprintf("%.*f %s", percision, bytes, units[index])
}

// FormatBytes into KB, MB, GB, ...
func FormatBytes(v uint64, w ...interface{}) string {

	return Format([]string{"B", "KB", "MB", "GB", "TB", "PB", "EB", "ZB", "YB"}, v, w...)
}

// FormatUnits into KB, MB, GB, ...
func FormatUnits(v uint64, w ...interface{}) string {

	return Format([]string{" ", "K", "M", "G", "T", "P", "E", "Z", "Y"}, v, w...)
}

// BitRate - return the network bit rate
func BitRate(ioPkts, ioBytes uint64) float64 {
	return float64(((ioPkts * PktOverheadSize) + ioBytes) * 8)
}
