// SPDX-License-Identifier: BSD-3-Clause
// Copyright(c) 2019-2020 Intel Corporation

package main

import (
	"unsafe"

	pcm "pmdt.org/pcm-shm"
)

// DumpSharedMemory offsets and sizes to debug shared memory layout
func DumpSharedMemory() string {
	foo := pcm.SharedPCMState{}

	s := ""
	s += sprintf("SharedPCMState Size", unsafe.Sizeof(foo))
	s += sprintf("PCMSystem", unsafe.Sizeof(foo.PCMCounters.System))
	s += sprintf("PCMCore", unsafe.Sizeof(foo.PCMCounters.Core))
	s += sprintf("PCMMemory", unsafe.Sizeof(foo.PCMCounters.Memory))
	s += sprintf("PCMQPI", unsafe.Sizeof(foo.PCMCounters.QPI))
	s += sprintf("Sample", unsafe.Sizeof(foo.Sample[0]))
	s += sprintf("PCIeEvents", unsafe.Sizeof(foo.Aggregate))
	s += sprintf("PCMCoreCounter", unsafe.Sizeof(foo.PCMCounters.Core.Cores[0]))
	s += "\n"

	s += sprintf("Header", unsafe.Sizeof(foo.Header), unsafe.Offsetof(foo.Header))
	s += sprintf("  Version", unsafe.Sizeof(foo.Header.Version), unsafe.Offsetof(foo.Header.Version))
	s += sprintf("  TscBegin", unsafe.Sizeof(foo.Header.TscBegin), unsafe.Offsetof(foo.Header.TscBegin))
	s += sprintf("  TscEnd", unsafe.Sizeof(foo.Header.TscEnd), unsafe.Offsetof(foo.Header.TscEnd))
	s += sprintf("  CyclesToGetState", unsafe.Sizeof(foo.Header.CyclesToGetState), unsafe.Offsetof(foo.Header.CyclesToGetState))
	s += sprintf("  TimeStamp", unsafe.Sizeof(foo.Header.TimeStamp), unsafe.Offsetof(foo.Header.TimeStamp))
	s += sprintf("  PollMs", unsafe.Sizeof(foo.Header.PollMs), unsafe.Offsetof(foo.Header.PollMs))
	s += sprintf("  DelayMs", unsafe.Sizeof(foo.Header.DelayMs), unsafe.Offsetof(foo.Header.DelayMs))

	s += sprintf("PCMCounters", unsafe.Sizeof(foo.PCMCounters), unsafe.Offsetof(foo.PCMCounters))
	s += sprintf("Sample", unsafe.Sizeof(foo.Sample), unsafe.Offsetof(foo.Sample))
	s += sprintf("Aggregate", unsafe.Sizeof(foo.Aggregate), unsafe.Offsetof(foo.Aggregate))
	s += "\n"

	off0 := unsafe.Offsetof(foo.PCMCounters)
	off := unsafe.Offsetof(foo.PCMCounters.System) + off0
	s += sprintf("PCMCounters.System", unsafe.Sizeof(foo.PCMCounters.System), off)
	s += sprintf("  System.NumOfCores", unsafe.Sizeof(foo.PCMCounters.System.NumOfCores), unsafe.Offsetof(foo.PCMCounters.System.NumOfCores)+off)
	s += sprintf("  System.NumOfOnlineCores", unsafe.Sizeof(foo.PCMCounters.System.NumOfOnlineCores), unsafe.Offsetof(foo.PCMCounters.System.NumOfOnlineCores)+off)
	s += sprintf("  System.NumOfSockets", unsafe.Sizeof(foo.PCMCounters.System.NumOfSockets), unsafe.Offsetof(foo.PCMCounters.System.NumOfSockets)+off)
	s += sprintf("  System.NumOfOnlineSockets", unsafe.Sizeof(foo.PCMCounters.System.NumOfOnlineSockets), unsafe.Offsetof(foo.PCMCounters.System.NumOfOnlineSockets)+off)
	s += sprintf("  System.NumOfAPILinksPerSocket", unsafe.Sizeof(foo.PCMCounters.System.NumOfQPILinksPerSocket), unsafe.Offsetof(foo.PCMCounters.System.NumOfQPILinksPerSocket)+off)
	s += sprintf("  System.CPUModel", unsafe.Sizeof(foo.PCMCounters.System.CPUModel), unsafe.Offsetof(foo.PCMCounters.System.CPUModel)+off)
	s += "\n"

	off = unsafe.Offsetof(foo.PCMCounters.Core) + off0
	s += sprintf("PCMCounters.Core", unsafe.Sizeof(foo.PCMCounters.Core), off)
	off = unsafe.Offsetof(foo.PCMCounters.Core) + off0
	s += sprintf("  SharedPCMCoreCounter", unsafe.Sizeof(foo.PCMCounters.Core.Cores[0]))
	s += sprintf("  Core.cores[0]", unsafe.Sizeof(foo.PCMCounters.Core.Cores), unsafe.Offsetof(foo.PCMCounters.Core.Cores)+off)
	s += sprintf("    cores.CoreID", unsafe.Sizeof(foo.PCMCounters.Core.Cores[0].CoreID), unsafe.Offsetof(foo.PCMCounters.Core.Cores)+off)
	s += sprintf("    cores.SocketID", unsafe.Sizeof(foo.PCMCounters.Core.Cores[0].SocketID), unsafe.Offsetof(foo.PCMCounters.Core.Cores[0].SocketID)+off)
	s += sprintf("    cores.L3CacheOccupancyAvail", unsafe.Sizeof(foo.PCMCounters.Core.Cores[0].L3CacheOccupancyAvailable), unsafe.Offsetof(foo.PCMCounters.Core.Cores[0].L3CacheOccupancyAvailable)+off)
	s += sprintf("    cores.LocalMemoryBWAvail", unsafe.Sizeof(foo.PCMCounters.Core.Cores[0].LocalMemoryBWAvailable), unsafe.Offsetof(foo.PCMCounters.Core.Cores[0].LocalMemoryBWAvailable)+off)
	s += sprintf("    cores.RemoteMemoryBWAvail", unsafe.Sizeof(foo.PCMCounters.Core.Cores[0].RemoteMemoryBWAvailable), unsafe.Offsetof(foo.PCMCounters.Core.Cores[0].RemoteMemoryBWAvailable)+off)
	s += sprintf("    cores.InstructionsPerCycles", unsafe.Sizeof(foo.PCMCounters.Core.Cores[0].InstructionsPerCycles), unsafe.Offsetof(foo.PCMCounters.Core.Cores[0].InstructionsPerCycles)+off)
	s += sprintf("    cores.Cycles", unsafe.Sizeof(foo.PCMCounters.Core.Cores[0].Cycles), unsafe.Offsetof(foo.PCMCounters.Core.Cores[0].Cycles)+off)
	s += sprintf("    cores.InstructionsRetried", unsafe.Sizeof(foo.PCMCounters.Core.Cores[0].InstructionsRetired), unsafe.Offsetof(foo.PCMCounters.Core.Cores[0].InstructionsRetired)+off)
	s += sprintf("    cores.ExecUsage", unsafe.Sizeof(foo.PCMCounters.Core.Cores[0].ExecUsage), unsafe.Offsetof(foo.PCMCounters.Core.Cores[0].ExecUsage)+off)
	s += sprintf("    cores.RelativeFrequency", unsafe.Sizeof(foo.PCMCounters.Core.Cores[0].RelativeFrequency), unsafe.Offsetof(foo.PCMCounters.Core.Cores[0].RelativeFrequency)+off)
	s += sprintf("    cores.ActiveRelativeFrequen", unsafe.Sizeof(foo.PCMCounters.Core.Cores[0].ActiveRelativeFrequency), unsafe.Offsetof(foo.PCMCounters.Core.Cores[0].ActiveRelativeFrequency)+off)
	s += sprintf("    cores.L3CacheMisses", unsafe.Sizeof(foo.PCMCounters.Core.Cores[0].L3CacheMisses), unsafe.Offsetof(foo.PCMCounters.Core.Cores[0].L3CacheMisses)+off)
	s += sprintf("    cores.L3CacheRefernce", unsafe.Sizeof(foo.PCMCounters.Core.Cores[0].L3CacheReference), unsafe.Offsetof(foo.PCMCounters.Core.Cores[0].L3CacheReference)+off)
	s += sprintf("    cores.L2CacheMisses", unsafe.Sizeof(foo.PCMCounters.Core.Cores[0].L2CacheMisses), unsafe.Offsetof(foo.PCMCounters.Core.Cores[0].L2CacheMisses)+off)
	s += sprintf("    cores.L3CacheHitRatio", unsafe.Sizeof(foo.PCMCounters.Core.Cores[0].L3CacheHitRatio), unsafe.Offsetof(foo.PCMCounters.Core.Cores[0].L3CacheHitRatio)+off)
	s += sprintf("    cores.L2CacheHitRatio", unsafe.Sizeof(foo.PCMCounters.Core.Cores[0].L2CacheHitRatio), unsafe.Offsetof(foo.PCMCounters.Core.Cores[0].L2CacheHitRatio)+off)
	s += sprintf("    cores.L3CacheMPI", unsafe.Sizeof(foo.PCMCounters.Core.Cores[0].L3CacheMPI), unsafe.Offsetof(foo.PCMCounters.Core.Cores[0].L3CacheMPI)+off)
	s += sprintf("    cores.L2CacheMPI", unsafe.Sizeof(foo.PCMCounters.Core.Cores[0].L2CacheMPI), unsafe.Offsetof(foo.PCMCounters.Core.Cores[0].L2CacheMPI)+off)
	s += sprintf("    cores.L3CacheOccupancy", unsafe.Sizeof(foo.PCMCounters.Core.Cores[0].L3CacheOccupancy), unsafe.Offsetof(foo.PCMCounters.Core.Cores[0].L3CacheOccupancy)+off)
	s += sprintf("    cores.LocalMemoryBW", unsafe.Sizeof(foo.PCMCounters.Core.Cores[0].LocalMemoryBW), unsafe.Offsetof(foo.PCMCounters.Core.Cores[0].LocalMemoryBW)+off)
	s += sprintf("    cores.RemoteMemoryBW", unsafe.Sizeof(foo.PCMCounters.Core.Cores[0].RemoteMemoryBW), unsafe.Offsetof(foo.PCMCounters.Core.Cores[0].RemoteMemoryBW)+off)
	s += sprintf("    cores.LocalMemoryAccesses", unsafe.Sizeof(foo.PCMCounters.Core.Cores[0].LocalMemoryAccesses), unsafe.Offsetof(foo.PCMCounters.Core.Cores[0].LocalMemoryAccesses)+off)
	s += sprintf("    cores.RemoteMemoryAccesses", unsafe.Sizeof(foo.PCMCounters.Core.Cores[0].RemoteMemoryAccesses), unsafe.Offsetof(foo.PCMCounters.Core.Cores[0].RemoteMemoryAccesses)+off)
	s += sprintf("    cores.ThermalHeadroom", unsafe.Sizeof(foo.PCMCounters.Core.Cores[0].ThermalHeadroom), unsafe.Offsetof(foo.PCMCounters.Core.Cores[0].ThermalHeadroom)+off)
	s += "\n"

	off += unsafe.Sizeof(foo.PCMCounters.Core.Cores[0])
	s += sprintf("  Core.cores[1]", unsafe.Sizeof(foo.PCMCounters.Core.Cores[1]), unsafe.Offsetof(foo.PCMCounters.Core.Cores)+off)
	s += sprintf("    cores.CoreID", unsafe.Sizeof(foo.PCMCounters.Core.Cores[1].CoreID), unsafe.Offsetof(foo.PCMCounters.Core.Cores)+off)
	s += sprintf("    cores.SocketID", unsafe.Sizeof(foo.PCMCounters.Core.Cores[1].SocketID), unsafe.Offsetof(foo.PCMCounters.Core.Cores[1].SocketID)+off)
	s += sprintf("    cores.L3CacheOccupancyAvail", unsafe.Sizeof(foo.PCMCounters.Core.Cores[1].L3CacheOccupancyAvailable), unsafe.Offsetof(foo.PCMCounters.Core.Cores[1].L3CacheOccupancyAvailable)+off)
	s += sprintf("    cores.LocalMemoryBWAvail", unsafe.Sizeof(foo.PCMCounters.Core.Cores[1].LocalMemoryBWAvailable), unsafe.Offsetof(foo.PCMCounters.Core.Cores[1].LocalMemoryBWAvailable)+off)
	s += sprintf("    cores.RemoteMemoryBWAvail", unsafe.Sizeof(foo.PCMCounters.Core.Cores[1].RemoteMemoryBWAvailable), unsafe.Offsetof(foo.PCMCounters.Core.Cores[1].RemoteMemoryBWAvailable)+off)
	s += sprintf("    cores.InstructionsPerCycles", unsafe.Sizeof(foo.PCMCounters.Core.Cores[1].InstructionsPerCycles), unsafe.Offsetof(foo.PCMCounters.Core.Cores[1].InstructionsPerCycles)+off)
	s += sprintf("    cores.Cycles", unsafe.Sizeof(foo.PCMCounters.Core.Cores[1].Cycles), unsafe.Offsetof(foo.PCMCounters.Core.Cores[1].Cycles)+off)
	s += sprintf("    cores.InstructionsRetried", unsafe.Sizeof(foo.PCMCounters.Core.Cores[1].InstructionsRetired), unsafe.Offsetof(foo.PCMCounters.Core.Cores[1].InstructionsRetired)+off)
	s += sprintf("    cores.ExecUsage", unsafe.Sizeof(foo.PCMCounters.Core.Cores[1].ExecUsage), unsafe.Offsetof(foo.PCMCounters.Core.Cores[1].ExecUsage)+off)
	s += sprintf("    cores.RelativeFrequency", unsafe.Sizeof(foo.PCMCounters.Core.Cores[1].RelativeFrequency), unsafe.Offsetof(foo.PCMCounters.Core.Cores[1].RelativeFrequency)+off)
	s += sprintf("    cores.ActiveRelativeFrequen", unsafe.Sizeof(foo.PCMCounters.Core.Cores[1].ActiveRelativeFrequency), unsafe.Offsetof(foo.PCMCounters.Core.Cores[1].ActiveRelativeFrequency)+off)
	s += sprintf("    cores.L3CacheMisses", unsafe.Sizeof(foo.PCMCounters.Core.Cores[1].L3CacheMisses), unsafe.Offsetof(foo.PCMCounters.Core.Cores[1].L3CacheMisses)+off)
	s += sprintf("    cores.L3CacheRefernce", unsafe.Sizeof(foo.PCMCounters.Core.Cores[1].L3CacheReference), unsafe.Offsetof(foo.PCMCounters.Core.Cores[1].L3CacheReference)+off)
	s += sprintf("    cores.L2CacheMisses", unsafe.Sizeof(foo.PCMCounters.Core.Cores[1].L2CacheMisses), unsafe.Offsetof(foo.PCMCounters.Core.Cores[1].L2CacheMisses)+off)
	s += sprintf("    cores.L3CacheHitRatio", unsafe.Sizeof(foo.PCMCounters.Core.Cores[1].L3CacheHitRatio), unsafe.Offsetof(foo.PCMCounters.Core.Cores[1].L3CacheHitRatio)+off)
	s += sprintf("    cores.L2CacheHitRatio", unsafe.Sizeof(foo.PCMCounters.Core.Cores[1].L2CacheHitRatio), unsafe.Offsetof(foo.PCMCounters.Core.Cores[1].L2CacheHitRatio)+off)
	s += sprintf("    cores.L3CacheMPI", unsafe.Sizeof(foo.PCMCounters.Core.Cores[1].L3CacheMPI), unsafe.Offsetof(foo.PCMCounters.Core.Cores[1].L3CacheMPI)+off)
	s += sprintf("    cores.L2CacheMPI", unsafe.Sizeof(foo.PCMCounters.Core.Cores[1].L2CacheMPI), unsafe.Offsetof(foo.PCMCounters.Core.Cores[1].L2CacheMPI)+off)
	s += sprintf("    cores.L3CacheOccupancy", unsafe.Sizeof(foo.PCMCounters.Core.Cores[1].L3CacheOccupancy), unsafe.Offsetof(foo.PCMCounters.Core.Cores[1].L3CacheOccupancy)+off)
	s += sprintf("    cores.LocalMemoryBW", unsafe.Sizeof(foo.PCMCounters.Core.Cores[1].LocalMemoryBW), unsafe.Offsetof(foo.PCMCounters.Core.Cores[1].LocalMemoryBW)+off)
	s += sprintf("    cores.RemoteMemoryBW", unsafe.Sizeof(foo.PCMCounters.Core.Cores[1].RemoteMemoryBW), unsafe.Offsetof(foo.PCMCounters.Core.Cores[1].RemoteMemoryBW)+off)
	s += sprintf("    cores.LocalMemoryAccesses", unsafe.Sizeof(foo.PCMCounters.Core.Cores[1].LocalMemoryAccesses), unsafe.Offsetof(foo.PCMCounters.Core.Cores[1].LocalMemoryAccesses)+off)
	s += sprintf("    cores.RemoteMemoryAccesses", unsafe.Sizeof(foo.PCMCounters.Core.Cores[1].RemoteMemoryAccesses), unsafe.Offsetof(foo.PCMCounters.Core.Cores[1].RemoteMemoryAccesses)+off)
	s += sprintf("    cores.ThermalHeadroom", unsafe.Sizeof(foo.PCMCounters.Core.Cores[1].ThermalHeadroom), unsafe.Offsetof(foo.PCMCounters.Core.Cores[1].ThermalHeadroom)+off)
	off = unsafe.Offsetof(foo.PCMCounters.Core) + off0
	s += sprintf("  Core.packageEnergyMetrisAvail", unsafe.Sizeof(foo.PCMCounters.Core.PackageEnergyMetricsAvailable), unsafe.Offsetof(foo.PCMCounters.Core.PackageEnergyMetricsAvailable)+off)
	s += sprintf("  Core.energyUsedBySockets", unsafe.Sizeof(foo.PCMCounters.Core.EnergyUsedBySockets), unsafe.Offsetof(foo.PCMCounters.Core.EnergyUsedBySockets)+off)
	s += "\n"

	off = unsafe.Offsetof(foo.PCMCounters.Memory) + off0
	s += sprintf("PCMCounters.Memory", unsafe.Sizeof(foo.PCMCounters.Memory), off)
	s += sprintf("  Memory.sockets", unsafe.Sizeof(foo.PCMCounters.Memory.Sockets), unsafe.Offsetof(foo.PCMCounters.Memory.Sockets)+off)
	s += sprintf("  Memory.system", unsafe.Sizeof(foo.PCMCounters.Memory.System), unsafe.Offsetof(foo.PCMCounters.Memory.System)+off)
	s += sprintf("  Memory.dramEnergyMetricsAvail", unsafe.Sizeof(foo.PCMCounters.Memory.DramEnergyMetricsAvailable), unsafe.Offsetof(foo.PCMCounters.Memory.DramEnergyMetricsAvailable)+off)
	s += "\n"

	off = unsafe.Offsetof(foo.PCMCounters.QPI) + off0
	s += sprintf("PCMCounters.QPI", unsafe.Sizeof(foo.PCMCounters.QPI), off)
	s += sprintf("  QPI.Incoming", unsafe.Sizeof(foo.PCMCounters.QPI.Incoming), unsafe.Offsetof(foo.PCMCounters.QPI.Incoming)+off)
	s += sprintf("  QPI.Outgoing", unsafe.Sizeof(foo.PCMCounters.QPI.Outgoing), unsafe.Offsetof(foo.PCMCounters.QPI.Outgoing)+off)
	s += sprintf("  QPI.IncomingTotal", unsafe.Sizeof(foo.PCMCounters.QPI.IncomingTotal), unsafe.Offsetof(foo.PCMCounters.QPI.IncomingTotal)+off)
	s += sprintf("  QPI.OutgoingTotal", unsafe.Sizeof(foo.PCMCounters.QPI.OutgoingTotal), unsafe.Offsetof(foo.PCMCounters.QPI.OutgoingTotal)+off)
	s += sprintf("  QPI.IncomingQPITrafficMetrics", unsafe.Sizeof(foo.PCMCounters.QPI.IncomingQPITrafficMetricsAvailable), unsafe.Offsetof(foo.PCMCounters.QPI.IncomingQPITrafficMetricsAvailable)+off)
	s += sprintf("  QPI.OutgoingQPITrafficMetrics", unsafe.Sizeof(foo.PCMCounters.QPI.OutgoingQPITrafficMetricsAvailable), unsafe.Offsetof(foo.PCMCounters.QPI.OutgoingQPITrafficMetricsAvailable)+off)
	s += "\n"

	off = unsafe.Offsetof(foo.Sample)
	s += sprintf("Sample", unsafe.Sizeof(foo.Sample), off)
	s += sprintf("  Total", unsafe.Sizeof(foo.Sample[0].Total), unsafe.Offsetof(foo.Sample[0].Total)+off)

	tot := foo.Sample[0].Total
	s += sprintf("    .PCIeRdCur", unsafe.Sizeof(tot.ReadCurrent), unsafe.Offsetof(tot.ReadCurrent)+off)
	s += sprintf("    .PCIeNSRd", unsafe.Sizeof(tot.NonSnoopRead), unsafe.Offsetof(tot.NonSnoopRead)+off)
	s += sprintf("    .PCIeWiLF", unsafe.Sizeof(tot.WriteNonAlloc), unsafe.Offsetof(tot.WriteNonAlloc)+off)
	s += sprintf("    .PCIeItoM", unsafe.Sizeof(tot.WriteAlloc), unsafe.Offsetof(tot.WriteAlloc)+off)
	s += sprintf("    .PCIeNSWr", unsafe.Sizeof(tot.NonSnoopWritePart), unsafe.Offsetof(tot.NonSnoopWritePart)+off)
	s += sprintf("    .PCIeNSWrF", unsafe.Sizeof(tot.NonSnoopWriteFull), unsafe.Offsetof(tot.NonSnoopWriteFull)+off)
	s += sprintf("    .RFO", unsafe.Sizeof(tot.ReadForOwnership), unsafe.Offsetof(tot.ReadForOwnership)+off)
	s += sprintf("    .CRd", unsafe.Sizeof(tot.DemandCodeRd), unsafe.Offsetof(tot.DemandCodeRd)+off)
	s += sprintf("    .DRd", unsafe.Sizeof(tot.DemandDataRd), unsafe.Offsetof(tot.DemandDataRd)+off)
	s += sprintf("    .PRd", unsafe.Sizeof(tot.PartialRead), unsafe.Offsetof(tot.PartialRead)+off)
	s += sprintf("    .WiL", unsafe.Sizeof(tot.WriteInvalidateLine), unsafe.Offsetof(tot.WriteInvalidateLine)+off)
	s += sprintf("    .ItoM", unsafe.Sizeof(tot.RequestInvalidateLine), unsafe.Offsetof(tot.RequestInvalidateLine)+off)
	s += sprintf("    .RdBw", unsafe.Sizeof(tot.ReadBandWidth), unsafe.Offsetof(tot.ReadBandWidth)+off)
	s += sprintf("    .WrBw", unsafe.Sizeof(tot.WriteBandWidth), unsafe.Offsetof(tot.WriteBandWidth)+off)
	off = unsafe.Offsetof(foo.Sample)
	s += sprintf("  Miss", unsafe.Sizeof(foo.Sample[0].Miss), unsafe.Offsetof(foo.Sample[0].Miss)+off)
	s += sprintf("  Hit", unsafe.Sizeof(foo.Sample[0].Hit), unsafe.Offsetof(foo.Sample[0].Hit)+off)

	off = unsafe.Offsetof(foo.Aggregate)
	s += sprintf("Aggregate", unsafe.Sizeof(foo.Aggregate), off)
	s += sprintf("Total Size", unsafe.Sizeof(foo))

	return s
}
