// SPDX-License-Identifier: BSD-3-Clause
// Copyright(c) 2019-2020 Intel Corporation

package main

import (
	"fmt"
	"os/exec"

	"github.com/rivo/tview"
	cz "pmdt.org/colorize"
	"pmdt.org/graphdata"
	pcm "pmdt.org/pcm"
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
	selectCore *SelectWindow
	//selectCoreRange  *SelectWindow
	CoreSystem       *tview.Table
	Core             *tview.Table
	CoreCharts       [2]*tview.TextView
	chart            *tview.Table
	selected         int
	selectionChanged bool

	system pcm.System
	header pcm.Header
	valid  bool

	charts                 *graphdata.GraphInfo
	ipc                    *graphdata.GraphInfo
	CoreRedraw, coreRedraw bool
}

const (
	corePanelName  string = "CoreCounters"
	corePanelLogID string = "CorePaneLogId"
	maxCorePoints  int    = 52
)

func init() {
	tlog.Register(corePanelLogID)
}

// setupCore - setup and init the main page
func setupCore() *PageCore {

	pg := &PageCore{pcmRunning: false}

	// create "graph" for each core
	pg.charts = graphdata.NewGraph(NumCPUs() * 2)
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
	flex1 := tview.NewFlex().SetDirection(tview.FlexColumn)
	flex2 := tview.NewFlex().SetDirection(tview.FlexRow)
	flex3 := tview.NewFlex().SetDirection(tview.FlexColumn)

	TitleBox(flex0)
	pg.topFlex = flex0

	pg.CoreSystem = CreateTableView(flex0, "System (1)", tview.AlignLeft, 4, 1, true)

	// Core selection window to be able to select a core to view
	table := CreateTableView(flex1, "Core (c)", tview.AlignLeft, 15, 1, true)

	// Select window setup and callback function when selection changes.
	pg.selectCore = NewSelectWindow(table, "CoreCounters", 0, func(row, col int) {

		if row != pg.selected {
			pg.selectCore.UpdateItem(row, col)

			pg.selectionChanged = true

			pg.selected = row
			// pg.chart.SetTitle(TitleColor(fmt.Sprintf("CPU %d (c)", pg.selected)))
		}
	})

	names := make([]interface{}, 0)

	for i := 0; i < NumCPUs(); i++ {
		s := fmt.Sprintf("%4d", i)
		names = append(names, s)
	}
	pg.selectCore.AddColumn(-1, names, cz.SkyBlueColor)

	pg.CoreCharts[0] = CreateTextView(flex2, "IPC Chart (2)", tview.AlignLeft, 0, 1, true)
	pg.CoreCharts[1] = CreateTextView(flex2, "Cycles Chart (3)", tview.AlignLeft, 0, 1, true)

	flex1.AddItem(flex2, 0, 2, true)

	flex0.AddItem(flex1, 0, 2, true)

	// Core range selection window to display counters
	//table1 := CreateTableView(flex3, "Core Range (r)", tview.AlignLeft, 15, 24, true)

	// Select window setup and callback function when selection changes.
	/*pg.selectCoreRange = NewSelectWindow(table1, "Core Counters", 0, func(row, col int) {

		if row != pg.selected {
			pg.selectCoreRange.UpdateItem(row, col)

			pg.selectionChanged = true

			pg.selected = row
			pg.Core.SetTitle(TitleColor(fmt.Sprintf("Core Counters %d (4)", pg.selected)))
		}
	})

	namesRange := make([]interface{}, 0)

	for i := 0; i < NumCPUs(); i++ {
		s := fmt.Sprintf("%4d", i)
		t := fmt.Sprintf("%4d", (i + 6))
		namesRange = append(names, s+"-"+t)
		i += 6
	}
	pg.selectCoreRange.AddColumn(-1, namesRange, cz.SkyBlueColor)
	*/

	pg.Core = CreateTableView(flex3, "Core Counters (4)", tview.AlignLeft, 0, 1, true)
	pg.Core.SetFixed(2, 1)
	pg.Core.SetSeparator(tview.Borders.Vertical)

	flex0.AddItem(flex3, 24, 4, true)

	to.Add(pg.CoreSystem, '1')
	to.Add(pg.selectCore.table, 'c')
	to.Add(pg.CoreCharts[0], '2')
	to.Add(pg.CoreCharts[1], '3')
	//to.Add(pg.selectCoreRange.table, 'r')
	to.Add(pg.Core, '4')

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

		// Collect Data for both charts
		pg.collectData()
		pg.displayCoreSystem(pg.CoreSystem)
		pg.displayCore(pg.Core)
		//pg.displayCharts()
		pg.displayCharts(pg.CoreCharts[0], 0, 0)
		pg.displayCharts(pg.CoreCharts[1], 1, 1)
	}
}

