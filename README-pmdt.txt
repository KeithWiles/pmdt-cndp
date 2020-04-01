Look at the pme/readme.txt file for more information.

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

