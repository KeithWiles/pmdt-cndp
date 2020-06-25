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

    pinfo_append(c, "{\"/pcm/header\":{");
    pinfo_append(c, "\"version\":\"%s\",", _shd->hdr.version);
    pinfo_append(c, "\"tscBegin\":%u,", _shd->hdr.tscBegin);
    pinfo_append(c, "\"tscEnd\":%u,", _shd->hdr.tscEnd);
    pinfo_append(c, "\"cyclesToGetState\":%u,", _shd->hdr.cyclesToGetPCMState);
    pinfo_append(c, "\"timestamp\":%u,", _shd->hdr.timestamp);
    pinfo_append(c, "\"socketfd\":%u,", _shd->hdr.socketfd);
    pinfo_append(c, "\"pollMs\":%u", _shd->hdr.pollMs);
    pinfo_append(c, "}}");

    return 0;
}

static int
systemInfo(void *_c)
{
    struct pinfo_client *c = _c;

    pinfo_append(c, "{\"/pcm/system\":{");
    pinfo_append(c, "\"numOfCores\":%u,", _shd->pcm.system.numOfCores);
    pinfo_append(c, "\"numOfOnlineCores\":%u,", _shd->pcm.system.numOfOnlineCores);
    pinfo_append(c, "\"numOfSockets\":%u,", _shd->pcm.system.numOfSockets);
    pinfo_append(c, "\"numOfOnlineSockets\":%u,", _shd->pcm.system.numOfOnlineSockets);
    pinfo_append(c, "\"numOfQPILinksPerSocket\":%u,", _shd->pcm.system.numOfQPILinksPerSocket);
    pinfo_append(c, "\"cpuModel\":%u", _shd->pcm.system.cpuModel);
    pinfo_append(c, "}}");

    return 0;
}

