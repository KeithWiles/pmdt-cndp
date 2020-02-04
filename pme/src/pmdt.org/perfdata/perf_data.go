// SPDX-License-Identifier: BSD-3-Clause
// Copyright(c) 2019-2020 Intel Corporation

package perfdata

import (
	"fmt"
	"os"
	"sync"

	"github.com/shirou/gopsutil/cpu"
	"golang.org/x/sys/unix"

	"pmdt.org/jevents"
	tlog "pmdt.org/ttylog"
)

// CPUInfo per CPU
type CPUInfo struct {
	Enabled bool
	CoreID  int
	Scaled  uint64
}

// CPUInfos of CPU data
type CPUInfos []CPUInfo

// PerfEventData - a list of  information with event name
type PerfEventData struct {
	Name         string   // Event name
	Sum          uint64   // Sum of all values for this event on all CPUs
	CPUsPerEvent CPUInfos // List of CPUs for this event
}

type ipcPerCPU struct {
	clksAny float64
	retired float64
	ipc     float64
}

// EventDataList a list of Events structures
type EventDataList []PerfEventData

// Info - Data for main page information
type Info struct {
	lock sync.Mutex

	opened  bool
	running bool

	nCPUs     int                  // Total number of CPUs in the system
	perfData  []*jevents.PerfEvent // Perf data for each event
	eventList EventDataList        // Slice of event data for each event
	ipcs      []ipcPerCPU
}

var perfInfo *Info

// Perf Cache access type
const (
	HwCacheReadAccess     uint64 = (((unix.PERF_COUNT_HW_CACHE_OP_READ) << 8) | ((unix.PERF_COUNT_HW_CACHE_RESULT_ACCESS) << 16))
	HwCacheWriteAccess           = (((unix.PERF_COUNT_HW_CACHE_OP_WRITE) << 8) | ((unix.PERF_COUNT_HW_CACHE_RESULT_ACCESS) << 16))
	HwCachePrefetchAccess        = (((unix.PERF_COUNT_HW_CACHE_OP_PREFETCH) << 8) | ((unix.PERF_COUNT_HW_CACHE_RESULT_ACCESS) << 16))
	HwCacheReadMiss              = (((unix.PERF_COUNT_HW_CACHE_OP_READ) << 8) | ((unix.PERF_COUNT_HW_CACHE_RESULT_MISS) << 16))
	HwCacheWriteMiss             = (((unix.PERF_COUNT_HW_CACHE_OP_WRITE) << 8) | ((unix.PERF_COUNT_HW_CACHE_RESULT_MISS) << 16))
	HwCachePrefetchMiss          = (((unix.PERF_COUNT_HW_CACHE_OP_PREFETCH) << 8) | ((unix.PERF_COUNT_HW_CACHE_RESULT_MISS) << 16))
)

type eventInfo struct {
	name   string
	config uint64
}

type eventInfos []eventInfo

// EventTypes map to hold known events
var EventTypes map[string]eventInfos
var numberCPUs int

// numCPUs is the number of CPUs in the system (logical cores)
func numCPUs() int {
	var once sync.Once

	once.Do(func() {
		num, err := cpu.Counts(true)
		if err != nil {
			os.Exit(1)
		}
		numberCPUs = num
	})

	return numberCPUs
}

