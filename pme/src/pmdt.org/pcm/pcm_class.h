// SPDX-License-Identifier: BSD-3-Clause
// Copyright(c) 2019-2020 Intel Corporation

#ifndef _PCM_CLASS_H
#define _PCM_CLASS_H

#include <stdint.h>

struct pcm_class {
    int msr_fd;
    
    int32_t cpu_family;
    int32_t cpu_model, original_cpu_model;
    int32_t cpu_stepping;
    int64_t cpu_microcode_level;
    int32_t max_cpuid;
    int32_t threads_per_core;
    int32_t num_cores;
    int32_t num_sockets;
    int32_t num_phys_cores_per_socket;
    int32_t num_online_cores;
    int32_t num_online_sockets;
    uint32_t core_gen_counter_num_max;
    uint32_t core_gen_counter_num_used;
    uint32_t core_gen_counter_width;
    uint32_t core_fixed_counter_num_max;
    uint32_t core_fixed_counter_num_used;
    uint32_t core_fixed_counter_width;
    uint32_t uncore_gen_counter_num_max;
    uint32_t uncore_gen_counter_num_used;
    uint32_t uncore_gen_counter_width;
    uint32_t uncore_fixed_counter_num_max;
    uint32_t uncore_fixed_counter_num_used;
    uint32_t uncore_fixed_counter_width;
    uint32_t perfmon_version;
    int32_t perfmon_config_anythread;
    uint64_t nominal_frequency;
    uint64_t max_qpi_speed; // in GBytes/second
    uint32_t L3ScalingFactor;
    int32_t pkgThermalSpecPower, pkgMinimumPower, pkgMaximumPower;

    int32_t socketRefCore[8];
};

extern struct pcm_class _pcm;

#endif
