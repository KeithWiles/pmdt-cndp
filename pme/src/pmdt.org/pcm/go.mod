module pmdt.org/pcm

replace pmdt.org/ttylog => ../ttylog

replace pmdt.org/encoding/raw => ../encoding/raw

replace pmdt.org/hexdump => ../hexdump

replace pmdt.org/semaphore => ../semaphore

go 1.13

require (
	pmdt.org/encoding/raw v0.0.0-00010101000000-000000000000
	pmdt.org/hexdump v0.0.0-00010101000000-000000000000
	pmdt.org/semaphore v0.0.0-00010101000000-000000000000
	pmdt.org/ttylog v0.0.0-00010101000000-000000000000
)