// Perf constants
func init() {
	EventTypes = make(map[string]eventInfos)

	EventTypes["KernelPmuEvents"] = []eventInfo{
		{name: "cpu-cycles", config: unix.PERF_COUNT_HW_CPU_CYCLES},
		{name: "instructions", config: unix.PERF_COUNT_HW_INSTRUCTIONS},
		{name: "cache-references", config: unix.PERF_COUNT_HW_CACHE_REFERENCES},
		{name: "cache-misses", config: unix.PERF_COUNT_HW_CACHE_MISSES},
		{name: "branches", config: unix.PERF_COUNT_HW_BRANCH_INSTRUCTIONS},
		{name: "branch-misses", config: unix.PERF_COUNT_HW_BRANCH_MISSES},
		{name: "bus-cycles", config: unix.PERF_COUNT_HW_BUS_CYCLES},
	}

	EventTypes["HwCacheEvents"] = []eventInfo{
		{name: "L1-dcache-loads", config: (unix.PERF_COUNT_HW_CACHE_L1D | HwCacheReadAccess)},
		{name: "L1-dcache-load-misses", config: (unix.PERF_COUNT_HW_CACHE_L1D | HwCacheReadMiss)},
		{name: "L1-dcache-stores", config: (unix.PERF_COUNT_HW_CACHE_L1D | HwCacheWriteAccess)},
		{name: "L1-dcache-store-misses", config: (unix.PERF_COUNT_HW_CACHE_L1D | HwCacheWriteMiss)},
		{name: "L1-dcache-prefetches", config: (unix.PERF_COUNT_HW_CACHE_L1D | HwCachePrefetchAccess)},
		{name: "L1-dcache-prefetch-misses", config: (unix.PERF_COUNT_HW_CACHE_L1D | HwCachePrefetchMiss)},

		{name: "L1-icache-loads", config: (unix.PERF_COUNT_HW_CACHE_L1I | HwCacheReadAccess)},
		{name: "L1-icache-load-misses", config: (unix.PERF_COUNT_HW_CACHE_L1I | HwCacheReadMiss)},
		{name: "L1-icache-prefetches", config: (unix.PERF_COUNT_HW_CACHE_L1I | HwCachePrefetchAccess)},
		{name: "L1-icache-prefetch-misses", config: (unix.PERF_COUNT_HW_CACHE_L1I | HwCachePrefetchMiss)},

		{name: "LLC-loads", config: (unix.PERF_COUNT_HW_CACHE_LL | HwCacheReadAccess)},
		{name: "LLC-load-misses", config: (unix.PERF_COUNT_HW_CACHE_LL | HwCacheReadMiss)},
		{name: "LLC-stores", config: (unix.PERF_COUNT_HW_CACHE_LL | HwCacheWriteAccess)},
		{name: "LLC-store-misses", config: (unix.PERF_COUNT_HW_CACHE_LL | HwCacheWriteMiss)},
		{name: "LLC-prefetches", config: (unix.PERF_COUNT_HW_CACHE_LL | HwCachePrefetchAccess)},
		{name: "LLC-prefetch-misses", config: (unix.PERF_COUNT_HW_CACHE_LL | HwCachePrefetchMiss)},

		{name: "dTLB-loads", config: (unix.PERF_COUNT_HW_CACHE_DTLB | HwCacheReadAccess)},
		{name: "dTLB-load-misses", config: (unix.PERF_COUNT_HW_CACHE_DTLB | HwCacheReadMiss)},
		{name: "dTLB-stores", config: (unix.PERF_COUNT_HW_CACHE_DTLB | HwCacheWriteAccess)},
		{name: "dTLB-store-misses", config: (unix.PERF_COUNT_HW_CACHE_DTLB | HwCacheWriteMiss)},
		{name: "dTLB-prefetches", config: (unix.PERF_COUNT_HW_CACHE_DTLB | HwCachePrefetchAccess)},
		{name: "dTLB-prefetch-misses", config: (unix.PERF_COUNT_HW_CACHE_DTLB | HwCachePrefetchMiss)},

		{name: "iTLB-loads", config: (unix.PERF_COUNT_HW_CACHE_ITLB | HwCacheReadAccess)},
		{name: "iTLB-load-misses", config: (unix.PERF_COUNT_HW_CACHE_ITLB | HwCacheReadMiss)},

		{name: "branch-loads", config: (unix.PERF_COUNT_HW_CACHE_BPU | HwCacheReadAccess)},
		{name: "branch-load-misses", config: (unix.PERF_COUNT_HW_CACHE_BPU | HwCacheReadMiss)},
	}

	EventTypes["SwEvents"] = []eventInfo{
		{name: "cpu-clock", config: unix.PERF_COUNT_SW_CPU_CLOCK},
		{name: "task-clock", config: unix.PERF_COUNT_SW_TASK_CLOCK},
		{name: "context-switches", config: unix.PERF_COUNT_SW_CONTEXT_SWITCHES},
		{name: "cpu-migrations", config: unix.PERF_COUNT_SW_CPU_MIGRATIONS},
		{name: "page-faults", config: unix.PERF_COUNT_SW_PAGE_FAULTS},
		{name: "minor-faults", config: unix.PERF_COUNT_SW_PAGE_FAULTS_MIN},
		{name: "major-faults", config: unix.PERF_COUNT_SW_PAGE_FAULTS_MAJ},
		{name: "alignment-faults", config: unix.PERF_COUNT_SW_ALIGNMENT_FAULTS},
		{name: "emulation-faults", config: unix.PERF_COUNT_SW_EMULATION_FAULTS},
	}

	tlog.Register("PerfDataLog", false)

	perfInfo = &Info{nCPUs: numCPUs(), opened: false, running: false}
}

// String to return CPUInfo
func (cpu CPUInfo) String() (str string) {
	return fmt.Sprintf("Enabled %t, CoreID: %d", cpu.Enabled, cpu.CoreID)
}

// String - return the event data as a string from Dataset
func (cpus CPUInfos) String() (str string) {

	for _, v := range cpus {
		str += fmt.Sprintf("%s", v)
	}

	return str
}

// String - return the event data as a string from Dataset
func (d PerfEventData) String() (str string) {

	str = fmt.Sprintf(">Name:%s Sum:%d", d.Name, d.Sum)
	str += fmt.Sprintf("%s", d.CPUsPerEvent)

	return str
}

