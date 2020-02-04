// SPDX-License-Identifier: BSD-3-Clause
// Copyright(c) 2019-2020 Intel Corporation

package main

import (
	"fmt"
	"sort"
	"strings"

	"github.com/rivo/tview"
	cz "pmdt.org/colorize"
	"pmdt.org/devbind"
	tab "pmdt.org/taborder"
	tlog "pmdt.org/ttylog"
)

// Display the dpdk_devbind.py information in windows to be examined.
// The code below parses the data similar to dpdk_devbind.py script without
// using the script.

// TableData for each view
type TableData struct {
	name       string
	classes    []*devbind.DeviceConfig
	align      int
	fixedSize  int
	proportion int
	focus      bool
	key        rune
}

// TableInfo for each Table window
type TableInfo struct {
	changed bool
	length  int
	name    string
	view    *tview.Table
	classes []*devbind.DeviceConfig
	devlist []*devbind.DeviceClass
}

// DevBindPanel - Data for main page information
type DevBindPanel struct {
	tabOrder *tab.Tab
	devbind  *devbind.BindInfo
	topFlex  *tview.Flex
	tables   []TableData

	tInfos map[string]*TableInfo
}

const (
	devbindPanelName string = "DevBind"
)

// SetupDevBind - setup and init the main page
func setupDevBind() *DevBindPanel {

	pg := &DevBindPanel{}

	pg.devbind = devbind.New()

	db := pg.devbind

	pg.tInfos = make(map[string]*TableInfo)

	// Create the set of tables to display each section in a different window
	pg.tables = []TableData{
		{
			name:       "Network",
			classes:    db.Groups[devbind.NetworkGroup],
			align:      tview.AlignLeft,
			fixedSize:  0,
			proportion: 2,
			focus:      true,
			key:        'n',
		}, {
			name:       "Crypto",
			classes:    db.Groups[devbind.CryptoGroup],
			align:      tview.AlignLeft,
			fixedSize:  0,
			proportion: 1,
			focus:      true,
			key:        'c',
		}, {

			name:       "Eventdev",
			classes:    db.Groups[devbind.EventdevGroup],
			align:      tview.AlignLeft,
			fixedSize:  0,
			proportion: 1,
			focus:      true,
			key:        'e',
		}, {
			name:       "Mempool",
			classes:    db.Groups[devbind.MempoolGroup],
			align:      tview.AlignLeft,
			fixedSize:  0,
			proportion: 1,
			focus:      true,
			key:        'm',
		}, {
			name:       "Compression",
			classes:    db.Groups[devbind.CompressGroup],
			align:      tview.AlignLeft,
			fixedSize:  0,
			proportion: 1,
			focus:      true,
			key:        'C',
		}, {
			name:       "DMA",
			classes:    db.Groups[devbind.DMAGroup],
			align:      tview.AlignLeft,
			fixedSize:  0,
			proportion: 1,
			focus:      true,
			key:        'd',
		},
	}

	// Add the table above into the TableInfo slice.
	for _, td := range pg.tables {
		pg.tInfos[td.name] = &TableInfo{classes: td.classes, name: td.name}
	}

	return pg
}

// DevBindPanelSetup setup the main event page
// Standard routine to setup the tview panel and displayed data
func DevBindPanelSetup(nextSlide func()) (pageName string, content tview.Primitive) {

	pg := setupDevBind()

	// Create a taborder object to setup tab order and single key selection
	to := tab.New(devbindPanelName, perfmon.app)
	pg.tabOrder = to

	top := tview.NewFlex().SetDirection(tview.FlexRow)
	flex := tview.NewFlex().SetDirection(tview.FlexRow)

	// Top window with title and version information about the tool
	TitleBox(top)

	ti := pg.tInfos

	// Create each table view for each of the device table entries
	for _, td := range pg.tables {
		s := fmt.Sprintf("%s Devices (%c)", td.name, td.key)

		ti[td.name].view = CreateTableView(flex, s, td.align, td.fixedSize, td.proportion, td.focus)

		// Add the single key and define the tab order.
		to.Add(ti[td.name].view, td.key)
	}

	to.SetInputDone()

	top.AddItem(flex, 0, 1, true)

	pg.topFlex = top

	// Set up the timers with callback to display the data in the windows
	perfmon.timers.Add(devbindPanelName, func(step int, ticks uint64) {
		switch step {
		case 0:
			for _, t := range pg.tInfos {
				pg.collectData(t)
			}
		case 2:
			if pg.topFlex.HasFocus() {
				perfmon.app.QueueUpdateDraw(func() {
					pg.displayDevBindPanel(step)
				})
			}
		}
	})

	return devbindPanelName, top
}

