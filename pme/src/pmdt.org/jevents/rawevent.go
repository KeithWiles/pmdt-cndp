// SPDX-License-Identifier: BSD-3-Clause
// Copyright(c) 2019-2020 Intel Corporation

package jevents

import (
	"fmt"
	"pmdt.org/perf"
)

// formatRawEvent - format a raw event in to a string
func formatRawEvent(attr *perf.Attr, name string) string {

	pmu := resolvePMU(int(attr.Type))

	if len(pmu) == 0 {
		return ""
	}

	str := fmt.Sprintf("%s/config=0x%x", pmu, attr.Config)

	if attr.Config1 > 0 {
		str += fmt.Sprintf(",config1=0x%x", attr.Config1)
	}
	if attr.Config2 > 0 {
		str += fmt.Sprintf(",config2=0x%x", attr.Config2)
	}
	if len(name) > 0 {
		str += fmt.Sprintf(",name=%s", name)
	}
	str += "/"

	return str
}
