Look at the pme/readme.txt file for more information.

The PMDT tool is written in Go and a C++ daemon to collect system information. The tool
is used to collect information about DPDK applications and display that data in a readable
format plus analyze the data and suggest changes to the system or application to improve 
performance.

The tool uses the new DPDK 20.05 telemetry library to expose internal
DPDK data in a simple to use and decodable format JSON.

PMDT (Performance Monitor Development Toolkit) is in two applications. One
application (pme) is the golang code to display metric information. The second
is (pcm-info) a daemon that runs in the background gathering system level metrics
and providing the metric data via a local domain socket.

pmdt/pme/pcm-info:
	A C++ daemon application running collecting PCM values. These values are
	accessed via a local domain socket located at /var/run/pcm-info/

pmdt/pme:
	The Performance monitor application written in Go.

Patches to the PMDT project should be sent to pmdt@dpdk.org 
All patches must follow the DPDK.org patch submission process.

Thanks

