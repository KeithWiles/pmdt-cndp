// SPDX-License-Identifier: BSD-3-Clause
// Copyright(c) 2019-2020 Intel Corporation

package jevents

import (
	"bufio"
	"fmt"
	"github.com/shirou/gopsutil/cpu"
	"os"
	"sync"
)

// Collect up the CPU type information into a string using the /proc/cpuinfo
// file information.

// cpuStringType - return cpu type string
func cpuStringType(t string) (string, string) {
	var vendor, idstr, res string

	model := 0
	fam := 0
	step := 0
	found := 0

	f, err := os.Open("/proc/cpuinfo")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		line := scanner.Text()
		n, _ := fmt.Sscanf(line, "vendor_id : %s", &vendor)
		if n == 1 {
			found++
		}
		n, _ = fmt.Sscanf(line, "model : %d", &model)
		if n == 1 {
			found++
		}
		n, _ = fmt.Sscanf(line, "cpu family : %d", &fam)
		if n == 1 {
			found++
		}
		n, _ = fmt.Sscanf(line, "stepping : %d", &step)
		if n == 1 {
			found++
		}
		if found == 4 {
			found = 0

			idstr = fmt.Sprintf("%s-%d-%X-%X%s", vendor, fam, model, step, t)
			res = fmt.Sprintf("%s-%d-%X%s", vendor, fam, model, t)
			break
		}
	}

	return idstr, res
}

// cpuString - return main CPU string without type
func cpuString() string {

	_, res := cpuStringType("-core")

	return res
}

var numberCPUs int

// numCPUs is the number of CPUs in the system (logical cores)
func numCPUs() int {
	var once sync.Once

	once.Do(func() {
		num, err := cpu.Counts(true)
		if err != nil {
			os.Exit(1)
		}
		numberCPUs = num
	})

	return numberCPUs
}
