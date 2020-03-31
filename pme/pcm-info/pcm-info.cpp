/* SPDX-License-Identifier: BSD-3-Clause
 * Copyright(c) 2020 Intel Corporation
 */

#include <cstdlib>
#include <iostream>
#include <cstring>
#include <algorithm>
#include <unistd.h>
#include <sys/types.h>
#include <sys/ipc.h>
#include <errno.h>
#include <time.h>
#include <sys/mman.h>

#ifndef CLOCK_MONOTONIC_RAW
#define CLOCK_MONOTONIC_RAW             (4) /* needed for SLES11 */
#endif

#include "pinfo.h"
#include "pcm-info.h"
#include "system-info.h"
#include "common.h"
#include "pcm.h"

namespace PCMDaemon {

	std::string Daemon::mmapLocation_;
	int Daemon::sharedMemoryId_;
	SharedPCMState* Daemon::sharedPCMState_;

	Daemon::Daemon(int argc, char *argv[])
	: debugMode_(false), pollIntervalMs_(0), groupName_(""), mode_(Mode::DIFFERENCE), pcmInstance_(NULL)
	{
        const char *err_str;

		allowedSubscribers_.push_back("core");
		allowedSubscribers_.push_back("memory");
		allowedSubscribers_.push_back("qpi");
		allowedSubscribers_.push_back("pcie");

		mmapLocation_ = std::string(DEFAULT_MMAP_LOCATION);
		sharedMemoryId_ = 0;
		sharedPCMState_ = NULL;

		readApplicationArguments(argc, argv);
		setupSharedMemory();

        if (pinfo_init(NULL, NULL, &err_str) < 0) {
            std::cerr << "PINFO failed to initialize. err: " << err_str << std::endl;
            return;
        }
		setupPCM();

		// Put the poll interval in shared memory so that the client knows
		sharedPCMState_->hdr.pollMs = pollIntervalMs_;

		updatePCMState(&systemStatesBefore_, &socketStatesBefore_, &coreStatesBefore_);
		systemStatesForQPIBefore_ = SystemCounterState(systemStatesBefore_);

		serverUncorePowerStatesBefore_ = new ServerUncorePowerState[pcmInstance_->getNumSockets()];
		serverUncorePowerStatesAfter_ = new ServerUncorePowerState[pcmInstance_->getNumSockets()];

        setupInfo(sharedPCMState_);
	}

	int Daemon::run()
	{
		std::cout << std::endl << "**** PCM Info Daemon Started *****" << std::endl;

		while(true) {
			if(debugMode_) {
				time_t rawtime;
				struct tm timeinfo;
				char timeBuffer[200];
				time(&rawtime);
				localtime_r(&rawtime, &timeinfo);

				snprintf(timeBuffer, 200, "[%02d %02d %04d %02d:%02d:%02d]",
					timeinfo.tm_mday, timeinfo.tm_mon + 1, timeinfo.tm_year + 1900,
					timeinfo.tm_hour, timeinfo.tm_min, timeinfo.tm_sec);

				std::cout << timeBuffer << "\tFetching counters..." << std::endl;
			}

			usleep(pollIntervalMs_ * 1000);

			getPCMCounters();
		}

		return EXIT_SUCCESS;
	}

	Daemon::~Daemon()
	{
		delete serverUncorePowerStatesBefore_;
		delete serverUncorePowerStatesAfter_;

        pinfo_remove();
	}

	void Daemon::setupPCM()
	{
		pcmInstance_ = PCM::getInstance();
		pcmInstance_->setBlocked(false);
		set_signal_handlers();
		set_post_cleanup_callback(&Daemon::cleanup);

		checkAccessAndProgramPCM();
	}

	void Daemon::checkAccessAndProgramPCM()
	{
	    PCM::ErrorCode status;

	    if(subscribers_.find("core") != subscribers_.end()) {
		    EventSelectRegister defEventSelectRegister;
		    defEventSelectRegister.value = 0;
		    defEventSelectRegister.fields.usr = 1;
		    defEventSelectRegister.fields.os = 1;
		    defEventSelectRegister.fields.enable = 1;

		    uint32 numOfCustomCounters = 4;

		    EventSelectRegister regs[numOfCustomCounters];
		    PCM::ExtendedCustomCoreEventDescription conf;
		    conf.nGPCounters = numOfCustomCounters;
		    conf.gpCounterCfg = regs;

			try {
				pcmInstance_->setupCustomCoreEventsForNuma(conf);
			}
			catch (UnsupportedProcessorException& e) {
		        std::cerr << std::endl << "PCM Info daemon does not support your processor currently." << std::endl << std::endl;
		        exit(EXIT_FAILURE);
			}

			// Set default values for event select registers
			for (uint32 i(0); i < numOfCustomCounters; ++i)
				regs[i] = defEventSelectRegister;

			regs[0].fields.event_select = 0xB7; // OFFCORE_RESPONSE 0 event
			regs[0].fields.umask = 0x01;
			regs[1].fields.event_select = 0xBB; // OFFCORE_RESPONSE 1 event
			regs[1].fields.umask = 0x01;
			regs[2].fields.event_select = ARCH_LLC_MISS_EVTNR;
			regs[2].fields.umask = ARCH_LLC_MISS_UMASK;
			regs[3].fields.event_select = ARCH_LLC_REFERENCE_EVTNR;
			regs[3].fields.umask = ARCH_LLC_REFERENCE_UMASK;

            if (pcmInstance_->getMaxCustomCoreEvents() == 3)
                conf.nGPCounters = 2; // drop LLC metrics

		    status = pcmInstance_->program(PCM::EXT_CUSTOM_CORE_EVENTS, &conf);
		} else
			status = pcmInstance_->program();

		switch (status) {
			case PCM::Success:
				break;
			case PCM::MSRAccessDenied:
				std::cerr << "Access to Intel(r) Performance Counter Monitor has denied (no MSR or PCI CFG space access)." << std::endl;
				exit(EXIT_FAILURE);
			case PCM::PMUBusy:
				std::cerr << "Access to Intel(r) Performance Counter Monitor has denied (Performance Monitoring Unit is occupied by other application). Try to stop the application that uses PMU." << std::endl;
				std::cerr << "Alternatively you can try to reset PMU configuration at your own risk. Try to reset? (y/n)" << std::endl;
				char yn;
				std::cin >> yn;
				if ('y' == yn) {
					pcmInstance_->resetPMU();
					std::cerr << "PMU configuration has been reset. Try to rerun the program again." << std::endl;
				}
				exit(EXIT_FAILURE);
			default:
				std::cerr << "Access to Intel(r) Performance Counter Monitor has denied (Unknown error)." << std::endl;
				exit(EXIT_FAILURE);
		}
	}

