// SPDX-License-Identifier: BSD-3-Clause
// Copyright(c) 2019-2020 Intel Corporation

package pcm

// SystemData information
type SystemData struct {
	NumOfCores             uint64 `json:"numOfCores"`
	NumOfOnlineCores       uint64 `json:"numOfOnlineCores"`
	NumOfSockets           uint64 `json:"numOfSockets"`
	NumOfOnlineSockets     uint64 `json:"numOfOnlineSockets"`
	NumOfQPILinksPerSocket uint64 `json:"numOfQPILinksPerSocket"`
	CPUModel               uint64 `json:"cpuModel"`
}

// System command structure
type System struct {
	Data SystemData `json:"/pcm/system"`
}

// CoreCounterData information
type CoreCounterData struct {
	CoreID                    uint64  `json:"coreId"`
	SocketID                  int64   `json:"socketId"`
	L3CacheOccupancyAvailable bool    `json:"l3CacheOccupancyAvailable"`
	LocalMemoryBWAvailable    bool    `json:"localMemoryBWAvailable"`
	RemoteMemoryBWAvailable   bool    `json:"remoteMemoryBWAvailable"`
	InstructionsPerCycle      float64 `json:"instructionsPerCycle"`
	Cycles                    uint64  `json:"cycles"`
	InstructionsRetired       uint64  `json:"instructionsRetired"`
	ExecUsage                 float64 `json:"execUsage"`
	RelativeFrequency         float64 `json:"relativeFrequency"`
	ActiveRelativeFrequency   float64 `json:"activeRelativeFrequency"`
	L3CacheMisses             uint64  `json:"l3CacheMisses"`
	L3CacheReference          uint64  `json:"l3CacheReference"`
	L2CacheMisses             uint64  `json:"l2CacheMisses"`
	L3CacheHitRatio           float64 `json:"l3CacheHitRatio"`
	L2CacheHitRatio           float64 `json:"l2CacheHitRatio"`
	L3CacheMPI                float64 `json:"l3CacheMPI"`
	L2CacheMPI                float64 `json:"l2CacheMPI"`
	L3CacheOccupancy          uint64  `json:"l3CacheOccupancy"`
	LocalMemoryBW             uint64  `json:"localMemoryBW"`
	RemoteMemoryBW            uint64  `json:"remoteMemoryBW"`
	LocalMemoryAccesses       uint64  `json:"localMemoryAccesses"`
	RemoteMemoryAccesses      uint64  `json:"remoteMemoryAccesses"`
	ThermalHeadroom           uint64  `json:"thermalHeadroom"`
}

// CoreCounters data structure
type CoreCounters struct {
	Data CoreCounterData `json:"/pcm/core"`
}

// SocketEnergy information
type SocketEnergy struct {
	PackageEnergyMetricsAvailable bool      `json:"packageEnergyMetricsAvailable"`
	EnergyUsedBySockets           []float64 `json:"energyUsedBySockets"`
}

// MemoryChannelCounter information
type MemoryChannelCounter struct {
	Read  float64 `json:"read"`
	Write float64 `json:"write"`
	Total float64 `json:"total"`
}

// MemorySocketCounter information
type MemorySocketCounter struct {
	SocketID     uint64                 `json:"socketId"`
	Channels     []MemoryChannelCounter `json:"channels"`
	Read         float64                `json:"read"`
	Write        float64                `json:"write"`
	PartialWrite float64                `json:"partialWrite"`
	Total        float64                `json:"total"`
	DramEnergy   float64                `json:"dramEnergy"`
}

// MemorySystemCounter information
type MemorySystemCounter struct {
	Read  float64 `json:"read"`
	Write float64 `json:"write"`
	Total float64 `json:"total"`
}

// SharedPCMMemory information
type SharedPCMMemory struct {
	Sockets                    []MemorySocketCounter `json:"sockets"`
	System                     MemorySystemCounter   `json:"system"`
	DramEnergyMetricsAvailable bool                  `json:"dramEnergyMetricsAvailable"`
}

// QPILinkCounter information
type QPILinkCounter struct {
	LinkID      uint64  `json:"linkId"`
	Bytes       uint64  `json:"bytes"`
	Utilization float64 `json:"utilization"`
}

// QPISocketCounter information
type QPISocketCounter struct {
	SocketID uint64           `json:"socketId"`
	Links    []QPILinkCounter `json:"links"`
	Total    uint64           `json:"total"`
}

// QPI information
type QPI struct {
	Incoming                           []QPISocketCounter `json:"incoming"`
	Outgoing                           []QPISocketCounter `json:"outgoing"`
	IncomingTotal                      uint64             `json:"incomingTotal"`
	OutgoingTotal                      uint64             `json:"outgoingTotal"`
	IncomingQPITrafficMetricsAvailable bool               `json:"incomingQPITrafficMetricsAvailable"`
	OutgoingQPITrafficMetricsAvailable bool               `json:"outgoingQPITrafficMetricsAvailable"`
}

