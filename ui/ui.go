package ui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func Main() {
	app := tview.NewApplication()

	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		// Handle global keys
		switch event.Key() {
		case tcell.KeyF1:
			if StatePanel != "first" {
				DrawNewDownloadPage(app)
			}
			return nil
		case tcell.KeyF2:
			if StatePanel != "second" {
				DrawAllDownloads(app)
			}
			return nil
		case tcell.KeyF3:
			if StatePanel != "third" {
				DrawMainQueuePage(app)
			}
			return nil
		case tcell.KeyEscape:
			// Handle Escape - go back
			return nil
		case tcell.KeyCtrlQ:
			app.Stop()
			return nil
		}
		return event
	})

	DrawAllDownloads(app)

	if err := app.EnableMouse(true).Run(); err != nil {
		panic(err)
	}
}
