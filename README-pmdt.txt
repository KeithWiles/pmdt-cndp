Look at the pme/README-pme.txt and pcm/README-pcm.txt files for more limited support.

The PMDT tool is written in Go and a C++ daemon to collect system information. The tool
is used to collect information about DPDK applications and display that data in a readable
format plus analyze the data and suggest changes to the system or application to improve 
performance.

The tool uses the process info library (defined by Bruce Richardson) to expose internal
DPDK data in a simple to use and decodable format JSON. The tool contains a patch for DPDK to add
the process info support and rebuilding the application with these changes adds the
process info support. The process info takes simple string commands and produces JSON formatted
strings on a local domain socket which is a filesystem based socket.

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