	void Daemon::readApplicationArguments(int argc, char *argv[])
	{
		int opt;
		int counterCount(0);

		if(argc == 1)
			printExampleUsageAndExit(argv);

		std::cout << std::endl;

        pollIntervalMs_ = 1000;
        debugMode_ = false;
		while ((opt = getopt(argc, argv, "p:c:s:d")) != -1) {
			switch (opt) {
			case 'p':
				pollIntervalMs_ = atoi(optarg);
				break;
			case 'c': {
                std::string subscriber(optarg);

                if(subscriber == "all") {
                    for(std::vector<std::string>::const_iterator it = allowedSubscribers_.begin(); it != allowedSubscribers_.end(); ++it) {
                        subscribers_.insert(std::pair<std::string, uint32>(*it, 1));
                        ++counterCount;
                    }
                } else {
                    if(std::find(allowedSubscribers_.begin(), allowedSubscribers_.end(), subscriber) == allowedSubscribers_.end())
                        printExampleUsageAndExit(argv);

                    subscribers_.insert(std::pair<std::string, uint32>(subscriber, 1));
                    ++counterCount;
                }

                if (debugMode_)
                    std::cout << "Listening to '" << subscriber << "' counters" << std::endl;
                }
				break;
			case 'd':
				debugMode_ = true;
				std::cout << "Debug mode enabled" << std::endl;
				break;
			case 's':
                mmapLocation_ = std::string(optarg);
                if (debugMode_)
                    std::cout << "Shared MMAP location: " << mmapLocation_ << " bool size " << sizeof(bool) << std::endl;
				break;
			default:
				printExampleUsageAndExit(argv);
				break;
			}
		}
        if (debugMode_)
            std::cout << "Polling every " << pollIntervalMs_ << "ms" << std::endl;

		if(pollIntervalMs_ <= 0 || counterCount == 0)
			printExampleUsageAndExit(argv);

        if (debugMode_)
    		std::cout << "PCM Info Daemon version: " << VERSION << std::endl << std::endl;
	}

	void Daemon::printExampleUsageAndExit(char *argv[])
	{
		std::cerr << std::endl;
		std::cerr << "-------------------------------------------------------------------" << std::endl;
		std::cerr << "Example usage: " << argv[0] << " -p 50 -c numa -c memory" << std::endl;
		std::cerr << "Poll every 50ms. Fetch counters for numa and memory" << std::endl << std::endl;

		std::cerr << "Example usage: " << argv[0] << " -p 250 -c all" << std::endl;
		std::cerr << "Poll every 250ms. Fetch all counters (core, numa & memory)." << std::endl;

		std::cerr << "-p <milliseconds> for poll frequency" << std::endl;
		std::cerr << "-c <counter> to request specific counters (Allowed counters: all ";

		for(std::vector<std::string>::const_iterator it = allowedSubscribers_.begin(); it != allowedSubscribers_.end(); ++it) {
			std::cerr << *it;

			if(it+1 != allowedSubscribers_.end())
				std::cerr << " ";
		}

		std::cerr << ")";
		std::cerr << std::endl;

		exit(EXIT_FAILURE);
	}

	void Daemon::setupSharedMemory()
	{
		sharedMemoryId_ = open(mmapLocation_.c_str(), O_RDWR | O_CREAT | O_TRUNC, (mode_t)0666);
		if (sharedMemoryId_ < 0) {
			std::cerr << "Failed to open shared memory file (errno=" << errno << ")" << std::endl;
			exit(EXIT_FAILURE);
		}

		__off_t off = (sizeof(SharedPCMState) + 4095) & ~4095;
		if (lseek(sharedMemoryId_, off - 1, SEEK_SET) < 0) {
			close(sharedMemoryId_);
			std::cerr << "Failed to grow file (errno=" << errno << ")" << std::endl;
				exit(EXIT_FAILURE);
		}
		if (write(sharedMemoryId_, "", 1) < 0) {
			close(sharedMemoryId_);
			std::cerr << "Failed to write last byte (errno=" << errno << ")" << std::endl;
				exit(EXIT_FAILURE);
		}

		sharedPCMState_ = (SharedPCMState *)mmap(0, sizeof(SharedPCMState), PROT_READ | PROT_WRITE, MAP_SHARED, sharedMemoryId_, 0);
		if (sharedPCMState_ == MAP_FAILED) {
			close(sharedMemoryId_);
			std::cerr << "Failed to mmap (errno=" << errno << ")" << std::endl;
			exit(EXIT_FAILURE);
		}
        if (debugMode_)
    		std::cout << "SharePCMState fd = " << sharedMemoryId_ << " @ " << sharedPCMState_ << " size " << off << std::endl;
	}

	gid_t Daemon::resolveGroupName(const std::string& groupName)
	{
		struct group* group = getgrnam(groupName.c_str());

		if(group == NULL) {
			std::cerr << "Failed to resolve group '" << groupName << "'" << std::endl;
			exit(EXIT_FAILURE);
		}

		return group->gr_gid;
	}

