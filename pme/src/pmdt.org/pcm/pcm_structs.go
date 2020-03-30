// SPDX-License-Identifier: BSD-3-Clause
// Copyright(c) 2019-2020 Intel Corporation

package pcm

import ()

// PCM system constants
const (
	Version              string = "1.0.6"
	MaxCPUCores          int    = 256
	MaxSockets           int    = 8
	QPIMaxLinks          int    = 32 // (MaxSockets * 4)
	MemoryMaxIMCChannels int    = 8

	MemoryRead        int = 0
	MemoryWrite       int = 1
	MemoryReadRankA   int = 0
	MemoryWriteRankA  int = 1
	MemoryReadRankB   int = 2
	MemoryWwriteRankB int = 3
	MemoryPartial     int = 2

	VersionSize int = 12
)

// SharedPCMSystem information
type SharedPCMSystem struct {
	NumOfCores             uint64
	NumOfOnlineCores       uint64
	NumOfSockets           uint64
	NumOfOnlineSockets     uint64
	NumOfQPILinksPerSocket uint64
	CPUModel               uint64
	System0                [2]uint64
}

// SharedPCMCoreCounter information
type SharedPCMCoreCounter struct {
	CoreID                    uint64
	SocketID                  int64
	L3CacheOccupancyAvailable bool
	LocalMemoryBWAvailable    bool
	RemoteMemoryBWAvailable   bool
	Filler0                   [5]byte
	InstructionsPerCycles     float64
	Cycles                    uint64
	InstructionsRetired       uint64
	ExecUsage                 float64
	RelativeFrequency         float64
	ActiveRelativeFrequency   float64
	L3CacheMisses             uint64
	L3CacheReference          uint64
	L2CacheMisses             uint64
	L3CacheHitRatio           float64
	L2CacheHitRatio           float64
	L3CacheMPI                float64
	L2CacheMPI                float64
	L3CacheOccupancy          uint64
	LocalMemoryBW             uint64
	RemoteMemoryBW            uint64
	LocalMemoryAccesses       uint64
	RemoteMemoryAccesses      uint64
	ThermalHeadroom           uint64
	Filler1                   [2]uint64
}

// SharedPCMCore information
type SharedPCMCore struct {
	Cores                         [MaxCPUCores]SharedPCMCoreCounter
	PackageEnergyMetricsAvailable bool
	Filler0                       [7]byte
	Filler7                       [7]uint64
	EnergyUsedBySockets           [MaxSockets]float64
}

// SharedPCMMemoryChannelCounter information
type SharedPCMMemoryChannelCounter struct {
	Read            float64
	Write           float64
	Total           float64
	MemChnlCounter0 [5]uint64
}

// SharedPCMMemorySocketCounter information
type SharedPCMMemorySocketCounter struct {
	SocketID     uint64
	MemSockCntr0 [7]uint64
	Channels     [MemoryMaxIMCChannels]SharedPCMMemoryChannelCounter
	Read         float64
	Write        float64
	PartialWrite float64
	Total        float64
	DramEnergy   float64
	MemSockCntr1 [3]uint64
}

// SharedPCMMemorySystemCounter information
type SharedPCMMemorySystemCounter struct {
	Read        float64
	Write       float64
	Total       float64
	MemSysCntr0 [5]uint64
}

// SharedPCMMemory information
type SharedPCMMemory struct {
	Sockets                    [MaxSockets]SharedPCMMemorySocketCounter
	System                     SharedPCMMemorySystemCounter
	DramEnergyMetricsAvailable bool
	Mem1                       bool
	Mem2                       bool
	Mem3                       bool
	Mem4                       bool
	Mem5                       bool
	Mem6                       bool
	Mem7                       bool
	Memory0                    [7]uint64
}

// SharedPCMQPILinkCounter information
type SharedPCMQPILinkCounter struct {
	Bytes       uint64
	Utilization float64
	QPILink0    [6]uint64
}

// SharedPCMQPISocketCounter information
type SharedPCMQPISocketCounter struct {
	SocketID uint64
	QPISock0 [7]uint64
	Links    [QPIMaxLinks]SharedPCMQPILinkCounter
	Total    uint64
	QPISock1 [7]uint64
}

// SharedPCMQPI information
type SharedPCMQPI struct {
	Incoming                           [MaxSockets]SharedPCMQPISocketCounter
	Outgoing                           [MaxSockets]SharedPCMQPISocketCounter
	IncomingTotal                      uint64
	OutgoingTotal                      uint64
	IncomingQPITrafficMetricsAvailable bool
	OutgoingQPITrafficMetricsAvailable bool
	QPIFiller0                         bool
	QPIFiller1                         bool
	QPIFiller2                         bool
	QPIFiller3                         bool
	QPIFiller4                         bool
	QPIFiller5                         bool
	QPICntr1                           [5]uint64
}

// PCIEvents information data
type PCIEvents struct {
	ReadCurrent           uint64	`json:"PCIeRdCur"`
	NonSnoopRead          uint64	`json:"PCIeNSRd"`
	WriteNonAlloc         uint64	`json:"PCIeWiLF"`
	WriteAlloc            uint64	`json:"PCIeItoM"`
	NonSnoopWritePart     uint64	`json:"PCIeNSWr"`	// Parial operation
	NonSnoopWriteFull     uint64	`json:"PCIeNSWrF"`

	ReadForOwnership      uint64	`json:"RFO"`
	DemandCodeRd          uint64	`json:"CRd"`
	DemandDataRd          uint64	`json:"DRd"`
	PartialRead           uint64	`json:"PRd"`
	WriteInvalidateLine   uint64	`json:"WiL"`	// Partial [MMIO write], PL: Not Documented in HSX/IVT
	RequestInvalidateLine uint64	`json:"ItoM"`	// PCIe write full cache line
	ReadBandWidth         uint64	`json:"RdBw"`
	WriteBandWidth        uint64	`json:"WrBw"`
}

// SampleData information
type SampleData struct {
	Total PCIEvents
	Miss  PCIEvents
	Hit   PCIEvents
}

// PCIeSampleData information
type PCIeSampleData struct {
	Sockets map[string]SampleData
	Aggregate   PCIEvents
}

// SharedPCMCounters data region
type SharedPCMCounters struct {
	System SharedPCMSystem
	Core   SharedPCMCore
	Memory SharedPCMMemory
	QPI    SharedPCMQPI
}

// SharedHeader values in shared memory
type SharedHeader struct {
	Version          string `rawlen:"16"`
	TscBegin         uint64
	TscEnd           uint64
	CyclesToGetState uint64
	TimeStamp        uint64
	PollMs           uint32
	DelayMs          uint32
	Header0          uint64
}

// SharedPCMState data in shared memory
type SharedPCMState struct {
	Header      SharedHeader
	PCMCounters SharedPCMCounters
	Sample      [MaxSockets]SampleData
	Aggregate   PCIEvents
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
	BDXModel:             "BDX",
	KNLModel:             "KNL",
	SKLModel:             "SKL",
	SKXModel:             "SKX",
}

// CPUModel string name
func CPUModel(id int) string {
	v, ok := CPUModels[id]
	if !ok {
		return ""
	}
	return v
}
