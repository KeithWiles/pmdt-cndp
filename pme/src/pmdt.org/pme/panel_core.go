// SPDX-License-Identifier: BSD-3-Clause
// Copyright(c) 2019-2020 Intel Corporation

package main

import (
	"fmt"
	"os/exec"

	"github.com/rivo/tview"
	cz "pmdt.org/colorize"
	"pmdt.org/graphdata"
	"pmdt.org/pcm"
	"pmdt.org/taborder"
	tlog "pmdt.org/ttylog"
)

// PageCore - Data for main page information
type PageCore struct {
	pcmRunning bool
	cmd        *exec.Cmd
	tabOrder   *taborder.Tab
	topFlex    *tview.Flex
	title      *tview.Box
	CoreSystem *tview.Table
	Core       *tview.Table
	CoreCharts  [2]*tview.TextView

	system     pcm.System
	header     pcm.Header
	valid      bool

	charts *graphdata.GraphInfo
	CoreRedraw, coreRedraw bool
}

const (
	corePanelName string = "Cores"
	maxCorePoints int    = 52
)

// setupCore - setup and init the main page
func setupCore() *PageCore {

	pg := &PageCore{pcmRunning: false}

	pg.charts = graphdata.NewGraph(2)
	for _, gd := range pg.charts.Graphs() {
		gd.SetMaxPoints(maxCorePoints)
	}
	pg.charts.SetFieldWidth(9)
	pg.valid = false

	return pg
}

// CorePanelSetup setup the main event page
func CorePanelSetup(nextSlide func()) (pageName string, content tview.Primitive) {

	pg := setupCore()

	to := taborder.New(corePanelName, perfmon.app)
	pg.tabOrder = to

	flex0 := tview.NewFlex().SetDirection(tview.FlexRow)
	flex1 := tview.NewFlex().SetDirection(tview.FlexRow)
	flex2 := tview.NewFlex().SetDirection(tview.FlexColumn)

	TitleBox(flex0)
	pg.topFlex = flex0

	pg.CoreSystem = CreateTableView(flex1, "System (1)", tview.AlignLeft, 4, 1, true)
	pg.Core = CreateTableView(flex1, "Core Counters (2)", tview.AlignLeft, 0, 1, true)
	pg.Core.SetFixed(2, 1)
	pg.Core.SetSeparator(tview.Borders.Vertical)

	flex0.AddItem(flex1, 0, 4, true)

	pg.CoreCharts[0] = CreateTextView(flex2, "IPC Charts (3)", tview.AlignLeft, 0, 1, true)
	pg.CoreCharts[1] = CreateTextView(flex2, "Core Charts (4)", tview.AlignLeft, 0, 1, true)

	flex0.AddItem(flex2, 0, 2, true)

	to.Add(pg.CoreSystem, '1')
	to.Add(pg.Core, '2')
	to.Add(pg.CoreCharts[0], '3')
	to.Add(pg.CoreCharts[1], '4')

	to.SetInputDone()

	pg.CoreRedraw = true
	pg.coreRedraw = true

	// Add the timer to update the display
	perfmon.timers.Add(corePanelName, func(step int, ticks uint64) {
		if pg.topFlex.HasFocus() {
			perfmon.app.QueueUpdateDraw(func() {
				pg.displayCorePage(step, ticks)
			})
		}
	})

	return corePanelName, pg.topFlex
}

// Display the Core information to the panel
func (pg *PageCore) displayCorePage(step int, ticks uint64) {

	switch step {
	case 0: // Display the data that was gathered
		pg.staticCoreData()

		pg.collectData()
		pg.displayCoreSystem(pg.CoreSystem)
		pg.displayCore(pg.Core)
		pg.displayCharts(pg.CoreCharts[0], 0, 0)
		pg.displayCharts(pg.CoreCharts[1], 1, 1)
	}
}

func (pg *PageCore) staticCoreData() {

	if pg.valid {
		return
	}

	if err := perfmon.pinfoPCM.Unmarshal("/pcm/system", &pg.system); err != nil {
		tlog.ErrorPrintf("Unable to get PCM system information\n")
		return
	}

	if err := perfmon.pinfoPCM.Unmarshal("/pcm/header", &pg.header); err != nil {
		tlog.ErrorPrintf("Unable to get PCM header information\n")
		return
	}

	pg.valid = true
}

func (pg *PageCore) collectData() {
/*
	core := pg.pcmState.PCMCounters.Core

	if Core.IncomingCoreTrafficMetricsAvailable {
		gd := pg.charts.WithIndex(0)
		gd.AddPoint(float64(Core.IncomingTotal))
		gd.SetName("Socket IN Total")
	}
	if Core.OutgoingCoreTrafficMetricsAvailable {
		gd := pg.charts.WithIndex(1)
		gd.AddPoint(float64(Core.OutgoingTotal))
		gd.SetName("Socket OUT Total")
	}
*/
}

