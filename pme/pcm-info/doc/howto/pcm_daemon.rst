..  SPDX-License-Identifier: BSD-3-Clause
    Copyright(c) 2020 Intel Corporation.


PMDT PCM-Daemon User Guide
==========================

The PCM Daemon provides users with the ability to query for information about 
the system, currently including information such as core stats, and memory 
usage. 


PCM-Daemon
----------

The PCM-Daemon opens a socket with path *<runtime_directory>/pcm-info.<pid>*. 
The pid represents the pid of the PCM Daemon. For example, a client would 
connect to a socket with path */var/run/pcm-info/\*/pinfo.7089* (when the 
process ID is 7089 and is run by a root user).


Running PCM-Daemon: Pinfo
-------------------------

The following steps show how to run the PCM-Daemon and query information using 
the pinfo client python script.

#. Launch PCM-Daemon in one window.

   .. code-block:: console
      cd pme/pcm-info
      ./build/pcm-info -c all

  Or by running, 

  .. code-block:: console
      cd pme/pcm-info
      ./run_pcm 

#. In a new window, launch the pinfo client script. 

   .. code-block:: console

      pcm-info/pinfo.py

#. When connected, the script displays the following, waiting for user input.

   .. code-block:: console

      Connecting to /var/run/pcm-info/pinfo.60285
      {"version": "1.0.5", "pid": 60285, "max_output_len": 16384}
      -->

#. The user can now input commands to send across the socket, and receive the
   response.

   .. code-block:: console

      --> /
      {"/": ["/", "/pcm/info", "/pcm/header", "/pcm/system", "/pcm/core", 
      "/pcm/memory", "/pcm/socket", "/pcm/qpi", "/pcm/pcie"]}
      --> /pcm/info
      {"/pcm/info": {"version": "1.0.5", "maxbuffer": 16384, "pid": 60285}}
      --> 
