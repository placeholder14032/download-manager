package cmd

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

var StatePanel string
var newDownloadFlex *tview.Flex
var urlDownload, nameDownload, queueDownload string

func DrawNewDownloadPage(app *tview.Application) {
	footer := tview.NewTextView().SetText("Press arrow keys to navigate | Enter to confirm | f[1,2,3] to chnage tabs | Ctrl+q to quit")
	var currentStep, maxStep int = 0, 3
	nameDownloadInput := tview.NewInputField().SetLabel("Name: ").SetFieldBackgroundColor(tcell.ColorBlack)
	urlDownloadInput := tview.NewInputField().SetLabel("Url: ").SetFieldBackgroundColor(tcell.ColorBlack)
	var inputFields []*tview.InputField = []*tview.InputField{nameDownloadInput, urlDownloadInput}
	queueDropDown := tview.NewDropDown().SetLabel("Queue: ").
		SetOptions([]string{"Queue1", "Queue2", "Queue3"}, nil).
		SetCurrentOption(0)
	queueDropDown.SetFieldBackgroundColor(tcell.ColorBlack)
	isQueueDropDownOpen := false
	queueDropDown.SetSelectedFunc(func(text string, index int) {
		app.Stop()
	})
	nameDownloadInput.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter {
			nameDownload = nameDownloadInput.GetText()
			currentStep++
			app.SetFocus(inputFields[currentStep])
		}
	})
	urlDownloadInput.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter {
			urlDownload = urlDownloadInput.GetText()
			currentStep++
			app.SetFocus(queueDropDown)
			queueDropDown.SetFieldBackgroundColor(tcell.ColorRed)
			isQueueDropDownOpen = true
		}
	})

	newDownloadFlex = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(inputFields[0], 1, 0, true).
		AddItem(inputFields[1], 1, 0, true).
		AddItem(queueDropDown, 1, 0, true).
		AddItem(tview.NewTextView().SetBackgroundColor(tcell.ColorBlack), 0, 1, false).
		AddItem(footer, 1, 0, false)

	newDownloadFlex.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyUp:
			if currentStep > 0 {
				currentStep--
				if currentStep < maxStep-1 {
					app.SetFocus(inputFields[currentStep])
					queueDropDown.SetFieldBackgroundColor(tcell.ColorBlack)
					isQueueDropDownOpen = false
				}
				return nil
			}
		case tcell.KeyDown:
			if currentStep < maxStep-1 {
				currentStep++
				if currentStep < maxStep-1 {
					app.SetFocus(inputFields[currentStep])
				} else {
					app.SetFocus(queueDropDown)
					queueDropDown.SetFieldBackgroundColor(tcell.ColorRed)
					isQueueDropDownOpen = true
				}
				return nil
			}
		}
		return event
	})

	queueDropDown.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEnter, tcell.KeyDown:
			if !isQueueDropDownOpen {
				return nil
			}
		}
		return event
	})

	app.SetRoot(newDownloadFlex, true).SetFocus(inputFields[0])
	StatePanel = "first"
}
