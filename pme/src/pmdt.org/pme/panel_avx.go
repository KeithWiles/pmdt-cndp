// SPDX-License-Identifier: BSD-3-Clause
// Copyright(c) 2019-2020 Intel Corporation

package main

import (
	"fmt"

	"github.com/rivo/tview"
	cz "pmdt.org/colorize"
	"pmdt.org/graphdata"
	pbf "pmdt.org/intelpbf"
	tab "pmdt.org/taborder"
)

// Create and display the AVX information.
// Older system like Ubuntu do not have the kernel modules loaded or avaiable.
// Ubuntu 18.04 does have the files in the /sys directory.

// PageAVX - Data for AVX - Power Base Frequency
type PageAVX struct {
	tabOrder   *tab.Tab
	topFlex    *tview.Flex
	title      *tview.Box
	selectCore *SelectWindow
	avxStats   *tview.Table
	avxThermal *tview.Table
	//turboCharts      [3]*tview.TextView
	chart            *tview.TextView
	turbo1           *tview.TextView
	turbo2           *tview.TextView
	turbo3           *tview.TextView
	selected         int
	selectionChanged bool
	freqs            *graphdata.GraphInfo
	turbo1freqs      *graphdata.GraphInfo
	turbo2freqs      *graphdata.GraphInfo
	turbo3freqs      *graphdata.GraphInfo
}

const (
	avxPanelName   string = "AVX"
	maxTurboPoints int    = 120
)

// Setup and create the AVX page structure
func setupAVX() *PageAVX {

	pg := &PageAVX{}

	pg.freqs = graphdata.NewGraph(NumCPUs())
	for _, gd := range pg.freqs.Graphs() {
		gd.SetMaxPoints(maxPBFPoints)
	}
	pg.freqs.SetFieldWidth(5)

	pg.turbo1freqs = graphdata.NewGraph(NumCPUs())
	for _, gd := range pg.turbo1freqs.Graphs() {
		gd.SetMaxPoints(maxTurboPoints)
	}
	pg.turbo1freqs.SetFieldWidth(5)

	pg.turbo2freqs = graphdata.NewGraph(NumCPUs())
	for _, gd := range pg.turbo2freqs.Graphs() {
		gd.SetMaxPoints(maxTurboPoints)
	}
	pg.turbo2freqs.SetFieldWidth(5)

	pg.turbo3freqs = graphdata.NewGraph(NumCPUs())
	for _, gd := range pg.turbo3freqs.Graphs() {
		gd.SetMaxPoints(maxTurboPoints)
	}
	pg.turbo3freqs.SetFieldWidth(5)

	pg.selectionChanged = true

	return pg
}

