/* SPDX-License-Identifier: BSD-3-Clause
 * Copyright(c) 2020 Intel Corporation
 */

#include <stdio.h>
#include <unistd.h>
#include <stdlib.h>
#include <stdarg.h>
#include <pthread.h>
#include <sys/socket.h>
#include <sys/un.h>
#include <dlfcn.h>
#include <errno.h>
#include <bsd/string.h>
#include <math.h>

#include "common.h"
#include "pinfo_private.h"
#include "pinfo.h"
#include "system-info.h"

static SharedPCMState *_shd;

static int
headerInfo(void *_c)
{
    struct pinfo_client *c = _c;

    pinfo_append(c, "{%Q:{", c->cmd);
    pinfo_append(c, "%Q:%Q,", "version", _shd->hdr.version);
    pinfo_append(c, "%Q:%u,", "tscBegin", _shd->hdr.tscBegin);
    pinfo_append(c, "%Q:%u,", "tscEnd", _shd->hdr.tscEnd);
    pinfo_append(c, "%Q:%u,", "cyclesToGetState", _shd->hdr.cyclesToGetPCMState);
    pinfo_append(c, "%Q:%u,", "timestamp", _shd->hdr.timestamp);
    pinfo_append(c, "%Q:%u,", "socketfd", _shd->hdr.socketfd);
    pinfo_append(c, "%Q:%u", "pollMs", _shd->hdr.pollMs);
    pinfo_append(c, "}}");

    return 0;
}

static int
systemInfo(void *_c)
{
    struct pinfo_client *c = _c;

    pinfo_append(c, "{%Q:{", c->cmd);
    pinfo_append(c, "%Q:%u,", "numOfCores", _shd->pcm.system.numOfCores);
    pinfo_append(c, "%Q:%u,", "numOfOnlineCores", _shd->pcm.system.numOfOnlineCores);
    pinfo_append(c, "%Q:%u,", "numOfSockets", _shd->pcm.system.numOfSockets);
    pinfo_append(c, "%Q:%u,", "numOfOnlineSockets", _shd->pcm.system.numOfOnlineSockets);
    pinfo_append(c, "%Q:%u,", "numOfQPILinksPerSocket", _shd->pcm.system.numOfQPILinksPerSocket);
    pinfo_append(c, "%Q:%u", "cpuModel", _shd->pcm.system.cpuModel);
    pinfo_append(c, "}}");

    return 0;
}

static int
pcmCore(void *_c)
{
    struct pinfo_client *c = _c;
    struct PCMCoreCounter *cc;
    int core;

	if (c->params == NULL) {
		pinfo_append(c, "{%Q:%Q}", c->cmd, "Missing core id");
        return 0;
	}

    core = atoi(c->params);
    if (core < 0 || core > (_shd->pcm.system.numOfCores * _shd->pcm.system.numOfSockets)) {
		pinfo_append(c, "{%Q:%Q}", c->cmd, "Invalid core id");
        return 0;
	}

    cc = &_shd->pcm.core.cores[core];

    pinfo_append(c, "{%Q:{", c->cmd);
    pinfo_append(c, "%Q:%lu,", "coreId", cc->coreId);
    pinfo_append(c, "%Q:%lu,", "socketId", cc->socketId);
    pinfo_append(c, "%Q:%f,", "instructionsPerCycle", cc->instructionsPerCycle);
    pinfo_append(c, "%Q:%lu,", "cycles", cc->cycles);
    pinfo_append(c, "%Q:%lu,", "instructionsRetired", cc->instructionsRetired);
    pinfo_append(c, "%Q:%f,", "execUsage", cc->execUsage);
    pinfo_append(c, "%Q:%f,", "relativeFrequency", cc->relativeFrequency);
    pinfo_append(c, "%Q:%f,", "activeRelativeFrequency", cc->activeRelativeFrequency);
    pinfo_append(c, "%Q:%lu,", "l3CacheMisses", cc->l3CacheMisses);
    pinfo_append(c, "%Q:%lu,", "l3CacheReference", cc->l3CacheReference);
    pinfo_append(c, "%Q:%lu,", "l2CacheMisses", cc->l2CacheMisses);
    pinfo_append(c, "%Q:%f,", "l3CacheHitRatio", isnan(cc->l3CacheHitRatio)? 0.0 : cc->l3CacheHitRatio);
    pinfo_append(c, "%Q:%f,", "l2CacheHitRatio", cc->l2CacheHitRatio);
    pinfo_append(c, "%Q:%f,", "l3CacheMPI", cc->l3CacheMPI);
    pinfo_append(c, "%Q:%f,", "l2CacheMPI", cc->l2CacheMPI);
    pinfo_append(c, "%Q:%s,", "l3CacheOccupancyAvailable", cc->l3CacheOccupancyAvailable? "true" : "false");
    pinfo_append(c, "%Q:%lu,", "l3CacheOccupancy", cc->l3CacheOccupancy);
    pinfo_append(c, "%Q:%s,", "localMemoryBWAvailable", cc->localMemoryBWAvailable? "true" : "false");
    pinfo_append(c, "%Q:%lu,", "localMemoryBW", cc->localMemoryBW);
    pinfo_append(c, "%Q:%s,", "remoteMemoryBWAvailable", cc->remoteMemoryBWAvailable? "true" : "false");
    pinfo_append(c, "%Q:%lu,", "remoteMemoryBW", cc->remoteMemoryBW);
    pinfo_append(c, "%Q:%lu,", "localMemoryAccesses", cc->localMemoryAccesses);
    pinfo_append(c, "%Q:%lu,", "remoteMemoryAccesses", cc->remoteMemoryAccesses);
    pinfo_append(c, "%Q:%lu,", "thermalHeadroom", cc->thermalHeadroom);
    pinfo_append(c,"%Q:%lu,", "branches", cc->branches);
    pinfo_append(c,"%Q:%lu", "branchMispredicts", cc->branchMispredicts);
    pinfo_append(c, "}}");

    return 0;
}

