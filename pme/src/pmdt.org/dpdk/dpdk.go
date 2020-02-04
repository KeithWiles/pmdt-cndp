// SPDX-License-Identifier: BSD-3-Clause
// Copyright(c) 2019-2020 Intel Corporation

// +build foobar

package dpdk

import (
	"encoding/json"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/shirou/gopsutil/cpu"
	"io/ioutil"
	"os"
	tlog "pmdt.org/ttylog"
	"strconv"
	"strings"
	"time"
)

const (
	// MaxPortCount for the number of ports
	MaxPortCount int = 16
	// MaxDataSize for the number of queues
	MaxDataSize int = 8
)

// QueueData containing
type QueueData []uint64

// PortStats - port stats
type PortStats struct {
	PortID      int       `json:"port_id"`
	InPackets   uint64    `json:"rx_good_packets"`
	OutPackets  uint64    `json:"tx_good_packets"`
	InBytes     uint64    `json:"rx_good_bytes"`
	OutBytes    uint64    `json:"tx_good_bytes"`
	InMissed    uint64    `json:"rx_missed_errors"`
	InErrors    uint64    `json:"rx_errors"`
	OutErrors   uint64    `json:"tx_errors"`
	RxNomBuf    uint64    `json:"rx_mbuf_allocation_errors"`
	InQPackets  QueueData `json:"q_ipackets"`
	OutQPackets QueueData `json:"q_opackets"`
	InQBytes    QueueData `json:"q_ibytes"`
	OutQBytes   QueueData `json:"q_obytes"`
	QErrors     QueueData `json:"q_errors"`
	RXQEmpty    QueueData `json:"rxq_empty_polls"`
	RXQBurst    QueueData `json:"rxq_burst"`
	TXQBurst    QueueData `json:"txq_burst"`
}

// EthdevPort information Ethdev
type EthdevPort struct {
	PortID     int
	BasicStats string
	Led        string
	LinkJSON   string
	LinkState  string
	Mtu        string
	SocketID   string
	Stats      string
}

// Ethdev - Ethdev information
type Ethdev struct {
	AvailCount string
	AvailTotal string
	Ports      []*EthdevPort
}

// AppInfo - Information about the app
type AppInfo struct {
	Pid         int
	AppName     string       // Name of DPDK applcation from the directory entry
	dpdkApp     bool         // True if a PDK application
	CoreList    string       // comma list of cores used by application
	cmdLine     string       // Command line string
	version     string       // DPDK version
	fuseVersion string       // FUSE version
	procType    string       // DPDK process type
	socketCnt   string       // Number of sockets in the system
	Ethdev      *Ethdev      // Ethdev information
	cmdData     *CmdLineData // command line data information
}

// Callback structure and data
type Callback struct {
	name string
	cb   func(event int)
}

// cfgInfo - Information about DPDK apps
type cfgInfo struct {
	basePath string
	currApps []*AppInfo
	mapNames map[string]*AppInfo
	callback map[string]*Callback
	numCPUs  int
	opened   bool
	watcher  *fsnotify.Watcher
}

const (
	dpdkDefaultPath = "/dpdk"
)

// Define the events the application callback will use
const (
	AppInited  = 0
	AppCreated = 1
	AppRemoved = 2
)

var cfg cfgInfo
var dpdkBasePath = dpdkDefaultPath

func init() {
	cfg.basePath = dpdkDefaultPath
	cfg.mapNames = make(map[string]*AppInfo)
	cfg.callback = make(map[string]*Callback)

	num, err := cpu.Counts(true)
	if err != nil {
		tlog.Log(tlog.FatalLog, "Unable to get number of CPUs: %v", err)
		os.Exit(1)
	}

	cfg.numCPUs = num

	a := setupSystem()

	cfg.currApps = append(cfg.currApps, a)
	cfg.mapNames[a.AppName] = a

	cfg.opened = false
}

func setupSystem() *AppInfo {

	// Add the default system level event monitor
	a := &AppInfo{AppName: "System", dpdkApp: false, Pid: -1, CoreList: "all"}
	a.cmdLine = fmt.Sprintf("system -l 0-%d", cfg.numCPUs-1)
	a.cmdData = ParseCmdLine(a.cmdLine)

	return a
}

