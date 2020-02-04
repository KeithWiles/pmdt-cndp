module pmdt.org/devgroup

replace pmdt.org/devbind => ../devbind

replace pmdt.org/ttylog => ../ttylog

go 1.13

require (
	github.com/BurntSushi/toml v0.3.1
	github.com/jessevdk/go-flags v1.4.0
	pmdt.org/devbind v0.0.0-00010101000000-000000000000
)
