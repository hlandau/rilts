# relip: VCS Repository Licence Prover Utilities

This is a tool written in Go which can prove the licence of repositories which
use the [RILTS](..) specification to track licencing information.

Currently, it supports only Git repositories.

## How to use

Make sure Go is installed.

```
# If you don't have a GOPATH set (you're not used to building Go programs):
$ mkdir some_temp_dir
$ cd some_temp_dir
$ export GOPATH=`pwd`

# Then, or if you already have a GOPATH set:
$ go get github.com/hlandau/rilts/relip
$ $GOPATH/bin/relip --help

# For example, to check that all contributions are MIT or Apache2-licenced:
$ $GOPATH/bin/relip -L MIT -L Apache2 -B master path/to/repo
$ echo $?
0             # <-- indicates licencing is valid
```

## Licence

MIT licenced. If you want proof of that, use `relip`!
