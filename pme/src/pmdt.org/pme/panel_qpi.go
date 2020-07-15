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

// PageQPI - Data for main page information
type PageQPI struct {
	pcmRunning bool
	cmd        *exec.Cmd
	tabOrder   *taborder.Tab
	topFlex    *tview.Flex
	title      *tview.Box
	qpi        *tview.Table
	qpiTotals  *tview.Table
	qpiCharts  [2]*tview.TextView

	system pcm.System
	header pcm.Header
	valid  bool

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
	pg.valid = false

	return pg
}

// QPIPanelSetup setup the main event page
func QPIPanelSetup(nextSlide func()) (pageName string, content tview.Primitive) {

	pg := setupQPI()

	to := taborder.New(qpiPanelName, perfmon.app)
	pg.tabOrder = to

	flex0 := tview.NewFlex().SetDirection(tview.FlexRow)
	flex1 := tview.NewFlex().SetDirection(tview.FlexColumn)
	flex2 := tview.NewFlex().SetDirection(tview.FlexColumn)

	TitleBox(flex0)
	pg.topFlex = flex0

	pg.qpi = CreateTableView(flex1, "QPI (1)", tview.AlignLeft, 0, 1, true)
	pg.qpi.SetSeparator(tview.Borders.Vertical)
	pg.qpiTotals = CreateTableView(flex1, "QPI Totals (2)", tview.AlignLeft, 0, 1, true)
	pg.qpiTotals.SetFixed(0, 0)
	pg.qpiTotals.SetSeparator(tview.Borders.Vertical)

	flex0.AddItem(flex1, 0, 1, true)

	pg.qpiCharts[0] = CreateTextView(flex2, "QPI Charts (3)", tview.AlignLeft, 0, 1, true)
	pg.qpiCharts[1] = CreateTextView(flex2, "QPI Charts (4)", tview.AlignLeft, 0, 1, true)

	flex0.AddItem(flex2, 0, 2, true)

	to.Add(pg.qpi, '1')
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
		pg.staticQPIData()

		pg.collectData()
		pg.displayQPI(pg.qpi)
		pg.displayQPITotals(pg.qpiTotals)
		pg.displayCharts(pg.qpiCharts[0], 0, 0)
		pg.displayCharts(pg.qpiCharts[1], 1, 1)
	}
}

func (pg *PageQPI) staticQPIData() {

	if pg.valid {
		return
	}

	if err := perfmon.pinfoPCM.Unmarshal(nil, "/pcm/system", &pg.system); err != nil {
		tlog.ErrorPrintf("Unable to get PCM system information\n")
		return
	}

	if err := perfmon.pinfoPCM.Unmarshal(nil, "/pcm/header", &pg.header); err != nil {
		tlog.ErrorPrintf("Unable to get PCM system information\n")
		return
	}

	pg.valid = true
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

	sys := pg.system.Data
	hdr := pg.header.Data

	SetCell(view, 0, 0, fmt.Sprintf("%s: %s", cz.Wheat("PCM Version"), cz.SkyBlue(hdr.Version)), tview.AlignLeft)
	SetCell(view, 0, 1, fmt.Sprintf("%s: %sms", cz.Wheat("PollRate"), cz.SkyBlue(hdr.PollMs)), tview.AlignLeft)
	SetCell(view, 0, 2, fmt.Sprintf("%s: %s", cz.Wheat("CPU Model "), cz.SkyBlue(pcm.CPUModel(int(sys.CPUModel)))), tview.AlignLeft)

	SetCell(view, 1, 0, fmt.Sprintf("%s: %s", cz.Wheat("NumCores   "), cz.SkyBlue(sys.NumOfCores)), tview.AlignLeft)
	SetCell(view, 1, 1, fmt.Sprintf("%s: %s", cz.Wheat("Online  "), cz.SkyBlue(sys.NumOfOnlineCores)), tview.AlignLeft)
	SetCell(view, 1, 2, fmt.Sprintf("%s: %s", cz.Wheat("QPILinks  "), cz.SkyBlue(sys.NumOfQPILinksPerSocket)), tview.AlignLeft)
	SetCell(view, 2, 0, fmt.Sprintf("%s: %s", cz.Wheat("NumSockets "), cz.SkyBlue(sys.NumOfSockets)), tview.AlignLeft)
	SetCell(view, 2, 1, fmt.Sprintf("%s: %s", cz.Wheat("Online  "), cz.SkyBlue(sys.NumOfOnlineSockets)), tview.AlignLeft)

	if pg.qpiRedraw {
		pg.qpiRedraw = false
		view.ScrollToBeginning()
	}
}

func (pg *PageQPI) fillQPITable(view *tview.Table, row int, sCntr []pcm.QPISocketCounter) int {

	col := 0
	for i := 0; i < len(sCntr); i++ {
		SetCell(view, row, col, cz.Wheat(fmt.Sprintf("Socket %d", sCntr[i].SocketID)), tview.AlignRight)
		col++
		SetCell(view, row, col+3, cz.SkyBlue(sCntr[i].Total, 10))
		for k := 0; k < len(sCntr[i].Links); k++ {
			SetCell(view, row, col, cz.SkyBlue(k), tview.AlignCenter)
			SetCell(view, row, col+1, cz.SkyBlue(sCntr[i].Links[k].Bytes, 10))
			SetCell(view, row, col+2, cz.SkyBlue(sCntr[i].Links[k].Utilization, 6, 6))
			row++
		}
		col = 0
	}
	return row
}

func (pg *PageQPI) displayQPITotals(view *tview.Table) {

	qpi := pcm.QPI{}
	if err := perfmon.pinfoPCM.Unmarshal(nil, "/pcm/qpi", &qpi); err != nil {
		tlog.ErrorPrintf("Unable to get QPI Totals: %v\n", err)
		return
	}

	for i, s := range []string{"Incoming QPI", "LinkID", "Bytes", "Utilization", "Total"} {
		SetCell(view, 0, i, cz.Orange(s), tview.AlignRight)
	}

	row := pg.fillQPITable(view, 1, qpi.Incoming)

	SetCell(view, row, 0, cz.Orange("Outgoing QPI"))
	row++

	pg.fillQPITable(view, row, qpi.Outgoing)

	if pg.coreRedraw {
		pg.coreRedraw = false
		view.ScrollToBeginning()
	}
}

func (pg *PageQPI) displayCharts(view *tview.TextView, start, end int) {

	view.SetText(pg.charts.MakeChart(view, start, end))
}
