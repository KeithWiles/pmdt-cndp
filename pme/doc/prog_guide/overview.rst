..  SPDX-License-Identifier: BSD-3-Clause
    Copyright(c) 2010-2014 Intel Corporation.

**Part 1: Architecture Overview**

Overview
========

This section gives a global overview of the architecture of Data Plane Development Kit (DPDK).

The main goal of the DPDK is to provide a simple,
complete framework for fast packet processing in data plane applications.
Users may use the code to understand some of the techniques employed,
to build upon for prototyping or to add their own protocol stacks.
Alternative ecosystem options that use the DPDK are available.

The framework creates a set of libraries for specific environments
through the creation of an Environment Abstraction Layer (EAL),
which may be specific to a mode of the Intel® architecture (32-bit or 64-bit),
Linux* user space compilers or a specific platform.
These environments are created through the use of make files and configuration files.
Once the EAL library is created, the user may link with the library to create their own applications.
Other libraries, outside of EAL, including the Hash,
Longest Prefix Match (LPM) and rings libraries are also provided.
Sample applications are provided to help show the user how to use various features of the DPDK.

The DPDK implements a run to completion model for packet processing,
where all resources must be allocated prior to calling Data Plane applications,
running as execution units on logical processing cores.
The model does not support a scheduler and all devices are accessed by polling.
The primary reason for not using interrupts is the performance overhead imposed by interrupt processing.

In addition to the run-to-completion model,
a pipeline model may also be used by passing packets or messages between cores via the rings.
This allows work to be performed in stages and may allow more efficient use of code on cores.

Development Environment
-----------------------

The DPDK project installation requires Linux and the associated toolchain,
such as one or more compilers, assembler, make utility,
editor and various libraries to create the DPDK components and libraries.

Once these libraries are created for the specific environment and architecture,
they may then be used to create the user's data plane application.

