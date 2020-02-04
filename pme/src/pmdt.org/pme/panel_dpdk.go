// SPDX-License-Identifier: BSD-3-Clause
// Copyright(c) 2019-2020 Intel Corporation

package main

import (
	"fmt"
	"path/filepath"
	"sync"

	"github.com/rivo/tview"

	"pmdt.org/graphdata"

	cz "pmdt.org/colorize"
	tab "pmdt.org/taborder"
	tlog "pmdt.org/ttylog"
)

const (
	dpdkPanelName string = "DPDK"
)

// Graph data points
type rxtxData struct {
	rxPoints *graphdata.GraphInfo
	txPoints *graphdata.GraphInfo
}

// DPDKPanel - Data for main page information
type DPDKPanel struct {
	tabOrder *tab.Tab
	topFlex  *tview.Flex

	once		sync.Once
	selectApp *SelectWindow

	dpdkInfo  *tview.TextView
	dpdkNet   *tview.Table
	totalRX   *tview.TextView
	totalTX   *tview.TextView
	dpdkQueue *tview.Table

	data *rxtxData
}

// Setup the DPDK Panel data structure
func setupDPDKPanel() *DPDKPanel {

	pg := &DPDKPanel{}

	pg.data = &rxtxData{}

	pg.data.rxPoints = graphdata.NewGraph(1)
	for _, gd := range pg.data.rxPoints.Graphs() {
		gd.SetMaxPoints(50)
	}
	pg.data.txPoints = graphdata.NewGraph(1)
	for _, gd := range pg.data.txPoints.Graphs() {
		gd.SetMaxPoints(50)
	}

	return pg
}

// DPDKPanelSetup setup the main event page
func DPDKPanelSetup(nextSlide func()) (pageName string, content tview.Primitive) {

	pg := setupDPDKPanel()

	to := tab.New(dpdkPanelName, perfmon.app)
	pg.tabOrder = to

	flex0 := tview.NewFlex().SetDirection(tview.FlexRow)
	flex1 := tview.NewFlex().SetDirection(tview.FlexColumn)
	flex2 := tview.NewFlex().SetDirection(tview.FlexRow)
	flex3 := tview.NewFlex().SetDirection(tview.FlexColumn)

	TitleBox(flex0)

	flex0.AddItem(flex1, 0, 1, true)

	table := CreateTableView(flex1, "Apps (a)", tview.AlignLeft, 18, 2, true)

	clearScrollText := func(view *tview.TextView, f func(*tview.TextView), flg bool) {
		if view == nil {
			return
		}
		f(view)
		if flg {
			view.Clear()
		} else {
			view.ScrollToBeginning()
		}
	}
	clearScrollTable := func(view *tview.Table, f func(*tview.Table), flg bool) {
		if view == nil {
			return
		}
		f(view)
		if flg {
			view.Clear()
		} else {
			view.ScrollToBeginning()
		}
	}

	pg.selectApp = NewSelectWindow(table, "DPDK", 0, func(row, col int) {
		pg.selectApp.UpdateItem(row, col)

		for _, gd := range pg.data.rxPoints.Graphs() {
			gd.Reset()
		}
		for _, gd := range pg.data.txPoints.Graphs() {
			gd.Reset()
		}

		clearScrollText(pg.dpdkInfo, pg.displayDPDKInfo, true)
		clearScrollTable(pg.dpdkNet, pg.displayDPDKNet, true)
		clearScrollTable(pg.dpdkQueue, pg.displayDPDKQueue, true)
	})

	perfmon.processInfo.Add("panel_dpdk", func(event int) {
		names := make([]interface{}, 0)

		for _, f := range perfmon.processInfo.Files() {
			names = append(names, filepath.Ext(f)[1:]) // Only display the PID
		}
		pg.selectApp.UpdateItem(-1, -1)
		pg.selectApp.AddColumn(-1, names)

		row := pg.selectApp.ItemIndex()

		if row == -1 {
			pg.selectApp.UpdateItem(0, -1)
		} else if row > len(names) {
			if len(names) == 0 {
				row = -1
			} else {
				row = len(names) - 1
			}
		}
		pg.selectApp.UpdateItem(row, -1)
	})

	flex1.AddItem(flex2, 0, 1, true)

	pg.dpdkInfo = CreateTextView(flex2, "DPDK Info (i)", tview.AlignLeft, 0, 2, true)
	pg.dpdkNet = CreateTableView(flex2, "DPDK Network Stats (n)", tview.AlignLeft, 0, 4, false)
	pg.dpdkNet.SetFixed(2, 0)
	flex2.AddItem(flex3, 0, 3, false)
	pg.dpdkQueue = CreateTableView(flex2, "DPDK Stats per Queue (s)", tview.AlignLeft, 0, 4, false)

	pg.totalRX = CreateTextView(flex3, "Total RX Mbps", tview.AlignLeft, 0, 1, false)
	pg.totalTX = CreateTextView(flex3, "Total TX Mbps", tview.AlignLeft, 0, 1, false)

	to.Add(pg.selectApp.table, 'a')
	to.Add(pg.dpdkInfo, 'i')
	to.Add(pg.dpdkNet, 'n')
	to.Add(pg.dpdkQueue, 's')

	to.SetInputDone()

	pg.topFlex = flex0

	// Time callback routine to dispaly or process data for the windows.
	perfmon.timers.Add(dpdkPanelName, func(step int, ticks uint64) {
		if pg.topFlex.HasFocus() {
			perfmon.app.QueueUpdateDraw(func() {
				pg.displayDPDKPanel(step, ticks)
			})
		}
	})

	return dpdkPanelName, pg.topFlex
}

