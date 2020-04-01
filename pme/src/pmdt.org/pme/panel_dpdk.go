// SPDX-License-Identifier: BSD-3-Clause
// Copyright(c) 2019-2020 Intel Corporation

package main

import (
	"fmt"
	"path/filepath"
	"sync"

	"github.com/rivo/tview"

	"pmdt.org/graphdata"
	"pmdt.org/pinfo"
	"pmdt.org/dpdk"

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

	pinfoDPDK *pinfo.ProcessInfo
	infoDPDK  dpdk.Information

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

	table := CreateTableView(flex1, "Apps (1)", tview.AlignLeft, 18, 2, true)

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
	})
	tlog.DoPrintf("Create DPDK ProcessInfo\n")
	// Setup and locate the telemery socket connections
	pg.pinfoDPDK = pinfo.NewProcessInfo("/var/run/dpdk", "dpdk_telemetry")
	if pg.pinfoDPDK == nil {
		panic("unable to setup pinfoDPDK")
	}

	if err := pg.pinfoDPDK.Open(); err != nil {
		panic(err)
	}
	defer pg.pinfoDPDK.Close()

	pg.pinfoDPDK.Add("panel_dpdk", func(event int) {
		names := make([]interface{}, 0)

		for _, f := range pg.pinfoDPDK.Files() {
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

	pg.dpdkInfo = CreateTextView(flex2, "DPDK Info (2)", tview.AlignLeft, 0, 2, true)
	pg.dpdkNet = CreateTableView(flex2, "DPDK Network Stats (3)", tview.AlignLeft, 0, 4, false)
	pg.dpdkNet.SetFixed(2, 0)
	pg.dpdkNet.SetSeparator(tview.Borders.Vertical)
	flex2.AddItem(flex3, 0, 3, false)

	pg.totalRX = CreateTextView(flex3, "Total RX Mbps", tview.AlignLeft, 0, 1, false)
	pg.totalTX = CreateTextView(flex3, "Total TX Mbps", tview.AlignLeft, 0, 1, false)

	to.Add(pg.selectApp.table, '1')
	to.Add(pg.dpdkInfo, '2')
	to.Add(pg.dpdkNet, '3')

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
		pi := pg.pinfoDPDK

		a := pi.AppInfoByIndex(pg.selectApp.ItemIndex())
		if a == nil {
			return
		}

		if err := pg.pinfoDPDK.Unmarshal("/eal/version", &pg.infoDPDK.Version); err != nil {
			tlog.ErrorPrintf("Unable to get EAL version: %v\n", err)
			return
		}
		tlog.DebugPrintf("EAL Version: %v\n", pg.infoDPDK.Version.Version)

		if err := pg.pinfoDPDK.Unmarshal("/eal/params", &pg.infoDPDK.Params); err != nil {
			tlog.ErrorPrintf("Unable to get EAL Parameters: %v\n", err)
			return
		}
		tlog.DebugPrintf("EAL Parameters: %v\n", pg.infoDPDK.Params.Params)

		if err := pg.pinfoDPDK.Unmarshal("/eal/app_params", &pg.infoDPDK.AppParams); err != nil {
			tlog.ErrorPrintf("Unable to get EAL Application Parameters: %v\n", err)
			return
		}
		tlog.DebugPrintf("EAL Application Parameters: %v\n", pg.infoDPDK.AppParams.Params)

		if err := pg.pinfoDPDK.Unmarshal("/", &pg.infoDPDK.Cmds); err != nil {
			tlog.ErrorPrintf("Unable to get EAL Commands: %v\n", err)
			return
		}
		tlog.DebugPrintf("EAL Commands: %v\n", pg.infoDPDK.Cmds)

		if err := pg.pinfoDPDK.Unmarshal("/ethdev/list", &pg.infoDPDK.PidList); err != nil {
			tlog.ErrorPrintf("Unable to get Ethdev List information: %v\n", err)
			return
		}
		tlog.DebugPrintf("EthdevList: %v\n", pg.infoDPDK.PidList)

		// Output the basic data for the stats and information of a port
		for _, pid := range pg.infoDPDK.PidList.Pids {

			eth := dpdk.EthdevStats{}
			cmd := fmt.Sprintf("/ethdev/stats,%d", pid)
			if err := pg.pinfoDPDK.Unmarshal(cmd, &eth); err != nil {
				tlog.WarnPrintf("Unable to get PCM system information\n")
				continue
			}
			eth.Stats.PortID = pid
			pg.infoDPDK.EthdevStats = append(pg.infoDPDK.EthdevStats, &eth)
			tlog.DoPrintf("/ethdev/stats,%d: %+v\n", pid, eth)

			// Update the previous stats
			pg.infoDPDK.PrevStats[eth.Stats.PortID].Stats = eth.Stats

			tlog.DebugPrintf("Prev: %+v\n", pg.infoDPDK.PrevStats[eth.Stats.PortID])
		}
	})

	switch step {
	case 0:
		// Display the screens each second
		pg.displayDPDKInfo(pg.dpdkInfo)
		pg.displayDPDKNet(pg.dpdkNet)
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
	a := pg.pinfoDPDK.AppInfoByIndex(pg.selectApp.ItemIndex())
	if a == nil {
		return
	}

	info := pg.infoDPDK
	// Set the speed/duplex and rate in the window
	str := fmt.Sprintf("%s: %s\n",
		cz.Orange("DPDK Verison", w), cz.LightGreen(info.Version.Version))

	// Dump out the DPDK and application args
	str += fmt.Sprintf("%s: %s\n", cz.Orange("DPDK Options", w), cz.LightGreen(info.Params.Params))

	str += fmt.Sprintf("%s: %s\n", cz.Orange("Application", w), cz.LightGreen(info.AppParams.Params))

	// Set the text into the window
	view.SetText(str)

}

// Display some Network information about the DPDK application
func (pg *DPDKPanel) displayDPDKNet(view *tview.Table) {

	if view == nil {
		tlog.DoPrintf("displayDPDKNet: view is nil\n")
		return
	}
	a := pg.pinfoDPDK.AppInfoByIndex(pg.selectApp.ItemIndex())
	if a == nil {
		return
	}

	if err := pg.pinfoDPDK.Unmarshal("/ethdev/list", &pg.infoDPDK.PidList); err != nil {
		tlog.ErrorPrintf("Unable to get Ethdev List information: %v\n", err)
		return
	}
	tlog.DebugPrintf("EthdevList: %v\n", pg.infoDPDK.PidList)

	// Output the basic data for the stats and information of a port
	for _, pid := range pg.infoDPDK.PidList.Pids {

		eth := dpdk.EthdevStats{}
		cmd := fmt.Sprintf("%s,%d", "/ethdev/stats", pid)
		if err := pg.pinfoDPDK.Unmarshal(cmd, &eth); err != nil {
			tlog.WarnPrintf("Unable to get PCM system information\n")
			continue
		}
		eth.Stats.PortID = pid
		pg.infoDPDK.EthdevStats = append(pg.infoDPDK.EthdevStats, &eth)
		tlog.DoPrintf("%s,%d: %+v\n", "/ethdev/stats", pid, eth)
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


	names := []string{"PortID", "ipackets", "opackets", "ibytes",
		"obytes", "imissed", "ierrors", "oerrors", "rx_nombuf"}

	for i, n := range names {
		// add the title names to the panel
		if i == 0 {
			setCell(row, 0, cz.Wheat(n), true)
			row++
		} else {
			setCell(row, 0, cz.Orange(n), true)
		}
		row++
	}

	var mbpsRx, mbpsTx float64

	// Output the basic data for the stats and information of a port
	for _, eth := range pg.infoDPDK.EthdevStats {

		col = int(eth.Stats.PortID) + 1
		row, _ = setCell(0, col, cz.Wheat(eth.Stats.PortID, 12), false)
		row++
		stats := &eth.Stats

		row, _ = setCell(row, col, cz.DeepPink(stats.InPackets), false)
		row, _ = setCell(row, col, cz.DeepPink(stats.OutPackets), false)
		row, _ = setCell(row, col, cz.DeepPink(stats.InBytes), false)
		row, _ = setCell(row, col, cz.DeepPink(stats.OutBytes), false)
		row, _ = setCell(row, col, cz.DeepPink(stats.InMissed), false)
		row, _ = setCell(row, col, cz.DeepPink(stats.InErrors), false)
		row, _ = setCell(row, col, cz.DeepPink(stats.OutErrors), false)
		row, _ = setCell(row, col, cz.DeepPink(stats.RxNomBuf), false)

		prevStats := &pg.infoDPDK.PrevStats[eth.Stats.PortID].Stats

		bytesIn := stats.InBytes - prevStats.InBytes
		bytesOut := stats.OutBytes - prevStats.OutBytes

		pktsIn := stats.InPackets - prevStats.InPackets
		pktsOut := stats.OutPackets - prevStats.OutPackets

		pg.infoDPDK.PrevStats[eth.Stats.PortID] = *eth

		mbpsRx += BitRate(pktsIn, bytesIn)
		mbpsTx += BitRate(pktsOut, bytesOut)

		// Update the previous stats
		pg.infoDPDK.PrevStats[eth.Stats.PortID].Stats = eth.Stats

		tlog.DoPrintf("Prev: %+v\n\n", pg.infoDPDK.PrevStats[eth.Stats.PortID])
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
/*
	pi := pg.pinfoDPDK
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
*/
}

// Display to update the graphs on the panel
func (pg *DPDKPanel) displayChart(view *tview.TextView, rx bool) {

	if rx {
		view.SetText(pg.data.rxPoints.MakeChart(view))
	} else {
		view.SetText(pg.data.txPoints.MakeChart(view))
	}
}
