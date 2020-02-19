Look at the pme/README and pme/readme.txt files for more limited support.

PMDT (Performance Monitor Development Toolkit) is in two applications. One
application (pme) is the golang code to display metric information. The second
is (pcm) a daemon that runs in the background gathering system level metrics
and placing that information in a shared memory region accessed by the PME tool.

perfmon/pcm is the PCM daemon application running collecting the value into shared memory.
perfmon/pme is the Performance monitor application written in Go.

Read the pcm/README.txt file for more information about build and running the pcm-daemon.

Reporting issues and bugs should be sent to dev@dpdk.org with a subject tag of '[PMDT]'.
Patches to the PMDT project should be sent to dev@dpdk.org with subject tag of '[PMDT PATCH]', unless a specific email address is created i.e. pmdt@dpdk.org.
All patches must follow the DPDK.org patch submission process.

Thanks

