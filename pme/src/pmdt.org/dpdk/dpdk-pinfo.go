// SPDX-License-Identifier: BSD-3-Clause
// Copyright(c) 2019-2020 Intel Corporation

// +build foo

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

// AppInfo - Information about the app
type AppInfo struct {
	valid  bool // true if the process info data is valid
	conn   *net.UnixConn
	Pid    int64     // Pid for the DPDK process
	Path   string    // Path of the process_info.<pid> file
	Params EALParams // Holds the EAL parameter data
	Info   RTEInfo   // Holds the DPDK process info data
	Cmds   EALCmds   // List of all known commands
	PrevStats [MaxPorts]PortStats
	PrevXStats [MaxPorts]PortXStats
}

// Callback structure and data
type Callback struct {
	name string          // string name of the callback used as key
	cb   func(event int) // function to callback the application for notifies
}

// AppData holds all of the process info data
type AppData map[string]*AppInfo

// CallbackData hold the watcher fsnotify callback information
type CallbackData map[string]*Callback

// ProcessInfo data for DPDK
type ProcessInfo struct {
	lock     sync.Mutex
	opened   bool              // true if process info open
	basePath string            // Base path to the dpdk run directory
	currApps AppData           // Indexed by path for each DPDK application
	callback CallbackData      // Callback routines for the fsnotify
	watcher  *fsnotify.Watcher // watcher for the directory notify
}

const (
	dpdkDefaultPath = "/var/run/dpdk" // The DPDK default path string
	dpdkBaseName    = "process_info." // The process info base filename
)

// Define the events the application callback will use
const (
	AppInited  = iota
	AppCreated = iota
	AppRemoved = iota
)

var dpdkBasePath = dpdkDefaultPath

// NewProcessInfo information structure
func NewProcessInfo(w ...string) *ProcessInfo {

	pi := &ProcessInfo{}

	pi.basePath = dpdkDefaultPath + "/rte"
	pi.currApps = make(map[string]*AppInfo)
	pi.callback = make(map[string]*Callback)

	pi.opened = false
	if len(w) > 0 {
		pi.basePath = w[0]
	}

	return pi
}

// Ethdev information
func (pi *ProcessInfo) doCmd(a *AppInfo, cmd string) ([]byte, error) {

	if _, err := a.conn.Write([]byte(cmd)); err != nil {
		return nil, fmt.Errorf("write on socket failed: %v", err)
	}

	buf := make([]byte, a.Info.MaxBuffer) // big buffer

	n, err := a.conn.Read(buf)
	if err != nil {
		return nil, err
	}
}

func (pi *ProcessInfo) getStaticData(a *AppInfo) error {

	// Setup the first time then use the returned value
	a.Info.MaxBuffer = 1024

	d, err := pi.doCmd(a, "/dpdk:info")
	if err != nil {
		return err
	}

	// Parse the information from the DPDK info string
	if err = json.Unmarshal(d, &a.Info); err != nil {
		return err
	}

	d, err = pi.doCmd(a, "/eal:params")
	if err != nil {
		return err
	}

	// Parse the information from the DPDK info string
	if err = json.Unmarshal(d, &a.Params); err != nil {
		return err
	}

	pi.Commands(a)
	d, err = pi.doCmd(a, "/")
	if err != nil {
		return err
	}

	// Parse the information from the DPDK info string
	if err = json.Unmarshal(d, &a.Params); err != nil {
		return err
	}
	return nil
}

func (pi *ProcessInfo) watchDir(path string, fi os.FileInfo, err error) error {

	if fi == nil {
		return nil
	}
	if fi.Mode().IsDir() {
		tlog.DoPrintf("watchDir: %s\n", path)
		return pi.watcher.Add(path)
	}
	return nil
}

// exists returns whether the given file or directory exists
func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return true, err
}