// Callback routine to display the windows in the panel called 4 times
// each step is part of a second, to allow for space out processing.
func (pg *DPDKPanel) displayDPDKPanel(step int, ticks uint64) {

	pg.once.Do(func () {
		pi := perfmon.processInfo

		a := pi.AppInfoByIndex(pg.selectApp.ItemIndex())
		if a == nil {
			return
		}

		// Find the list of devices currently handled by the DPDK application
		eth, err := pi.EthdevList(a)
		if err != nil {
			tlog.ErrorPrintf("EthdevList failed: %v\n", err)
			return
		}

		// Output the basic data for the stats and information of a port
		for _, eth := range eth.Ports {

			_, err := pi.EthdevStats(a, eth.PortID)
			if err != nil {
				tlog.WarnPrintf("Calling EthdevStats failed: %v\n", err)
				return
			}
		}
	})

	switch step {
	case 0:
		// Display the screens each second
		pg.displayDPDKInfo(pg.dpdkInfo)
		pg.displayDPDKNet(pg.dpdkNet)
		pg.displayDPDKQueue(pg.dpdkQueue)
		pg.displayChart(pg.totalRX, true)
		pg.displayChart(pg.totalTX, false)
	}
}

// Display thebasic DPDK application information
func (pg *DPDKPanel) displayDPDKInfo(view *tview.TextView) {

	if view == nil {
		tlog.DoPrintf("displayDPDKInfo: called\n")
		return
	}

	w := -14

	// Find the current selected application if any are available
	a := perfmon.processInfo.AppInfoByIndex(pg.selectApp.ItemIndex())
	if a == nil {
		return
	}
	info := perfmon.processInfo.Info(a)

	// Set the speed/duplex and rate in the window
	str := fmt.Sprintf("%s: %s, %s: %s\n",
		cz.Orange("DPDK Verison", w), cz.LightGreen(info.Version),
		cz.Orange("Pid"), cz.DeepPink(a.Pid))

	// Dump out the DPDK and application args
	str += fmt.Sprintf("%s: %s\n", cz.Orange("DPDK Options", w), cz.LightGreen(a.Params.EALArgs))

	str += fmt.Sprintf("%s: %s\n", cz.Orange("Application", w), cz.LightGreen(a.Params.AppArgs))

	// Set the text into the window
	view.SetText(str)
}