// AVXPanelSetup setup the main event page
func AVXPanelSetup(nextSlide func()) (pageName string, content tview.Primitive) {

	pg := setupAVX()

	to := tab.New(avxPanelName, perfmon.app)
	pg.tabOrder = to

	// Flex boxes used to hold tview window types
	flex0 := tview.NewFlex().SetDirection(tview.FlexRow)
	flex1 := tview.NewFlex().SetDirection(tview.FlexColumn)
	flex2 := tview.NewFlex().SetDirection(tview.FlexRow)
	flex3 := tview.NewFlex().SetDirection(tview.FlexColumn)

	// Create the top window for basic information about tool and panel
	TitleBox(flex0)
	pg.topFlex = flex0

	// Core selection window to be able to select a core to view
	table := CreateTableView(flex1, "Core (c)", tview.AlignLeft, 12, 1, true)

	// Select window setup and callback function when selection changes.
	pg.selectCore = NewSelectWindow(table, "AVX", 0, func(row, col int) {

		if row != pg.selected {
			pg.selectCore.UpdateItem(row, col)

			pg.selectionChanged = true

			pg.selected = row
			pg.chart.SetTitle(TitleColor(fmt.Sprintf("CPU %d (C)", pg.selected)))
		}
	})

	names := make([]interface{}, 0)

	for i := 0; i < NumCPUs(); i++ {
		s := fmt.Sprintf("%4d", i)
		names = append(names, s)
	}
	pg.selectCore.AddColumn(-1, names, cz.SkyBlueColor)

	pg.avxStats = CreateTableView(flex1, "AVX Power Base Frequency (p)", tview.AlignLeft, 0, 2, true)
	pg.avxStats.SetFixed(1, 0)
	pg.avxStats.SetSeparator(tview.Borders.Vertical)

	pg.avxThermal = CreateTableView(flex1, "AVX Thermal & Busy Freq (t)", tview.AlignLeft, 0, 2, true)
	pg.avxThermal.SetFixed(1, 0)
	pg.avxThermal.SetSeparator(tview.Borders.Vertical)

	flex0.AddItem(flex1, 0, 3, true)

	pg.chart = CreateTextView(flex2, "CPU 0 (C)", tview.AlignLeft, 0, 1, true)
	flex0.AddItem(flex2, 0, 1, true)

	pg.turbo1 = CreateTextView(flex3, "Turbo AVX2 Light (1)", tview.AlignLeft, 0, 1, true)
	pg.turbo2 = CreateTextView(flex3, "Turbo AVX512 Light (2)", tview.AlignLeft, 0, 1, false)
	pg.turbo3 = CreateTextView(flex3, "Turbo AVX512 Heavy (3)", tview.AlignLeft, 0, 1, false)
	flex0.AddItem(flex3, 0, 2, true)

	to.Add(pg.selectCore.table, 'c')
	to.Add(pg.avxStats, 'p')
	to.Add(pg.chart, 'C')

	to.Add(pg.turbo1, '1')
	to.Add(pg.turbo2, '2')
	to.Add(pg.turbo3, '3')

	to.SetInputDone()

	// Create timer and callback function to display and process AVX data
	perfmon.timers.Add(avxPanelName, func(step int, ticks uint64) {
		// up to 4 cases, done every second
		switch step {
		case 0:
			pg.collectChartData()
		case 1:
			if pg.topFlex.HasFocus() {
				perfmon.app.QueueUpdateDraw(func() {
					pg.displayAVXPage()
				})
			}
		}

	})

	return avxPanelName, pg.topFlex
}

// Display the AVX data in the windows created
func (pg *PageAVX) displayAVXPage() {

	pg.displayAVX(pg.avxStats)
	pg.displayAVXTherm(pg.avxThermal)
	pg.displayFreqChart()

	if pg.selectionChanged {
		pg.selectionChanged = false
		pg.avxStats.ScrollToBeginning()
		pg.chart.ScrollToBeginning()
		pg.chart.SetTitle(TitleColor(fmt.Sprintf("CPU %d (C)", pg.selectCore.ItemIndex())))
		pg.turbo1.ScrollToBeginning()
		pg.turbo2.ScrollToBeginning()
		pg.turbo3.ScrollToBeginning()
	}
}

// Collect the graph data to be displayed in the chart window
func (pg *PageAVX) collectChartData() {

	for cpu, gd := range pg.freqs.Graphs() {
		p := pbf.InfoPerCPU(cpu)

		// Append the frequency data to the list for the graphing in a chart
		gd.AddPoint(float64(p.CurFreq))
	}

	for cpu, gd := range pg.turbo1freqs.Graphs() {
		p := pbf.AVXInfoPerCPU(cpu)

		// Append the frequency data to the list for the graphing in a chart
		gd.AddPoint(float64(p.Turbo1))
	}

	for cpu, gd := range pg.turbo2freqs.Graphs() {
		p := pbf.AVXInfoPerCPU(cpu)

		// Append the frequency data to the list for the graphing in a chart
		gd.AddPoint(float64(p.Turbo2))
	}

	for cpu, gd := range pg.turbo3freqs.Graphs() {
		p := pbf.AVXInfoPerCPU(cpu)

		// Append the frequency data to the list for the graphing in a chart
		gd.AddPoint(float64(p.Turbo3))
	}

}

