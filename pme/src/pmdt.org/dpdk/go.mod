module pmdt.org/dpdk

replace pmdt.org/ttylog => ../ttylog

go 1.13

require (
	github.com/fsnotify/fsnotify v1.4.7
	github.com/shirou/gopsutil v2.19.9+incompatible
	golang.org/x/sys v0.0.0-20191105231009-c1f44814a5cd // indirect
	pmdt.org/ttylog v0.0.0-00010101000000-000000000000
)