static int
memoryCounters(void *_c)
{
    struct pinfo_client *c = _c;
    struct PCMMemorySystemCounter *mc;
    struct PCMMemorySocketCounter *sc;
    struct PCMMemoryChannelCounter *cc;

    pinfo_append(c, "{%Q:{", c->cmd);
    pinfo_append(c, "%Q:%s,", "dramEnergyMetricsAvailable",
        _shd->pcm.memory.dramEnergyMetricsAvailable? "true" : "false");

    mc = &_shd->pcm.memory.system;
    pinfo_append(c, "%Q:[%f,%f,%f],", "system", mc->read, mc->write, mc->total);

    pinfo_append(c, "%Q:{", "sockets");
    for(int i = 0; i < _shd->pcm.system.numOfSockets; i++) {
        sc = &_shd->pcm.memory.sockets[i];

        pinfo_append(c, "\"%d\":{");
        pinfo_append(c, "%Q:%lu,", "socketId", sc->socketId);
        pinfo_append(c, "%Q:%lu,", "numbOfChannels", sc->numOfChannels);
        pinfo_append(c, "%Q:%lf,", "read", sc->read);
        pinfo_append(c, "%Q:%lf,", "write", sc->write);
        pinfo_append(c, "%Q:%lf,", "partialWrite", sc->partialWrite);
        pinfo_append(c, "%Q:%lf,", "total", sc->total);
        pinfo_append(c, "%Q:%lf,", "dramEnergy", sc->dramEnergy);

        pinfo_append(c, "%Q:{", "channels");
        for(int j = 0; j < MEMORY_MAX_IMC_CHANNELS; j++) {
            cc = &sc->channels[j];

            pinfo_append(c, "\"%d\":[%f,%f,%f]%s", j,
                cc->read, cc->write, cc->total,
                ((j + 1) < MEMORY_MAX_IMC_CHANNELS)? "," : "");
        }
        pinfo_append(c, "}}%s",
            ((i + 1) < _shd->pcm.system.numOfSockets)? "," : "");
    }
    pinfo_append(c, "}}}");
    return 0;
}

static int
pcmCoreEnergyAvailable(void *_c)
{
    struct pinfo_client *c = _c;

    pinfo_append(c, "{%Q:{", c->cmd);
    pinfo_append(c, "%Q:%s,", "packageEnergyMetricsAvailable",
        _shd->pcm.core.packageEnergyMetricsAvailable? "true" : "false");
    pinfo_append(c, "%Q:[", "energyUsedBySocket");
    for(int i = 0; i < _shd->pcm.system.numOfSockets; i++) {
        pinfo_append(c, "%lf%s", _shd->pcm.core.energyUsedBySockets[i],
            ((i + 1) < _shd->pcm.system.numOfSockets)? "," : "");
    }
    pinfo_append(c, "]}}");
    return 0;
}