	void Daemon::getPCMCounters()
	{
		memcpy (sharedPCMState_->hdr.version, VERSION, sizeof(VERSION));
		sharedPCMState_->hdr.version[sizeof(VERSION)] = '\0';

        sharedPCMState_->hdr.tscBegin = RDTSC();

		updatePCMState(&systemStatesAfter_, &socketStatesAfter_, &coreStatesAfter_);

		getPCMSystem();

		if(subscribers_.find("core") != subscribers_.end())
			getPCMCore();
		if(subscribers_.find("memory") != subscribers_.end())
			getPCMMemory();
		bool fetchQPICounters = subscribers_.find("qpi") != subscribers_.end();
		if(fetchQPICounters)
			getPCMQPI();
		if(subscribers_.find("pcie") != subscribers_.end())
			getPCIeCounters();

		const auto tscEnd = RDTSC();
		sharedPCMState_->hdr.cyclesToGetPCMState = tscEnd - sharedPCMState_->hdr.tscBegin;
		sharedPCMState_->hdr.timestamp = getTimestamp();

		// As the client polls this timestamp (lastUpdateTsc)
		sharedPCMState_->hdr.tscEnd = tscEnd;
		if(mode_ == Mode::DIFFERENCE)
			swapPCMBeforeAfterState();
		if(fetchQPICounters)
			systemStatesForQPIBefore_ = SystemCounterState(systemStatesAfter_);

		std::swap(collectionTimeBefore_, collectionTimeAfter_);
	}

	void Daemon::updatePCMState(SystemCounterState* systemStates, std::vector<SocketCounterState>* socketStates, std::vector<CoreCounterState>* coreStates)
	{
		if(subscribers_.find("core") != subscribers_.end())
			pcmInstance_->getAllCounterStates(*systemStates, *socketStates, *coreStates);
		else {
			if(subscribers_.find("memory") != subscribers_.end() || subscribers_.find("qpi") != subscribers_.end())
				pcmInstance_->getUncoreCounterStates(*systemStates, *socketStates);
		}
		collectionTimeAfter_ = pcmInstance_->getTickCount();
	}

	void Daemon::swapPCMBeforeAfterState()
	{
		//After state now becomes before state (for the next iteration)
		std::swap(coreStatesBefore_, coreStatesAfter_);
		std::swap(socketStatesBefore_, socketStatesAfter_);
		std::swap(systemStatesBefore_, systemStatesAfter_);
		std::swap(serverUncorePowerStatesBefore_, serverUncorePowerStatesAfter_);
	}

	void Daemon::getPCMSystem()
	{
		PCMSystem& system = sharedPCMState_->pcm.system;
		system.numOfCores = pcmInstance_->getNumCores();
		system.numOfOnlineCores = pcmInstance_->getNumOnlineCores();
		system.numOfSockets = pcmInstance_->getNumSockets();
		system.numOfOnlineSockets = pcmInstance_->getNumOnlineSockets();
		system.numOfQPILinksPerSocket = pcmInstance_->getQPILinksPerSocket();
		system.cpuModel = pcmInstance_->getCPUModel();
	}

	void Daemon::getPCMCore()
	{
		PCMCore& core = sharedPCMState_->pcm.core;

		const uint32 numCores = sharedPCMState_->pcm.system.numOfCores;

		uint32 onlineCoresI(0);
		for(uint32 coreI(0); coreI < numCores ; ++coreI) {
			if(!pcmInstance_->isCoreOnline(coreI))
				continue;

			PCMCoreCounter& coreCounters = core.cores[onlineCoresI];

			int32 socketId = pcmInstance_->getSocketId(coreI);
			double instructionsPerCycle = getIPC(coreStatesBefore_[coreI], coreStatesAfter_[coreI]);
			uint64 cycles = getCycles(coreStatesBefore_[coreI], coreStatesAfter_[coreI]);
			uint64 instructionsRetired = getInstructionsRetired(coreStatesBefore_[coreI], coreStatesAfter_[coreI]);
			double execUsage = getExecUsage(coreStatesBefore_[coreI], coreStatesAfter_[coreI]);
			double relativeFrequency = getRelativeFrequency(coreStatesBefore_[coreI], coreStatesAfter_[coreI]);
			double activeRelativeFrequency = getActiveRelativeFrequency(coreStatesBefore_[coreI], coreStatesAfter_[coreI]);
			uint64 l3CacheMisses = getNumberOfCustomEvents(2, coreStatesBefore_[coreI], coreStatesAfter_[coreI]);
			uint64 l3CacheReference = getNumberOfCustomEvents(3, coreStatesBefore_[coreI], coreStatesAfter_[coreI]);
			uint64 l2CacheMisses = getL2CacheMisses(coreStatesBefore_[coreI], coreStatesAfter_[coreI]);
			double l3CacheHitRatio = getL3CacheHitRatio(coreStatesBefore_[coreI], coreStatesAfter_[coreI]);
			double l2CacheHitRatio = getL2CacheHitRatio(coreStatesBefore_[coreI], coreStatesAfter_[coreI]);
			double l3CacheMPI = double(l3CacheMisses) / instructionsRetired;
			double l2CacheMPI = double(l2CacheMisses) / instructionsRetired;
			int32 thermalHeadroom = coreStatesAfter_[coreI].getThermalHeadroom();

			coreCounters.coreId = coreI;
			coreCounters.socketId = socketId;
			coreCounters.instructionsPerCycle = instructionsPerCycle;
			coreCounters.cycles = cycles;
			coreCounters.instructionsRetired = instructionsRetired;
			coreCounters.execUsage = execUsage;
			coreCounters.relativeFrequency = relativeFrequency;
			coreCounters.activeRelativeFrequency = activeRelativeFrequency;
			coreCounters.l3CacheMisses = l3CacheMisses;
			coreCounters.l3CacheReference = l3CacheReference;
			coreCounters.l2CacheMisses = l2CacheMisses;
			coreCounters.l3CacheHitRatio = l3CacheHitRatio;
			coreCounters.l2CacheHitRatio = l2CacheHitRatio;
			coreCounters.l3CacheMPI = l3CacheMPI;
			coreCounters.l2CacheMPI = l2CacheMPI;
			coreCounters.thermalHeadroom = thermalHeadroom;

			coreCounters.l3CacheOccupancyAvailable = pcmInstance_->L3CacheOccupancyMetricAvailable();
			if (coreCounters.l3CacheOccupancyAvailable) {
				uint64 l3CacheOccupancy = getL3CacheOccupancy(coreStatesAfter_[coreI]);
				coreCounters.l3CacheOccupancy = l3CacheOccupancy;
			}

			coreCounters.localMemoryBWAvailable = pcmInstance_->CoreLocalMemoryBWMetricAvailable();
			if (coreCounters.localMemoryBWAvailable) {
				uint64 localMemoryBW = getLocalMemoryBW(coreStatesBefore_[coreI], coreStatesAfter_[coreI]);
				coreCounters.localMemoryBW = localMemoryBW;
			}

			coreCounters.remoteMemoryBWAvailable = pcmInstance_->CoreRemoteMemoryBWMetricAvailable();
			if (coreCounters.remoteMemoryBWAvailable) {
				uint64 remoteMemoryBW = getRemoteMemoryBW(coreStatesBefore_[coreI], coreStatesAfter_[coreI]);
				coreCounters.remoteMemoryBW = remoteMemoryBW;
			}

			coreCounters.localMemoryAccesses = getNumberOfCustomEvents(0, coreStatesBefore_[coreI], coreStatesAfter_[coreI]);
			coreCounters.remoteMemoryAccesses = getNumberOfCustomEvents(1, coreStatesBefore_[coreI], coreStatesAfter_[coreI]);

			++onlineCoresI;
		}

		const uint32 numSockets = sharedPCMState_->pcm.system.numOfSockets;

		core.packageEnergyMetricsAvailable = pcmInstance_->packageEnergyMetricsAvailable();
		if(core.packageEnergyMetricsAvailable) {
			for (uint32 i(0); i < numSockets; ++i)
				core.energyUsedBySockets[i] = getConsumedJoules(socketStatesBefore_[i], socketStatesAfter_[i]);
		}
	}