// Open the dpdk directory and read the first set of directories
func (pi *ProcessInfo) Open(w ...string) error {

	// Create the watcher to watch all sub-directories
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	pi.watcher = watcher

	if ok, _ := exists(dpdkDefaultPath); !ok {
		os.MkdirAll(dpdkDefaultPath, os.ModePerm)
	}

	tlog.DoPrintf("Watch: %s\n", dpdkDefaultPath)

	watcher.Add(dpdkDefaultPath)

	if err := filepath.Walk(dpdkDefaultPath, pi.watchDir); err != nil {
		return fmt.Errorf("%s: %v", dpdkDefaultPath, err)
	}

	pi.scan()

	go func() {
		// Callback the user level functions for changes the first time
		for _, c := range pi.callback {
			c.cb(AppInited)
		}

		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					continue
				}
				//				time.Sleep(time.Second / 2)
				switch {
				case (event.Op & fsnotify.Create) == fsnotify.Create:

					tlog.DoPrintf("Event: %s\n", event.String())

					if fi, err := os.Stat(event.Name); err == nil {
						if fi.Mode().IsDir() {
							tlog.DoPrintf("Add Watcher for %s\n", event.Name)
							watcher.Add(event.Name)
						}
					}

					pi.scan()

					for _, c := range pi.callback {
						c.cb(AppCreated)
					}
				case (event.Op & fsnotify.Remove) == fsnotify.Remove:
					tlog.DoPrintf("Event: %s\n", event.String())

					tlog.DoPrintf("Remove Watcher for %s\n", event.Name)
					watcher.Remove(event.Name)

					pi.scan()

					for _, c := range pi.callback {
						c.cb(AppRemoved)
					}
				}

			case err, ok := <-watcher.Errors:
				if !ok {
					continue
				}
				tlog.ErrorPrintf("fsnotify: error %v", err)
			}
		}
	}()

	pi.opened = true

	return nil
}

// Close the connection
func (pi *ProcessInfo) Close() {

	if !pi.opened {
		return
	}

	pi.watcher.Close()

	pi.watcher = nil
	pi.opened = false
}

// Add callback function when /dpdk directory changes
func (pi *ProcessInfo) Add(name string, f func(event int)) {
	pi.callback[name] = &Callback{name: name, cb: f}

	f(AppInited) // Call it the first time it is setup
}

// Remove callback function
func (pi *ProcessInfo) Remove(name string) {
	_, ok := pi.callback[name]
	if ok {
		delete(pi.callback, name)
	}
}

// SetPath to the currect base path for dpdk apps
func (pi *ProcessInfo) SetPath(path string) {
	pi.basePath = path
}

// Scan for the DPDK process info socket files
func (pi *ProcessInfo) scan() {

	dpdkDirs, err := ioutil.ReadDir(dpdkBasePath)
	if err != nil {
		log.Fatalf("ReadDir failed: %v\n", err)
	}

	pi.lock.Lock()
	defer pi.lock.Unlock()

	// Set all of the current files to false, to allow for removal later
	// When we find the same one in the scan we mark it as true, then
	// remove the ones that are not valid anymore
	for _, a := range pi.currApps {
		a.valid = false
	}

	for _, entry := range dpdkDirs {

		dpdkDir := entry.Name()

		if entry.IsDir() {
			appFiles, err := ioutil.ReadDir(dpdkBasePath + "/" + dpdkDir)
			if err != nil {
				log.Fatalf("Unable to open %s\n", dpdkBasePath+"/"+dpdkDir)
			}

			for _, file := range appFiles {
				name := file.Name()

				if strings.HasPrefix(filepath.Base(name), dpdkBaseName) {
					ext := filepath.Ext(name)

					pid, err := strconv.ParseInt(ext[1:], 10, 64)
					if err != nil {
						tlog.WarnPrintf("uable to parse pid from filename %s\n", name)
						continue
					}

					path := dpdkBasePath + "/" + dpdkDir + "/" + name
					if a, ok := pi.currApps[path]; ok {
						a.valid = true
						continue
					}

					// Open the connection to the DPDK application
					t := "unixpacket"
					laddr := net.UnixAddr{Name: path, Net: t}
					conn, err := net.DialUnix(t, nil, &laddr)
					if err != nil {
						log.Fatalf("connection to socket failed: %v", err)
					}

					ap := &AppInfo{valid: true, Pid: pid, Path: path, conn: conn}

					if err := pi.getStaticData(ap); err != nil {
						break
					}

					// Add the AppInfo to the internal map structures
					pi.currApps[path] = ap
				}
			}
		}
	}

	// release AppInfo data for old process info files/pids
	for k, a := range pi.currApps {
		if !a.valid {
			a.conn.Close()
			delete(pi.currApps, k)
		}
	}
}

// Scan the directory for new or removed process info files
func (pi *ProcessInfo) Scan() {

	if pi != nil {
		pi.scan()
	}
}