static int
qpiCounters(void *_c)
{
    struct pinfo_client *c = _c;

    pinfo_append(c, "{%Q:{", c->cmd);
    pinfo_append(c, "%Q:%s,", "incomingQPITrafficMetricsAvailable",
        _shd->pcm.qpi.incomingQPITrafficMetricsAvailable? "true" : "false");
    pinfo_append(c, "%Q:%s,", "outgoingQPITrafficMetricsAvailable",
        _shd->pcm.qpi.outgoingQPITrafficMetricsAvailable? "true" : "false");

    pinfo_append(c, "%Q:%lu,", "incomingTotal", _shd->pcm.qpi.incomingTotal);
    pinfo_append(c, "%Q:%lu,", "outgoingTotal", _shd->pcm.qpi.outgoingTotal);

    pinfo_append(c, "%Q:[", "incoming");
    for(int i = 0; i < _shd->pcm.system.numOfSockets; i++) {
        pinfo_append(c, "{%Q:%d,%Q:%lu,%Q:[",
            "socketID", i, "total", _shd->pcm.qpi.incoming[i].total, "links");
        for(int j = 0; j < _shd->pcm.system.numOfQPILinksPerSocket; j++) {
            pinfo_append(c, "{%Q:%d,%Q:%lu,%Q:%lf}%s",
                "linkID", i,
                "bytes", _shd->pcm.qpi.incoming[i].links[j].bytes,
                "utilization", _shd->pcm.qpi.incoming[i].links[j].utilization,
                ((j + 1) < _shd->pcm.system.numOfQPILinksPerSocket)? "," : "");
        }
        pinfo_append(c, "]");
        pinfo_append(c, "}%s", ((i + 1) < _shd->pcm.system.numOfSockets)? "," : "");
    }
    pinfo_append(c, "],");

    pinfo_append(c, "%Q:[", "outgoing");
    for(int i = 0; i < _shd->pcm.system.numOfSockets; i++) {
        pinfo_append(c, "{%Q:%d,%Q:%lu,%Q:[",
            "socketID", i, "total", _shd->pcm.qpi.outgoing[i].total, "links");
        for(int j = 0; j < _shd->pcm.system.numOfQPILinksPerSocket; j++) {
            pinfo_append(c, "{%Q:%d,%Q:%lu,%Q:%lf}%s",
                "linkID", j,
                "bytes", _shd->pcm.qpi.outgoing[i].links[j].bytes,
                "utilization", _shd->pcm.qpi.outgoing[i].links[j].utilization,
                ((j + 1) < _shd->pcm.system.numOfQPILinksPerSocket)? "," : "");
        }
        pinfo_append(c, "]");
        pinfo_append(c, "}%s", ((i + 1) < _shd->pcm.system.numOfSockets)? "," : "");
    }

    pinfo_append(c, "]}}");
    return 0;
}

static void
pcie_events(struct pinfo_client *c, struct PCIeEvents *e)
{
    pinfo_append(c, "%Q:%lu,", "PCIeRdCur", e->PCIeRdCur);
    pinfo_append(c, "%Q:%lu,", "PCIeNSRd", e->PCIeNSRd);
    pinfo_append(c, "%Q:%lu,", "PCIeWiLF", e->PCIeWiLF);
    pinfo_append(c, "%Q:%lu,", "PCIeItoM", e->PCIeItoM);
    pinfo_append(c, "%Q:%lu,", "PCIeNSWr", e->PCIeNSWr);
    pinfo_append(c, "%Q:%lu,", "PCIeNSWrF", e->PCIeNSWrF);
    pinfo_append(c, "%Q:%lu,", "RFO", e->RFO);
    pinfo_append(c, "%Q:%lu,", "CRd", e->CRd);
    pinfo_append(c, "%Q:%lu,", "DRd", e->DRd);
    pinfo_append(c, "%Q:%lu,", "PRd", e->PRd);
    pinfo_append(c, "%Q:%lu,", "WiL", e->WiL);
    pinfo_append(c, "%Q:%lu,", "ItoM", e->ItoM);
    pinfo_append(c, "%Q:%lu,", "RdBw", e->RdBw);
    pinfo_append(c, "%Q:%lu", "WrBw", e->WrBw);
}

static int
pcieEvents(void *_c)
{
    struct pinfo_client *c = _c;
    struct Sample_s *ps;

    pinfo_append(c, "{%Q:", c->cmd);

    pinfo_append(c, "{%Q:{", "sockets");
    for(int i = 0; i < _shd->pcm.system.numOfSockets; i++) {
        ps = &_shd->sample[i];

        pinfo_append(c, "\"%d\":{", i);
        pinfo_append(c, "%Q:{", "total");
        pcie_events(c, &ps->total);
        pinfo_append(c, "},");
        pinfo_append(c, "%Q:{", "miss");
        pcie_events(c, &ps->miss);
        pinfo_append(c, "},");
        pinfo_append(c, "%Q:{", "hit");
        pcie_events(c, &ps->hit);
        pinfo_append(c, "}}%s",
            ((i + 1) < _shd->pcm.system.numOfSockets)? "," : "");
    }
    pinfo_append(c, "},%Q:{", "aggregate");
    pcie_events(c, &_shd->aggregate);
    pinfo_append(c, "}}}");
    return 0;
}

int
setupInfo(void *_s)
{
    _shd = _s;

    pinfo_register("/pcm/header", headerInfo);
    pinfo_register("/pcm/system", systemInfo);
    pinfo_register("/pcm/core", pcmCore);
    pinfo_register("/pcm/memory", memoryCounters);
    pinfo_register("/pcm/socket", pcmCoreEnergyAvailable);
    pinfo_register("/pcm/qpi", qpiCounters);
    pinfo_register("/pcm/pcie", pcieEvents);

    return 0;
}
