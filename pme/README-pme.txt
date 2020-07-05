                   PMDT (Performance Monitor Data Toolkit)

This directory contains the 'pme' performance monitor code written in Go.

The top level directory contains a script to help run the tool and you can do
./pme_run
or
./pme_run -p N

Where N is the /dev/pts/N device, the script uses /dev/pts/0 for some crash
reporting and will need to change if you use some other pts device via an xterm.

To get a screen of panels to view. The pme tool only needs an xterm to run as
long as it supports VT100 ANSI escape codes and color is suggested for a better
view of the data.

The tool contains a number of go packages to support the tool. Packages from Go
are used fairly lighty or at least the basic system of tools. The primary display
in this semi-graphic format is provided by gdamore/tcell and rivo/tview packages
hosted on github.com.

The tool will attempt to locate DPDK applications if these application are using
the DPDK 20.05 telemetry library to gather information about the DPDK apps.
The telemetry library creates a socket at location /var/run/dpdk/rte/*

The pme tools also needs access to the PMU/MSR registers and needs to run as sudo application.
The 'pme_run' script handles building and executing the Go application.

Read the setup-build.txt file for more install instructions in the PME directory.

Thanks