static int
pcmCore(void *_c)
{
    struct pinfo_client *c = _c;
    struct PCMCoreCounter *cc;
    int core;

    pinfo_append(c, "{\"/pcm/core\":{");
    
    if (c->params == NULL)
        pinfo_append(c, "null}");
        return 0; 

    core = atoi(c->params);
    if (core < 0 || core > (_shd->pcm.system.numOfCores * _shd->pcm.system.numOfSockets))
        pinfo_append(c, "null}");
        return 0; 

    cc = &_shd->pcm.core.cores[core];

    pinfo_append(c, "\"coreId\":%lu,", cc->coreId);
    pinfo_append(c, "\"socketId\":%lu,", cc->socketId);
    pinfo_append(c, "\"instructionsPerCycle\":%f,", cc->instructionsPerCycle);
    pinfo_append(c, "\"cycles\":%lu,", cc->cycles);
    pinfo_append(c, "\"instructionsRetired\":%lu,", cc->instructionsRetired);
    pinfo_append(c, "\"execUsage\":%f,", cc->execUsage);
    pinfo_append(c, "\"relativeFrequency\":%f,", cc->relativeFrequency);
    pinfo_append(c, "\"activeRelativeFrequency\":%f,", cc->activeRelativeFrequency);
    pinfo_append(c, "\"l3CacheMisses\":%lu,", cc->l3CacheMisses);
    pinfo_append(c, "\"l3CacheReference\":%lu,", cc->l3CacheReference);
    pinfo_append(c, "\"l2CacheMisses\":%lu,", cc->l2CacheMisses);
    pinfo_append(c, "\"l3CacheHitRatio\":%f,", isnan(cc->l3CacheHitRatio)? 0.0 : cc->l3CacheHitRatio);
    pinfo_append(c, "\"l2CacheHitRatio\":%f,", cc->l2CacheHitRatio);
    pinfo_append(c, "\"l3CacheMPI\":%f,", cc->l3CacheMPI);
    pinfo_append(c, "\"l2CacheMPI\":%f,", cc->l2CacheMPI);
    pinfo_append(c, "\"l3CacheOccupancyAvailable\":%s,", cc->l3CacheOccupancyAvailable? "true" : "false");
    pinfo_append(c, "\"l3CacheOccupancy\":%lu,", cc->l3CacheOccupancy);
    pinfo_append(c, "\"localMemoryBWAvailable\":%s,", cc->localMemoryBWAvailable? "true" : "false");
    pinfo_append(c, "\"localMemoryBW\":%lu,", cc->localMemoryBW);
    pinfo_append(c, "\"remoteMemoryBWAvailable\":%s,", cc->remoteMemoryBWAvailable? "true" : "false");
    pinfo_append(c, "\"remoteMemoryBW\":%lu,", cc->remoteMemoryBW);
    pinfo_append(c, "\"localMemoryAccesses\":%lu,", cc->localMemoryAccesses);
    pinfo_append(c, "\"remoteMemoryAccesses\":%lu,", cc->remoteMemoryAccesses);
    pinfo_append(c, "\"thermalHeadroom\":%lu", cc->thermalHeadroom);
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

    pinfo_append(c, "{\"/pcm/memory\":{");
    pinfo_append(c, "\"dramEnergyMetricsAvailable\":%s,",
        _shd->pcm.memory.dramEnergyMetricsAvailable? "true" : "false");

    mc = &_shd->pcm.memory.system;
    pinfo_append(c, "\"system\":[%f,%f,%f],",
        mc->read, mc->write, mc->total);

    pinfo_append(c, "\"sockets\":{");
    for(int i = 0; i < _shd->pcm.system.numOfSockets; i++) {
        sc = &_shd->pcm.memory.sockets[i];

        pinfo_append(c, "\"%d\":{");
        pinfo_append(c, "\"socketId\":%lu,", sc->socketId);
        pinfo_append(c, "\"numbOfChannels\":%lu,", sc->numOfChannels);
        pinfo_append(c, "\"read\":%lf,", sc->read);
        pinfo_append(c, "\"write\":%lf,", sc->write);
        pinfo_append(c, "\"partialWrite\":%lf,", sc->partialWrite);
        pinfo_append(c, "\"total\":%lf,", sc->total);
        pinfo_append(c, "\"dramEnergy\":%lf,", sc->dramEnergy);

        pinfo_append(c, "\"channels\":{");
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

    pinfo_append(c, "{\"/pcm/socket\":{");
    pinfo_append(c, "\"packageEnergyMetricsAvailable\":%s,",
        _shd->pcm.core.packageEnergyMetricsAvailable? "true" : "false");
    pinfo_append(c, "\"energyUsedBySocket\":[");
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

    pinfo_append(c, "{\"/pcm/qpi\":{");
    pinfo_append(c, "\"incomingQPITrafficMetricsAvailable\":%s,",
        _shd->pcm.qpi.incomingQPITrafficMetricsAvailable? "true" : "false");
    pinfo_append(c, "\"outgoingQPITrafficMetricsAvailable\":%s,",
        _shd->pcm.qpi.outgoingQPITrafficMetricsAvailable? "true" : "false");

    pinfo_append(c, "\"incomingTotal\":%lu,", _shd->pcm.qpi.incomingTotal);
    pinfo_append(c, "\"outgoingTotal\":%lu,", _shd->pcm.qpi.outgoingTotal);

    pinfo_append(c, "\"incoming\":[");
    for(int i = 0; i < _shd->pcm.system.numOfSockets; i++) {
        pinfo_append(c, "{\"socketID\":%d,\"total\":%lu,\"links\":[",
            i, _shd->pcm.qpi.incoming[i].total);
        for(int j = 0; j < _shd->pcm.system.numOfQPILinksPerSocket; j++) {
            pinfo_append(c, "{\"linkID\":%d,\"bytes\":%lu,\"utilization\":%lf}%s", j,
                _shd->pcm.qpi.incoming[i].links[j].bytes,
                _shd->pcm.qpi.incoming[i].links[j].utilization,
                ((j + 1) < _shd->pcm.system.numOfQPILinksPerSocket)? "," : "");
        }
        pinfo_append(c, "]");
        pinfo_append(c, "}%s", ((i + 1) < _shd->pcm.system.numOfSockets)? "," : "");
    }
    pinfo_append(c, "],");

    pinfo_append(c, "\"outgoing\":[");
    for(int i = 0; i < _shd->pcm.system.numOfSockets; i++) {
        pinfo_append(c, "{\"socketID\":%d,\"total\":%lu,\"links\":[",
            i, _shd->pcm.qpi.outgoing[i].total);
        for(int j = 0; j < _shd->pcm.system.numOfQPILinksPerSocket; j++) {
            pinfo_append(c, "{\"linkID\":%d,\"bytes\":%lu,\"utilization\":%lf}%s", j,
                _shd->pcm.qpi.outgoing[i].links[j].bytes,
                _shd->pcm.qpi.outgoing[i].links[j].utilization,
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
    pinfo_append(c, "\"PCIeRdCur\":%lu,", e->PCIeRdCur);
    pinfo_append(c, "\"PCIeNSRd\":%lu,", e->PCIeNSRd);
    pinfo_append(c, "\"PCIeWiLF\":%lu,", e->PCIeWiLF);
    pinfo_append(c, "\"PCIeItoM\":%lu,", e->PCIeItoM);
    pinfo_append(c, "\"PCIeNSWr\":%lu,", e->PCIeNSWr);
    pinfo_append(c, "\"PCIeNSWrF\":%lu,", e->PCIeNSWrF);
    pinfo_append(c, "\"RFO\":%lu,", e->RFO);
    pinfo_append(c, "\"CRd\":%lu,", e->CRd);
    pinfo_append(c, "\"DRd\":%lu,", e->DRd);
    pinfo_append(c, "\"PRd\":%lu,", e->PRd);
    pinfo_append(c, "\"WiL\":%lu,", e->WiL);
    pinfo_append(c, "\"ItoM\":%lu,", e->ItoM);
    pinfo_append(c, "\"RdBw\":%lu,", e->RdBw);
    pinfo_append(c, "\"WrBw\":%lu", e->WrBw);
}

static int
pcieEvents(void *_c)
{
    struct pinfo_client *c = _c;
    struct Sample_s *ps;

    pinfo_append(c, "{\"/pcm/pcie\":");

    pinfo_append(c, "{\"sockets\":{");
    for(int i = 0; i < _shd->pcm.system.numOfSockets; i++) {
        ps = &_shd->sample[i];

        pinfo_append(c, "\"%d\":{", i);
        pinfo_append(c, "\"total\":{");
        pcie_events(c, &ps->total);
        pinfo_append(c, "},");
        pinfo_append(c, "\"miss\":{");
        pcie_events(c, &ps->miss);
        pinfo_append(c, "},");
        pinfo_append(c, "\"hit\":{");
        pcie_events(c, &ps->hit);
        pinfo_append(c, "}}%s",
            ((i + 1) < _shd->pcm.system.numOfSockets)? "," : "");
    }
    pinfo_append(c, "},\"aggregate\":{");
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
