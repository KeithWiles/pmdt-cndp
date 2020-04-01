// SPDX-License-Identifier: BSD-3-Clause
// Copyright(c) 2019-2020 Intel Corporation

package pinfo

import (
	"encoding/json"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"net"
	"sync"

	tlog "pmdt.org/ttylog"
)

// ConnInfo - Information about the app
type ConnInfo struct {
	valid bool // true if the process info data is valid
	conn  *net.UnixConn
	Pid   int64  // Pid for the process
	Path  string // Path of the process_pinfo.<pid> file
}

// ConnInfoMap holds all of the process info data
type ConnInfoMap map[int64]*ConnInfo

// CallbackMap holds the watcher fsnotify callback information
type CallbackMap map[string]*Callback

// ProcessInfo data for applications
type ProcessInfo struct {
	lock     sync.Mutex
	opened   bool              // true if process info open
	basePath string            // Base path to the run directory
	baseName string            // Base file name
	connInfo ConnInfoMap       // Indexed by pid for each application
	callback CallbackMap       // Callback routines for the fsnotify
	watcher  *fsnotify.Watcher // watcher for the directory notify
}

// Define the buffer size to be used for incoming data
const (
	maxBufferSize = (16 * 1024)
)

// New information structure
func New(bpath, bname string) *ProcessInfo {

	pi := &ProcessInfo{basePath: bpath, baseName: bname}

	pi.connInfo = make(ConnInfoMap)
	pi.callback = make(CallbackMap)

	pi.opened = false

	return pi
}

// doCmd information
func (pi *ProcessInfo) doCmd(a *ConnInfo, cmd string) ([]byte, error) {

	if _, err := a.conn.Write([]byte(cmd)); err != nil {
		return nil, fmt.Errorf("write on socket failed: %v", err)
	}

	buf := make([]byte, maxBufferSize) // big buffer

	n, err := a.conn.Read(buf)
	if err != nil {
		return nil, err
	}
	return buf[:n], nil
}

// ConnectionList returns the list of ConnInfo structures
func (pi *ProcessInfo) ConnectionList() []*ConnInfo {

	p := make([]*ConnInfo, 0)

	for _, a := range pi.connInfo {
		p = append(p, a)
	}
	return p
}

// Files returns a string slice of application process info data
func (pi *ProcessInfo) Files() []string {

	files := []string{}
	for _, a := range pi.connInfo {
		files = append(files, a.Path)
	}

	return files
}

// Pids returns a int64 slice of application process info data
func (pi *ProcessInfo) Pids() []int64 {

	pids := make([]int64, 0)
	for _, a := range pi.connInfo {
		pids = append(pids, a.Pid)
	}

	return pids
}

// ConnectionByPid returns the ConnInfo pointer using the Pid
func (pi *ProcessInfo) ConnectionByPid(pid int64) *ConnInfo {

	for _, a := range pi.connInfo {
		if a.Pid == pid {
			return a
		}
	}
	return nil
}

// Unmarshal the JSON data into a structure
func (pi *ProcessInfo) Unmarshal(p *ConnInfo, command string, data interface{}) error {

	if len(pi.connInfo) == 0 {
		return nil
	}
	if p == nil {
		// Get the first element of a map
		for _, m := range pi.connInfo {
			p = m
			break
		}
	}
	d, err := pi.doCmd(p, command)
	if err != nil {
		return err
	}
	tlog.DebugPrintf("Data: %v\n", string(d))

	if err := json.Unmarshal(d, data); err != nil {
		return err
	}

	return nil
}

// Marshal the structure into a JSON string
func (pi *ProcessInfo) Marshal(data interface{}) ([]byte, error) {

	return json.MarshalIndent(data, "", "  ")
}
