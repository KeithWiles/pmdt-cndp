..  SPDX-License-Identifier: BSD-3-Clause
    Copyright(c) 2020 Intel Corporation.


PMDT PCM-Daemon User Guide
==========================

The PCM Daemon provides users with the ability to query for information about 
the system, currently including information such as ethdev stats, and ethdev 
port list. 


PCM-Daemon Interface
--------------------

The :doc:`../prog_guide/pcm_daemon` opens a socket with path
*<runtime_directory>/pcm-info.<pid>*. The pid represents the
telemetry version, the latest is v2. For example, a client would connect to a
socket with path  */var/run/dpdk/\*/dpdk_telemetry.v2* (when the primary process
is run by a root user).


Running PCM-Daemon
------------------

The following steps show how to run the PCM-Daemon and query information using 
the pinfo client python script.

#. Launch PCM-Daemon.

   .. code-block:: console

      ./build/pcm-info -c all

#. In a new window, launch the pinfo client script. This requires root 
privileges.

   .. code-block:: console

      python usertools/pinfo.py

#. When connected, the script displays the following, waiting for user input.

   .. code-block:: console

      Connecting to /var/run/
      {"version": "DPDK 20.05.0-rc0", "pid": 60285, "max_output_len": 16384}
      -->

#. The user can now input commands to send across the socket, and receive the
   response.

   .. code-block:: console

      --> /
      {"/": ["/", "/eal/app_params", "/eal/params", "/ethdev/list",
      "/ethdev/link_status", "/ethdev/xstats", "/help", "/info"]}
      --> /
      {"": []}