	void Daemon::getPCMMemory()
	{
		pcmInstance_->disableJKTWorkaround();

		PCMMemory& memory = sharedPCMState_->pcm.memory;
		memory.dramEnergyMetricsAvailable = pcmInstance_->dramEnergyMetricsAvailable();

		const uint32 numSockets = sharedPCMState_->pcm.system.numOfSockets;

        for(uint32 i(0); i < numSockets; ++i)
        	serverUncorePowerStatesAfter_[i] = pcmInstance_->getServerUncorePowerState(i);

        uint64 elapsedTime = collectionTimeAfter_ - collectionTimeBefore_;

		float iMC_Rd_socket_chan[MAX_SOCKETS][MEMORY_MAX_IMC_CHANNELS];
		float iMC_Wr_socket_chan[MAX_SOCKETS][MEMORY_MAX_IMC_CHANNELS];
		float iMC_Rd_socket[MAX_SOCKETS];
		float iMC_Wr_socket[MAX_SOCKETS];
		uint64 partial_write[MAX_SOCKETS];

		for(uint32 skt(0); skt < numSockets; ++skt) {
			iMC_Rd_socket[skt] = 0.0;
			iMC_Wr_socket[skt] = 0.0;
			partial_write[skt] = 0;

			for(uint32 channel(0); channel < MEMORY_MAX_IMC_CHANNELS; ++channel) {
				//In case of JKT-EN, there are only three channels. Skip one and continue.
				bool memoryReadAvailable = getMCCounter(channel,MEMORY_READ,serverUncorePowerStatesBefore_[skt],serverUncorePowerStatesAfter_[skt]) == 0.0;
				bool memoryWriteAvailable = getMCCounter(channel,MEMORY_WRITE,serverUncorePowerStatesBefore_[skt],serverUncorePowerStatesAfter_[skt]) == 0.0;
				if(memoryReadAvailable && memoryWriteAvailable) {
					iMC_Rd_socket_chan[skt][channel] = -1.0;
					iMC_Wr_socket_chan[skt][channel] = -1.0;
					continue;
				}

				iMC_Rd_socket_chan[skt][channel] = (float) (getMCCounter(channel,MEMORY_READ,serverUncorePowerStatesBefore_[skt],serverUncorePowerStatesAfter_[skt]) * 64 / 1000000.0 / (elapsedTime/1000.0));
				iMC_Wr_socket_chan[skt][channel] = (float) (getMCCounter(channel,MEMORY_WRITE,serverUncorePowerStatesBefore_[skt],serverUncorePowerStatesAfter_[skt]) * 64 / 1000000.0 / (elapsedTime/1000.0));

				iMC_Rd_socket[skt] += iMC_Rd_socket_chan[skt][channel];
				iMC_Wr_socket[skt] += iMC_Wr_socket_chan[skt][channel];

				partial_write[skt] += (uint64) (getMCCounter(channel,MEMORY_PARTIAL,serverUncorePowerStatesBefore_[skt],serverUncorePowerStatesAfter_[skt]) / (elapsedTime/1000.0));
			}
		}

	    float systemRead(0.0);
	    float systemWrite(0.0);

	    uint32 onlineSocketsI(0);
	    for(uint32 skt (0); skt < numSockets; ++skt) {
			if(!pcmInstance_->isSocketOnline(skt))
				continue;

			uint64 currentChannelI(0);
	    	for(uint64 channel(0); channel < MEMORY_MAX_IMC_CHANNELS; ++channel) {
				//If the channel read neg. value, the channel is not working; skip it.
				if(iMC_Rd_socket_chan[0][skt*MEMORY_MAX_IMC_CHANNELS+channel] < 0.0 && iMC_Wr_socket_chan[0][skt*MEMORY_MAX_IMC_CHANNELS+channel] < 0.0)
					continue;

				float socketChannelRead = iMC_Rd_socket_chan[0][skt*MEMORY_MAX_IMC_CHANNELS+channel];
				float socketChannelWrite = iMC_Wr_socket_chan[0][skt*MEMORY_MAX_IMC_CHANNELS+channel];

				memory.sockets[onlineSocketsI].channels[currentChannelI].read = socketChannelRead;
				memory.sockets[onlineSocketsI].channels[currentChannelI].write = socketChannelWrite;
				memory.sockets[onlineSocketsI].channels[currentChannelI].total = socketChannelRead + socketChannelWrite;

				++currentChannelI;
			}

			memory.sockets[onlineSocketsI].socketId = skt;
			memory.sockets[onlineSocketsI].numOfChannels = currentChannelI;
			memory.sockets[onlineSocketsI].read = iMC_Rd_socket[skt];
			memory.sockets[onlineSocketsI].write = iMC_Wr_socket[skt];
			memory.sockets[onlineSocketsI].partialWrite = partial_write[skt];
			memory.sockets[onlineSocketsI].total= iMC_Rd_socket[skt] + iMC_Wr_socket[skt];
			if(memory.dramEnergyMetricsAvailable)
				memory.sockets[onlineSocketsI].dramEnergy = getDRAMConsumedJoules(socketStatesBefore_[skt], socketStatesAfter_[skt]);

			systemRead += iMC_Rd_socket[skt];
			systemWrite += iMC_Wr_socket[skt];

			++onlineSocketsI;
	    }

	    memory.system.read = systemRead;
	    memory.system.write = systemWrite;
	    memory.system.total = systemRead + systemWrite;
	}