func (pg *PageCore) staticCoreData() {

	if pg.valid == false {
		if err := perfmon.pinfoPCM.Unmarshal(nil, "/pcm/system", &pg.system); err != nil {
			tlog.ErrorPrintf("Unable to get PCM system information\n")
			return
		}

		if err := perfmon.pinfoPCM.Unmarshal(nil, "/pcm/header", &pg.header); err != nil {
			tlog.ErrorPrintf("Unable to get PCM header information\n")
			return
		}

		pg.valid = true
	}
}

/*
func (pg *PagePBF) collectChartData() {

	for cpu, gd := range pg.freqs.Graphs() {
		p := pbf.InfoPerCPU(cpu)

		// Append the frequency data to the list for the graphing in a chart
		gd.AddPoint(float64(p.CurFreq))
	}
}

*/
func (pg *PageCore) collectData() {

	core := pcm.CoreCounters{}
	if err := perfmon.pinfoPCM.Unmarshal(nil, fmt.Sprintf("/pcm/core,%d", pg.selected), &core); err != nil {
		tlog.ErrorPrintf("Unable to get PCM system information\n")
		return
	}

	gd1 := pg.charts.WithIndex(0)
	gd1.AddPoint(float64(core.Data.InstructionsPerCycle))

	gd2 := pg.charts.WithIndex(1)
	gd2.AddPoint(float64(core.Data.Cycles))

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

	sys := pg.system.Data
	hdr := pg.header.Data

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

	p := perfmon.pinfoPCM.ConnectionList()
	if len(p) == 0 {
		return
	}

	row := 0
	col := 0
	num := int(pg.system.Data.NumOfCores)
	label := []string{
		"Core/Socket", "", "IPC", "Cycles", "Retired", "Exec", "R-Freq",
		"L3CacheMiss", "L3CacheRef", "L2CacheMiss", "L3CacheHit", "L2CacheHit",
		"L2CacheMPI", "L2CacheMPIHit", "L3CacheOccAvail", "L3CacheOcc",
		"LocalMemoryBW", "RemoteMemoryBW", "LocalMemeoryAcc", "RemoteMAcc", "ThermalHR",
		"Branches", "BranchMispredicts",
	}
	for i, t := range label {
		SetCell(view, row+i, col, cz.Wheat(t))
	}
	col++

	for i, j := 0, row; i < num; i++ {

		data := pcm.CoreCounters{}
		if err := perfmon.pinfoPCM.Unmarshal(nil, fmt.Sprintf("/pcm/core,%d", i), &data); err != nil {
			tlog.ErrorPrintf("Unable to get PCM system information\n")
			return
		}

		core := data.Data

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
		SetCell(view, j+21, col, cz.SkyBlue(core.Branches))
		SetCell(view, j+22, col, cz.SkyBlue(core.BranchMispredicts))
		col++
	}
}

func (pg *PageCore) displayCharts(view *tview.TextView, start, end int) {

	view.SetText(pg.charts.MakeChart(view, start, end))
	//pg.CoreCharts[0].SetText(pg.charts.MakeChart(pg.CoreCharts[0], pg.selected, pg.selected))
	//pg.CoreCharts[1].SetText(pg.charts.MakeChart(pg.CoreCharts[1], pg.selected, pg.selected))
}