// Display the avxStats data in the table view
func (pg *PageAVX) displayAVX(view *tview.Table) {

	// create the headers for each column
	SetCell(pg.avxStats, 0, 0, cz.Orange("CPU", 4))
	SetCell(pg.avxStats, 0, 1, cz.Orange("128BLight", 6))
	SetCell(pg.avxStats, 0, 2, cz.Orange("128BHeavy", 6))
	SetCell(pg.avxStats, 0, 3, cz.Orange("256BLight", 6))
	SetCell(pg.avxStats, 0, 4, cz.Orange("256BHeavy", 6))
	SetCell(pg.avxStats, 0, 3, cz.Orange("512BLight", 6))
	SetCell(pg.avxStats, 0, 4, cz.Orange("512BHeavy", 6))

	// For the number of CPUs display the data one CPU per line
	for i := 0; i < NumCPUs(); i++ {
		//p := pbf.AVXInfoPerCPU(i)

		SetCell(pg.avxStats, i+1, 0, cz.LightGreen(i))
		/*
			SetCell(pg.avxStats, i+1, 1, cz.SkyBlue(p.128BLight))
			SetCell(pg.avxStats, i+1, 2, cz.SkyBlue(p.128BHeavy))
			SetCell(pg.avxStats, i+1, 3, cz.LightGreen(p.256BLight))
			SetCell(pg.avxStats, i+1, 4, cz.CornSilk(p.256BHeavy))
			SetCell(pg.avxStats, i+1, 3, cz.LightGreen(p.512BLight))
			SetCell(pg.avxStats, i+1, 4, cz.CornSilk(p.512BHeavy))
		*/

	}
}

// Display the PBF data in the table view
func (pg *PageAVX) displayAVXTherm(view *tview.Table) {

	// create the headers for each column
	SetCell(pg.avxThermal, 0, 0, cz.Orange("CPU", 4))
	SetCell(pg.avxThermal, 0, 1, cz.Orange("Busy Core Mhz", 6))
	SetCell(pg.avxThermal, 0, 2, cz.Orange("Avg Core Mhz", 6))
	SetCell(pg.avxThermal, 0, 3, cz.Orange("Core Temp", 6))
	SetCell(pg.avxThermal, 0, 4, cz.Orange("Pkg Temp", 10))

	//avx.mPerf = ReadMPerf(cpu)
	//avx.aPerf = ReadAPerf(cpu)
	//avx.thermStatus = ReadThermStatus(cpu)
	//avx.pkgThermStatus = ReadPkgThermStatus(cpu)

	// For the number of CPUs display the data one CPU per line
	for i := 0; i < NumCPUs(); i++ {
		p := pbf.AVXInfoPerCPU(i)

		SetCell(pg.avxThermal, i+1, 0, cz.LightGreen(i))
		SetCell(pg.avxThermal, i+1, 1, cz.SkyBlue(p.MPerf))
		SetCell(pg.avxThermal, i+1, 2, cz.SkyBlue(p.APerf))
		SetCell(pg.avxThermal, i+1, 3, cz.LightGreen(p.ThermStatus))
		SetCell(pg.avxThermal, i+1, 4, cz.CornSilk(p.PkgThermStatus))

	}
}

// Display the Frequency values in the text view using a graph or chart
func (pg *PageAVX) displayFreqChart() {

	pg.chart.SetText(pg.freqs.MakeChart(pg.chart, pg.selected, pg.selected))
	pg.chart.SetText(pg.turbo1freqs.MakeChart(pg.turbo1, pg.selected, pg.selected))
	pg.chart.SetText(pg.turbo2freqs.MakeChart(pg.turbo2, pg.selected, pg.selected))
	pg.chart.SetText(pg.turbo3freqs.MakeChart(pg.turbo3, pg.selected, pg.selected))
}
