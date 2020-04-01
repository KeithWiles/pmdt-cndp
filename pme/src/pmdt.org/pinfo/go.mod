module pmdt.org/pinfo

replace pmdt.org/pinfo => ./pinfo

replace pmdt.org/ttylog => ../ttylog

go 1.14

require (
	github.com/dc0d/dirwatch v0.4.3
	github.com/dc0d/retry v1.2.0 // indirect
	github.com/farmergreg/rfsnotify v0.0.0-20150112005255-94bf32bab0af
	github.com/fsnotify/fsnotify v1.4.7
	github.com/pkg/errors v0.8.1 // indirect
	golang.org/x/sys v0.0.0-20191105231009-c1f44814a5cd // indirect
	gopkg.in/fsnotify.v1 v1.4.7 // indirect
	pmdt.org/ttylog v0.0.0-00010101000000-000000000000
)
