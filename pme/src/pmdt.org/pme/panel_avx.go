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

// Create and display the PBF or Power Base Frequency information.
// Older system like Ubuntu do not have the kernel modules loaded or avaiable.
// Ubuntu 18.04 does have the files in the /sys directory.

// PageAVX - Data for AVX - Power Base Frequency
type PageAVX struct {
	tabOrder         *tab.Tab
	topFlex          *tview.Flex
	title            *tview.Box
	selectCore       *SelectWindow
	pbf              *tview.Table
	chart            *tview.TextView
	selected         int
	selectionChanged bool
	freqs            *graphdata.GraphInfo
}
       
const (
	avxPanelName string = "AVX"
	maxAVXPoints int    = 120
)

// Setup and create the PBF page structure
func setupAVX() *PageAVX {

	pg := &PageAVX{}

	pg.freqs = graphdata.NewGraph(NumCPUs())
	for _, gd := range pg.freqs.Graphs() {
		gd.SetMaxPoints(maxAVXPoints)
	}
	pg.freqs.SetFieldWidth(5)

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

	pg.pbf = CreateTableView(flex1, "AVX Power Base Frequency (p)", tview.AlignLeft, 0, 2, true)
	pg.pbf.SetFixed(1, 0)
	pg.pbf.SetSeparator(tview.Borders.Vertical)

	flex0.AddItem(flex1, 0, 3, true)

	pg.chart = CreateTextView(flex2, "CPU 0 (C)", tview.AlignLeft, 0, 1, true)

	flex0.AddItem(flex2, 0, 1, true)

	to.Add(pg.selectCore.table, 'c')
	to.Add(pg.pbf, 'p')
	to.Add(pg.chart, 'C')

	to.SetInputDone()

	// Create timer and callback function to display and process PBF data
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

// Display the PBF data in the windows created
func (pg *PageAVX) displayAVXPage() {

	pg.displayAVX(pg.pbf)
	pg.displayFreqChart()

	if pg.selectionChanged {
		pg.selectionChanged = false
		pg.pbf.ScrollToBeginning()
		pg.chart.ScrollToBeginning()
		pg.chart.SetTitle(TitleColor(fmt.Sprintf("CPU %d (C)", pg.selectCore.ItemIndex())))
	}
}

// Collect the graph data to be displayed in the chart window
func (pg *PageAVX) collectChartData() {

	for cpu, gd := range pg.freqs.Graphs() {
		p := pbf.InfoPerCPU(cpu)

		// Append the frequency data to the list for the graphing in a chart
		gd.AddPoint(float64(p.CurFreq))
	}
}

// Display the PBF data in the table view
func (pg *PageAVX) displayAVX(view *tview.Table) {

	// create the headers for each column
	SetCell(pg.pbf, 0, 0, cz.Orange("CPU", 4))
	SetCell(pg.pbf, 0, 1, cz.Orange("Max", 6))
	SetCell(pg.pbf, 0, 2, cz.Orange("Min", 6))
	SetCell(pg.pbf, 0, 3, cz.Orange("Curr", 6))
	SetCell(pg.pbf, 0, 4, cz.Orange("Governor", 10))

	// Display the CState names as columns
	p := pbf.InfoPerCPU(0)
	for j, v := range p.CStateNames {
		SetCell(pg.pbf, 0, 5+j, cz.Orange(v, 6))
	}

	// For the number of CPUs display the data one CPU per line
	for i := 0; i < NumCPUs(); i++ {
		p := pbf.InfoPerCPU(i)

		SetCell(pg.pbf, i+1, 0, cz.LightGreen(i))
		SetCell(pg.pbf, i+1, 1, cz.SkyBlue(p.MaxFreq))
		SetCell(pg.pbf, i+1, 2, cz.SkyBlue(p.MinFreq))
		SetCell(pg.pbf, i+1, 3, cz.LightGreen(p.CurFreq))
		SetCell(pg.pbf, i+1, 4, cz.CornSilk(p.Governor))

		// Output the CStates per CPU per line
		for j, v := range p.CStates {
			SetCell(pg.pbf, i+1, 5+j, cz.LightGreen(v, 6))
		}
	}
}

// Display the Frequency values in the text view using a graph or chart
func (pg *PageAVX) displayFreqChart() {

	pg.chart.SetText(pg.freqs.MakeChart(pg.chart, pg.selected, pg.selected))
}
