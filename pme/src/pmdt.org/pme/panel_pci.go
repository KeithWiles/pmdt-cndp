// SPDX-License-Identifier: BSD-3-Clause
// Copyright(c) 2019-2020 Intel Corporation

package main

import (
	"bufio"
	"fmt"
	"os/exec"
	"sync"

	"github.com/rivo/tview"
	cz "pmdt.org/colorize"
	"pmdt.org/graphdata"
	"pmdt.org/pcm"
	tab "pmdt.org/taborder"
	tlog "pmdt.org/ttylog"
)

// Display the PCIe data from the shared memory region.
// Encode the shared memory data for PCIe into a Go structure and display the data

// PagePCI - Data for main page information
type PagePCI struct {
	pcmRunning bool
	cmd        *exec.Cmd
	tabOrder   *tab.Tab
	topFlex    *tview.Flex
	title      *tview.Box
	pci        *tview.Table
	legend     [2]*tview.TextView
	mmio       *tview.TextView
	note       *tview.TextView
	pciCharts  [2]*tview.TextView
	pciScanner *bufio.Scanner
	pcmState   *pcm.SharedPCMState
	charts     *graphdata.GraphInfo
	pciRedraw  bool
	once       sync.Once
}

const (
	pciPanelName string = "PCI"
	maxPCIPoints int    = 56
)

// setupPCI - setup and init the main page
func setupPCI() *PagePCI {

	pg := &PagePCI{pcmRunning: false}

	pg.charts = graphdata.NewGraph(2)
	for _, gd := range pg.charts.Graphs() {
		gd.SetMaxPoints(maxPCIPoints)
	}
	pg.charts.SetFieldWidth(6)

	return pg
}

// PCIPanelSetup setup the main event page
func PCIPanelSetup(nextSlide func()) (pageName string, content tview.Primitive) {

	pg := setupPCI()

	to := tab.New(pciPanelName, perfmon.app)
	pg.tabOrder = to

	// Create all of the flex boxes to be use
	flex0 := tview.NewFlex().SetDirection(tview.FlexRow)
	flex1 := tview.NewFlex().SetDirection(tview.FlexRow)
	flex1c := tview.NewFlex().SetDirection(tview.FlexColumn)
	flex2 := tview.NewFlex().SetDirection(tview.FlexColumn)
	flex3 := tview.NewFlex().SetDirection(tview.FlexColumn)

	TitleBox(flex0)
	pg.topFlex = flex0

	// The PCI window is the view into the PCI stats and counters
	pg.pci = CreateTableView(flex1, "Stats (p)", tview.AlignLeft, 11, 2, true)
	pg.pci.SetSeparator(tview.Borders.Vertical)

	// Chart some of the PCI counters
	pg.pciCharts[0] = CreateTextView(flex3, "PCI Charts (1)", tview.AlignLeft, 0, 1, true)
	pg.pciCharts[1] = CreateTextView(flex3, "PCI Charts (2)", tview.AlignLeft, 0, 1, true)

	flex1.AddItem(flex3, 0, 2, true)
	pg.legend[0] = CreateTextView(flex1c, "Legend", tview.AlignLeft, 0, 1, true)
	pg.legend[1] = CreateTextView(flex1c, "Legend", tview.AlignLeft, 0, 1, true)

	flex1.AddItem(flex1c, 15, 2, true)

	pg.note = CreateTextView(flex2, "Note", tview.AlignLeft, 0, 1, true)
	pg.mmio = CreateTextView(flex2, "MMIO", tview.AlignLeft, 0, 1, true)

	flex1.AddItem(flex2, 5, 1, true)

	flex0.AddItem(flex1, 0, 2, true)

	to.Add(pg.pci, 'p')

	// pg.note does not have a focus key
	to.Add(pg.pciCharts[0], '1')
	to.Add(pg.pciCharts[1], '2')

	to.SetInputDone()

	pg.addLegend()
	pg.pciRedraw = true
	perfmon.timers.Add(pciPanelName, func(step int, ticks uint64) {
		if pg.topFlex.HasFocus() {
			perfmon.app.QueueUpdateDraw(func() {
				pg.displayPCIPage(step, ticks)
			})
		}
	})

	return pciPanelName, pg.topFlex
}