// AppList returns the list of process info structures
func (pi *ProcessInfo) AppList() AppData {

	return pi.currApps
}

// AppItem returns the application info structure for the given name
func (pi *ProcessInfo) AppItem(name string) *AppInfo {

	if a, ok := pi.currApps[name]; ok {
		return a
	}
	return nil
}

// Files returns a string slice of DPDK application process info data
func (pi *ProcessInfo) Files() []string {

	files := []string{}
	for _, a := range pi.currApps {
		files = append(files, a.Path)
	}

	return files
}

// AppInfoByIndex returns the AppInfo pointer by the index
func (pi *ProcessInfo) AppInfoByIndex(idx int) *AppInfo {

	if idx >= 0 {
		files := pi.Files()
		if a, ok := pi.currApps[files[idx]]; ok {
			return a
		}
	}
	return nil
}

// Args returns the current DPDK command line
func (pi *ProcessInfo) Args(a *AppInfo) (EALParams, error) {

	return a.Params, nil
}

// CommandList returns the list of commands in DPDK
func (pi *ProcessInfo) CommandList(a *AppInfo) ([]byte, error) {

	return pi.doCmd(a, "/")
}

// Commands returns the list of commands
func (pi *ProcessInfo) Commands(a *AppInfo) ([]string, error) {

	d, err := pi.doCmd(a, "/")
	if err != nil {
		return nil, err
	}

	data := struct {
		Cmds []string
	}{}

	if err := json.Unmarshal(d, &data); err != nil {
		return nil, err
	}

	return data.Cmds, nil
}

// list of devices as port ids
func (pi *ProcessInfo) list(a *AppInfo, cmd string) (*DevList, error) {

	b, err := pi.doCmd(a, cmd)
	if err != nil {
		return nil, err
	}

	list := &DevList{}

	if err := json.Unmarshal(b, list); err != nil {
		return nil, err
	}
	return list, nil
}

// stats information
func (pi *ProcessInfo) stats(a *AppInfo, cmd string, portID int) (*PortStats, error) {

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
func (pi *ProcessInfo) xstats(a *AppInfo, cmd string, portID int) (*PortXStats, error) {

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
func (pi *ProcessInfo) Info(a *AppInfo) *RTEInfo {

	if a == nil {
		return nil
	}
	return &a.Info
}

// EthdevList of devices as port ids
func (pi *ProcessInfo) EthdevList(a *AppInfo) (*DevList, error) {

	return pi.list(a, "/ethdev:list")
}

// EthdevStats information
func (pi *ProcessInfo) EthdevStats(a *AppInfo, portID int) (*PortStats, error) {

	pstats, err := pi.stats(a, "/ethdev:stats", portID)
	if err != nil {
		return nil, fmt.Errorf("/ethdev:stats failed: %v", err)
	}
	p := *pstats
	a.PrevStats[portID] = p

	return pstats, err
}

// PreviousStats is the save stats data
func (pi *ProcessInfo) PreviousStats(a *AppInfo, portID int) (PortStats, error) {

	if portID >= len(a.PrevStats) {
		return PortStats{}, fmt.Errorf("invalid port id")
	}
	return a.PrevStats[portID], nil
}

// EthdevXStats information
func (pi *ProcessInfo) EthdevXStats(a *AppInfo, portID int) (*PortXStats, error) {

	return pi.xstats(a, "/ethdev:xstats", portID)
}

// PreviousXStats is the save stats data
func (pi *ProcessInfo) PreviousXStats(a *AppInfo, portID int) (*PortXStats, error) {
	if portID >= len(a.PrevStats) {
		return nil, fmt.Errorf("invalid port id")
	}
	pstats := &a.PrevXStats[portID]

	return pstats, nil
}

// RawdevList of devices as port ids
func (pi *ProcessInfo) RawdevList(a *AppInfo) (*DevList, error) {

	return pi.list(a, "/rawdev:list")
}

// RawdevStats information
func (pi *ProcessInfo) RawdevStats(a *AppInfo, portID int) (*PortStats, error) {

	return pi.stats(a, "/rawdev:stats", portID)
}

// RawdevXStats information
func (pi *ProcessInfo) RawdevXStats(a *AppInfo, portID int) (*PortXStats, error) {

	return pi.xstats(a, "/rawdev:stats", portID)
}