	void Daemon::getPCMQPI()
	{
		const uint32 numSockets = sharedPCMState_->pcm.system.numOfSockets;
		const uint32 numLinksPerSocket = sharedPCMState_->pcm.system.numOfQPILinksPerSocket;

		sharedPCMState_->pcm.qpi.incomingQPITrafficMetricsAvailable = pcmInstance_->incomingQPITrafficMetricsAvailable();
        if (debugMode_) {
            std::cout << "incomingQPITrafficMetricsAvailable " << sharedPCMState_->pcm.qpi.incomingQPITrafficMetricsAvailable;
            std::cout << "  (" << &sharedPCMState_->pcm.qpi.incomingQPITrafficMetricsAvailable << ")" << std::endl;
        }

		if (sharedPCMState_->pcm.qpi.incomingQPITrafficMetricsAvailable) {
			uint32 onlineSocketsI(0);

			for (uint32 i(0); i < numSockets; ++i) {
				if(!pcmInstance_->isSocketOnline(i))
					continue;

				sharedPCMState_->pcm.qpi.incoming[onlineSocketsI].socketId = i;

				uint64 total(0);
				for (uint32 l(0); l < numLinksPerSocket; ++l) {
					uint64 bytes = getIncomingQPILinkBytes(i, l, systemStatesBefore_, systemStatesAfter_);
					sharedPCMState_->pcm.qpi.incoming[onlineSocketsI].links[l].bytes = bytes;
					sharedPCMState_->pcm.qpi.incoming[onlineSocketsI].links[l].utilization = getIncomingQPILinkUtilization(i, l, systemStatesForQPIBefore_, systemStatesAfter_);

					total+=bytes;
				}
				sharedPCMState_->pcm.qpi.incoming[i].total = total;

				++onlineSocketsI;
			}

			sharedPCMState_->pcm.qpi.incomingTotal = getAllIncomingQPILinkBytes(systemStatesBefore_, systemStatesAfter_);
		}

		sharedPCMState_->pcm.qpi.outgoingQPITrafficMetricsAvailable = pcmInstance_->outgoingQPITrafficMetricsAvailable();
        if (debugMode_) {
            std::cout << "outgoingQPITrafficMetricsAvailable " << sharedPCMState_->pcm.qpi.outgoingQPITrafficMetricsAvailable;
            std::cout << "  (" << &sharedPCMState_->pcm.qpi.outgoingQPITrafficMetricsAvailable << ")" << std::endl;
        }
		if (sharedPCMState_->pcm.qpi.outgoingQPITrafficMetricsAvailable) {
			uint32 onlineSocketsI(0);

			for (uint32 i(0); i < numSockets; ++i) {
				if(!pcmInstance_->isSocketOnline(i))
					continue;

				sharedPCMState_->pcm.qpi.outgoing[onlineSocketsI].socketId = i;

				uint64 total(0);
				for (uint32 l(0); l < numLinksPerSocket; ++l) {
					uint64 bytes = getOutgoingQPILinkBytes(i, l, systemStatesBefore_, systemStatesAfter_);
					sharedPCMState_->pcm.qpi.outgoing[onlineSocketsI].links[l].bytes = bytes;
					sharedPCMState_->pcm.qpi.outgoing[onlineSocketsI].links[l].utilization = getOutgoingQPILinkUtilization(i, l, systemStatesForQPIBefore_, systemStatesAfter_);

					total+=bytes;
				}
				sharedPCMState_->pcm.qpi.outgoing[i].total = total;

				++onlineSocketsI;
			}

			sharedPCMState_->pcm.qpi.outgoingTotal = getAllOutgoingQPILinkBytes(systemStatesBefore_, systemStatesAfter_);
		}
	}

#define NUM_SAMPLES (1)
#define BANDWIDTH_CNT	2

    uint32 num_events = (sizeof(PCIeEvents_t)/sizeof(uint64)) - BANDWIDTH_CNT;

