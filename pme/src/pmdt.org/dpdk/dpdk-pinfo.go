// SPDX-License-Identifier: BSD-3-Clause
// Copyright(c) 2019-2020 Intel Corporation

package dpdk

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	//	"time"
	"github.com/fsnotify/fsnotify"
	"io/ioutil"
	"net"
	"path/filepath"
	"sync"

	tlog "pmdt.org/ttylog"
)

const (
	// MaxPorts is the number ports supported
	MaxPorts	int = 16
)

// RTEInfo is the data returned from the dpdk:info command
type RTEInfo struct {
	Version   string `json:"version"`
	MaxBuffer int64  `json:"maxbuffer"`
	ProcType  string `json:"proctype"`
}

// EALParams is the data structure to hold EAL Parameters
type EALParams struct {
	EALArgs []string `json:"ealargs"`
	AppArgs []string `json:"appargs"`
}

// EALCmds host the list of commands
type EALCmds struct {
	Cmds []string
}

// PortState is the information about link state
type PortState struct {
	PortID int    `json:"port"`
	Duplex string `json:"duplex"`
	State  string `json:"state"`
	Rate   int    `json:"rate"`
}

// DevList information
type DevList struct {
	Ports []PortState `json:"ports"`
	Avail int         `json:"avail"`
	Total int         `json:"total"`
}

// PortStats is the data structure to hold the counters.
type PortStats struct {
	PortID     int    `json:"portid"`
	PacketsIn  uint64 `json:"ipackets"`
	PacketsOut uint64 `json:"opackets"`
	BytesIn    uint64 `json:"ibytes"`
	BytesOut   uint64 `json:"obytes"`
	MissedIn   uint64 `json:"imissed"`
	ErrorsIn   uint64 `json:"ierrors"`
	ErrorsOut  uint64 `json:"oerrors"`
	RxNoMbuf   uint64 `json:"rx_nombuf"`

	QInPackets  []uint64 `json:"q_ipackets"`
	QOutPackets []uint64 `json:"q_opackets"`
	QInBytes    []uint64 `json:"q_ibytes"`
	QOutBytes   []uint64 `json:"q_obytes"`
	QErrors     []uint64 `json:"q_errors"`
}

// PortXStats is the data structure to hold data
type PortXStats struct {
	PortID int
	XStats map[string]uint64
}

// dpdkInfo - Information about the app
type dpdkInfo struct {
	Params EALParams // Holds the EAL parameter data
	Info   RTEInfo   // Holds the DPDK process info data
	Cmds   EALCmds   // List of all known commands
	PrevStats [MaxPorts]PortStats
	PrevXStats [MaxPorts]PortXStats
}

// stats information
func (pi *ProcessInfo) stats(a *dpdkInfo, cmd string, portID int) (*PortStats, error) {

	b, err := pi.doCmd(a, fmt.Sprintf("%s,%d", cmd, portID))
	if err != nil {
		return nil, err
	}

	data := &PortStats{}

	data.PortID = portID
	if err := json.Unmarshal(b, &data); err != nil {
		return nil, err
	}

	return data, nil
}

// xstats information
func (pi *ProcessInfo) xstats(a *dpdkInfo, cmd string, portID int) (*PortXStats, error) {

	b, err := pi.doCmd(a, fmt.Sprintf("%s,%d", cmd, portID))
	if err != nil {
		return nil, err
	}

	data := &PortXStats{}

	data.PortID = portID
	if err := json.Unmarshal(b, &data.XStats); err != nil {
		return nil, err
	}

	return data, nil
}

// Info returns the RTEInfo structure
func (pi *ProcessInfo) Info(a *dpdkInfo) *RTEInfo {

	if a == nil {
		return nil
	}
	return &a.Info
}

// EthdevList of devices as port ids
func (pi *ProcessInfo) EthdevList(a *dpdkInfo) (*DevList, error) {

	return pi.list(a, "/ethdev/list")
}

// EthdevStats information
func (pi *ProcessInfo) EthdevStats(a *dpdkInfo, portID int) (*PortStats, error) {

	pstats, err := pi.stats(a, "/ethdev/stats", portID)
	if err != nil {
		return nil, fmt.Errorf("/ethdev/stats failed: %v", err)
	}
	p := *pstats
	a.PrevStats[portID] = p

	return pstats, err
}

// PreviousStats is the save stats data
func (pi *ProcessInfo) PreviousStats(a *dpdkInfo, portID int) (PortStats, error) {

	if portID >= len(a.PrevStats) {
		return PortStats{}, fmt.Errorf("invalid port id")
	}
	return a.PrevStats[portID], nil
}

// EthdevXStats information
func (pi *ProcessInfo) EthdevXStats(a *dpdkInfo, portID int) (*PortXStats, error) {

	return pi.xstats(a, "/ethdev/xstats", portID)
}

// PreviousXStats is the save stats data
func (pi *ProcessInfo) PreviousXStats(a *dpdkInfo, portID int) (*PortXStats, error) {
	if portID >= len(a.PrevStats) {
		return nil, fmt.Errorf("invalid port id")
	}
	pstats := &a.PrevXStats[portID]

	return pstats, nil
}

// RawdevList of devices as port ids
func (pi *ProcessInfo) RawdevList(a *dpdkInfo) (*DevList, error) {

	return pi.list(a, "/rawdev/list")
}

// RawdevStats information
func (pi *ProcessInfo) RawdevStats(a *dpdkInfo, portID int) (*PortStats, error) {

	return pi.stats(a, "/rawdev/stats", portID)
}

// RawdevXStats information
func (pi *ProcessInfo) RawdevXStats(a *dpdkInfo, portID int) (*PortXStats, error) {

	return pi.xstats(a, "/rawdev/stats", portID)
}
