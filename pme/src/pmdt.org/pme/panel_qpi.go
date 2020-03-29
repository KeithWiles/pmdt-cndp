// SPDX-License-Identifier: BSD-3-Clause
// Copyright(c) 2019-2020 Intel Corporation

package main

import (
//	"fmt"
	"os/exec"

	"github.com/rivo/tview"
//	cz "pmdt.org/colorize"
	"pmdt.org/graphdata"
//	"pmdt.org/pcm"
	"pmdt.org/taborder"
)

// PageQPI - Data for main page information
type PageQPI struct {
	pcmRunning bool
	cmd        *exec.Cmd
	tabOrder   *taborder.Tab
	topFlex    *tview.Flex
	title      *tview.Box
	qpi        *tview.Table
	qpiCore    *tview.Table
	qpiTotals  *tview.Table
	qpiCharts  [2]*tview.TextView

//	pcmState *pcm.SharedPCMState

	charts                *graphdata.GraphInfo
	qpiRedraw, coreRedraw bool
}

const (
	qpiPanelName string = "QPI"
	maxQPIPoints int    = 52
)

// setupQPI - setup and init the main page
func setupQPI() *PageQPI {

	pg := &PageQPI{pcmRunning: false}

	pg.charts = graphdata.NewGraph(2)
	for _, gd := range pg.charts.Graphs() {
		gd.SetMaxPoints(maxQPIPoints)
	}
	pg.charts.SetFieldWidth(9)

	return pg
}

// QPIPanelSetup setup the main event page
func QPIPanelSetup(nextSlide func()) (pageName string, content tview.Primitive) {

	pg := setupQPI()

	to := taborder.New(qpiPanelName, perfmon.app)
	pg.tabOrder = to

	flex0 := tview.NewFlex().SetDirection(tview.FlexRow)
	flex1 := tview.NewFlex().SetDirection(tview.FlexRow)
	flex2 := tview.NewFlex().SetDirection(tview.FlexColumn)

	TitleBox(flex0)
	pg.topFlex = flex0

	pg.qpi = CreateTableView(flex1, "QPI (1)", tview.AlignLeft, 4, 1, true)

	pg.qpiCore = CreateTableView(flex1, "QPI Core (2)", tview.AlignLeft, 24, 1, true)
	pg.qpiCore.SetFixed(0, 1)

	pg.qpiTotals = CreateTableView(flex1, "QPI Totals (3)", tview.AlignLeft, 5, 1, true)
	pg.qpiTotals.SetFixed(0, 0)

	flex0.AddItem(flex1, 0, 5, true)

	pg.qpiCharts[0] = CreateTextView(flex2, "QPI Charts (4)", tview.AlignLeft, 0, 1, true)
	pg.qpiCharts[1] = CreateTextView(flex2, "QPI Charts (5)", tview.AlignLeft, 0, 1, true)

	flex0.AddItem(flex2, 0, 3, true)

	to.Add(pg.qpi, '1')
	to.Add(pg.qpiCore, '2')
	to.Add(pg.qpiTotals, '3')
	to.Add(pg.qpiCharts[0], '4')
	to.Add(pg.qpiCharts[1], '5')

	to.SetInputDone()

	pg.qpiRedraw = true
	pg.coreRedraw = true

	// Add the timer to update the display
	perfmon.timers.Add(qpiPanelName, func(step int, ticks uint64) {
		if pg.topFlex.HasFocus() {
			perfmon.app.QueueUpdateDraw(func() {
				pg.displayQPIPage(step, ticks)
			})
		}
	})

	return qpiPanelName, pg.topFlex
}

// Display the QPI information to the panel
func (pg *PageQPI) displayQPIPage(step int, ticks uint64) {

	switch step {
	case 0: // Display the data that was gathered
		pg.collectData()
		pg.displayQPI(pg.qpi)
		pg.displayQPICore(pg.qpiCore)
		pg.displayQPITotals(pg.qpiTotals)
		pg.displayCharts(pg.qpiCharts[0], 0, 0)
		pg.displayCharts(pg.qpiCharts[1], 1, 1)
	}
}

