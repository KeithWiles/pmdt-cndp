// SPDX-License-Identifier: BSD-3-Clause
// Copyright(c) 2019-2020 Intel Corporation

package pinfo

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

// AppInfo - Information about the app
type AppInfo struct {
	valid  bool		// true if the process info data is valid
	conn   *net.UnixConn
	Pid    int64    // Pid for the process
	Path   string   // Path of the process_pinfo.<pid> file
}

// Callback structure and data
type Callback struct {
	name string          // string name of the callback used as key
	cb   func(event int) // function to callback the application for notifies
}

// AppMapByPath holds all of the process info data
type AppMapByPath map[string]*AppInfo

// AppMapByPid holds all of the process info data
type AppMapByPid map[int64]*AppInfo

// CallbackMap holds the watcher fsnotify callback information
type CallbackMap map[string]*Callback

// ProcessInfo data for applications
type ProcessInfo struct {
	lock     sync.Mutex
	opened   bool              // true if process info open
	basePath string            // Base path to the run directory
	baseName string			   // Base file name
	appsByPath AppMapByPath    // Indexed by path for each application
	appsByPid AppMapByPid
	callback CallbackMap       // Callback routines for the fsnotify
	watcher  *fsnotify.Watcher // watcher for the directory notify
}

// Define the events the application callback will use
const (
	AppInited  = iota
	AppCreated = iota
	AppRemoved = iota

	maxBufferSize = (16 * 1024)
)

// NewProcessInfo information structure
func NewProcessInfo(bpath, bname string) *ProcessInfo {

	pi := &ProcessInfo{ basePath: bpath, baseName: bname }

	pi.appsByPath = make(AppMapByPath)
	pi.appsByPid = make(AppMapByPid)
	pi.callback = make(CallbackMap)

	pi.opened = false

	return pi
}

// doCmd information
func (pi *ProcessInfo) doCmd(a *AppInfo, cmd string) ([]byte, error) {

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

func (pi *ProcessInfo) watchDir(path string, fi os.FileInfo, err error) error {

	if fi == nil {
		return nil
	}
	if fi.Mode().IsDir() {
		tlog.DebugPrintf("watchDir: %s\n", path)
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

// Open the directory and read the first set of directories
func (pi *ProcessInfo) Open() error {

	// Create the watcher to watch all sub-directories
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	pi.watcher = watcher

	if ok, _ := exists(pi.basePath); !ok {
		os.MkdirAll(pi.basePath, os.ModePerm)
	}

	tlog.DebugPrintf("Watch: %s\n", pi.basePath)

	watcher.Add(pi.basePath)

	if err := filepath.Walk(pi.basePath, pi.watchDir); err != nil {
		return fmt.Errorf("%s: %v", pi.basePath, err)
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

					//tlog.DoPrintf("Event: %s\n", event.String())

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
					tlog.DebugPrintf("Event: %s\n", event.String())

					tlog.DebugPrintf("Remove Watcher for %s\n", event.Name)
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

// Add callback function when directory changes
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

// SetPath to the currect base path for apps
func (pi *ProcessInfo) SetPath(path string) {
	pi.basePath = path
}

func (pi *ProcessInfo) addFile(name, dir string) {

	if strings.HasPrefix(filepath.Base(name), pi.baseName) {
		ext := filepath.Ext(name)

		pid, err := strconv.ParseInt(ext[1:], 10, 64)
		if err != nil {
			tlog.WarnPrintf("uable to parse pid from filename %s\n", name)
			return
		}

		var path string
		if dir == "" {
			path = pi.basePath + "/" + name
		} else {
			path = pi.basePath + "/" + dir + "/" + name
		}
		if a, ok := pi.appsByPath[path]; ok {
			a.valid = true
			return
		}

		// Open the connection to the application
		t := "unixpacket"
		laddr := net.UnixAddr{Name: path, Net: t}
		conn, err := net.DialUnix(t, nil, &laddr)
		if err != nil {
			log.Fatalf("connection to socket failed: %v", err)
		}

		ap := &AppInfo{valid: true, Pid: pid, Path: path, conn: conn}

		// Add the AppInfo to the internal map structures
		pi.appsByPath[path] = ap
		pi.appsByPid[pid] = ap
	}
}

// Scan for the process info socket files
func (pi *ProcessInfo) scan() {

	dirs, err := ioutil.ReadDir(pi.basePath)
	if err != nil {
		log.Fatalf("ReadDir failed: %v\n", err)
	}

	pi.lock.Lock()
	defer pi.lock.Unlock()

	// Set all of the current files to false, to allow for removal later
	// When we find the same one in the scan we mark it as true, then
	// remove the ones that are not valid anymore
	for _, a := range pi.appsByPath {
		a.valid = false
	}

	for _, entry := range dirs {

		if entry.IsDir() {
			appFiles, err := ioutil.ReadDir(pi.basePath + "/" + entry.Name())
			if err != nil {
				log.Fatalf("Unable to open %s\n", pi.basePath + "/" + entry.Name())
			}

			for _, file := range appFiles {
				if strings.HasPrefix(filepath.Base(file.Name()), pi.baseName) {
					pi.addFile(file.Name(), entry.Name())
				}
			}
		} else {
			pi.addFile(entry.Name(), "")
		}
	}

	// release AppInfo data for old process info files/pids
	for k, a := range pi.appsByPath {
		if !a.valid {
			a.conn.Close()
			delete(pi.appsByPid, a.Pid)
			delete(pi.appsByPath, k)
		}
	}
}

// Scan the directory for new or removed process info files
func (pi *ProcessInfo) Scan() {

	if pi != nil {
		pi.scan()
	}
}

// AppsList returns the list of AppInfo structures
func (pi *ProcessInfo) AppsList() []*AppInfo {

	p := make([]*AppInfo, 0)

	for _, a := range pi.appsByPath {
		p = append(p, a)
	}
	return p
}

// Files returns a string slice of application process info data
func (pi *ProcessInfo) Files() []string {

	files := []string{}
	for _, a := range pi.appsByPath {
		files = append(files, a.Path)
	}

	return files
}

// Pids returns a int64 slice of application process info data
func (pi *ProcessInfo) Pids() []int64 {

	pids := make([]int64, 0)
	for _, a := range pi.appsByPid {
		pids = append(pids, a.Pid)
	}

	return pids
}

// AppInfoByPid returns the AppInfo pointer using the Pid
func (pi *ProcessInfo) AppInfoByPid(pid int64) *AppInfo {

	for _, a := range pi.appsByPath {
		if a.Pid == pid {
			return a
		}
	}
	return nil
}

// AppInfoByIndex returns the AppInfo pointer by the index
func (pi *ProcessInfo) AppInfoByIndex(idx int) *AppInfo {

	if idx >= 0 {
		files := pi.Files()
		if a, ok := pi.appsByPath[files[idx]]; ok {
			return a
		}
	}
	return nil
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

// IssueCommand to the process socket
func (pi *ProcessInfo) IssueCommand(a *AppInfo, str string) ([]byte, error) {

	return pi.doCmd(a, str)
}

// Unmarshal the JSON data into a structure
func (pi *ProcessInfo) Unmarshal(command string, data interface{}) error {

	p := pi.AppsList()
	if len(p) == 0 {
		return fmt.Errorf("No PCM data")
	}

	d, err := pi.IssueCommand(p[0], command)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(d, data); err != nil {
		return err
	}

	return nil
}

// Marshal the structure into a JSON string
func (pi *ProcessInfo) Marshal(data interface{}) ([]byte, error) {

	return json.MarshalIndent(data, "", "  ")
}