// Display the given devbind data panel for each window
func (pg *DevBindPanel) displayDevBindPanel(step int) {
	for _, ti := range pg.tInfos {
		if ti.changed {
			ti.changed = false
			pg.displayView(ti)
		}
	}
}

// Collect the data to be displayed in the different device windows
func (pg *DevBindPanel) collectData(ti *TableInfo) {

	deviceList := make([]*devbind.DeviceClass, 0)

	tlog.DebugPrintf("Name: %s, Classes: %+v\n", ti.name, ti.classes)

	// Convert the map into a slice to be able to sort it
	for _, l := range pg.devbind.FindDevicesByDeviceClass(ti.name, ti.classes) {
		deviceList = append(deviceList, l)
	}
	tlog.DebugPrintf("Name: %s, deviceList: %+v\n", ti.name, deviceList)

	sort.Slice(deviceList, func(i, j int) bool {
		return deviceList[j].Slot > deviceList[i].Slot
	})

	// Set the device list and set the changed flag to force update of window
	ti.devlist = deviceList
	if ti.length != len(deviceList) {
		ti.changed = true
		ti.length = len(deviceList)
	}
}

// Display the formation into the given table, all windows use this routine
func (pg *DevBindPanel) displayView(ti *TableInfo) {

	view := ti.view

	// Setup the header for columns and rows
	SetCell(view, 0, 0, cz.CornSilk("Slot"), tview.AlignLeft)
	SetCell(view, 0, 1, cz.CornSilk("Vendor ID"), tview.AlignLeft)
	SetCell(view, 0, 2, cz.CornSilk("Vendor Name"), tview.AlignLeft)
	SetCell(view, 0, 3, cz.CornSilk("Device Description"), tview.AlignLeft)
	SetCell(view, 0, 4, cz.CornSilk("Interface"), tview.AlignLeft)
	SetCell(view, 0, 5, cz.CornSilk("Driver"), tview.AlignLeft)
	SetCell(view, 0, 6, cz.CornSilk("Active"), tview.AlignLeft)
	SetCell(view, 0, 7, cz.CornSilk("Numa"), tview.AlignLeft)

	// Add each device information to the table based on the devlist
	row := 1
	for _, d := range ti.devlist {
		col := 0

		SetCell(view, row, col, cz.DeepPink(d.Slot), tview.AlignLeft)
		col++

		s := fmt.Sprintf("[%s:%s]",
			cz.SkyBlue(d.Vendor.ID), cz.SkyBlue(d.Device.ID))
		SetCell(view, row, col, s, tview.AlignLeft)
		col++

		str := d.Vendor.Str
		idx := strings.Index(str, "[")
		if idx != -1 {
			str = str[:idx-1]
		}
		SetCell(view, row, col, cz.SkyBlue(str), tview.AlignLeft)
		col++

		str = d.SDevice.Str
		idx = strings.Index(str, "[")
		if idx != -1 {
			str = str[:idx-1]
		}
		SetCell(view, row, col, cz.LightGreen(str), tview.AlignLeft)
		col++

		str = d.Interface
		SetCell(view, row, col, cz.ColorWithName("Tomato", str), tview.AlignLeft)
		col++

		str = d.Driver
		SetCell(view, row, col, cz.LightYellow(str), tview.AlignLeft)
		col++

		str = ""
		if d.Active {
			str = cz.Orange("*Active*")
		}
		SetCell(view, row, col, str, tview.AlignLeft)
		col++

		str = d.NumaNode
		idx = strings.Index(str, "[")
		if idx != -1 {
			str = str[:idx-1]
		}
		SetCell(view, row, col, cz.MistyRose(str), tview.AlignLeft)
		col++

		row++
	}

	ti.view.ScrollToBeginning()
}
