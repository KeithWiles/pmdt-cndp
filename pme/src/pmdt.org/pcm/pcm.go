// SPDX-License-Identifier: BSD-3-Clause
// Copyright(c) 2019-2020 Intel Corporation

package pcm

// CPUModel string name
func CPUModel(id int) string {
	v, ok := CPUModels[id]
	if !ok {
		return ""
	}
	return v
}
