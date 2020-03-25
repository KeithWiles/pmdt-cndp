// SPDX-License-Identifier: BSD-3-Clause
// Copyright(c) 2019-2020 Intel Corporation

package pcm

/*
#include <stdint.h>

void
pcm_cpuid_subleaf(int leaf, const unsigned int subleaf, uint32_t regs[])
{
	__asm__ __volatile__ ("cpuid" : "=a" (regs[0]), "=b" (regs[1]), "=c" (regs[2]), "=d" (regs[3]) : "a" (leaf), "c" (subleaf));
}

void
pcm_cpuid(int leaf, uint32_t regs[])
{
	pcm_cpuid_subleaf(leaf, 0, regs);
}

#if 0
double pop_mean(int numPoints, double a[]) {
    if (a == NULL || numPoints == 0) {
        return 0;
    }
    double mean = 0;
    for (int i = 0; i < numPoints; i++) {
        mean += a[i];
    }
	return mean / numPoints;
}
#endif
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

// CPUid fills in the CPU id information
func (p *Info) CPUid(leaf int) *CPURegs {

	arr := make([]C.uint32_t, 4, 4)
	fv := &(arr[0])

	C.pcm_cpuid(C.int(leaf), fv)

	a := &CPURegs{}

	a.eax = uint32(arr[0])
	a.ebx = uint32(arr[1])
	a.ecx = uint32(arr[2])
	a.edx = uint32(arr[3])

	return a
}

/*
// MyTest to figure out cgo
func (p *Info) MyTest() {
	// Get a basic function to work, while passing in an ARRAY

	// Create a dummy array of (10,20,30), the mean of which is 20.
	arr := make([]C.double, 0)
	arr = append(arr, C.double(10.0))
	arr = append(arr, C.double(20.0))
	arr = append(arr, C.double(30.0))
	firstValue := &(arr[0]) // this notation seems to be pretty important... Re-use this!
	// if you don't make it a pointer right away, then you make a whole new object in a different location, so the contiguous-ness of the array is jeopardized.
	// Because we have IMMEDIATELY made a pointer to the original value,the first value in the array, we have preserved the contiguous-ness of the array.
	fmt.Println("array length: ", len(arr))

	var arrayLength C.int
	arrayLength = C.int(len(arr))
	// arrayLength = C.int(2)

	fmt.Println("array length we are using: ", arrayLength)

	arrayMean := C.pop_mean(arrayLength, firstValue)
	fmt.Println("pop_mean (10, 20, 30): ", arrayMean)
}
*/