// SPDX-License-Identifier: BSD-3-Clause
// Copyright(c) 2019-2020 Intel Corporation

package main

import (
	"fmt"
	"math"
	"sync"

	"github.com/rivo/tview"

	"pmdt.org/dpdk"
	"pmdt.org/graphdata"
	pcm "pmdt.org/pcm"
	"pmdt.org/pinfo"

	cz "pmdt.org/colorize"
	// pbf "pmdt.org/intelpbf"
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
	dpdkBusy *tview.TextView //Table
	totalRX  *tview.TextView
	totalTX  *tview.TextView

	pinfoDPDK *pinfo.ProcessInfo
	infoDPDK  dpdk.Information

	system     pcm.System
	data       *rxtxData
	appCoreMap map[uint16][]uint16 //TODO
	percent    []float64
}

// Setup the DPDK Panel data structure
func setupDPDKPanel() *DPDKPanel {

	pg := &DPDKPanel{}

	pg.data = &rxtxData{}

	pg.appCoreMap = make(map[uint16][]uint16)

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
	pg.dpdkBusy = CreateTextView(flex2, "DPDK Core Busy Stats (b)", tview.AlignLeft, 0, 4, false)
	pg.dpdkNet.SetFixed(2, 0)
	pg.dpdkNet.SetSeparator(tview.Borders.Vertical)
	flex2.AddItem(flex3, 0, 3, false)

	pg.totalRX = CreateTextView(flex3, "Total RX Mbps", tview.AlignLeft, 0, 1, false)
	pg.totalTX = CreateTextView(flex3, "Total TX Mbps", tview.AlignLeft, 0, 1, false)

	to.Add(pg.selectApp.table, '1')
	to.Add(pg.dpdkInfo, '2')
	to.Add(pg.dpdkNet, '3')
	to.Add(pg.dpdkBusy, 'b')

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

// selectedConnection returns the DPDK app name that is selected
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
		pg.displayDPDKBusy(pg.dpdkBusy)
		pg.displayChart(pg.totalRX, true)
		pg.displayChart(pg.totalTX, false)
	}
}

