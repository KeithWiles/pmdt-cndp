/* SPDX-License-Identifier: BSD-3-Clause
 * Copyright(c) 2020 Intel Corporation
 */

#ifndef COMMON_H_
#define COMMON_H_

#ifdef __cplusplus
#include <cstring>
#else
typedef unsigned char bool;
#endif
#include <stdint.h>

static const char DEFAULT_MMAP_LOCATION[] = "/tmp/opcm-info-mmap";
static const char VERSION[] = "1.0.5";

#define MAX_CPU_CORES 256
#define MAX_SOCKETS 4
#define MEMORY_MAX_IMC_CHANNELS 8
#define MEMORY_READ 0
#define MEMORY_WRITE 1
#define MEMORY_READ_RANK_A 0
#define MEMORY_WRITE_RANK_A 1
#define MEMORY_READ_RANK_B 2
#define MEMORY_WRITE_RANK_B 3
#define MEMORY_PARTIAL 2
#define QPI_MAX_LINKS MAX_SOCKETS * 4

#define VERSION_SIZE 12

#define ALIGNMENT 64
#define ALIGN(x) __attribute__((aligned((x))))

#ifdef __cplusplus
namespace PCMDaemon {
#endif
	typedef int int32;
	typedef long int64;
	typedef unsigned int uint32;
	typedef unsigned long uint64;

	struct PCMSystem {
		uint32 numOfCores;
		uint32 numOfOnlineCores;
		uint32 numOfSockets;
		uint32 numOfOnlineSockets;
		uint32 numOfQPILinksPerSocket;
		uint32 cpuModel;
#ifdef __cplusplus
	public:
		PCMSystem() :
			numOfCores(0),
			numOfOnlineCores(0),
			numOfSockets(0),
			numOfOnlineSockets(0),
			numOfQPILinksPerSocket(0),
			cpuModel(0) {}
#endif
	} ALIGN(ALIGNMENT);

	typedef struct PCMSystem PCMSystem;

	struct PCMCoreCounter {
		uint32 coreId;
		uint32 socketId;
		double instructionsPerCycle;
		uint64 cycles;
		uint64 instructionsRetired;
		double execUsage;
		double relativeFrequency;
		double activeRelativeFrequency;
		uint64 l3CacheMisses;
		uint64 l3CacheReference;
		uint64 l2CacheMisses;
		double l3CacheHitRatio;
		double l2CacheHitRatio;
		double l3CacheMPI;
		double l2CacheMPI;
		bool l3CacheOccupancyAvailable;
		uint64 l3CacheOccupancy;
		bool localMemoryBWAvailable;
		uint64 localMemoryBW;
		bool remoteMemoryBWAvailable;
		uint64 remoteMemoryBW;
		uint64 localMemoryAccesses;
		uint64 remoteMemoryAccesses;
		int32 thermalHeadroom;

#ifdef __cplusplus
	public:
		PCMCoreCounter() :
			l3CacheOccupancyAvailable(false),
			l3CacheOccupancy(0),
			localMemoryBWAvailable(false),
			localMemoryBW(0),
			remoteMemoryBWAvailable(false),
			remoteMemoryBW(0) {}
#endif
	} ALIGN(ALIGNMENT);

	typedef struct PCMCoreCounter PCMCoreCounter;

	struct PCMCore {
		PCMCoreCounter cores[MAX_CPU_CORES];
		bool packageEnergyMetricsAvailable;
		double energyUsedBySockets[MAX_SOCKETS] ALIGN(ALIGNMENT);

#ifdef __cplusplus
	public:
		PCMCore() :
			packageEnergyMetricsAvailable(false) {
			for(int i = 0; i < MAX_SOCKETS; ++i)
			{
				energyUsedBySockets[i] = -1.0;
			}
		}
#endif
	} ALIGN(ALIGNMENT);

	typedef struct PCMCore PCMCore;

	struct PCMMemoryChannelCounter {
		float read;
		float write;
		float total;

#ifdef __cplusplus
	public:
		PCMMemoryChannelCounter() :
			read(-1.0),
			write(-1.0),
			total(-1.0) {}
#endif
	} ALIGN(ALIGNMENT);

	typedef struct PCMMemoryChannelCounter PCMMemoryChannelCounter;

	struct PCMMemorySocketCounter {
		uint64 socketId;
		PCMMemoryChannelCounter channels[MEMORY_MAX_IMC_CHANNELS];
		uint32 numOfChannels;
		float read;
		float write;
		float partialWrite;
		float total;
		double dramEnergy;

#ifdef __cplusplus
	public:
		PCMMemorySocketCounter() :
			numOfChannels(0),
			read(-1.0),
			write(-1.0),
			partialWrite(-1.0),
			total(-1.0),
			dramEnergy(0.0) {}
#endif
	} ALIGN(ALIGNMENT);

	typedef struct PCMMemorySocketCounter PCMMemorySocketCounter;

