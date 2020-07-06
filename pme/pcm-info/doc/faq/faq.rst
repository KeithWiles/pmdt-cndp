..  SPDX-License-Identifier: BSD-3-Clause
    Copyright(c) 2010-2014 Intel Corporation.

How can I check if PCM-Daemon is running?
-----------------------------------------

Here are two ways to check if PCM is running in the background: 

* Check if the process is running in the background by searching by name for 
  the process: 

   .. code-block:: console
      ps -ef | grep pcm-info
  
      #If the process is running you will see something similar to the following:
      root       7889   7888  0 20:05 pts/3    00:00:00 sudo ./build/pcm-info -c all
      root       7890   7889  1 20:05 pts/3    00:00:05 ./build/pcm-info -c all

* Try running pcm, one way is by using the script ./run_pcm, and if you get 
  the response:
     
     Error while reading perf data. Result is 0
     Check if you run other competing Linux perf clients.

  then PCM is already running in the background. 
