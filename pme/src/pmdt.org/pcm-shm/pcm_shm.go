// SPDX-License-Identifier: BSD-3-Clause
// Copyright(c) 2019-2020 Intel Corporation

package pcmshm

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"time"
	"unsafe"

	"pmdt.org/encoding/raw"
	hx "pmdt.org/hexdump"
	sem "pmdt.org/semaphore"
	tlog "pmdt.org/ttylog"
)

// Process PCM data by retriving the data from a shared memory region which is
// encoded into a Go structure.
// The shared memory is raw data and we must convert/encode the raw data into
// standard Go structures for the rest of the PME tool

const (
	pcmDefaultPath = "/tmp/opcm-daemon-mmap"
	semaName       = "/opcm-daemon-sema"
)

// Data information
type Data struct {
	path    string
	f       *os.File
	data    []byte
	Opened  bool
	tsc     *SharedHeader
	timo    time.Duration
	action  chan string
	ticker  *time.Ticker
	state   *SharedPCMState
	valid   bool
	Sema	*sem.Semaphore
}

func e(format string, args ...interface{}) error {
	return fmt.Errorf("pcm-shm: "+format, args...)
}

// Reader is the data []bytes in bytes.Buffer (io.Reader)
func (d *Data) Reader() *bytes.Buffer {
	return bytes.NewBuffer(d.data)
}

// Map the file into a mmap page of memory.
func (d *Data) mmap() error {

	f, err := os.Open(d.path)
	if err != nil {
		return err
	}
	d.f = f
	return d.mmapFile()
}

// Len of the shared memory region
func (d *Data) Len() int {

	return len(d.data)
}

// Bytes of the shared memory region
func (d *Data) Bytes() []byte {
	return d.data
}

// Open a file with mmap
func Open(path string) (*Data, error) {

	if len(path) == 0 {
		path = pcmDefaultPath
	}
	d := &Data{path: path}

	d.timo = time.Second
	d.action = make(chan string, 16)
	d.ticker = time.NewTicker(d.timo)
	d.tsc = &SharedHeader{}
	d.state = &SharedPCMState{}

	if err := d.mmap(); err != nil {
		return &Data{}, err
	}

	sema, err := sem.Open(semaName, false, 0, 1)
	if err != nil {
		return nil, err
	}
	d.Sema = sema

	d.Opened = true
	return d, nil
}

// Close a file via munmap
func (d *Data) Close() error {

	if d == nil {
		return nil
	}
	if d.Opened {
		d.Sema.Wait()

		d.Opened = false
		d.action <- "quit"
	}

	d.Sema.Close()
	return d.munmapFile()
}

// getSharedHeader strucure information
func (d *Data) getSharedHeader() bool {

	tsc := raw.NewEncoder(d.Reader())

	if err := tsc.Encode(d.tsc); err != nil {
		tlog.DebugPrintf("encode of SharedHeader failed\n")
		return false
	}

	return true
}

// Start the timer and parsing of the mmap data into structure
// Parsing the shared memory data is done every second
func (d *Data) Start() {

	if !d.Opened {
		fmt.Printf("mmap not opened\n")
		return
	}

	d.Lock()
	d.getSharedHeader()
	d.Unlock()

	go func() {
	ForLoop:
		for {
			select {
			case event := <-d.action:
				if strings.ToLower(event) == "quit" {
					break ForLoop
				}
			case <-d.ticker.C:
				d.doUpdate()
			}
		}
	}()
}

// Update the shared Go structure after validating the memory exist and is valid
func (d *Data) doUpdate() {

	d.Lock()
	defer d.Unlock()

	if !d.getSharedHeader() {
		return
	}

	// Perform the encoding of the raw data to a SharedPMUState structure
	enc := raw.NewEncoder(d.Reader())

	tlog.DebugPrintf(">>>> Start\n")

	if err := enc.Encode(d.state); err != nil {
		tlog.DoPrintf("encode: state %v\n", err)
		return
	}

	tlog.DebugPrintf(">>>> End\n")
}

// Lock data for the caller to use and return true if locked
func (d *Data) Lock() bool {

	d.Sema.Wait()

	return true
}

// Unlock data for the next user
func (d *Data) Unlock() {

	d.Sema.Post()
}

// State data from the PCM shared memory
func (d *Data) State() *SharedPCMState {
	return d.state
}

// DumpPCI a section of the data
func (d *Data) DumpPCI(len int) string {
	return hx.HexDump("PCI", d.data, int(unsafe.Offsetof(d.state.Sample)), len)
}

// DumpQPI a section of the data
func (d *Data) DumpQPI(len int) string {
	return hx.HexDump("QPI", d.data, int(unsafe.Offsetof(d.state.Sample)), len)
}