	struct PCMMemorySystemCounter {
		float read;
		float write;
		float total;

#ifdef __cplusplus
	public:
		PCMMemorySystemCounter() :
			read(-1.0),
			write(-1.0),
			total(-1.0) {}
#endif
	} ALIGN(ALIGNMENT);

	typedef struct PCMMemorySystemCounter PCMMemorySystemCounter;

	struct PCMMemory {
		PCMMemorySocketCounter sockets[MAX_SOCKETS];
		PCMMemorySystemCounter system;
		bool dramEnergyMetricsAvailable;

#ifdef __cplusplus
	public:
		PCMMemory() :
			dramEnergyMetricsAvailable(false) {}
#endif
	} ALIGN(ALIGNMENT);

	typedef struct PCMMemory PCMMemory;

	struct PCMQPILinkCounter {
		uint64 bytes;
		double utilization;

#ifdef __cplusplus
	public:
		PCMQPILinkCounter() :
			bytes(0),
			utilization(-1.0) {}
#endif
	} ALIGN(ALIGNMENT);

	typedef struct PCMQPILinkCounter PCMQPILinkCounter;

	struct PCMQPISocketCounter {
		uint64 socketId;
		PCMQPILinkCounter links[QPI_MAX_LINKS];
		uint64 total;

#ifdef __cplusplus
	public:
		PCMQPISocketCounter() :
			total(0) {}
#endif
	} ALIGN(ALIGNMENT);

	typedef struct PCMQPISocketCounter PCMQPISocketCounter;

	struct PCMQPI {
		PCMQPISocketCounter incoming[MAX_SOCKETS];
		uint64 incomingTotal;
		PCMQPISocketCounter outgoing[MAX_SOCKETS];
		uint64 outgoingTotal;
		bool incomingQPITrafficMetricsAvailable;
		bool outgoingQPITrafficMetricsAvailable;

#ifdef __cplusplus
	public:
		PCMQPI() :
			incomingTotal(0),
			outgoingTotal(0),
			incomingQPITrafficMetricsAvailable(false),
			outgoingQPITrafficMetricsAvailable(false) {}
#endif
	} ALIGN(ALIGNMENT);

	typedef struct PCMQPI PCMQPI;

	struct SharedPCMCounters {
		PCMSystem system;
		PCMCore core;
		PCMMemory memory;
		PCMQPI qpi;
	} ALIGN(ALIGNMENT);

	typedef struct SharedPCMCounters SharedPCMCounters;

	struct PCIeEvents {
		// PCIe read events (PCI devices reading from memory)
		uint64 PCIeRdCur; // PCIe read current
		uint64 PCIeNSRd;  // PCIe non-snoop read
		// PCIe write events (PCI devices writing to memory)
		uint64 PCIeWiLF;  // PCIe Write (non-allocating)
		uint64 PCIeItoM;  // PCIe Write (allocating)
		uint64 PCIeNSWr;  // PCIe Non-snoop write (partial)
		uint64 PCIeNSWrF; // PCIe Non-snoop write (full)
		// events shared by CPU and IO
		uint64 RFO;       // Demand Data RFO [PCIe write partial cache line]
		uint64 CRd;       // Demand Code Read
		uint64 DRd;       // Demand Data read
		uint64 PRd;       // Partial Reads (UC) [MMIO Read]
		uint64 WiL;       // Write Invalidate Line - partial [MMIO write], PL: Not documented in HSX/IVT
		uint64 ItoM;      // Request Invalidate Line [PCIe write full cache line]
		uint64 RdBw;	  // PCIe Rd bandwidth */
		uint64 WrBw;	  // PCIe Wr bandwidth */
	} ALIGN(ALIGNMENT);

	typedef struct PCIeEvents PCIeEvents_t;

	struct Sample_s {
		PCIeEvents_t total;
		PCIeEvents_t miss;
		PCIeEvents_t hit;
	} ALIGN(ALIGNMENT);

	typedef struct Sample_s Sample_t;

	struct SharedHeader {
		char version[VERSION_SIZE];
		uint64 tscBegin;
		uint64 tscEnd;
		uint64 cyclesToGetPCMState;
		uint64 timestamp;
        int socketfd;
		uint32 pollMs;

#ifdef __cplusplus
	public:
		SharedHeader() :
			tscBegin(0),
            tscEnd(0),
			cyclesToGetPCMState(0),
			timestamp(0),
            socketfd(-1),
			pollMs(-1)
			{
				memset(this->version, '\0', sizeof(char)*VERSION_SIZE);
			}
#endif
	} ALIGN(ALIGNMENT);

    typedef struct SharedHeader SharedHeader;

	struct SharedPCMState {
		SharedHeader hdr;
		SharedPCMCounters pcm;
		Sample_t sample[MAX_SOCKETS];
		PCIeEvents_t aggregate;
	} ALIGN(ALIGNMENT);

	typedef struct SharedPCMState SharedPCMState;
#ifdef __cplusplus
}
#endif

#endif /* COMMON_H_ */