// Open the dpdk directory and read the first set of directories
func Open(path string) error {
	if len(path) == 0 {
		path = cfg.basePath
	} else {
		cfg.basePath = path
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}

	cfg.watcher = watcher

	Apps() // Grab a current copy

	go func() {

		for _, c := range cfg.callback {
			c.cb(AppInited)
		}

		for {
			select {
			case event, ok := <-watcher.Events:
				// tlog.DoPrintf("fsnotify event %+v, ok %v\n", event, ok)
				if !ok {
					continue
				}
				time.Sleep(time.Second / 2)
				switch {
				case (event.Op & fsnotify.Create) == fsnotify.Create:
					Apps()

					for _, c := range cfg.callback {
						c.cb(AppCreated)
					}
				case (event.Op & fsnotify.Remove) == fsnotify.Remove:
					Apps()

					for _, c := range cfg.callback {
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

	if err := watcher.Add(cfg.basePath); err != nil {
		return err
	}
	cfg.opened = true

	return nil
}

// Close the DPDK path and kill the go thread for watching the directory
func Close() {
	if !cfg.opened {
		return
	}

	cfg.watcher.Close()

	cfg.watcher = nil

	cfg.callback = make(map[string]*Callback)
	cfg.mapNames = make(map[string]*AppInfo)

	// Add the default system level event monitor
	a := setupSystem()

	cfg.currApps = append(cfg.currApps, a)
	cfg.mapNames[a.AppName] = a
}

// Add callback function when /dpdk directory changes
func Add(name string, f func(event int)) {
	cfg.callback[name] = &Callback{name: name, cb: f}

	f(AppInited) // Call it the first time it is setup
}

// Remove callback function
func Remove(name string) {
	_, ok := cfg.callback[name]
	if ok {
		delete(cfg.callback, name)
	}
}

// SetPath to the currect base path for dpdk apps
func SetPath(path string) {
	cfg.basePath = path
}

// Apps return all of the DPDK applications in the system, by reading the
// '/dpdk' directory and creating a AppInfo for each entry in the directory
func Apps() error {

	// read the directory entries in the /dpdk/*
	dirEntries, err := ioutil.ReadDir(cfg.basePath)
	if err != nil {
		tlog.DoPrintf("ReadDir failed: %s\n", err)
		return err
	}

	info := []*AppInfo{}
	names := make(map[string]*AppInfo)

	// Add the default system level event monitor
	a := setupSystem()

	info = append(info, a)
	names[a.AppName] = a

	for _, entry := range dirEntries {

		if entry.IsDir() {
			appName := entry.Name()

			version := readFile(appName, "version")
			if len(version) == 0 {
				continue
			}

			app := &AppInfo{AppName: appName, version: version, dpdkApp: true, CoreList: "all"}

			app.Ethdev = &Ethdev{}

			app.cmdLine = readFile(appName, "command-line")
			if len(app.cmdLine) > 0 {
				app.cmdData = ParseCmdLine(app.cmdLine)
			}
			app.procType = readFile(appName, "proc-type")
			app.fuseVersion = readFile(appName, "fuse-version")
			app.socketCnt = readFile(appName, "eal/socket-cnt")

			if strings.Contains(appName, "dpdk-") {
				pid, err := strconv.ParseInt(appName[5:], 0, 0)
				app.Pid = -1
				if err == nil {
					app.Pid = int(pid)
				}
			}

			info = append(info, app)
			names[appName] = app
		}
	}
	cfg.currApps = info
	cfg.mapNames = names

	return nil
}

// IsAppDPDK is a valid DPDK application
func IsAppDPDK(a *AppInfo) bool {
	return a.dpdkApp
}

// ConvertJSON - convert JSON stats data to PortStats structure
func ConvertJSON(str string) (*PortStats, error) {
	portStats := &PortStats{}

	dat := []byte(str)

	err := json.Unmarshal(dat, portStats)
	if err != nil {
		return nil, fmt.Errorf("unmarshal failed: %v\n%s", err, str)
	}

	return portStats, nil
}

func readFile(dpdkDir, file string) string {

	path := fmt.Sprintf("%s/%s/%s", cfg.basePath, dpdkDir, file)

	if strings.Contains(dpdkDir, "/dpdk/") {
		path = fmt.Sprintf("%s/%s", dpdkDir, file)
	}

	dat, err := ioutil.ReadFile(path)
	if err != nil {
		tlog.ErrorPrintf("%s\n", path, err)
		return ""
	}
	s := strings.TrimSpace(string(dat))

	return s
}

type fileFunc func(interface{}, os.FileInfo, string) error

func processDir(v interface{}, dir string, f fileFunc) error {

	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return err
	}

	for _, file := range files {
		err = f(v, file, dir)
		if err != nil {
			return err
		}
	}
	return nil
}

// EthdevData - get the ethdev information
func EthdevData(app *AppInfo) (bool, error) {

	if app.AppName == "System" {
		return false, nil
	}

	dir := fmt.Sprintf("%s/%s/ethdev", cfg.basePath, app.AppName)

	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return false, err
	}
	app.Ethdev = &Ethdev{}
	app.Ethdev.Ports = []*EthdevPort{}

	ethdevDirFunc := func(v interface{}, fi os.FileInfo, dirPath string) error {
		ethdev := v.(*Ethdev)

		if fi.Mode().IsRegular() {
			dat := readFile(app.AppName, "ethdev/"+fi.Name())

			switch fi.Name() {
			case "avail_count":
				ethdev.AvailCount = dat

			case "total_count":
				ethdev.AvailTotal = dat
			}
		}
		return nil
	}

	portDirFunc := func(v interface{}, fi os.FileInfo, dirPath string) error {
		port := v.(*EthdevPort)

		if fi.Mode().IsRegular() {
			dat := readFile(dirPath, fi.Name())

			switch fi.Name() {
			case "basic-stats":
				port.BasicStats = dat

			case "led":
				port.Led = dat

			case "link.json":
				port.LinkJSON = dat

			case "link":
				port.LinkState = dat

			case "mtu":
				port.Mtu = dat

			case "socket_id":
				port.SocketID = dat

			case "stats.json":
				port.Stats = dat
			}
		}
		return nil
	}

	err = processDir(app.Ethdev, dir, ethdevDirFunc)
	if err != nil {
		tlog.ErrorPrintf("ethdevDirFunc error: %s\n", err)
		return false, err
	}

	for _, file := range files {
		if file.Mode().IsDir() && strings.Contains(file.Name(), "port-") {
			port := &EthdevPort{}
			pid, _ := strconv.ParseInt(file.Name()[5:], 0, 32)
			port.PortID = int(pid)

			d := fmt.Sprintf("%s/%s/ethdev/%s", cfg.basePath, app.AppName, file.Name())

			err = processDir(port, d, portDirFunc)
			if err != nil {
				tlog.ErrorPrintf("portDirFunc error: %s\n", err)
				return false, err
			}

			app.Ethdev.Ports = append(app.Ethdev.Ports, port)
		}
	}

	return true, nil
}

func isAppsEqual(oldAppsInfo, newAppsInfo []*AppInfo) bool {

	if len(newAppsInfo) != len(oldAppsInfo) {
		return false
	}
	// Verify the list is the same list or set true if not the same
	for i, v := range oldAppsInfo {
		if v.AppName != newAppsInfo[i].AppName {
			return false
		}
	}

	return true
}

// NameStrings - Display the application names in the apps window table
func NameStrings() []string {

	names := make([]string, 0)

	for _, info := range cfg.currApps {
		names = append(names, info.AppName)
	}

	return names
}

// Names - Display the application names in the apps window table
func Names() []interface{} {

	names := make([]interface{}, 0)

	for _, info := range cfg.currApps {
		names = append(names, info.AppName)
	}

	return names
}

// Pid value of the selected application
func Pid(name string) int {
	a, ok := cfg.mapNames[name]
	if !ok {
		return -1
	}

	return a.Pid
}

// Name from the application selected
func Name(sel int) string {

	if sel < 0 || sel > len(cfg.currApps) {
		return ""
	}

	return cfg.currApps[sel].AppName
}

// Data from the application selected
func Data(sel int) *AppInfo {

	if sel < 0 || sel > len(cfg.currApps) {
		return cfg.currApps[0]
	}

	return cfg.currApps[sel]
}

// CmdLine information
func CmdLine(a *AppInfo) string {

	if a.dpdkApp && len(a.cmdLine) == 0 {
		if a.AppName != "System" {
			a.cmdLine = readFile(a.AppName, "command-line")
		}
	}
	return a.cmdLine
}

// ProcType function to return the DPDK process type
func ProcType(a *AppInfo) string {

	if a.dpdkApp && len(a.procType) == 0 {
		a.procType = readFile(a.AppName, "proc-type")
	}

	return a.procType
}

// FuseVersion function to return the DPDK FUSE version
func FuseVersion(a *AppInfo) string {

	if a.dpdkApp && len(a.fuseVersion) == 0 {
		a.fuseVersion = readFile(a.AppName, "fuse-version")
	}
	return a.fuseVersion
}

// Version function to return the DPDK FUSE version
func Version(a *AppInfo) string {

	if a.dpdkApp && len(a.version) == 0 {
		a.version = readFile(a.AppName, "version")
	}
	return a.version
}

// Sockets is the number of socket in the system
func Sockets(a *AppInfo) string {

	if a.dpdkApp && len(a.socketCnt) == 0 {
		a.socketCnt = readFile(a.AppName, "eal/socket-cnt")
	}
	return a.socketCnt
}

// GetCmdLineData for the given application
func GetCmdLineData(a *AppInfo) *CmdLineData {
	return a.cmdData
}
