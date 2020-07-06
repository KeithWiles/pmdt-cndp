..  SPDX-License-Identifier: BSD-3-Clause
    Copyright(c) 2010-2014 Intel Corporation.

**Part 1: Architecture Overview**

Overview
========

This section gives a global overview of the architecture of Performance Monitor Data Toolkit (PMDT).

The main goal of the PMDT is to provide a simple, lightweight tool
to monitor performance of DPDK (Data Plane Development Kit) applications.
Users may use the code to understand some of the techniques employed,
to build upon for prototyping or to add their own values to monitor.



Development Environment
-----------------------

The PMDT project installation requires Linux and the associated toolchain,
such as one or more compilers, assembler, make utility, editor and various 
libraries to create the PMDT tools.

The tool will attempt to locate DPDK applications if these application are using
the DPDK 20.05 telemetry library to gather information about the DPDK apps.
The telemetry library creates a socket at location /var/run/dpdk/rte/*

.. Note:: 

   The telemetry socket may be located in a differnt directory depending on the 
   value of "--file-prefix=" set with the DPDK application. The default value is 
   rte (as shown above). Example: if a user sets the DPDK parameter to 
   --file-prefix=foo then the DPDK telemetry socket location for that application 
   is /var/run/dpdk/foo/*
