// SPDX-License-Identifier: BSD-3-Clause
// Copyright(c) 2019-2020 Intel Corporation

package main

import (
	"fmt"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	cz "pmdt.org/colorize"
	tlog "pmdt.org/ttylog"
)

const (
	defaultColumn int = 1
)

// SelectWindow to hold the table and Application information
type SelectWindow struct {
	name   string        // Name of the selection window
	table  *tview.Table  // Table for the selection window
	values []interface{} // slice of values for the sCol column selection
	offset int           // Number of rows to offset to show selection values
	item   int           // used to select the values slice
	sRow   int           // Current row pointer value
	sCol   int           // The column to use for selection values
}

// NewSelectWindow structure
func NewSelectWindow(table *tview.Table, name string, offset int, f func(row, col int)) *SelectWindow {

	tlog.DebugPrintf("NewSelectWindow:%s: Entry\n", name)
	defer tlog.DebugPrintf("NewSelectWindow:%s: Exit\n", name)

	w := &SelectWindow{
		name:   name,
		table:  table,
		values: make([]interface{}, 0),
		offset: offset,
		item:   0,
		sRow:   -1,
		sCol:   defaultColumn,
	}

	table.SetSelectable(true, false)
	table.SetSelectedStyle(tcell.StyleDefault)

	tlog.DebugPrintf("NewSelectWindow: %s: sCol %d, offset %d\n", w.name, w.sCol, w.offset)
	table.Select(w.offset, 0)
	w.UpdatePointer()
	table.SetSelectionChangedFunc(f)

	return w
}

// AddColumn of data if the selection column then save the slice of data
func (w *SelectWindow) AddColumn(col int, values []interface{}, color ...string) {

	if values == nil || w == nil || w.table == nil {
		return
	}

	// when col is negative then we restore the value back to the selected column
	if col < 0 {
		col = w.sCol
	}

	// if the column to be setup matches the selection column update the slice
	// of values for the column
	if w.sCol == col {
		w.values = values
	}

	// Clear the cells to clean up the display
	for row := len(values) + w.offset; row < w.table.GetRowCount(); row++ {
		w.table.RemoveRow(row)
	}

	row := (w.offset - 1)

	// Output all of the values for a given column
	for _, name := range values {
		s := fmt.Sprintf("%v", name)

		// Set the color is the color argument is given
		if len(color) > 0 {
			s = cz.ColorWithName(color[0], name)
		}
		row++
		tableCell := tview.NewTableCell(s).
			SetAlign(tview.AlignLeft).
			SetSelectable(false)
		w.table.SetCell(row, col, tableCell)
	}

	if w.sRow > row {
		tlog.DebugPrintf("AddColumn: sRow %d, row %d\n", w.sRow, row)
		w.item = row - w.offset
		w.UpdatePointer()
		w.table.Select(w.sRow, w.sCol)
	}

	// Scroll the list to the beginning of the list
	w.table.ScrollToBeginning()
}

// SetColumn for the value returned on selection of a row.
func (w *SelectWindow) SetColumn(col int) {

	if w == nil {
		return
	}
	w.sCol = col
}

// UpdatePointer by replacing the current select item with '->' string
func (w *SelectWindow) UpdatePointer() {

	tlog.DebugPrintf("UpdatePointer: %s: old sRow %d\n", w.name, w.sRow)

	if w.sRow >= 0 {
		SetCell(w.table, w.sRow, 0, "  ", tview.AlignLeft, true)
	}

	w.sRow = w.item + w.offset

	if w.sRow >= 0 {
		SetCell(w.table, w.sRow, 0, "->", tview.AlignLeft, true)
	}

	tlog.DebugPrintf("UpdatePointer: %s: new sRow %d\n", w.name, w.sRow)
}

// Offset value is returned
func (w *SelectWindow) Offset() int {

	return w.offset
}

// ItemIndex value is returned
func (w *SelectWindow) ItemIndex() int {

	if w.item >= len(w.values) {
		w.item = len(w.values) - 1
	}
	return w.item
}

// ItemValue returns the current selected item value
func (w *SelectWindow) ItemValue() interface{} {

	if w == nil || w.item < 0 || w.values == nil || len(w.values) == 0 {
		return nil
	}
	if w.item >= len(w.values) {
		return nil
	}

	return w.values[w.item]
}

// UpdateItem for the apps pointer
func (w *SelectWindow) UpdateItem(row, col int) {

	tlog.DebugPrintf("UpdateItem: Entry\n")
	defer tlog.DebugPrintf("UpdateItem: Exit\n")

	if w == nil || w.table == nil {
		tlog.DebugPrintf("UpdateItem: w is nil\n")
		return
	}

	tlog.DebugPrintf("UpdateItem: %s: row %d, col %d\n", w.name, row, col)

	tlog.DebugPrintf("UpdateItem: %s: sRow %d, item %d, len(w.values) %d\n", w.name, w.sRow, w.item, len(w.values))

	if w.sRow != row {
		row -= w.offset
		tlog.DebugPrintf("UpdateItem: %s: old sRow %d\n", w.name, w.sRow)
		if row >= 0 && row < len(w.values) {
			tlog.DebugPrintf("UpdateItem: new row %d\n", row)
			w.item = row
			w.UpdatePointer()
			tlog.DebugPrintf("UpdateItem: %s: sRow %d, item %d\n", w.name, w.sRow, w.item)
		}
	}
}