func (pg *DPDKPanel) getFixedData(a *pinfo.ConnInfo) {

	pg.infoDPDK.Version = pg.pinfoDPDK.Version(a)
	tlog.DebugPrintf("EAL Version: %s\n", pg.infoDPDK.Version)

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

// collectBusyData collect the cores branch and missed branches stats
func (pg *PageCore) collectBusyData() {

	core := pcm.CoreCounters{}
	if err := perfmon.pinfoPCM.Unmarshal(nil, fmt.Sprintf("/pcm/core,%d", pg.selected), &core); err != nil {
		tlog.ErrorPrintf("Unable to get PCM system information\n")
		return
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

// displayDPDKInfo display the basic DPDK application information
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

// displayDPDKNet display some Network information about the DPDK application
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

// parseCoreMask parse the DPDK cores
//func (pg *DPDKPanel) parseCoremask(int mask) {

// DPDK list of cores
//x := []int { 1, 2, 3 }

// use bit operator
//dpdkCores := append(dpdkCores, m)

//}

/*
// getDPDKCores parse the coremask from the eal option of the selected DPDK app
func (pg *DPDKPanel) getDPDKCores() {
	//info := pg.infoDPDK
	//
	//dpdkCores := []int {0}

	// Check that a coremask was passed, if not then use value 1
	if err := strings.Contains(info.params.params, "-c"); err != nil {
		tlog.WarnPrintf("Unable to get Cores for DPDK App: using default 0 value")
	//dpdkCores := 0
	//return dpdkCores
	}
	//if err == nil {
	dpdkCoresParam := strings.SplitAfter(info.params.params, "-c ")
	//}
	// Grab core value, chop off the rest of the params
	// until next ' '
	dpdkCoresParam = strings.SplitAfter(dpdkCoresParam, " ")

	st := 1
	// For the length of the string add the values to
	for st != NULL { //st != ERROR && c && (! end of st) ) {
		//debug_printf("%c %x -> ", c, c);
		switch (st) {
			case 1: //START
				if dpdkCoresParam.Contains(x) {
					st = 2
				}
			case 2: // HEX
				// If x exists, coremask
				// Parse the coremask from the selected app
				// replace 0x or 0X with empty String
				mask := strings.Replace(dpdkCoresParam, "0x", "", -1)
				mask = strings.Replace(mask, "0X", "", -1)

				// convert to decimal then binary
				mask = strconv.FormatInt(mask, 10)
				mask = strconv.FormatInt(mask, 2)

				temp := parseCoremask(mask)
				// Add the core values of the params to the list of cores
				// for the length of temp, add to the dpdkCores list


				// Append dpdkCores from the coremask parser
				dpdkCores.append(temp)
		}
	}
	// Eal params stored in info.params.params

}
*/

// Display some Busy/Branch information about the DPDK application
func (pg *DPDKPanel) displayDPDKBusy(view *tview.TextView) {
	if view == nil {
		tlog.DoPrintf("displayDPDKBusy: view is nil\n")
		return
	}
	/*
		p := perfmon.pinfoPCM.ConnectionList()
		if len(p) == 0 {
			return
		} */

	pg.percent[0] = 100.0

	// Pass the beginning and ending of the coremask (0-4 for poc)
	pg.displayBusy(pg.percent, 0, 4, view)

	/*
		row := 0
		col := 0
		// num is EAL coremask of selected app
		num := int(pg.system.Data.NumOfCores)
		label := []string{
			"Core/Socket", "", "Branches", "BranchMispredicts", "PercentBusy",
		}
		for i, t := range label {
			SetCell(view, row+i, col, cz.Wheat(t))
		}
		row++
		for i, j := 0, row; i < num; i++ {
			data := pcm.CoreCounters{}
			if err := perfmon.pinfoPCM.Unmarshal(nil, fmt.Sprintf("/pcm/core,%d", i), &data); err != nil {
				tlog.ErrorPrintf("Unable to get PCM system information\n")
				return
			}
			core := data.Data
			total := 100.0
			percent := (core.BranchMispredicts / core.Branches)
			SetCell(view, row, j+0, cz.Orange(fmt.Sprintf("%d/%d", core.CoreID, core.SocketID)))
			SetCell(view, row, j+2, cz.SkyBlue(core.Branches))
			SetCell(view, row, j+3, cz.SkyBlue(core.BranchMispredicts))
			p := clamp(float64(percent), 0.0, total)
			SetCell(view, row, j+4, cz.SkyBlue(p))
			row++
		}
	*/
}

// Display the busy meters
func (pg *DPDKPanel) displayBusy(percent []float64, start, end int16, view *tview.TextView) {
	_, _, width, _ := view.GetInnerRect()
	width -= 14
	if width <= 0 {
		return
	}
	str := ""
	str += fmt.Sprintf("%s\n", cz.Orange("Busy Percent          Load Meter"))
	for i := start; i < end; i++ {
		str += pg.drawBusyMeter(i, percent[i], width)
	}
	view.SetText(str)
	view.ScrollToBeginning()
}

// Draw the meter for the busy ratio
func (pg *DPDKPanel) drawBusyMeter(id int16, percent float64, width int) string {
	total := 100.0

	p := clamp(percent, 0.0, total)
	if p > 0 {
		p = math.Ceil((p / total) * float64(width))
	}

	bar := make([]byte, width)

	for i := 0; i < width; i++ {
		if i <= int(p) {
			bar[i] = '|'
		} else {
			bar[i] = ' '
		}
	}
	str := fmt.Sprintf(" %2d:%s%% [%s]\n",
		id, cz.Red(percent, 5, 1), cz.LightGreen(string(bar)))

	return str
}

// Display to update the graphs on the panel
func (pg *DPDKPanel) displayChart(view *tview.TextView, rx bool) {
	if rx {
		view.SetText(pg.data.rxPoints.MakeChart(view))
	} else {
		view.SetText(pg.data.txPoints.MakeChart(view))
	}
}
