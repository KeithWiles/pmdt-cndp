module pmdt.org/pme

replace pmdt.org/asciichart => ../asciichart

replace pmdt.org/colorize => ../colorize

replace pmdt.org/devbind => ../devbind

replace pmdt.org/dpdk => ../dpdk

replace pmdt.org/ttylog => ../ttylog

replace pmdt.org/perfdata => ../perfdata

replace pmdt.org/jevents => ../jevents

replace pmdt.org/perf => ../perf

replace pmdt.org/taborder => ../taborder

replace pmdt.org/intelpbf => ../intel-pbf

replace pmdt.org/etimers => ../etimers

replace pmdt.org/profiles => ../profiles

replace pmdt.org/pcm => ../pcm

replace pmdt.org/encoding/raw => ../encoding/raw

replace pmdt.org/hexdump => ../hexdump

replace pmdt.org/semaphore => ../semaphore

replace pmdt.org/graphdata => ../graphdata

replace pmdt.org/pinfo => ../pinfo

go 1.13

require (
	github.com/StackExchange/wmi v0.0.0-20190523213315-cbe66965904d // indirect
	github.com/gdamore/tcell v1.3.0
	github.com/go-ole/go-ole v1.2.4 // indirect
	github.com/jessevdk/go-flags v1.4.0
	github.com/rivo/tview v0.0.0-20200204110323-ae3d8cac5e4b
	github.com/shirou/gopsutil v2.20.1+incompatible
	github.com/stretchr/testify v1.5.1 // indirect
	pmdt.org/colorize v0.0.0-00010101000000-000000000000
	pmdt.org/devbind v0.0.0-00010101000000-000000000000
	pmdt.org/etimers v0.0.0-00010101000000-000000000000
	pmdt.org/graphdata v0.0.0-00010101000000-000000000000
	pmdt.org/intelpbf v0.0.0-00010101000000-000000000000
	pmdt.org/pcm v0.0.0-00010101000000-000000000000
	pmdt.org/pinfo v0.0.0-00010101000000-000000000000
	pmdt.org/profiles v0.0.0-00010101000000-000000000000
	pmdt.org/taborder v0.0.0-00010101000000-000000000000
	pmdt.org/ttylog v0.0.0-00010101000000-000000000000
)