func (pg *PageCore) displayCoreSystem(view *tview.Table) {

	sys := pg.system
	hdr := pg.header

	SetCell(view, 0, 0, fmt.Sprintf("%s %s", cz.Wheat("PCM Version", 12), cz.SkyBlue(hdr.Version)), tview.AlignLeft)
	SetCell(view, 0, 1, fmt.Sprintf("%s %sms", cz.Wheat("PollRate", 12), cz.SkyBlue(hdr.PollMs)), tview.AlignLeft)
	SetCell(view, 0, 2, fmt.Sprintf("%s %s", cz.Wheat("CPU Model", 12), cz.SkyBlue(pcm.CPUModel(int(sys.CPUModel)))), tview.AlignLeft)

	SetCell(view, 1, 0, fmt.Sprintf("%s %s", cz.Wheat("NumCores", 12), cz.SkyBlue(sys.NumOfCores)), tview.AlignLeft)
	SetCell(view, 1, 1, fmt.Sprintf("%s %s", cz.Wheat("Online", 12), cz.SkyBlue(sys.NumOfOnlineCores)), tview.AlignLeft)
	SetCell(view, 1, 2, fmt.Sprintf("%s %s", cz.Wheat("CoreLinks", 12), cz.SkyBlue(sys.NumOfQPILinksPerSocket)), tview.AlignLeft)
	SetCell(view, 1, 3, fmt.Sprintf("%s %s", cz.Wheat("NumSockets", 12), cz.SkyBlue(sys.NumOfSockets)), tview.AlignLeft)
	SetCell(view, 1, 4, fmt.Sprintf("%s %s", cz.Wheat("Online", 12), cz.SkyBlue(sys.NumOfOnlineSockets)), tview.AlignLeft)

	if pg.CoreRedraw {
		pg.CoreRedraw = false
		view.ScrollToBeginning()
	}
}

func (pg *PageCore) displayCore(view *tview.Table) {

	p := perfmon.pinfoPCM.AppsList()
	if len(p) == 0 {
		return
	}

	row := 0
	col := 0
	num := int(pg.system.NumOfCores)
	label := []string{
		"Core/Socket", "", "IPC", "Cycles", "Retired", "Exec", "R-Freq",
		"L3CacheMiss", "L3CacheRef", "L2CacheMiss", "L3CacheHit", "L2CacheHit",
		"L2CacheMPI", "L2CacheMPIHit", "L3CacheOccAvail", "L3CacheOcc",
		"LocalMemoryBW", "RemoteMemoryBW", "LocalMemeoryAcc", "RemoteMAcc", "ThermalHR",
	}
	for i, t := range label {
		SetCell(view, row+i, col, cz.Wheat(t))
	}
	col++

	for i, j := 0, row; i < num; i++ {

		core := pcm.CoreCounters{}
		if err := perfmon.pinfoPCM.Unmarshal(fmt.Sprintf("/pcm/core,%d", i), &core); err != nil {
			tlog.ErrorPrintf("Unable to get PCM system information\n")
			return
		}

		SetCell(view, j+0, col, cz.Orange(fmt.Sprintf("%d/%d", core.CoreID, core.SocketID)))

		SetCell(view, j+2, col, cz.SkyBlue(core.InstructionsPerCycle, 10))
		SetCell(view, j+3, col, cz.SkyBlue(core.Cycles, 10))
		SetCell(view, j+4, col, cz.SkyBlue(core.InstructionsRetired))
		SetCell(view, j+5, col, cz.SkyBlue(core.ExecUsage))
		SetCell(view, j+6, col, cz.SkyBlue(core.RelativeFrequency))
		SetCell(view, j+7, col, cz.SkyBlue(core.L3CacheMisses))
		SetCell(view, j+8, col, cz.SkyBlue(core.L3CacheReference))
		SetCell(view, j+9, col, cz.SkyBlue(core.L2CacheMisses))
		SetCell(view, j+10, col, cz.SkyBlue(core.L3CacheHitRatio))
		SetCell(view, j+11, col, cz.SkyBlue(core.L2CacheHitRatio))
		SetCell(view, j+12, col, cz.SkyBlue(core.L2CacheMPI))
		SetCell(view, j+13, col, cz.SkyBlue(core.L2CacheHitRatio))
		SetCell(view, j+14, col, cz.SkyBlue(core.L3CacheOccupancyAvailable))
		SetCell(view, j+15, col, cz.SkyBlue(core.L3CacheOccupancy))
		if core.LocalMemoryBWAvailable {
			SetCell(view, j+16, col, cz.SkyBlue(core.LocalMemoryBW))
		} else {
			SetCell(view, j+16, col, cz.SkyBlue("disabled"))
		}
		if core.RemoteMemoryBWAvailable {
			SetCell(view, j+17, col, cz.SkyBlue(core.RemoteMemoryBW))
		} else {
			SetCell(view, j+17, col, cz.SkyBlue("disabled"))
		}
		SetCell(view, j+18, col, cz.SkyBlue(core.LocalMemoryAccesses))
		SetCell(view, j+19, col, cz.SkyBlue(core.RemoteMemoryAccesses))
		SetCell(view, j+20, col, cz.SkyBlue(core.ThermalHeadroom))
		col++
	}
}

func (pg *PageCore) displayCharts(view *tview.TextView, start, end int) {

	view.SetText(pg.charts.MakeChart(view, start, end))
}
