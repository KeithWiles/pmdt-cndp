// SPDX-License-Identifier: BSD-3-Clause
// Copyright(c) 2019-2020 Intel Corporation

package main

import (
	"fmt"
	"sync"

	"github.com/rivo/tview"

	"pmdt.org/dpdk"
	"pmdt.org/graphdata"
	"pmdt.org/pinfo"

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

	once      sync.Once
	selectApp *SelectWindow

	dpdkInfo *tview.TextView
	dpdkNet  *tview.Table
	totalRX  *tview.TextView
	totalTX  *tview.TextView

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

	// Setup and locate the telemery socket connections
	pg.pinfoDPDK = pinfo.New("/var/run/dpdk", "dpdk_telemetry")
	if pg.pinfoDPDK == nil {
		panic("unable to setup pinfoDPDK")
	}

	if err := pg.pinfoDPDK.StartWatching(); err != nil {
		panic(err)
	}
	defer pg.pinfoDPDK.StopWatching()

	// Add a callback for this watcher
	pg.pinfoDPDK.Add("panel_dpdk", func(event int) {
		names := make([]interface{}, 0)

		for _, f := range pg.pinfoDPDK.Processes() {
			names = append(names, f) // Only display the ProcessName
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

func (pg *DPDKPanel) selectedConnection() (*pinfo.ConnInfo, error) {

	tlog.DebugPrintf("Value: %v\n", pg.selectApp.ItemValue()) // name of DPDK App
	selectedName := pg.selectApp.ItemValue()
	if selectedName == nil {
		return nil, fmt.Errorf("I am not selected")
	}

	// Find the current selected application if any are available
	a := pg.pinfoDPDK.ConnectionByProcessName(selectedName.(string))
	if a == nil {
		return nil, fmt.Errorf("failed to get connection pointer")
	}

	return a, nil
}

func (pg *DPDKPanel) displayDPDKPanel(step int, ticks uint64) {

	switch step {
	case 0:
		pg.collectStats()

	case 3:
		// Display the screens each second
		pg.displayDPDKInfo(pg.dpdkInfo)
		pg.displayDPDKNet(pg.dpdkNet)
		pg.displayChart(pg.totalRX, true)
		pg.displayChart(pg.totalTX, false)
	}
}

func (pg *DPDKPanel) getFixedData(a *pinfo.ConnInfo) {

	pg.infoDPDK.Version = pg.pinfoDPDK.Version(a)
	tlod.DebugPrintf("EAL Version: %s\n", pg.infoDPDKVersion)

	if err := pg.pinfoDPDK.Unmarshal(a, "/eal/params", &pg.infoDPDK.Params); err != nil {
		tlog.ErrorPrintf("Unable to get EAL Parameters: %v\n", err)
		return
	}
	tlog.DebugPrintf("EAL Parameters: %v\n", pg.infoDPDK.Params.Params)

	if err := pg.pinfoDPDK.Unmarshal(a, "/eal/app_params", &pg.infoDPDK.AppParams); err != nil {
		tlog.ErrorPrintf("Unable to get EAL Application Parameters: %v\n", err)
		return
	}
	tlog.DebugPrintf("EAL Application Parameters: %v\n", pg.infoDPDK.AppParams.Params)

	if err := pg.pinfoDPDK.Unmarshal(a, "/", &pg.infoDPDK.Cmds); err != nil {
		tlog.ErrorPrintf("Unable to get EAL Commands: %v\n", err)
		return
	}
	tlog.DebugPrintf("EAL Commands: %v\n", pg.infoDPDK.Cmds)

	if err := pg.pinfoDPDK.Unmarshal(a, "/ethdev/list", &pg.infoDPDK.PidList); err != nil {
		tlog.ErrorPrintf("Unable to get Ethdev List information: %v\n", err)
		return
	}
	tlog.DebugPrintf("EthdevList: %v\n", pg.infoDPDK.PidList)
}

func (pg *DPDKPanel) getEthdevStats(a *pinfo.ConnInfo) {

	// Clear the previous stats
	pg.infoDPDK.EthdevStats = nil

	// Output the basic data for the stats and information of a port
	for _, pid := range pg.infoDPDK.PidList.Pids {

		eth := dpdk.EthdevStats{}
		cmd := fmt.Sprintf("/ethdev/stats,%d", pid)
		if err := pg.pinfoDPDK.Unmarshal(a, cmd, &eth); err != nil {
			tlog.WarnPrintf("Unable to get Ethdev Stats for Port %d\n", pid)
			continue
		}
		eth.Stats.PortID = pid
		pg.infoDPDK.EthdevStats = append(pg.infoDPDK.EthdevStats, &eth)
		tlog.DebugPrintf("/ethdev/stats,%d: %+v\n", pid, eth)

		// Update the previous stats
		pg.infoDPDK.PrevStats[eth.Stats.PortID].Stats = eth.Stats

		tlog.DebugPrintf("Prev: %+v\n", pg.infoDPDK.PrevStats[eth.Stats.PortID])
	}
}

func (pg *DPDKPanel) collectStats() {

	a, err := pg.selectedConnection()
	if err != nil {
		tlog.DebugPrintf("No connection selected %s\n", err)
		return
	}
	pg.getFixedData(a)
	pg.getEthdevStats(a)
}

// Display the basic DPDK application information
func (pg *DPDKPanel) displayDPDKInfo(view *tview.TextView) {

	if view == nil {
		tlog.DoPrintf("displayDPDKInfo: called\n")
		return
	}

	w := -14

	info := pg.infoDPDK
	// Set the speed/duplex and rate in the window
	str := fmt.Sprintf("%s: %s\n", cz.Orange("DPDK Version", w), cz.LightGreen(info.Version))

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

		mbpsRx += BitRate(pktsIn, bytesIn)
		mbpsTx += BitRate(pktsOut, bytesOut)

		tlog.DebugPrintf("%d: Bytes in/out %d/%d, Pkts in/out %d/%d, Mbps in/out %.2f/%.2f\n",
			eth.Stats.PortID, bytesIn, bytesOut, pktsIn, pktsOut, mbpsRx, mbpsTx)

		pg.infoDPDK.PrevStats[eth.Stats.PortID] = *eth

		// Update the previous stats
		pg.infoDPDK.PrevStats[eth.Stats.PortID].Stats = eth.Stats

		tlog.DebugPrintf("Prev: %+v\n\n", pg.infoDPDK.PrevStats[eth.Stats.PortID])
	}
	tlog.DebugPrintf("\n")

	pg.data.rxPoints.GraphPoints(0).AddPoint(mbpsRx / (1024.0 * 1024.0))
	pg.data.txPoints.GraphPoints(0).AddPoint(mbpsTx / (1024.0 * 1024.0))
}

// Display to update the graphs on the panel
func (pg *DPDKPanel) displayChart(view *tview.TextView, rx bool) {

	if rx {
		view.SetText(pg.data.rxPoints.MakeChart(view))
	} else {
		view.SetText(pg.data.txPoints.MakeChart(view))
	}
}
