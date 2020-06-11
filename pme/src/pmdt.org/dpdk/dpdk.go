// SPDX-License-Identifier: BSD-3-Clause
// Copyright(c) 2019-2020 Intel Corporation

package dpdk

const (
	// MaxPortCount for the number of ports
	MaxPortCount int = 16
)

// EthdevPortStats - port stats
type EthdevPortStats struct {
	PortID     uint16
	InPackets  uint64 `json:"rx_good_packets"`
	OutPackets uint64 `json:"tx_good_packets"`
	InBytes    uint64 `json:"rx_good_bytes"`
	OutBytes   uint64 `json:"tx_good_bytes"`
	InMissed   uint64 `json:"rx_missed_errors"`
	InErrors   uint64 `json:"rx_errors"`
	OutErrors  uint64 `json:"tx_errors"`
	RxNomBuf   uint64 `json:"rx_mbuf_allocation_errors"`
	/*
		RxQ0Packets uint64	  `json:"rx_q0packets"`
		RxQ0Bytes   uint64    `json:"rx_q0bytes"`
		RxQ0Errors  uint64    `json:"rx_q0errors"`
		TxQ0Packets uint64    `json:"tx_q0packets"`
		TxQ0Bytes   uint64    `json:"tx_q0bytes"`
		RxUnicastPackets uint64 `json:"rx_unicast_packets"`
		RxMulticastPackets uint64 `json:"rx_multicast_packets"`
		RxBroadcastPackets uint64 `json:"rx_broadcast_packets"`
		RxDroppedPackets uint64 `json:"rx_dropped_packets"`
		RxUnknownProtocolPackets uint64 `json:"rx_unknown_protocol_packets"`
		TxUnicastPackets uint64 `json:"tx_unicast_packets"`
		TxMulticastPackets uint64 `json:"tx_multicast_packets"`
		TxBroadcastPackets uint64 `json:"tx_broadcast_packets"`
		TxDroppedPackets uint64 `json:"tx_dropped_packets"`
		TxLinkDownDroppped uint64 `json:"tx_link_down_dropped"`
		RxCRCErrors uint64 `json:"rx_crc_errors"`
		RxIllegalByteErrors uint64 `json:"rx_illegal_byte_errors"`
		RxErrorBytes uint64 `json:"rx_error_bytes"`
		MacLocalErrors uint64 `json:"mac_local_errors"`
		MacRemoteErrors uint64 `json:"mac_remote_errors"`
		RxLengthErrors uint64 `json:"rx_length_errors"`
		TxXonPackets uint64 `json:"tx_xon_packets"`
		RxXonPackets uint64 `json:"rx_xon_packets"`
		TxXoffPackets uint64 `json:"tx_xoff_packets"`
		RxXoffPackets uint64 `json:"rx_xoff_packets"`
		RxSize64Packets uint64 `json:"rx_size_64_packets"`
		RxSize65To127Packets uint64 `json:"rx_size_65_to_127_packets"`
		RxSize128To255Packets uint64 `json:"rx_size_128_to_255_packets"`
		RxSize256To511Packets uint64 `json:"rx_size_256_to_511_packets"`
		RxSize512To1023Packets uint64 `json:"rx_size_512_to_1023_packets"`
		RxSize1024To1522Packets uint64 `json:"rx_size_1024_to_1522_packets"`

		uint64 `json:"rx_size_1523_to_max_packets"`
		uint64 `json:"rx_undersized_errors"`
		uint64 `json:"rx_oversize_errors"`
		uint64 `json:"rx_mac_short_dropped"`
		uint64 `json:"rx_fragmented_errors"`
		uint64 `json:"rx_jabber_errors"`
		uint64 `json:"tx_size_64_packets"`
		uint64 `json:"tx_size_65_to_127_packets"`
		uint64 `json:"tx_size_128_to_255_packets"`
		uint64 `json:"tx_size_256_to_511_packets"`
		uint64 `json:"tx_size_512_to_1023_packets"`
		uint64 `json:"tx_size_1024_to_1522_packets"`
		uint64 `json:"tx_size_1523_to_max_packets"`
		uint64 `json:"rx_flow_director_atr_match_packets"`
		uint64 `json:"rx_flow_director_sb_match_packets"`
		uint64 `json:"tx_low_power_idle_status"`
		uint64 `json:"rx_low_power_idle_status"`
		uint64 `json:"tx_low_power_idle_count"`
		uint64 `json:"rx_low_power_idle_count"`
		uint64 `json:"rx_priority0_xon_packets"`
		uint64 `json:"rx_priority1_xon_packets"`
		uint64 `json:"rx_priority2_xon_packets"`
		uint64 `json:"rx_priority3_xon_packets"`
		uint64 `json:"rx_priority4_xon_packets"`
		uint64 `json:"rx_priority5_xon_packets"`
		uint64 `json:"rx_priority6_xon_packets"`
		uint64 `json:"rx_priority7_xon_packets"`
		uint64 `json:"rx_priority0_xoff_packets"`
		uint64 `json:"rx_priority1_xoff_packets"`
		uint64 `json:"rx_priority2_xoff_packets"`
		uint64 `json:"rx_priority3_xoff_packets"`
		uint64 `json:"rx_priority4_xoff_packets"`
		uint64 `json:"rx_priority5_xoff_packets"`
		uint64 `json:"rx_priority6_xoff_packets"`
		uint64 `json:"rx_priority7_xoff_packets"`
		uint64 `json:"tx_priority0_xon_packets"`
		uint64 `json:"tx_priority1_xon_packets"`
		uint64 `json:"tx_priority2_xon_packets"`
		uint64 `json:"tx_priority3_xon_packets"`
		uint64 `json:"tx_priority4_xon_packets"`
		uint64 `json:"tx_priority5_xon_packets"`
		uint64 `json:"tx_priority6_xon_packets"`
		uint64 `json:"tx_priority7_xon_packets"`
		uint64 `json:"tx_priority0_xoff_packets"`
		uint64 `json:"tx_priority1_xoff_packets"`
		uint64 `json:"tx_priority2_xoff_packets"`
		uint64 `json:"tx_priority3_xoff_packets"`
		uint64 `json:"tx_priority4_xoff_packets"`
		uint64 `json:"tx_priority5_xoff_packets"`
		uint64 `json:"tx_priority6_xoff_packets"`
		uint64 `json:"tx_priority7_xoff_packets"`
		uint64 `json:"tx_priority0_xon_to_xoff_packets"`
		uint64 `json:"tx_priority1_xon_to_xoff_packets"`
		uint64 `json:"tx_priority2_xon_to_xoff_packets"`
		uint64 `json:"tx_priority3_xon_to_xoff_packets"`
		uint64 `json:"tx_priority4_xon_to_xoff_packets"`
		uint64 `json:"tx_priority5_xon_to_xoff_packets"`
		uint64 `json:"tx_priority6_xon_to_xoff_packets"`
		uint64 `json:"tx_priority7_xon_to_xoff_packets"`
	*/
}

// EthdevStats holds the port stats
type EthdevStats struct {
	Stats EthdevPortStats `json:"/ethdev/stats"`
}

// EthdevPidList of port IDs
type EthdevPidList struct {
	Pids []uint16 `json:"/ethdev/list"`
}

// EALParams is the data structure to hold EAL Parameters
type EALParams struct {
	Params []string `json:"/eal/params"`
}

// AppParams is the data structure to hold EAL Parameters
type AppParams struct {
	Params []string `json:"/eal/app_params"`
}

// CmdList host the list of commands
type CmdList struct {
	Cmds []string `json:"/"`
}

// Information and information about a DPDK instance
type Information struct {
	Version     string
	Cmds        CmdList   // List of all known commands
	Params      EALParams // Holds the EAL parameter data
	AppParams   AppParams // Holds the EAL parameter data
	PidList     EthdevPidList
	EthdevStats []*EthdevStats
	PrevStats   [MaxPortCount]EthdevStats
}
