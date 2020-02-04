module pmdt.org/perfdata

replace pmdt.org/ttylog => ../ttylog

replace pmdt.org/jevents => ../jevents

replace pmdt.org/perf => ../perf

go 1.13

require (
	github.com/shirou/gopsutil v2.19.9+incompatible
	golang.org/x/sys v0.0.0-20191009170203-06d7bd2c5f4f
	pmdt.org/jevents v0.0.0-00010101000000-000000000000
	pmdt.org/ttylog v0.0.0-00010101000000-000000000000
)