	void Daemon::getPCIeEvents(PCM *m, PCM::PCIeEventCode opcode, uint32 delay_ms, Sample_t *sample, const uint32 tid = 0, const uint32 q = 0, const uint32 nc = 0)
	{
		PCIeCounterState * before = new PCIeCounterState[m->getNumSockets()];
		PCIeCounterState * after = new PCIeCounterState[m->getNumSockets()];
		PCIeCounterState * before2 = new PCIeCounterState[m->getNumSockets()];
		PCIeCounterState * after2 = new PCIeCounterState[m->getNumSockets()];
		SharedPCMState *mm = sharedPCMState_;
		uint32 i;

		m->programPCIeCounters(opcode, tid, 0, q, nc);
		for(i=0; i<m->getNumSockets(); ++i)
			before[i] = m->getPCIeCounterState(i);
		MySleepMs(delay_ms);
		for(i=0; i<m->getNumSockets(); ++i)
			after[i] = m->getPCIeCounterState(i);

		m->programPCIeMissCounters(opcode, tid, q, nc);
		for(i=0; i<m->getNumSockets(); ++i)
			before2[i] = m->getPCIeCounterState(i);
		MySleepMs(delay_ms);
		for(i=0; i<m->getNumSockets(); ++i)
			after2[i] = m->getPCIeCounterState(i);

		for(i=0; i<m->getNumSockets(); ++i) {
			switch(opcode) {
				case PCM::PCIeRdCur:
				case PCM::SKX_RdCur:
					mm->sample[i].total.PCIeRdCur += (sizeof(PCIeEvents_t)/sizeof(uint64)) * getNumberOfEvents(before[i], after[i]);
					mm->sample[i].miss.PCIeRdCur += (sizeof(PCIeEvents_t)/sizeof(uint64)) * getNumberOfEvents(before2[i], after2[i]);
					mm->sample[i].hit.PCIeRdCur += (mm->sample[i].total.PCIeRdCur > mm->sample[i].miss.PCIeRdCur) ? mm->sample[i].total.PCIeRdCur - mm->sample[i].miss.PCIeRdCur : 0;
					mm->aggregate.PCIeRdCur += mm->sample[i].total.PCIeRdCur;
					break;
				case PCM::PCIeNSRd:
					mm->sample[i].total.PCIeNSRd += (sizeof(PCIeEvents_t)/sizeof(uint64)) * getNumberOfEvents(before[i], after[i]);
					mm->sample[i].miss.PCIeNSRd += (sizeof(PCIeEvents_t)/sizeof(uint64)) * getNumberOfEvents(before2[i], after2[i]);
					mm->sample[i].hit.PCIeNSRd += (mm->sample[i].total.PCIeNSRd > mm->sample[i].miss.PCIeNSRd) ? mm->sample[i].total.PCIeNSRd - mm->sample[i].miss.PCIeNSRd : 0;
					mm->aggregate.PCIeNSRd += mm->sample[i].total.PCIeNSRd;
					break;
				case PCM::PCIeWiLF:
					mm->sample[i].total.PCIeWiLF += (sizeof(PCIeEvents_t)/sizeof(uint64)) * getNumberOfEvents(before[i], after[i]);
					mm->sample[i].miss.PCIeWiLF += (sizeof(PCIeEvents_t)/sizeof(uint64)) * getNumberOfEvents(before2[i], after2[i]);
					mm->sample[i].hit.PCIeWiLF += (mm->sample[i].total.PCIeWiLF > mm->sample[i].miss.PCIeWiLF) ? mm->sample[i].total.PCIeWiLF - mm->sample[i].miss.PCIeWiLF : 0;
					mm->aggregate.PCIeWiLF += mm->sample[i].total.PCIeWiLF;
					break;
				case PCM::PCIeItoM:
					mm->sample[i].total.PCIeItoM += (sizeof(PCIeEvents_t)/sizeof(uint64)) * getNumberOfEvents(before[i], after[i]);
					mm->sample[i].miss.PCIeItoM += (sizeof(PCIeEvents_t)/sizeof(uint64)) * getNumberOfEvents(before2[i], after2[i]);
					mm->sample[i].hit.PCIeItoM += (mm->sample[i].total.PCIeItoM > mm->sample[i].miss.PCIeItoM) ? mm->sample[i].total.PCIeItoM - mm->sample[i].miss.PCIeItoM : 0;
					mm->aggregate.PCIeItoM += mm->sample[i].total.PCIeItoM;
					break;
				case PCM::PCIeNSWr:
					mm->sample[i].total.PCIeNSWr += (sizeof(PCIeEvents_t)/sizeof(uint64)) * getNumberOfEvents(before[i], after[i]);
					mm->sample[i].miss.PCIeNSWr += (sizeof(PCIeEvents_t)/sizeof(uint64)) * getNumberOfEvents(before2[i], after2[i]);
					mm->sample[i].hit.PCIeNSWr += (mm->sample[i].total.PCIeNSWr > mm->sample[i].miss.PCIeNSWr) ? mm->sample[i].total.PCIeNSWr - mm->sample[i].miss.PCIeNSWr : 0;
					mm->aggregate.PCIeNSWr += mm->sample[i].total.PCIeNSWr;
					break;
				case PCM::PCIeNSWrF:
					mm->sample[i].total.PCIeNSWrF += (sizeof(PCIeEvents_t)/sizeof(uint64)) * getNumberOfEvents(before[i], after[i]);
					mm->sample[i].miss.PCIeNSWrF += (sizeof(PCIeEvents_t)/sizeof(uint64)) * getNumberOfEvents(before2[i], after2[i]);
					mm->sample[i].hit.PCIeNSWrF += (mm->sample[i].total.PCIeNSWrF > mm->sample[i].miss.PCIeNSWrF) ? mm->sample[i].total.PCIeNSWrF - mm->sample[i].miss.PCIeNSWrF : 0;
					mm->aggregate.PCIeNSWrF += mm->sample[i].total.PCIeNSWrF;
					break;
				case PCM::SKX_RFO:
				case PCM::RFO:
					if(opcode == PCM::SKX_RFO || tid == PCM::RFOtid) { //Use tid to filter only PCIe traffic
						mm->sample[i].total.RFO += (sizeof(PCIeEvents_t)/sizeof(uint64)) * getNumberOfEvents(before[i], after[i]);
						mm->sample[i].miss.RFO += (sizeof(PCIeEvents_t)/sizeof(uint64)) * getNumberOfEvents(before2[i], after2[i]);
						mm->sample[i].hit.RFO += (mm->sample[i].total.RFO > mm->sample[i].miss.RFO) ? mm->sample[i].total.RFO - mm->sample[i].miss.RFO : 0;
						mm->aggregate.RFO += mm->sample[i].total.RFO;
					}
					break;
				case PCM::SKX_ItoM:
				case PCM::ItoM:
					if(opcode == PCM::SKX_ItoM || tid == PCM::ItoMtid) { //Use tid to filter only PCIe traffic
						mm->sample[i].total.ItoM += (sizeof(PCIeEvents_t)/sizeof(uint64)) * getNumberOfEvents(before[i], after[i]);
						mm->sample[i].miss.ItoM += (sizeof(PCIeEvents_t)/sizeof(uint64)) * getNumberOfEvents(before2[i], after2[i]);
						mm->sample[i].hit.ItoM += (mm->sample[i].total.ItoM > mm->sample[i].miss.ItoM) ? mm->sample[i].total.ItoM - mm->sample[i].miss.ItoM : 0;
						mm->aggregate.ItoM += mm->sample[i].total.ItoM;
					}
					break;
				case PCM::SKX_WiL:
				case PCM::WiL:
					mm->sample[i].total.WiL += (sizeof(PCIeEvents_t)/sizeof(uint64)) * getNumberOfEvents(before[i], after[i]);
					mm->sample[i].miss.WiL += (sizeof(PCIeEvents_t)/sizeof(uint64)) * getNumberOfEvents(before2[i], after2[i]);
					mm->sample[i].hit.WiL += (mm->sample[i].total.WiL > mm->sample[i].miss.WiL) ? mm->sample[i].total.WiL - mm->sample[i].miss.WiL : 0;
					mm->aggregate.WiL += mm->sample[i].total.WiL;
					break;
				case PCM::SKX_PRd:
				case PCM::PRd:
					mm->sample[i].total.PRd += (sizeof(PCIeEvents_t)/sizeof(uint64)) * getNumberOfEvents(before[i], after[i]);
					mm->sample[i].miss.PRd += (sizeof(PCIeEvents_t)/sizeof(uint64)) * getNumberOfEvents(before2[i], after2[i]);
					mm->sample[i].hit.PRd += (mm->sample[i].total.PRd > mm->sample[i].miss.PRd) ? mm->sample[i].total.PRd - mm->sample[i].miss.PRd : 0;
					mm->aggregate.PRd += mm->sample[i].total.PRd;
					break;
				case PCM::SKX_CRd:
				case PCM::CRd:
					mm->sample[i].total.CRd += (sizeof(PCIeEvents_t)/sizeof(uint64)) * getNumberOfEvents(before[i], after[i]);
					mm->sample[i].miss.CRd += (sizeof(PCIeEvents_t)/sizeof(uint64)) * getNumberOfEvents(before2[i], after2[i]);
					mm->sample[i].hit.CRd += (mm->sample[i].total.CRd > mm->sample[i].miss.CRd) ? mm->sample[i].total.CRd - mm->sample[i].miss.CRd : 0;
					mm->aggregate.CRd += mm->sample[i].total.CRd;
					break;
				case PCM::SKX_DRd:
				case PCM::DRd:
					mm->sample[i].total.DRd += (sizeof(PCIeEvents_t)/sizeof(uint64)) * getNumberOfEvents(before[i], after[i]);
					mm->sample[i].miss.DRd += (sizeof(PCIeEvents_t)/sizeof(uint64)) * getNumberOfEvents(before2[i], after2[i]);
					mm->sample[i].hit.DRd += (mm->sample[i].total.DRd > mm->sample[i].miss.DRd) ? mm->sample[i].total.DRd - mm->sample[i].miss.DRd : 0;
					mm->aggregate.DRd += mm->sample[i].total.DRd;
					break;
			}
		}

		delete[] before;
		delete[] after;
		delete[] before2;
		delete[] after2;
	}

