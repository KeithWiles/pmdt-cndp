..  SPDX-License-Identifier: BSD-3-Clause
    Copyright(c) 2020 Intel Corporation.


PMDT Pinfo User Guide
=====================

The pinfo script provides users with the ability to query the PCM-Daemon for
information about the system, currently including information such as ethdev 
stats, and ethdev port list.

.. Note::

   This script is meant as a way to verify the PMDT tool and can be used for 
   troubleshooting.


PCM-Daemon Interface
--------------------

The :doc:`../prog_guide/telemetry_lib` opens a socket with path
*<runtime_directory>/dpdk_telemetry.<version>*. The version represents the
telemetry version, the latest is v2. For example, a client would connect to a
socket with path  */var/run/dpdk/\*/dpdk_telemetry.v2* (when the primary process
is run by a root user).


Running Pinfo
-------------

The following steps show how to run an application with telemetry support,
and query information using the telemetry client python script.

#. Launch testpmd as the primary application with telemetry.

   .. code-block:: console

      ./app/dpdk-testpmd

#. Launch the telemetry client script.

   .. code-block:: console

      python usertools/dpdk-telemetry.py

#. When connected, the script displays the following, waiting for user input.

   .. code-block:: console

      Connecting to /var/run/dpdk/rte/dpdk_telemetry.v2
      {"version": "DPDK 20.05.0-rc0", "pid": 60285, "max_output_len": 16384}
      -->

#. The user can now input commands to send across the socket, and receive the
   response.

   .. code-block:: console

      --> /
      {"/": ["/", "/eal/app_params", "/eal/params", "/ethdev/list",
      "/ethdev/link_status", "/ethdev/xstats", "/help", "/info"]}
      --> /ethdev/list
      {"/ethdev/list": [0, 1]}

