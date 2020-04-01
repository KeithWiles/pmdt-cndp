// SPDX-License-Identifier: BSD-3-Clause
// Copyright(c) 2019-2020 Intel Corporation

package pinfo

import (
	"testing"
)

var pi *ProcessInfo

func TestOpen(t *testing.T) {

	pi = NewProcessInfo("/var/run/pcm", "pinfo.")

	if err := pi.Open(); err != nil {
		t.Errorf("Open() failed: %v", err)
	}

	t.Log("Process  Open: OK")
}

func TestInfo(t *testing.T) {

	for _, a := range pi.AppsList() {
		cmds, err := pi.Commands(a)
		if err != nil {
			t.Logf("Info for %s\n", a.Path)
			continue
		}

		t.Logf("Info    : %+v\n", cmds)
	}
}

func TestCommands(t *testing.T) {

	for _, a := range pi.AppsList() {
		cmds, err := pi.Commands(a)
		if err != nil {
			t.Errorf("unable to retrive commands: %v", err)
		} else {
			t.Logf("Commands     : %v\n", cmds)
		}
	}
}

func TestFiles(t *testing.T) {

	files := pi.Files()
	if len(files) > 0 {
		t.Logf("Files        : %v\n", files)
	}
}

func TestClose(t *testing.T) {

	if pi == nil {
		t.Errorf("ProcessInfo pointer is nil")
	} else {
		pi.Close()
	}
	t.Log("Process Close: OK")
}
