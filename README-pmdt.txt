Look at the pme/readme.txt file for more information.

PMDT (Performance Monitor Development Toolkit) is in two applications. One
application (pme) is the golang code to display metric information. The second
is (pcm-info) a daemon that runs in the background gathering system level metrics
and placing that information in a shared memory region accessed by the PME tool.

pmdt/pme/pcm-info:
	A C++ daemon application running collecting PCM values. These values are
	accessed via a local domain socket located at /var/run/pcm-info/

pmdt/pme:
	The Performance monitor application written in Go.

Patches to the PMDT project should be sent to pmdt@dpdk.org 
All patches must follow the DPDK.org patch submission process.

Thanks