func (pg *PageQPI) collectData() {
	/*
	qpi := pg.pcmState.PCMCounters.QPI

	if qpi.IncomingQPITrafficMetricsAvailable {
		gd := pg.charts.WithIndex(0)
		gd.AddPoint(float64(qpi.IncomingTotal))
		gd.SetName("Socket IN Total")
	}
	if qpi.OutgoingQPITrafficMetricsAvailable {
		gd := pg.charts.WithIndex(1)
		gd.AddPoint(float64(qpi.OutgoingTotal))
		gd.SetName("Socket OUT Total")
	}
	*/
}

func (pg *PageQPI) displayQPI(view *tview.Table) {

	/*
	state := pg.pcmState
	sys := state.PCMCounters.System

	SetCell(view, 0, 0, fmt.Sprintf("%s: %s", cz.Wheat("PCM Version"), cz.SkyBlue(state.Header.Version)), tview.AlignLeft)
	SetCell(view, 0, 1, fmt.Sprintf("%s: %sms", cz.Wheat("PollRate"), cz.SkyBlue(state.Header.PollMs)), tview.AlignLeft)
	SetCell(view, 0, 2, fmt.Sprintf("%s: %sB", cz.Wheat("Size    "), cz.SkyBlue(perfmon.pcmData.Len())), tview.AlignLeft)
	SetCell(view, 0, 3, fmt.Sprintf("%s: %s", cz.Wheat("CPU Model "), cz.SkyBlue(pcm.CPUModel(int(sys.CPUModel)))), tview.AlignLeft)

	SetCell(view, 1, 0, fmt.Sprintf("%s: %s", cz.Wheat("NumCores   "), cz.SkyBlue(sys.NumOfCores)), tview.AlignLeft)
	SetCell(view, 1, 1, fmt.Sprintf("%s: %s", cz.Wheat("Online  "), cz.SkyBlue(sys.NumOfOnlineCores)), tview.AlignLeft)
	SetCell(view, 1, 2, fmt.Sprintf("%s: %s", cz.Wheat("QPILinks"), cz.SkyBlue(sys.NumOfQPILinksPerSocket)), tview.AlignLeft)
	SetCell(view, 1, 3, fmt.Sprintf("%s: %s", cz.Wheat("NumSockets"), cz.SkyBlue(sys.NumOfSockets)), tview.AlignLeft)
	SetCell(view, 1, 4, fmt.Sprintf("%s: %s", cz.Wheat("Online"), cz.SkyBlue(sys.NumOfOnlineSockets)), tview.AlignLeft)
*/
	if pg.qpiRedraw {
		pg.qpiRedraw = false
		view.ScrollToBeginning()
	}
}

