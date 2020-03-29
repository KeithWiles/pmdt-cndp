// SPDX-License-Identifier: BSD-3-Clause
// Copyright(c) 2019-2020 Intel Corporation

// +build !linux

package pcm

/*
#include <stdint.h>
#include <unistd.h>
#include <sys/types.h>
#include <sys/stat.h>
#include <fcntl.h>

#include "pcm_class.h"

struct pcm_class _pcm;

void
pcm_cpuid_subleaf(int leaf, const int subleaf, uint32_t regs[])
{
	__asm__ __volatile__ (
		"cpuid" : "=a" (regs[0]),
		"=b" (regs[1]),
		"=c" (regs[2]),
		"=d" (regs[3]) : "a" (leaf),
		"c" (subleaf));
}

void
pcm_cpuid(int leaf, uint32_t regs[])
{
	pcm_cpuid_subleaf(leaf, 0, regs);
}

int
MsrHandle_open(uint32_t cpu)
{
	int handle;
    char path[256];
	snprintf(path, sizeof(path), "/dev/cpu/%d/msr", cpu);

	handle = open(path, O_RDWR);

    if (handle < 0)
		 printf("PCM Error: can't open MSR handle for core\n");
	else
		_pcm.msr_fd = handle;

    return handle;
}

void
MsrHandle_close(int fd)
{
	if (fd >= 0)
		close(fd);
}

ssize_t
MsrHandle_write(uint64_t msr_number, uint64_t value)
{
	if (_pcm.msr_fd <= 0)
		return -1;
    return pwrite(_pcm.msr_fd, (const void *)&value, sizeof(uint64_t), msr_number);
}

ssize_t
MsrHandle_read(uint64_t msr_number, uint64_t * value)
{
	if (_pcm.msr_fd <= 0)
		return -1;
    return pread(_pcm.msr_fd, (void *)value, sizeof(uint64_t), msr_number);
}
*/
import "C"

import (
)

// CPURegs to store the cpuid values
type CPURegs struct {
	eax, ebx, ecx, edx uint32
}

func buildBit(beg, end uint32) uint32 {
	var v uint32 = 0

    if end == 31 {
        v = uint32(0xFFFFFFFF);
	} else {
        v = (1 << (end + 1)) - 1;
    }
    v = v >> beg;
    return v;
}

func extractBits(val uint32, beg, end uint32) uint32 {

	var v uint32
	var beg1, end1 uint32

	if beg <= end {
		beg1 = beg
		end1 = end
	} else {
		beg1 = end
		end1 = beg
	}

	v = val >> beg
	v = v & buildBit(beg1, end1)

	return v
}

// CPUidSubleaf to get data for leaf and subleaf
func (p *Info) CPUidSubleaf(leaf, subleaf int) *CPURegs {

	arr := make([]C.uint32_t, 4, 4)
	fv := &(arr[0])

	C.pcm_cpuid_subleaf(C.int(leaf), C.int(subleaf), fv)

	a := &CPURegs{}

	a.eax = uint32(arr[0])
	a.ebx = uint32(arr[1])
	a.ecx = uint32(arr[2])
	a.edx = uint32(arr[3])

	return a
}

// CPUid fills in the CPU id information
func (p *Info) CPUid(leaf int) *CPURegs {

	return p.CPUidSubleaf(leaf, 0)
}

// SetCPUModel in the C code
func (p *Info) SetCPUModel() {

	C._pcm.cpu_model = C.int(p.cpuModel)
}