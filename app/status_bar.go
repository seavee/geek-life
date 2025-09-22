package main

import (
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// StatusBar displays hints and messages at the bottom of app
type StatusBar struct {
	*tview.Pages
	message   *tview.TextView
	container *tview.Application
}

// Name of page keys
const (
	defaultPage = "default"
	messagePage = "message"
)

// Used to skip queued restore of statusBar
// in case of new showForSeconds within waiting period
var restorInQ = 0

func prepareStatusBar(app *tview.Application) *StatusBar {
	messageView := tview.NewTextView().SetDynamicColors(true).SetText("Loading...")
	messageView.SetBackgroundColor(tcell.NewHexColor(0x0c0c0c))
	
	pages := tview.NewPages()
	pages.SetBackgroundColor(tcell.NewHexColor(0x0c0c0c))
	
	statusBar = &StatusBar{
		Pages:     pages,
		message:   messageView,
		container: app,
	}

	statusBar.AddPage(messagePage, statusBar.message, true, true)
	grid := tview.NewGrid().
		SetColumns(0, 0, 0, 0).
		SetRows(0).
		AddItem(tview.NewTextView().SetText("Navigate List: ↓,↑ / j,k").SetBackgroundColor(tcell.NewHexColor(0x0c0c0c)), 0, 0, 1, 1, 0, 0, false).
		AddItem(tview.NewTextView().SetText("New Task/Project: n").SetTextAlign(tview.AlignCenter).SetBackgroundColor(tcell.NewHexColor(0x0c0c0c)), 0, 1, 1, 1, 0, 0, false).
		AddItem(tview.NewTextView().SetText("Step back: Esc").SetTextAlign(tview.AlignCenter).SetBackgroundColor(tcell.NewHexColor(0x0c0c0c)), 0, 2, 1, 1, 0, 0, false).
		AddItem(tview.NewTextView().SetText("Quit: Ctrl+C").SetTextAlign(tview.AlignRight).SetBackgroundColor(tcell.NewHexColor(0x0c0c0c)), 0, 3, 1, 1, 0, 0, false)
	
	grid.SetBackgroundColor(tcell.NewHexColor(0x0c0c0c))
	statusBar.AddPage(defaultPage, grid, true, true)

	return statusBar
}

func (bar *StatusBar) restore() {
	bar.container.QueueUpdateDraw(func() {
		bar.SwitchToPage(defaultPage)
	})
}

func (bar *StatusBar) showForSeconds(message string, timeout int) {
	if bar.container == nil {
		return
	}

	bar.message.SetText(message)
	bar.SwitchToPage(messagePage)
	restorInQ++

	go func() {
		time.Sleep(time.Second * time.Duration(timeout))

		// Apply restore only if this is the last pending restore
		if restorInQ == 1 {
			bar.restore()
		}
		restorInQ--
	}()
}