func (pg *PageQPI) displayQPICore(view *tview.Table) {
/*
	row := 0
	col := 0
	num := int(pg.pcmState.PCMCounters.System.NumOfOnlineCores)
	core := pg.pcmState.PCMCounters.Core.Cores
	label := []string{
		"CoreID:", "SocketID:", "", "IPC:", "Cycles:", "Retired:", "Exec:", "R-Freq:",
		"L3CacheMiss:", "L3CacheRef:", "L2CacheMiss:", "L3CacheHit:", "L2CacheHit:",
		"L2CacheMPI:", "L2CacheMPIHit:", "L3CacheOccAvail:", "L3CacheOcc:",
		"LocalMemoryBW:", "RemoteMemoryBW:", "LocalMemeoryAcc:", "RemoteMAcc:", "ThermalHR:",
	}
	for i, t := range label {
		SetCell(view, row+i, col, cz.Wheat(t))
	}
	col++

	for i, j := 0, row; i < num; i++ {
		SetCell(view, j+0, col, cz.SkyBlue(core[i].CoreID))
		SetCell(view, j+1, col, cz.SkyBlue(core[i].SocketID))

		SetCell(view, j+3, col, cz.SkyBlue(core[i].InstructionsPerCycles, 10))
		SetCell(view, j+4, col, cz.SkyBlue(core[i].Cycles, 10))
		SetCell(view, j+5, col, cz.SkyBlue(core[i].InstructionsRetired))
		SetCell(view, j+6, col, cz.SkyBlue(core[i].ExecUsage))
		SetCell(view, j+7, col, cz.SkyBlue(core[i].RelativeFrequency))
		SetCell(view, j+8, col, cz.SkyBlue(core[i].L3CacheMisses))
		SetCell(view, j+9, col, cz.SkyBlue(core[i].L3CacheReference))
		SetCell(view, j+10, col, cz.SkyBlue(core[i].L2CacheMisses))
		SetCell(view, j+11, col, cz.SkyBlue(core[i].L3CacheHitRatio))
		SetCell(view, j+12, col, cz.SkyBlue(core[i].L2CacheHitRatio))
		SetCell(view, j+13, col, cz.SkyBlue(core[i].L2CacheMPI))
		SetCell(view, j+14, col, cz.SkyBlue(core[i].L2CacheHitRatio))
		SetCell(view, j+15, col, cz.SkyBlue(core[i].L3CacheOccupancyAvailable))
		SetCell(view, j+16, col, cz.SkyBlue(core[i].L3CacheOccupancy))
		if core[i].LocalMemoryBWAvailable {
			SetCell(view, j+17, col, cz.SkyBlue(core[i].LocalMemoryBW))
		} else {
			SetCell(view, j+17, col, cz.SkyBlue("disabled"))
		}
		if core[i].RemoteMemoryBWAvailable {
			SetCell(view, j+18, col, cz.SkyBlue(core[i].RemoteMemoryBW))
		} else {
			SetCell(view, j+18, col, cz.SkyBlue("disabled"))
		}
		SetCell(view, j+19, col, cz.SkyBlue(core[i].LocalMemoryAccesses))
		SetCell(view, j+20, col, cz.SkyBlue(core[i].RemoteMemoryAccesses))
		SetCell(view, j+21, col, cz.SkyBlue(core[i].ThermalHeadroom))
		col++
	}
	*/
}

func (pg *PageQPI) displayQPITotals(view *tview.Table) {
/*
	qpi := pg.pcmState.PCMCounters.QPI

	row := 0
	col := 0

	num := int(pg.pcmState.PCMCounters.System.NumOfSockets)
	SetCell(view, row, 0, "    ") // Add some space between the results
	SetCell(view, row, 3, "    ") // Add some space between the results
	if qpi.IncomingQPITrafficMetricsAvailable {
		col = 1
		num = int(pg.pcmState.PCMCounters.System.NumOfSockets)
		for i := 0; i < num; i++ {
			SetCell(view, row+i, col, cz.Wheat(fmt.Sprintf("Socket %d IN Count:", i)))
		}
		for i := 0; i < num; i++ {
			SetCell(view, row+i, col+1, cz.SkyBlue(qpi.Incoming[i].Total))
		}
		row += num
		SetCell(view, row, col, cz.Wheat("Total:"))
		SetCell(view, row, col+1, cz.SkyBlue(qpi.IncomingTotal))
	}

	if qpi.OutgoingQPITrafficMetricsAvailable {
		row = 0
		col = 4
		num = int(pg.pcmState.PCMCounters.System.NumOfSockets)

		for i := 0; i < num; i++ {
			SetCell(view, row+i, col, cz.Wheat(fmt.Sprintf("Socket %d OUT Count:", i)))
		}
		for i := 0; i < num; i++ {
			SetCell(view, row+i, col+1, cz.SkyBlue(qpi.Outgoing[i].Total))
		}
		row += num
		SetCell(view, row, col, cz.Wheat("Total:"))
		SetCell(view, row, col+1, cz.SkyBlue(qpi.OutgoingTotal))
	}

	row += num
*/
	if pg.coreRedraw {
		pg.coreRedraw = false
		view.ScrollToBeginning()
	}
}

func (pg *PageQPI) displayCharts(view *tview.TextView, start, end int) {

	view.SetText(pg.charts.MakeChart(view, start, end))
}