// Display some Network information about the DPDK application
func (pg *DPDKPanel) displayDPDKNet(view *tview.Table) {

	if view == nil {
		tlog.DoPrintf("displayDPDKNet: view is nil\n")
		return
	}

	pi := perfmon.processInfo
	a := pi.AppInfoByIndex(pg.selectApp.ItemIndex())
	if a == nil {
		return
	}

	// Routine to help set the table cell
	setCell := func(row, col int, value string, left bool) (int, int) {

		cell := SetCell(view, row, col, value, true)
		if left {
			cell.SetAlign(tview.AlignLeft)
		}

		return row + 1, col + 1
	}

	row := 0
	col := 0

	// Find the list of devices currently handled by the DPDK application
	eth, err := pi.EthdevList(a)
	if err != nil {
		tlog.ErrorPrintf("EthdevList failed: %v\n", err)
		return
	}

	// Setup and display the number of ports total and the number available
	_, col = setCell(row, col, cz.LightSkyBlue("Ports Avail/Total"), true)
	_, col = setCell(row, col, fmt.Sprintf("%s/%s",
		cz.DeepPink(eth.Avail), cz.DeepPink(eth.Total)), true)

	row++
	setCell(row, 0, cz.Lavender("Link:Port"), true)

	names := []string{"ipackets", "opackets", "ibytes", "obytes", "imissed", "ierrors", "oerrors", "rx_nombuf"}

	row++
	for _, n := range names {
		// add the title names to the panel
		setCell(row, 0, cz.Orange(n), true)
		row++
	}

	var mbpsRx, mbpsTx float64

	// Output the basic data for the stats and information of a port
	for _, eth := range eth.Ports {

		col = eth.PortID + 1
		row = 1

		// Output to the table the current rate values
		str := fmt.Sprintf("%s-%s-%d:%d", eth.Duplex, eth.State, eth.Rate, eth.PortID)
		row, _ = setCell(row, col, cz.Orange(str), false)

		prevStats, err := pi.PreviousStats(a, eth.PortID)
		if err != nil {
			tlog.WarnPrintf("Calling PreviousStats failed: %v\n", err)
			return
		}

		stats, err := pi.EthdevStats(a, eth.PortID)
		if err != nil {
			tlog.WarnPrintf("Calling EthdevStats failed: %v\n", err)
			return
		}

		row, _ = setCell(row, col, cz.DeepPink(stats.PacketsIn), false)
		row, _ = setCell(row, col, cz.DeepPink(stats.PacketsOut), false)
		row, _ = setCell(row, col, cz.DeepPink(stats.BytesIn), false)
		row, _ = setCell(row, col, cz.DeepPink(stats.BytesOut), false)
		row, _ = setCell(row, col, cz.DeepPink(stats.MissedIn), false)
		row, _ = setCell(row, col, cz.DeepPink(stats.ErrorsIn), false)
		row, _ = setCell(row, col, cz.DeepPink(stats.ErrorsOut), false)
		row, _ = setCell(row, col, cz.DeepPink(stats.RxNoMbuf), false)


		bytesIn := stats.BytesIn - prevStats.BytesIn
		bytesOut := stats.BytesOut - prevStats.BytesOut

		pktsIn := stats.PacketsIn - prevStats.PacketsIn
		pktsOut := stats.PacketsOut - prevStats.PacketsOut

		mbpsRx += BitRate(pktsIn, bytesIn)
		mbpsTx += BitRate(pktsOut, bytesOut)
	}

	pg.data.rxPoints.GraphPoints(0).AddPoint(mbpsRx/(1024.0 * 1024.0))
	pg.data.txPoints.GraphPoints(0).AddPoint(mbpsTx/(1024.0 * 1024.0))
}


// Display theper queue stats per port
func (pg *DPDKPanel) displayDPDKQueue(view *tview.Table) {

	if view == nil {
		tlog.DoPrintf("displayDPDKNet: view is nil\n")
		return
	}

	pi := perfmon.processInfo
	a := pi.AppInfoByIndex(pg.selectApp.ItemIndex())
	if a == nil {
		return
	}

	// Table cell setup routine
	setCell := func(row, col int, value string, left bool) (int, int) {

		cell := SetCell(view, row, col, value, true)
		if left {
			cell.SetAlign(tview.AlignLeft)
		}

		return row + 1, col + 1
	}

	row := 0
	col := 0

	//	eth, _ := pi.EthdevList(a)

	queues := []string{"port", "q_ipackets", "q_opackets", "q_ibytes", "q_obytes", "q_errors"}

	for i, n := range queues {
		setCell(row+i, col, cz.Orange(n), false)
	}

}

// Display to update the graphs on the panel
func (pg *DPDKPanel) displayChart(view *tview.TextView, rx bool) {

	if rx {
		view.SetText(pg.data.rxPoints.MakeChart(view))
	} else {
		view.SetText(pg.data.txPoints.MakeChart(view))
	}
}