// String - return the event data as a string from Dataset
func (list EventDataList) String() (str string) {

	for _, ds := range list {
		str += fmt.Sprintf("%s\n", ds)
	}

	return str
}

// EventString - return the event data as a string
func EventString(events EventDataList) (str string) {
	return fmt.Sprintf("%s\n", events)
}

// Open based on the string of events
// eventStr is the semi-comma seperated list of events
func Open(eventStr string, all bool, pid int) error {

	pd := perfInfo

	tlog.DebugPrintf("Open events called %+v\n", eventStr)

	// Close the current events if opened
	if pd.opened {
		Close()

		pd.perfData = nil
	}

	perfData, err := jevents.Parse(eventStr)
	if err != nil {
		return fmt.Errorf("perf data is nil")
	}

	pd.perfData = perfData
	pd.eventList = make(EventDataList, len(perfData))
	pd.ipcs = make([]ipcPerCPU, pd.nCPUs)

	if err := jevents.Open(pd.perfData, all, pid); err != nil {
		tlog.ErrorPrintf("jevents.Open() failed: %s\n", err)
		return err
	}
	pd.opened = true

	return nil
}

// Close by calling jevents.Close to close the file descriptors
func Close() error {

	pd := perfInfo

	if !pd.opened {
		return nil
	}

	if pd.running {
		if err := Stop(); err != nil {
			return err
		}
	}

	if err := jevents.Close(pd.perfData); err != nil {
		return err
	}
	pd.opened = false

	return nil
}

// Start running to measure events
func Start() error {

	pd := perfInfo

	if pd.running {
		tlog.DebugPrintf("perfdata.Start() already running\n")
		return nil
	}
	if err := jevents.Start(pd.perfData); err != nil {
		tlog.ErrorPrintf("jevents.Start() failed: %s\n", err)
		return err
	}
	pd.running = true

	return nil
}

// Stop running
func Stop() error {

	pd := perfInfo

	if !pd.running {
		return nil
	}
	if err := jevents.Stop(pd.perfData); err != nil {
		return err
	}
	pd.running = false

	return nil
}

// Collect - grad the event data
func Collect() (err error) {

	pd := perfInfo

	if pd.opened == false {
		return fmt.Errorf("events not opened")
	}

	if pd.perfData == nil {
		return fmt.Errorf("perfData is nil")
	}

	// Collect the event data
	for k, e := range pd.perfData {
		pEvent := &pd.eventList[k]

		pEvent.Name = e.EventName
		pEvent.CPUsPerEvent = make(CPUInfos, pd.nCPUs)

		sum := uint64(0)
		for i, ce := range e.CPUEvents {
			cpe := &pEvent.CPUsPerEvent[i]

			cpe.Enabled = true
			cpe.CoreID = i
			cpe.Scaled = ce.Scaled

			if pEvent.Name == "cpu_clk_unhalted.thread_any" {
				pd.ipcs[i].clksAny = float64(ce.Scaled)
			} else if pEvent.Name == "inst_retired.any" {
				pd.ipcs[i].retired = float64(ce.Scaled)
			}
			sum += ce.Scaled
		}

		pEvent.Sum = sum
	}

	// calculate the IPC value
	for _, ev := range pd.eventList {
		for _, cpe := range ev.CPUsPerEvent {
			p := &pd.ipcs[cpe.CoreID]
			p.ipc = 0.0
			if p.clksAny > 0.0 && (p.clksAny/2) > 0.0 {
				p.ipc = (p.retired / (p.clksAny / 2))
			}
		}
	}

	return nil
}

// CPUScaledValue value from event and cpu
func CPUScaledValue(eIdx, cpuID int) uint64 {

	pd := perfInfo

	if pd.eventList == nil ||
		eIdx > len(pd.eventList) ||
		pd.eventList[eIdx].CPUsPerEvent == nil ||
		cpuID > len(pd.eventList[eIdx].CPUsPerEvent) {
		return 0
	}
	return pd.eventList[eIdx].CPUsPerEvent[cpuID].Scaled
}

// AllIPCs slice for all CPUs
func AllIPCs() []float64 {

	pd := perfInfo

	ipcs := []float64{}

	if pd.ipcs == nil {
		ipcs = nil
	} else {
		for _, v := range pd.ipcs {
			ipcs = append(ipcs, v.ipc)
		}
	}

	return ipcs
}

// EventList returns the list of events
func EventList() EventDataList {

	return perfInfo.eventList
}

// PerfData returns the perfData slice
func PerfData() []*jevents.PerfEvent {

	return perfInfo.perfData
}

// Opened flag is returned
func Opened() bool {
	return perfInfo.opened
}

// Running flag is returned
func Running() bool {
	return perfInfo.running
}
