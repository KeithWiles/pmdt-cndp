// SPDX-License-Identifier: BSD-3-Clause
// Copyright(c) 2019-2020 Intel Corporation

package intelpbf

import (
	"encoding/binary"
	"fmt"
	"os"
	"syscall"

	tlog "pmdt.org/ttylog"
)

// A package to read the MSR values from the system

// MSRfd is the file descriptor for MSR registers
type MSRfd struct {
	fd int
}

// Open a MSR register
// Using the FD open the requested CPU MSR file
func Open(cpu int) (MSRfd, error) {

	file := fmt.Sprintf(CPUMsrFile, cpu)

	fd, err := syscall.Open(file, syscall.O_RDWR, 777)
	if err != nil {
		return MSRfd{}, err
	}

	return MSRfd{fd: fd}, nil
}

// Close the MSR CPU file
func (f *MSRfd) Close() error {
	return syscall.Close(f.fd)
}

// ReadAt a give msr register offset
// Read at the given MSR register location the MSR value
func (f MSRfd) ReadAt(msr int64) (uint64, error) {

	b := make([]byte, 8)

	// System call pread to read the MSR special register
	n, err := syscall.Pread(f.fd, b, msr)
	if err != nil {
		tlog.ErrorPrintf("MSR syscall.Pread() failed\n")
		return 0, err
	}
	// If length not value return error
	if n != len(b) {
		return 0, fmt.Errorf("msr read len is not 8 != %d", n)
	}

	// Covert the value to little endain format
	return binary.LittleEndian.Uint64(b), nil
}

// WriteAt a give MSR register
// Writ the given uint64 value to a given MSR register
func (f *MSRfd) WriteAt(msr int64, val uint64) error {
	b := make([]byte, 8)

	// Convert the value to little endian not sure this required
	binary.LittleEndian.PutUint64(b, val)

	if n, err := syscall.Pwrite(f.fd, b, msr); err != nil || n != len(b) {
		if n != len(b) {
			return fmt.Errorf("write of msr length 8 != %d", n)
		}
		return err
	}

	return nil
}

// ReadMsr register by opening, reading and closing the CPU file
// Single routine to read an MSR for a CPU in one go function
func ReadMsr(cpu int, msr int64) (uint64, error) {

	f, err := Open(cpu)
	if err != nil {
		tlog.ErrorPrintf("ReadMsr: MSR Open() failed\n")
		return 0, err
	}
	defer f.Close()

	var val uint64

	if val, err = f.ReadAt(msr); err != nil {
		tlog.ErrorPrintf("MSR ReadAt() failed\n")
		return 0, err
	}
	return val, nil
}

// WriteMsr register by opening, writing and closing the CPU file
// Single routine to write a MSR in one go routine
func WriteMsr(cpu int, msr int64, val uint64) error {

	f, err := Open(cpu)
	if err != nil {
		tlog.ErrorPrintf("WriteMsr: MSR Open() failed\n")
		return err
	}
	defer f.Close()

	if err := f.WriteAt(msr, val); err != nil {
		tlog.ErrorPrintf("MSR WriteAt() failed\n")
		return err
	}

	return nil
}

// fileExists reports whether the named file or directory exists.
// Check to see if the MSR file exists
func fileExists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}
