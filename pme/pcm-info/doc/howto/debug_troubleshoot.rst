..  SPDX-License-Identifier: BSD-3-Clause
    Copyright(c) 2018 Intel Corporation.

Debug & Troubleshoot guide
==========================

PMDT is designed to work with various DPDK applications that have been designed 
to have simple or complex pipeline processing stages making use of single or 
multiple threads. Applications can use poll mode hardware devices which helps in
offloading CPU cycles too. It is common to find solutions designed with

* single or multiple primary processes

* single primary and single secondary

* single primary and multiple secondaries

In all the above cases, it is tedious to isolate, debug, and understand various
behaviors which occur randomly or periodically. The goal of the guide is to
consolidate a few commonly seen issues for reference. Then, isolate to identify
the root cause through step by step debug at various stages.


Is the tool not displaying any data? 
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

TODO: Add troubleshooting advice. 


Is the tool not finding your DPDK application? 
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

TODO: Add troubleshooting advice. 
