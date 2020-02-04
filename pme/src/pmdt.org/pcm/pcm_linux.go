// SPDX-License-Identifier: BSD-3-Clause
// Copyright(c) 2019-2020 Intel Corporation

package pcm

import (
	"fmt"
	"log"
	"os"
	"syscall"
)

// mmap a file into memory and set the location in the Data structure
func (d *Data) mmapFile() error {
	st, err := d.f.Stat()
	if err != nil {
		log.Fatalf("pcm_linux: %v\n", err)
	}
	size := st.Size()
	if int64(int(size+4095)) != size+4095 {
		log.Fatalf("%s: too large for mmap", d.f.Name())
	}
	n := int(size)
	if n == 0 {
		return fmt.Errorf("file size is zero")
	}

	pagesize := os.Getpagesize()
	datalen := int((n + (pagesize - 1)) & ^(pagesize - 1))

	data, err := syscall.Mmap(int(d.f.Fd()), 0, datalen, syscall.PROT_READ, syscall.MAP_SHARED)
	if err != nil {
		log.Fatalf("mmap %s: %v", d.f.Name(), err)
	}
	d.data = data[:n]
	return nil
}

// release a mmap file and its resources
func (d *Data) munmapFile() error {

	if d.data == nil {
		return nil
	}
	if err := syscall.Munmap(d.data); err != nil {
		return err
	}
	return d.f.Close()
}
