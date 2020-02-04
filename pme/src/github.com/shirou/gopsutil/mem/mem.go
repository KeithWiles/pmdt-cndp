package mem

import (
	"encoding/json"

	"github.com/shirou/gopsutil/internal/common"
)

var invoke common.Invoker = common.Invoke{}

// VirtualMemoryPerSocketStat information
type VirtualMemoryPerSocketStat struct {
	Total			uint64		`json:"total"`
	Free			uint64		`json:"free"`
	Used			uint64		`json:"used"`
	Active			uint64		`json:"active"`
	Inactive		uint64		`json:"inactive"`
	ActiveAnon		uint64		`json:"activeanon"`
	InactiveAnon	uint64		`json:"inactiveannon"`
	ActiveFile		uint64		`json:"activefile"`
	InactiveFile	uint64		`json:"inactivefile"`
	Unevictable		uint64		`json:"unevictable"`
	Mlocked			uint64		`json:"mlocked"`
	Dirty			uint64		`json:"dirty"`
	Writeback		uint64		`json:"writeback"`
	FilePages		uint64		`json:"filepages"`
	Mapped			uint64		`json:"mapped"`
	AnonPages		uint64		`json:"anonpages"`
	Shmem			uint64		`json:"shmem"`
	KernelStack		uint64		`json:"kernelstack"`
	PageTables		uint64		`json:"pagetables"`
	NFSUnstable		uint64		`json:"nfsunsable"`
	Bounce			uint64		`json:"bounce"`
	WritebackTmp	uint64		`json:"writebacktmp"`
	KReclaimable	uint64		`json:"kreclaimble"`
	Slab			uint64		`json:"slab"`
	SReclaimable	uint64		`json:"srecliamable"`
	SUnreclaim		uint64		`json:"sunreclaim"`
	AnonHugePages	uint64		`json:"anonhugepages"`
	ShmemHugePages	uint64		`json:"shmemhugepages"`
	ShmemPmdMapped	uint64		`json:"shmempmdmapped"`
	HugePagesTotal	uint64		`json:"hugepagestotal"`
	HugePagesFree	uint64		`json:"hugepagesfree"`
	HugePagesSurp	uint64		`json:"hugepages_surp"`
}

// VirtualMemoryStat should have a comment
// Memory usage statistics. Total, Available and Used contain numbers of bytes
// for human consumption.
//
// The other fields in this struct contain kernel specific values.
type VirtualMemoryStat struct {
	// Total amount of RAM on this system
	Total uint64 `json:"total"`

	// RAM available for programs to allocate
	//
	// This value is computed from the kernel specific values.
	Available uint64 `json:"available"`

	// RAM used by programs
	//
	// This value is computed from the kernel specific values.
	Used uint64 `json:"used"`

	// Percentage of RAM used by programs
	//
	// This value is computed from the kernel specific values.
	UsedPercent float64 `json:"usedPercent"`

	// This is the kernel's notion of free memory; RAM chips whose bits nobody
	// cares about the value of right now. For a human consumable number,
	// Available is what you really want.
	Free uint64 `json:"free"`

	// OS X / BSD specific numbers:
	// http://www.macyourself.com/2010/02/17/what-is-free-wired-active-and-inactive-system-memory-ram/
	Active   uint64 `json:"active"`
	Inactive uint64 `json:"inactive"`
	Wired    uint64 `json:"wired"`

	// FreeBSD specific numbers:
	// https://reviews.freebsd.org/D8467
	Laundry uint64 `json:"laundry"`

	// Linux specific numbers
	// https://www.centos.org/docs/5/html/5.1/Deployment_Guide/s2-proc-meminfo.html
	// https://www.kernel.org/doc/Documentation/filesystems/proc.txt
	// https://www.kernel.org/doc/Documentation/vm/overcommit-accounting
	Buffers        uint64 `json:"buffers"`
	Cached         uint64 `json:"cached"`
	Writeback      uint64 `json:"writeback"`
	Dirty          uint64 `json:"dirty"`
	WritebackTmp   uint64 `json:"writebacktmp"`
	Shared         uint64 `json:"shared"`
	Slab           uint64 `json:"slab"`
	SReclaimable   uint64 `json:"sreclaimable"`
	SUnreclaim     uint64 `json:"sunreclaim"`
	PageTables     uint64 `json:"pagetables"`
	SwapCached     uint64 `json:"swapcached"`
	CommitLimit    uint64 `json:"commitlimit"`
	CommittedAS    uint64 `json:"committedas"`
	HighTotal      uint64 `json:"hightotal"`
	HighFree       uint64 `json:"highfree"`
	LowTotal       uint64 `json:"lowtotal"`
	LowFree        uint64 `json:"lowfree"`
	SwapTotal      uint64 `json:"swaptotal"`
	SwapFree       uint64 `json:"swapfree"`
	Mapped         uint64 `json:"mapped"`
	VMallocTotal   uint64 `json:"vmalloctotal"`
	VMallocUsed    uint64 `json:"vmallocused"`
	VMallocChunk   uint64 `json:"vmallocchunk"`
	HugePagesTotal uint64 `json:"hugepagestotal"`
	HugePagesFree  uint64 `json:"hugepagesfree"`
	HugePageSize   uint64 `json:"hugepagesize"`
	HugePagesRsvd  uint64 `json:"hugepages_rsvd"`
	HugePagesSurp  uint64 `json:"hugepages_surp"`
}

// SwapMemoryStat should have a comment
type SwapMemoryStat struct {
	Total       uint64  `json:"total"`
	Used        uint64  `json:"used"`
	Free        uint64  `json:"free"`
	UsedPercent float64 `json:"usedPercent"`
	Sin         uint64  `json:"sin"`
	Sout        uint64  `json:"sout"`
	PgIn        uint64  `json:"pgin"`
	PgOut       uint64  `json:"pgout"`
	PgFault     uint64  `json:"pgfault"`
}

// String should have a comment
func (m VirtualMemoryStat) String() string {
	s, _ := json.Marshal(m)
	return string(s)
}

// String should have a comment
func (m SwapMemoryStat) String() string {
	s, _ := json.Marshal(m)
	return string(s)
}
