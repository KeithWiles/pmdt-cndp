// SPDX-License-Identifier: BSD-3-Clause
// Copyright(c) 2019-2020 Intel Corporation

package main

import (
	"fmt"
	flags "github.com/jessevdk/go-flags"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	cz "pmdt.org/colorize"
	tlog "pmdt.org/ttylog"

	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
	"pmdt.org/etimers"
	"pmdt.org/pinfo"
	"pmdt.org/profiles"
)

const (
	// pmeVersion string
	pmeVersion = "20.02.0"
)

// PanelInfo for title and primitive
type PanelInfo struct {
	title     string
	primitive tview.Primitive
}

// Panels is a function which returns the feature's main primitive and its title.
// It receives a "nextFeature" function which can be called to advance the
// presentation to the next slide.
type Panels func(nextPanel func()) (title string, content tview.Primitive)

// PerfMonitor for monitoring DPDK and system performance data
type PerfMonitor struct {
	version      string             // Version of PMDT
	app          *tview.Application // Application or top level application
	timers       *etimers.EventTimers
	panels       []PanelInfo

	pinfoPCM     *pinfo.ProcessInfo
}

// Options command line options
type Options struct {
	Ptty        string `short:"p" long:"ptty" description:"path to ptty /dev/pts/X"`
	Dbg         bool   `short:"D" long:"debug" description:"Wait 15 seconds (default) to connect debugger"`
	WaitTime    uint   `short:"W" long:"wait-time" description:"N seconds before startup" default:"15"`
	ShowVersion bool   `short:"V" long:"version" description:"Print out version and exit"`
	Verbose     bool   `short:"v" long:"Verbose output for debugging"`
}

// Global to the main package for the tool
var perfmon PerfMonitor
var options Options
var parser = flags.NewParser(&options, flags.Default)

const (
	mainLog = "MainLogID"
)

func buildPanelString(idx int) string {
	// Build the panel selection string at the bottom of the xterm and
	// highlight the selected tab/panel item.
	s := ""
	for index, p := range perfmon.panels {
		if index == idx {
			s += fmt.Sprintf("F%d:[orange::r]%s[white::-]", index+1, p.title)
		} else {
			s += fmt.Sprintf("F%d:[orange::-]%s[white::-]", index+1, p.title)
		}
		if (index + 1) < len(perfmon.panels) {
			s += " "
		}
	}
	return s
}

// Setup the tool's global information and startup the process info connection
func init() {
	tlog.Register(mainLog, true)

	perfmon = PerfMonitor{}
	perfmon.version = pmeVersion

	// Parse the event profiles JSON file
	if err := profiles.Parse(""); err != nil {
		log.Fatalf("get config data failed: %s", err)
	}

	// Setup and locate the process info socket connections
	perfmon.pinfoPCM = pinfo.New("/var/run/pcm-info", "pinfo")
	if perfmon.pinfoPCM == nil {
		panic("unable to setup pinfoPCM")
	}

	// Create the main tveiw application.
	perfmon.app = tview.NewApplication()
}

// Version number string
func Version() string {
	return perfmon.version
}

func main() {

	cz.SetDefault("ivory", "", 0, 2, "")

	_, err := parser.Parse()
	if err != nil {
		fmt.Printf("*** invalid arguments %v\n", err)
		os.Exit(1)
	}

	if options.ShowVersion {
		fmt.Printf("PME Version: %s\n", perfmon.version)
		return
	}

	if len(options.Ptty) > 0 {
		err = tlog.Open(options.Ptty)
		if err != nil {
			fmt.Printf("ttylog open failed: %s\n", err)
			os.Exit(1)
		}
	}
	tlog.Log(mainLog, "\n===== %s =====\n", PerfmonInfo(false))
	fmt.Printf("\n===== %s =====\n", PerfmonInfo(false))

	app := perfmon.app

	perfmon.timers = etimers.New(time.Second/4, 4)
	perfmon.timers.Start()

	panels := []Panels{
		ProcessPanelSetup,
		SysInfoPanelSetup,
		DevBindPanelSetup,
		DPDKPanelSetup,
		CorePanelSetup,
		PCIPanelSetup,
		QPIPanelSetup,
		PBFPanelSetup,
	}

	// The bottom row has some info on where we are.
	info := tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true).
		SetWrap(false)

	currentPanel := 0
	info.Highlight(strconv.Itoa(currentPanel))

	pages := tview.NewPages()

	previousPanel := func() {
		currentPanel = (currentPanel - 1 + len(panels)) % len(panels)
		info.Highlight(strconv.Itoa(currentPanel)).
			ScrollToHighlight()
		pages.SwitchToPage(strconv.Itoa(currentPanel))
		info.SetText(buildPanelString(currentPanel))
	}

	nextPanel := func() {
		currentPanel = (currentPanel + 1) % len(panels)
		info.Highlight(strconv.Itoa(currentPanel)).
			ScrollToHighlight()
		pages.SwitchToPage(strconv.Itoa(currentPanel))
		info.SetText(buildPanelString(currentPanel))
	}

	for index, f := range panels {
		title, primitive := f(nextPanel)
		pages.AddPage(strconv.Itoa(index), primitive, true, index == currentPanel)
		perfmon.panels = append(perfmon.panels, PanelInfo{title: title, primitive: primitive})
	}
	info.SetText(buildPanelString(0))

	// Create the main panel.
	panel := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(pages, 0, 1, true).
		AddItem(info, 1, 1, false)

	// Shortcuts to navigate the panels.
	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyCtrlN {
			nextPanel()
		} else if event.Key() == tcell.KeyCtrlP {
			previousPanel()
		} else if event.Key() == tcell.KeyCtrlQ {
			app.Stop()
		} else if event.Key() == tcell.KeyCtrlH {

		} else {
			var idx int

			switch {
			case tcell.KeyF1 <= event.Key() && event.Key() <= tcell.KeyF19:
				idx = int(event.Key() - tcell.KeyF1)
			case event.Rune() == 'q':
				app.Stop()
			default:
				idx = -1
			}
			if idx != -1 {
				if idx < len(panels) {
					currentPanel = idx
					info.Highlight(strconv.Itoa(currentPanel)).ScrollToHighlight()
					pages.SwitchToPage(strconv.Itoa(currentPanel))
				}
				info.SetText(buildPanelString(idx))
			}
		}
		return event
	})

	setupSignals(syscall.SIGINT, syscall.SIGTERM, syscall.SIGSEGV)

	if options.Dbg {
		fmt.Printf("Waiting %d seconds for dlv to attach\n", options.WaitTime)
		time.Sleep(time.Second * time.Duration(options.WaitTime))
	}

	if err := perfmon.pinfoPCM.StartWatching(); err != nil {
		panic(err)
	}
	defer perfmon.pinfoPCM.StopWatching()

	// Start the application.
	if err := app.SetRoot(panel, true).Run(); err != nil {
		panic(err)
	}

	tlog.Log(mainLog, "===== Done =====\n")
}

func setupSignals(signals ...os.Signal) {
	app := perfmon.app

	sigs := make(chan os.Signal, 1)

	signal.Notify(sigs, signals...)
	go func() {
		sig := <-sigs

		tlog.Log(mainLog, "Signal: %v\n", sig)
		time.Sleep(time.Second)

		app.Stop()
		os.Exit(1)
	}()
}
