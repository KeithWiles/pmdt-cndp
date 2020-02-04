// SPDX-License-Identifier: BSD-3-Clause
// Copyright(c) 2019-2020 Intel Corporation

package dpdk

import (
	tlog "pmdt.org/ttylog"
	"strconv"
	"strings"
)

type dpdkOpts struct {
	opts []string
	arg  bool
}

// Option for the command lines
type Option struct {
	flg string
	val string
}

// Options list
type Options []*Option

// CmdLineData values parsed up into flags/option
type CmdLineData struct {
	cmdline    string
	optList    []string
	executable string
	dpdkOpts   Options
	appOpts    Options
}

var dpdkOptions []dpdkOpts

func init() {
	dpdkOptions = []dpdkOpts{
		{[]string{"-b", "--pci-blacklist"}, true},
		{[]string{"-c"}, true},
		{[]string{"-s"}, true},
		{[]string{"-d"}, true},
		{[]string{"-h", "--help"}, true},
		{[]string{"-l", "--lcores"}, true},
		{[]string{"-S", "--lcores"}, true},
		{[]string{"-m", "--lcores"}, true},
		{[]string{"-n", "--lcores"}, true},
		{[]string{"-r", "--lcores"}, true},
		{[]string{"-v", "--lcores"}, false},
		{[]string{"-w", "--pci-whitelist"}, true},

		{[]string{"--base-virtaddr"}, true},
		{[]string{"--create-uio-dev"}, false},
		{[]string{"--file-prefix"}, true},
		{[]string{"--huge-dir"}, true},
		{[]string{"--huge-unlink"}, false},
		{[]string{"--iova-mode"}, true},
		{[]string{"--iova-mode"}, true},
		{[]string{"--log-level"}, true},
		{[]string{"--master-lcore"}, true},
		{[]string{"--mbuf-pool-opts-name"}, true},
		{[]string{"--no-hpet"}, false},
		{[]string{"--no-huge"}, false},
		{[]string{"--no-pci"}, false},
		{[]string{"--no-shconf"}, false},
		{[]string{"--in-memory"}, false},
		{[]string{"--proc-type"}, true},
		{[]string{"--socket-mem"}, true},
		{[]string{"--socket-limit"}, true},
		{[]string{"--syslog"}, true},
		{[]string{"--vdev"}, true},
		{[]string{"--vfio-intr"}, true},
		{[]string{"--vmware-tsc-map"}, false},
		{[]string{"--legacy-mem"}, false},
		{[]string{"--single-file-segments"}, false},
		{[]string{"--match-allocations"}, false},
	}
}

// findOption and pair up options and args
func findOption(opts []string, opt string, next int) []string {
	ret := make([]string, 0)

	ret = append(ret, opt)

	for _, o := range dpdkOptions {
		for _, s := range o.opts {
			if s == opt {
				if o.arg {
					if next < len(opts) {
						ret = append(ret, opts[next])
						return ret
					}
				}
			}
		}
	}

	return ret
}

// ParseCmdLine into the CmdLineData structure
func ParseCmdLine(cmdline string) *CmdLineData {

	dat := &CmdLineData{cmdline: cmdline}

	dat.optList = strings.Split(dat.cmdline, " ")

	optType := "exe"

	for i := 0; i < len(dat.optList); i++ {
		opt := dat.optList[i]

		tlog.DebugPrintf("optType %v\n", optType)
		switch optType {
		case "exe":
			dat.executable = opt
			optType = "dpdk"

		case "dpdk":
			if opt == "--" {
				optType = "appl"
				break
			}

			o := &Option{flg: opt}
			s := findOption(dat.optList, opt, i+1)
			if len(s) > 1 {
				o.flg = s[0]
				o.val = s[1]
				i++
			} else {
				o.flg = opt
			}
			dat.dpdkOpts = append(dat.dpdkOpts, o)

		case "appl":
			o := &Option{flg: opt}

			tlog.DebugPrintf("opt %v %+v i %d len(optListi) %d\n",
				opt, dat.optList, i, len(dat.optList))
			if opt[:1] == "-" && i < (len(dat.optList)-1) {
				o.val = dat.optList[i+1]
				i++
			}
			dat.appOpts = append(dat.appOpts, o)
		}
	}

	return dat
}

// String of options
func (opts Options) String() string {

	s := ""

	for i, o := range opts {
		s += o.flg
		if len(o.val) > 0 {
			s += " " + o.val
		}
		if i < len(opts) {
			s += " "
		}
	}

	return s
}

// Executable name that was run for the DPDK application
func (d *CmdLineData) Executable() string {
	if d == nil {
		return ""
	}
	return d.executable
}

// DPDKOptions from the running command list
func (d *CmdLineData) DPDKOptions() Options {
	if d == nil {
		return nil
	}
	return d.dpdkOpts
}

// AppOptions used to start the applicaiton
func (d *CmdLineData) AppOptions() Options {
	if d == nil {
		return nil
	}
	return d.appOpts
}

// String of the cmdline
func (d *CmdLineData) String() string {
	if d == nil {
		return ""
	}
	return d.cmdline
}

// OptionList of the full cmdline
func (d *CmdLineData) OptionList() []string {
	if d == nil {
		return nil
	}
	return d.optList
}

// FindOption string or strings
// opt - is the option flag value to find
// dpdkOpts - if true will search the DPDK options or the Application options
func (d *CmdLineData) FindOption(opt string, dpdkOpts bool) []string {

	if d == nil {
		return nil
	}
	opts := make([]string, 0)

	if d == nil {
		return opts
	}

	cmdOpts := d.dpdkOpts
	if !dpdkOpts {
		cmdOpts = d.appOpts
	}

	for _, o := range cmdOpts {
		if o.flg == opt {
			opts = append(opts, o.val)
		}
	}

	return opts
}

// LCores used by application
func (d *CmdLineData) LCores() []int {

	if d == nil {
		tlog.ErrorPrintf("*CmdLineData value is nil\n")
		return nil
	}
	lcores := make([]int, 0)

	v := d.FindOption("-l", true)
	if len(v) > 0 {
		tlog.DebugPrintf("CPU: -l %s\n", v)
		lst := strings.Split(v[0], ",")
		for _, c := range lst {
			if strings.Contains(c, "-") {
				dash := strings.Split(c, "-")
				lo, _ := strconv.Atoi(dash[0])
				hi, _ := strconv.Atoi(dash[1])
				for ; lo < hi; lo++ {
					lcores = append(lcores, lo)
				}
			} else {
				lo, _ := strconv.Atoi(c)
				lcores = append(lcores, lo)
			}
		}
	} else {
		v := d.FindOption("-c", true)
		if len(v) == 0 {
			return nil
		}
		tlog.DebugPrintf("CPU: -c %s\n", v)
		cores, _ := strconv.ParseUint(v[0], 16, 64)
		for i := 0; i < 64; i++ {
			if (cores & (1 << uint(i))) > 0 {
				lcores = append(lcores, i)
			}
		}
	}

	return lcores
}

// IsLCoreUsed in this DPDK application
func IsLCoreUsed(cpu int, lcores []int) bool {

	for _, core := range lcores {
		if core == cpu {
			return true
		}
	}
	return false
}