	void Daemon::getPCIeCounters()
	{
		SharedPCMState *mm = sharedPCMState_;
	    double delay = -1.0;
		uint32 delay_ms = uint32(1000 / num_events / NUM_SAMPLES);
		uint32 i;

		if(delay_ms * num_events * NUM_SAMPLES < delay * 1000) ++delay_ms; //Adjust the delay_ms if it's less than delay time

	    PCM * m = PCM::getInstance();

        memset(mm->sample, 0 ,sizeof(Sample_t));
        memset(&mm->aggregate, 0 ,sizeof(mm->aggregate));

		if (!m->hasPCICFGUncore()) {
			std::cerr << "Jaketown, Ivytown, Haswell, Broadwell-DE Server CPU is required for this tool! Program aborted" << std::endl;
			exit(EXIT_FAILURE);
		}

        if(!(m->getCPUModel() == PCM::JAKETOWN) && !(m->getCPUModel() == PCM::IVYTOWN)) {
            for(i=0;i<NUM_SAMPLES;i++) {
                if(m->getCPUModel() == PCM::SKX) {
                    getPCIeEvents(m, m->SKX_RdCur, delay_ms, mm->sample, 0, m->PRQ);
                    getPCIeEvents(m, m->SKX_RFO, delay_ms, mm->sample, 0, m->PRQ);
                    getPCIeEvents(m, m->SKX_CRd, delay_ms, mm->sample, 0, m->PRQ);
                    getPCIeEvents(m, m->SKX_DRd, delay_ms, mm->sample, 0, m->PRQ);
                    getPCIeEvents(m, m->SKX_ItoM, delay_ms, mm->sample, 0, m->PRQ);
                    getPCIeEvents(m, m->SKX_PRd, delay_ms, mm->sample, 0, m->IRQ, 1);
                    getPCIeEvents(m, m->SKX_WiL, delay_ms, mm->sample, 0, m->IRQ, 1);
                } else {
                    getPCIeEvents(m, m->PCIeRdCur, delay_ms, mm->sample);
                    getPCIeEvents(m, m->RFO, delay_ms, mm->sample, m->RFOtid);
                    getPCIeEvents(m, m->CRd, delay_ms, mm->sample);
                    getPCIeEvents(m, m->DRd, delay_ms, mm->sample);
                    getPCIeEvents(m, m->ItoM, delay_ms, mm->sample, m->ItoMtid);
                    getPCIeEvents(m, m->PRd, delay_ms, mm->sample);
                    getPCIeEvents(m, m->WiL, delay_ms, mm->sample);
                }
            }

			for(i=0; i<m->getNumSockets(); ++i) {
				mm->sample[i].total.RdBw = ((mm->sample[i].total.PCIeRdCur + mm->sample[i].total.RFO + mm->sample[i].total.CRd + mm->sample[i].total.DRd)*64ULL);
				mm->sample[i].total.WrBw = ((mm->sample[i].total.ItoM + mm->sample[i].total.RFO)*64ULL);
				mm->sample[i].miss.RdBw = ((mm->sample[i].miss.PCIeRdCur + mm->sample[i].miss.RFO + mm->sample[i].miss.CRd + mm->sample[i].miss.DRd)*64ULL);
				mm->sample[i].miss.WrBw = ((mm->sample[i].miss.ItoM + mm->sample[i].miss.RFO)*64ULL);
				mm->sample[i].hit.RdBw = ((mm->sample[i].hit.PCIeRdCur + mm->sample[i].hit.RFO + mm->sample[i].hit.CRd + mm->sample[i].hit.DRd)*64ULL);
				mm->sample[i].hit.WrBw = ((mm->sample[i].hit.ItoM + mm->sample[i].hit.RFO)*64ULL);
			}

			mm->aggregate.RdBw = ((mm->aggregate.PCIeRdCur + mm->aggregate.CRd + mm->aggregate.DRd + mm->aggregate.RFO)*64ULL);
			mm->aggregate.WrBw = ((mm->aggregate.ItoM + mm->aggregate.RFO)*64ULL);
        } else { // Ivytown and Older Architectures
            for(i=0;i<NUM_SAMPLES;i++) {
                getPCIeEvents(m, m->PCIeRdCur, delay_ms, mm->sample,0);
                getPCIeEvents(m, m->PCIeNSRd, delay_ms, mm->sample,0);
                getPCIeEvents(m, m->PCIeWiLF, delay_ms, mm->sample,0);
                getPCIeEvents(m, m->PCIeItoM, delay_ms, mm->sample,0);
                getPCIeEvents(m, m->PCIeNSWr, delay_ms, mm->sample,0);
                getPCIeEvents(m, m->PCIeNSWrF, delay_ms, mm->sample,0);
            }

            //report extrapolated read and write PCIe bandwidth per socket using the data from the sample
			for(i=0; i<m->getNumSockets(); ++i) {
				mm->sample[i].total.RdBw = ((mm->sample[i].total.PCIeRdCur+ mm->sample[i].total.PCIeNSWr)*64ULL);
				mm->sample[i].total.WrBw = ((mm->sample[i].total.PCIeWiLF+mm->sample[i].total.PCIeItoM+mm->sample[i].total.PCIeNSWr+mm->sample[i].total.PCIeNSWrF)*64ULL);
				mm->sample[i].miss.RdBw = ((mm->sample[i].miss.PCIeRdCur+ mm->sample[i].miss.PCIeNSWr)*64ULL);
				mm->sample[i].miss.WrBw = ((mm->sample[i].miss.PCIeWiLF+mm->sample[i].miss.PCIeItoM+mm->sample[i].miss.PCIeNSWr+mm->sample[i].miss.PCIeNSWrF)*64ULL);
				mm->sample[i].hit.RdBw = ((mm->sample[i].hit.PCIeRdCur+ mm->sample[i].hit.PCIeNSWr)*64ULL);
				mm->sample[i].hit.WrBw = ((mm->sample[i].hit.PCIeWiLF+mm->sample[i].hit.PCIeItoM+mm->sample[i].hit.PCIeNSWr+mm->sample[i].hit.PCIeNSWrF)*64ULL);
			}

			mm->aggregate.RdBw = ((mm->aggregate.PCIeRdCur+ mm->aggregate.PCIeNSWr)*64ULL);
			mm->aggregate.WrBw = ((mm->aggregate.PCIeWiLF+mm->aggregate.PCIeItoM+mm->aggregate.PCIeNSWr+mm->aggregate.PCIeNSWrF)*64ULL);
		}
	}

	uint64 Daemon::getTimestamp()
	{
		struct timespec now;

		clock_gettime(CLOCK_MONOTONIC_RAW, &now);

		uint64 epoch = (uint64)now.tv_sec * 1E9;
		epoch+=(uint64)now.tv_nsec;

		return epoch;
	}

	void Daemon::cleanup()
	{
		if(sharedPCMState_ != NULL) {
			int success;

			success = munmap(sharedPCMState_, sizeof(SharedPCMState));
			if (success != 0)
				std::cerr << "Failed to unmap shared memory location: " << mmapLocation_ << " (errno=" << errno << ")" << std::endl;

			close(sharedMemoryId_);
			//Delete shared memory ID file
			success = remove(mmapLocation_.c_str());
			if(success != 0)
				std::cerr << "Failed to delete mmap location: " << mmapLocation_ << " (errno=" << errno << ")" << std::endl;

            pinfo_remove();
		}
	}
}
