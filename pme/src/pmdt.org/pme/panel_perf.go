// SPDX-License-Identifier: BSD-3-Clause
// Copyright(c) 2019-2020 Intel Corporation

package main

import (
	"github.com/rivo/tview"
	tab "pmdt.org/taborder"
)

// PagePerf - Data for main page information
type PagePerf struct {
	tabOrder *tab.Tab
	topFlex  *tview.Flex
	title    *tview.Box
	perf     *tview.Table
}

const (
	perfPanelName string = "Perf"
)

// setupPerf - setup and init the main page
func setupPerf() *PagePerf {

	pg := &PagePerf{}

	return pg
}

// PerfPanelSetup setup the main event page
func PerfPanelSetup(nextSlide func()) (pageName string, content tview.Primitive) {

	pg := setupPerf()

	to := tab.New(perfPanelName, perfmon.app)
	pg.tabOrder = to

	top := tview.NewFlex().SetDirection(tview.FlexRow)
	flex := tview.NewFlex().SetDirection(tview.FlexRow)

	TitleBox(top)
	pg.topFlex = top
	pg.perf = CreateTableView(flex, "Perf (p)", tview.AlignLeft, 0, 2, true)

	top.AddItem(flex, 0, 3, true)

	to.Add(pg.perf, 'p')

	to.SetInputDone()

	perfmon.timers.Add(perfPanelName, func(step int, ticks uint64) {
		if pg.topFlex.HasFocus() {
			perfmon.app.QueueUpdateDraw(func() {
				pg.displayPerfPage(step, ticks)
			})
		}
	})

	return perfPanelName, top
}

// Display the perf information
func (pg *PagePerf) displayPerfPage(step int, ticks uint64) {

	pg.displayPerf(pg.perf)
}

// Display the perf information
func (pg *PagePerf) displayPerf(view *tview.Table) {

}