// PCIEvents information data
type PCIEvents struct {
	ReadCurrent       uint64 `json:"PCIeRdCur"`
	NonSnoopRead      uint64 `json:"PCIeNSRd"`
	WriteNonAlloc     uint64 `json:"PCIeWiLF"`
	WriteAlloc        uint64 `json:"PCIeItoM"`
	NonSnoopWritePart uint64 `json:"PCIeNSWr"` // Parial operation
	NonSnoopWriteFull uint64 `json:"PCIeNSWrF"`

	ReadForOwnership      uint64 `json:"RFO"`
	DemandCodeRd          uint64 `json:"CRd"`
	DemandDataRd          uint64 `json:"DRd"`
	PartialRead           uint64 `json:"PRd"`
	WriteInvalidateLine   uint64 `json:"WiL"`  // Partial [MMIO write], PL: Not Documented in HSX/IVT
	RequestInvalidateLine uint64 `json:"ItoM"` // PCIe write full cache line
	ReadBandWidth         uint64 `json:"RdBw"`
	WriteBandWidth        uint64 `json:"WrBw"`
}

// SampleData information
type SampleData struct {
	Total PCIEvents `json:"total"`
	Miss  PCIEvents `json:"miss"`
	Hit   PCIEvents `json:"hit"`
}

// PCIeSampleData information
type PCIeSampleData struct {
	Sockets   map[string]SampleData `json:"sockets"`
	Aggregate PCIEvents             `json:"aggregate"`
}

// HeaderData values in shared memory
type HeaderData struct {
	Version          string `rawlen:"16"`
	TscBegin         uint64 `json:"tscBegin"`
	TscEnd           uint64 `json:"tscEnd"`
	CyclesToGetState uint64 `json:"cyclesToGetPCMState"`
	TimeStamp        uint64 `json:"timestamp"`
	SocketFd         int32  `json:"socketFd"`
	PollMs           uint32 `json:"pollMs"`
}

// Header struct for the PCM data
type Header struct {
	Data HeaderData `json:"/pcm/header"`
}

// CPU Model IDs
const (
	NehalemEPModel       = 26
	NehalemModel         = 30
	AtomModel            = 28
	Atom2Model           = 53
	AtomCentertonModel   = 54
	AtomBaytrailModel    = 55
	AtomAvotonModel      = 77
	AtomCherrytrailModel = 76
	AtomApolloLakeModel  = 92
	AtomDenvertonModel   = 95
	ClarkdaleModel       = 37
	WestmereEPModel      = 44
	NehalemEXModel       = 46
	WestmereEXModel      = 47
	SandyBridgeModel     = 42
	JaketownModel        = 45
	IvyBridgeModel       = 58
	HaswellModel         = 60
	HaswellULTModel      = 69
	Haswell2MNodel       = 70
	IvytownModel         = 62
	HaswellxModel        = 63
	BroadwellModel       = 61
	BroadwellXeonE3Model = 71
	BDXDEModel           = 86
	SklUYMdel            = 78
	KBLModel             = 158
	KBL1Model            = 142
	BDXModel             = 79
	KNLModel             = 87
	SKLModel             = 94
	SKXModel             = 85
)

// CPUModels is a list of known Intel CPU Ids
var CPUModels = map[int]string{
	NehalemEPModel:       "Nehalem EP",
	NehalemModel:         "Nehalem",
	AtomModel:            "Atom",
	Atom2Model:           "Atom2",
	AtomCentertonModel:   "Atom Centerton",
	AtomBaytrailModel:    "Atom Baytrail",
	AtomAvotonModel:      "Atom Avoton",
	AtomCherrytrailModel: "Atom Cherrytrail",
	AtomApolloLakeModel:  "Atom Apollo Lake",
	AtomDenvertonModel:   "Atom Denverton",
	ClarkdaleModel:       "Clarkdale",
	WestmereEPModel:      "Westmere EP",
	NehalemEXModel:       "Nehalem EX",
	WestmereEXModel:      "Westmere EX",
	SandyBridgeModel:     "Sandy Bridge",
	JaketownModel:        "Jaketown",
	IvyBridgeModel:       "Ivy Bridge",
	HaswellModel:         "Haswell",
	HaswellULTModel:      "Haswell ULT",
	Haswell2MNodel:       "Haswell 2",
	IvytownModel:         "Ivytown",
	HaswellxModel:        "Hasewell X",
	BroadwellModel:       "Broadwell",
	BroadwellXeonE3Model: "Broadwell Xeon E3",
	BDXDEModel:           "BDX DE",
	SklUYMdel:            "SKL UY",
	KBLModel:             "KBL",
	KBL1Model:            "KBL 1",
	BDXModel:             "Broadwell",
	KNLModel:             "KNL",
	SKLModel:             "Skylake",
	SKXModel:             "SKX",
}