// Display the legend text to help understand the data on the page
func (pg *PagePCI) addLegend() {

	legend1 := pg.legend[0]
	legend2 := pg.legend[1]
	mmio := pg.mmio
	note := pg.note

	s := cz.SkyBlue("PCIe event definitions (each event counts as a transfer):\n\n")
	s += fmt.Sprintf(" %s (Devices reading from memory -\n", cz.DeepPink("PCIe Read Events"))
	s += fmt.Sprintf("      application writes to disk/network/PCIe device):\n\n")
	s += fmt.Sprintf("   %s - UC read transfer (partial cache line)\n", cz.DeepPink("PartialRd"))
	s += fmt.Sprintf("   %s%s - Read Current transfer (full cache line)\n", cz.DeepPink("ReadCurr"), cz.Yellow("*"))
	s += fmt.Sprintf("               On Haswell Server Read Current counts both\n")
	s += fmt.Sprintf("               full/partial cache lines\n")
	s += fmt.Sprintf("   %s%s - Demand Data RFO\n", cz.DeepPink("RdOwner"), cz.Yellow("*"))
	s += fmt.Sprintf("   %s%s - Demand Code Read\n", cz.DeepPink("CodeRd"), cz.Yellow("*"))
	s += fmt.Sprintf("   %s - Demand Data Read\n", cz.DeepPink("DataRd"))
	s += fmt.Sprintf("   %s - Non-snoop write transfer (partial cache line)\n", cz.DeepPink("NonSnoopWr"))
	legend1.SetText(s)

	s = cz.SkyBlue("PCIe event definitions (each event counts as a transfer):\n\n")
	s += fmt.Sprintf(" %s (Devices writing to memory -\n", cz.DeepPink("PCIe Write Events"))
	s += fmt.Sprintf("      application reads from disk/network/PCIe device):\n\n")
	s += fmt.Sprintf("   %s - Write transfer (non-allocating) full cacheline\n", cz.DeepPink("WrNonAlloc"))
	s += fmt.Sprintf("   %s - Write transfer (allocating) (full cacheline)\n", cz.DeepPink("WrAlloc"))
	s += fmt.Sprintf("   %s - Non-snoop write transfer partial cacheline\n", cz.DeepPink("NonSnoopWrPart"))
	s += fmt.Sprintf("   %s - Non-snoop write transfer (full cache line)\n", cz.DeepPink("NonSnoopWrFull"))
	s += fmt.Sprintf("   %s - Write full cache line\n", cz.DeepPink("WrInvalid"))
	s += fmt.Sprintf("   %s - Partial Write\n", cz.DeepPink("Rd4Owner"))
	s += fmt.Sprintf("   %s - Read Bandwidth\n", cz.DeepPink("RdBandWidth"))
	s += fmt.Sprintf("   %s - Write Bandwidth\n", cz.DeepPink("WrBandWidth"))
	legend2.SetText(s)

	s = ""
	s += fmt.Sprintf(" %s (CPU reading/writing to PCIe devices):\n", cz.DeepPink("CPU MMIO Events"))
	s += fmt.Sprintf("   %s - MMIO Read [Haswell Server only] Partial Cacheline\n", cz.DeepPink("PartialRd"))
	s += fmt.Sprintf("   %s - MMIO Write (Full/Partial)\n\n", cz.DeepPink("ReqInvalid"))
	mmio.SetText(s)

	s = ""
	s += cz.SkyBlue(fmt.Sprintf("%s %s\n", cz.Yellow("*"), cz.SkyBlue("Depending on the configuration of your BIOS, this tool")))
	s += cz.SkyBlue("  may report '0' if the message has not been selected.")
	note.SetText(s)
}

