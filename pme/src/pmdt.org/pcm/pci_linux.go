// SPDX-License-Identifier: BSD-3-Clause
// Copyright(c) 2019-2020 Intel Corporation

// +build !linux

package pcm

/*
#include <stdint.h>

#include "pcm_class.h"
#include "counter_types.h"

uint32_t
getMaxNumOfCBoxes(void)
{
	//  on other supported CPUs there is one CBox per physical core.
	//  This calculation will get us the number of physical cores per socket
	//  which is the expected qvalue to be returned.
	return (uint32_t)_pcm.num_phys_cores_per_socket;
}

void
programCboOpcodeFilter(const uint32_t opc0, uint32_t *filter, const uint32_t nc_, const uint32_t opc1)
{
	if(JAKETOWN == _pcm.cpu_model)
		filter[0] = JKT_CBO_MSR_PMON_BOX_FILTER_OPC(opc0);
	else if(IVYTOWN == _pcm.cpu_model || HASWELLX == _pcm.cpu_model ||
			BDX_DE == _pcm.cpu_model || BDX == _pcm.cpu_model)
		filter[1] = IVTHSX_CBO_MSR_PMON_BOX_FILTER1_OPC(opc0);
	else if(SKX == _pcm.cpu_model)
	{
		filter[1] = SKX_CHA_MSR_PMON_BOX_FILTER1_OPC0(opc0) +
				 SKX_CHA_MSR_PMON_BOX_FILTER1_OPC1(opc1) +
				 SKX_CHA_MSR_PMON_BOX_FILTER1_REM(1) +
				 SKX_CHA_MSR_PMON_BOX_FILTER1_LOC(1) +
				 SKX_CHA_MSR_PMON_BOX_FILTER1_NM(1) +
				 SKX_CHA_MSR_PMON_BOX_FILTER1_NOT_NM(1) +
				 (nc_?SKX_CHA_MSR_PMON_BOX_FILTER1_NC(1):0ULL);
	}
}

void
programCbo(const uint64_t * events, const uint32_t opCode, const uint32_t nc_,
	const uint32_t tid_)
{
	uint32_t filters[2];

    for (int32_t i = 0; i < _pcm.num_sockets; ++i) {
        uint32_t refCore = _pcm.socketRefCore[i];

        for(uint32_t cbo = 0; cbo < getMaxNumOfCBoxes(); ++cbo) {
            // freeze enable
            *cboPMUs[i][cbo].unitControl = UNC_PMON_UNIT_CTL_FRZ_EN;
            // freeze
            *cboPMUs[i][cbo].unitControl = UNC_PMON_UNIT_CTL_FRZ_EN + UNC_PMON_UNIT_CTL_FRZ;

            programCboOpcodeFilter(opCode, filters, nc_, 0);

			if((HASWELLX == _pcm.cpu_model || BDX_DE == _pcm.cpu_model ||
					BDX == _pcm.cpu_model) && tid_ != 0)
                filters[0] = tid_;

            for (int c = 0; c < 4; ++c) {
                *cboPMUs[i][cbo].counterControl[c] = CBO_MSR_PMON_CTL_EN;
                *cboPMUs[i][cbo].counterControl[c] = CBO_MSR_PMON_CTL_EN + events[c];
            }

            // reset counter values
            *cboPMUs[i][cbo].unitControl = UNC_PMON_UNIT_CTL_FRZ_EN + UNC_PMON_UNIT_CTL_FRZ + UNC_PMON_UNIT_CTL_RST_COUNTERS;

            // unfreeze counters
            *cboPMUs[i][cbo].unitControl = UNC_PMON_UNIT_CTL_FRZ_EN;

            for (int c = 0; c < 4; ++c)
                *cboPMUs[i][cbo].counterValue[c] = 0;
        }
    }
}

void
programPCIeCounters(const enum PCIeEventCode event_, const uint32_t tid_,
	const uint32_t miss_, const uint32_t q_, const uint32_t nc_)
{
    const uint32_t opCode = (uint32_t)event_;

    uint64_t event0 = 0;
    // TOR_INSERTS.OPCODE event
    if (SKX == _pcm.cpu_model) {
        uint64_t umask = 0;
        switch (q_)
        {
        case PRQ:
            umask |= (uint64_t)(SKX_CHA_TOR_INSERTS_UMASK_PRQ(1));
            break;
        case IRQ:
            umask |= (uint64_t)(SKX_CHA_TOR_INSERTS_UMASK_IRQ(1));
            break;
        }
        switch (miss_)
        {
        case 0:
            umask |= (uint64_t)(SKX_CHA_TOR_INSERTS_UMASK_HIT(1));
            umask |= (uint64_t)(SKX_CHA_TOR_INSERTS_UMASK_MISS(1));
            break;
        case 1:
            umask |= (uint64_t)(SKX_CHA_TOR_INSERTS_UMASK_MISS(1));
            break;
        }

        event0 = CBO_MSR_PMON_CTL_EVENT(0x35) + CBO_MSR_PMON_CTL_UMASK(umask);
    }
    else
        event0 = CBO_MSR_PMON_CTL_EVENT(0x35) + (CBO_MSR_PMON_CTL_UMASK(1) | (miss_ ? CBO_MSR_PMON_CTL_UMASK(0x3) : 0ULL)) + (tid_ ? CBO_MSR_PMON_CTL_TID_EN : 0ULL);

    uint64_t events[4] = { event0, 0, 0, 0 };
    programCbo(events, opCode, nc_, tid_);
}
*/
import "C"

import (
)
