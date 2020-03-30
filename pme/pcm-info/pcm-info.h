/* SPDX-License-Identifier: BSD-3-Clause
 * Copyright(c) 2020 Intel Corporation
 */

#ifndef PCM_INFO_H_
#define PCM_INFO_H_

#include <sys/types.h>
#include <map>
#include <string>
#include <grp.h>

#include "common.h"
#include "pcm.h"

namespace PCMDaemon {

	enum Mode { DIFFERENCE, ABSOLUTE };

	class Daemon {
	public:
		Daemon(int argc, char *argv[]);
		~Daemon();
		int run();
	private:
		void setupPCM();
		void checkAccessAndProgramPCM();
		void readApplicationArguments(int argc, char *argv[]);
		void printExampleUsageAndExit(char *argv[]);
		void setupSharedMemory();
		gid_t resolveGroupName(const std::string& groupName);
		void getPCMCounters();
		void updatePCMState(SystemCounterState* systemStates, std::vector<SocketCounterState>* socketStates, std::vector<CoreCounterState>* coreStates);
		void swapPCMBeforeAfterState();
		void getPCMSystem();
		void getPCMCore();
		void getPCMMemory();
		void getPCMQPI();
		void getPCIeEvents(PCM *m, PCM::PCIeEventCode opcode, uint32 delay_ms, Sample_t *sample, const uint32 tid, const uint32 q, const uint32 nc);
		void getPCIeCounters();
		uint64 getTimestamp();
		static void cleanup();

		bool debugMode_;
		uint32 pollIntervalMs_;
		std::string groupName_;
		Mode mode_;
		static std::string mmapLocation_;

		static int sharedMemoryId_;
		static SharedPCMState* sharedPCMState_;
		PCM* pcmInstance_;
		std::map<std::string, uint32> subscribers_;
		std::vector<std::string> allowedSubscribers_;

		//Data for core, socket and system state
		uint64 collectionTimeBefore_, collectionTimeAfter_;
		std::vector<CoreCounterState> coreStatesBefore_, coreStatesAfter_;
		std::vector<SocketCounterState> socketStatesBefore_, socketStatesAfter_;
		SystemCounterState systemStatesBefore_, systemStatesForQPIBefore_, systemStatesAfter_;
		ServerUncorePowerState* serverUncorePowerStatesBefore_;
		ServerUncorePowerState* serverUncorePowerStatesAfter_;
	};
}

#endif /* PCM_INFO_H_ */