// Display the PCI data if valid and collect the points for the charts
func (pg *PagePCI) displayPCIPage(step int, ticks uint64) {

	pg.once.Do(func() {
		tlog.DebugPrintf(DumpSharedMemory())
	})

	switch step {
	case 0: // Display the data that was gathered every second
		if perfmon.pcmData == nil {
			break
		}
		if ok := perfmon.pcmData.Lock(); ok {
			defer perfmon.pcmData.Unlock()
			pg.pcmState = perfmon.pcmData.State()

			if pg.pcmState == nil {
				return
			}
			pg.collectChartData()
			pg.displayPCI(pg.pci)
			pg.displayCharts(pg.pciCharts[0], 0, 0)
			pg.displayCharts(pg.pciCharts[1], 1, 1)
		}
	}
}

// Collect the data from for charts
func (pg *PagePCI) collectChartData() {

	for i, gd := range pg.charts.Graphs() {
		p := pg.pcmState

		gd.AddPoint(float64(p.Sample[i].Total.ReadForOwnership))
		gd.SetName(fmt.Sprintf("ReadForOwnership %d", i))
	}
}

// Display the PCI information into the window or table view object
func (pg *PagePCI) displayPCI(view *tview.Table) {

	row := 0

	// Set the column headers for the PCIe data
	for i, s := range []string{"Socket", "ReadCurr",
		"Rd4Owner", "CodeRd", "DataRd", "ReqInvalid", "PartialRd", "WrInvalid", "RdBandWidth", "WrBandWidth"} {
		SetCell(view, row, i, cz.Orange(s, 9), tview.AlignRight)
	}
	row++

	// Put the counters in the tabel view
	num := int(pg.pcmState.PCMCounters.System.NumOfSockets)
	for i := 0; i < num; i++ {

		// Add the Total PCI counter values in the table
		s := pg.pcmState.Sample[i].Total
		SetCell(view, row, 0, cz.Orange(fmt.Sprintf("Total %d", i)), tview.AlignLeft)
		SetCell(view, row, 1, cz.SkyBlue(FormatUnits(s.ReadCurrent)), tview.AlignRight)
		SetCell(view, row, 2, cz.SkyBlue(FormatUnits(s.ReadForOwnership)), tview.AlignRight)
		SetCell(view, row, 3, cz.SkyBlue(FormatUnits(s.DemandCodeRd)), tview.AlignRight)
		SetCell(view, row, 4, cz.SkyBlue(FormatUnits(s.DemandDataRd)), tview.AlignRight)
		SetCell(view, row, 5, cz.SkyBlue(FormatUnits(s.RequestInvalidateLine)), tview.AlignRight)
		SetCell(view, row, 6, cz.SkyBlue(FormatUnits(s.PartialRead)), tview.AlignRight)
		SetCell(view, row, 7, cz.SkyBlue(FormatUnits(s.WriteInvalidateLine)), tview.AlignRight)
		SetCell(view, row, 8, cz.SkyBlue(FormatUnits(s.ReadBandWidth)), tview.AlignRight)
		SetCell(view, row, 9, cz.SkyBlue(FormatUnits(s.WriteBandWidth)), tview.AlignRight)
		row++

		// Add the Missed PCI counter values in the table
		s = pg.pcmState.Sample[i].Miss
		SetCell(view, row, 0, cz.Orange(fmt.Sprintf("Miss  %d", i)), tview.AlignLeft)
		SetCell(view, row, 1, cz.SkyBlue(FormatUnits(s.ReadCurrent)), tview.AlignRight)
		SetCell(view, row, 2, cz.SkyBlue(FormatUnits(s.ReadForOwnership)), tview.AlignRight)
		SetCell(view, row, 3, cz.SkyBlue(FormatUnits(s.DemandCodeRd)), tview.AlignRight)
		SetCell(view, row, 4, cz.SkyBlue(FormatUnits(s.DemandDataRd)), tview.AlignRight)
		SetCell(view, row, 5, cz.SkyBlue(FormatUnits(s.RequestInvalidateLine)), tview.AlignRight)
		SetCell(view, row, 6, cz.SkyBlue(FormatUnits(s.PartialRead)), tview.AlignRight)
		SetCell(view, row, 7, cz.SkyBlue(FormatUnits(s.WriteInvalidateLine)), tview.AlignRight)
		SetCell(view, row, 8, cz.SkyBlue(FormatUnits(s.ReadBandWidth)), tview.AlignRight)
		SetCell(view, row, 9, cz.SkyBlue(FormatUnits(s.WriteBandWidth)), tview.AlignRight)
		row++

		// Add the cache Hit PCI counter values in the table
		s = pg.pcmState.Sample[i].Hit
		SetCell(view, row, 0, cz.Orange(fmt.Sprintf("Hit   %d", i)), tview.AlignLeft)
		SetCell(view, row, 1, cz.SkyBlue(FormatUnits(s.ReadCurrent)), tview.AlignRight)
		SetCell(view, row, 2, cz.SkyBlue(FormatUnits(s.ReadForOwnership)), tview.AlignRight)
		SetCell(view, row, 3, cz.SkyBlue(FormatUnits(s.DemandCodeRd)), tview.AlignRight)
		SetCell(view, row, 4, cz.SkyBlue(FormatUnits(s.DemandDataRd)), tview.AlignRight)
		SetCell(view, row, 5, cz.SkyBlue(FormatUnits(s.RequestInvalidateLine)), tview.AlignRight)
		SetCell(view, row, 6, cz.SkyBlue(FormatUnits(s.PartialRead)), tview.AlignRight)
		SetCell(view, row, 7, cz.SkyBlue(FormatUnits(s.WriteInvalidateLine)), tview.AlignRight)
		SetCell(view, row, 8, cz.SkyBlue(FormatUnits(s.ReadBandWidth)), tview.AlignRight)
		SetCell(view, row, 9, cz.SkyBlue(FormatUnits(s.WriteBandWidth)), tview.AlignRight)

		row++
	}
	row++

	// Add the Aggregate PCI counter values in the table
	s := pg.pcmState.Aggregate
	SetCell(view, row, 0, cz.Orange("Aggregate"), tview.AlignLeft)
	SetCell(view, row, 1, cz.Orange(FormatUnits(s.ReadCurrent)), tview.AlignRight)
	SetCell(view, row, 2, cz.Orange(FormatUnits(s.ReadForOwnership)), tview.AlignRight)
	SetCell(view, row, 3, cz.Orange(FormatUnits(s.DemandCodeRd)), tview.AlignRight)
	SetCell(view, row, 4, cz.Orange(FormatUnits(s.DemandDataRd)), tview.AlignRight)
	SetCell(view, row, 5, cz.Orange(FormatUnits(s.RequestInvalidateLine)), tview.AlignRight)
	SetCell(view, row, 6, cz.Orange(FormatUnits(s.PartialRead)), tview.AlignRight)
	SetCell(view, row, 7, cz.Orange(FormatUnits(s.WriteInvalidateLine)), tview.AlignRight)
	SetCell(view, row, 8, cz.Orange(FormatUnits(s.ReadBandWidth)), tview.AlignRight)
	SetCell(view, row, 9, cz.Orange(FormatUnits(s.WriteBandWidth)), tview.AlignRight)

	// If required redraw the display data
	if pg.pciRedraw {
		pg.pciRedraw = false
		view.ScrollToBeginning()
	}
}

// Create and display charts of the PCIe data
func (pg *PagePCI) displayCharts(view *tview.TextView, start, end int) {
	view.SetText(pg.charts.MakeChart(view, start, end))
}
