package ui

import (
	"strconv"

	"github.com/gdamore/tcell/v2"
	"github.com/placeholder14032/download-manager/internal/controller"
	"github.com/rivo/tview"
)

var StatePanel string
var newDownloadFlex *tview.Flex
var urlDownload, nameDownload, queueDownload string

func DrawNewDownloadPage(app *tview.Application) {
	errorView := ""
	tabHeader := tview.NewFlex().SetDirection(tview.FlexColumn)
	tab1 := tview.NewTextView().
		SetText("Tab 1")
	// had to write these seperate for text to be shown
	tab1.SetTextAlign(tview.AlignCenter).
		SetBackgroundColor(tcell.ColorBlue)

	tab2 := tview.NewTextView().
		SetText("tab 2").
		SetTextAlign(tview.AlignCenter)

	tab3 := tview.NewTextView().
		SetText("Tab 3").
		SetTextAlign(tview.AlignCenter)

	tabHeader.AddItem(tview.NewTextView().SetBackgroundColor(tcell.ColorBlack), 0, 1, false).
		AddItem(tab1, 10, 0, false).
		AddItem(tab2, 10, 0, false).
		AddItem(tab3, 10, 0, false).
		AddItem(tview.NewTextView().SetBackgroundColor(tcell.ColorBlack), 0, 1, false)

	header := tview.NewTextView().
		SetText("[::b]NEW DOWNLOAD[::-]").
		SetDynamicColors(true)

	footer := tview.NewTextView().SetText("Press arrow keys to navigate | Enter to confirm | f[1,2,3] to chnage tabs | Ctrl+q to quit")
	var currentStep, maxStep int = 0, 3
	nameDownloadInput := tview.NewInputField().SetLabel("Name: ").SetFieldBackgroundColor(tcell.ColorBlack)
	urlDownloadInput := tview.NewInputField().SetLabel("Url: ").SetFieldBackgroundColor(tcell.ColorBlack)
	var inputFields []*tview.InputField = []*tview.InputField{nameDownloadInput, urlDownloadInput}

	allQueues := controller.GetQueues()
	if len(allQueues) == 0 {
		errorView += "No Queues Available! ** "
	}
	var allQueueNames []string
	for _, q := range allQueues {
		allQueueNames = append(allQueueNames, strconv.FormatInt(q.ID, 32))
	}
	queueDropDown := tview.NewDropDown().SetLabel("Queue: ").
		SetOptions(allQueueNames, nil).
		SetCurrentOption(0)
	queueDropDown.SetFieldBackgroundColor(tcell.ColorBlack)
	isQueueDropDownOpen := false
	queueDropDown.SetSelectedFunc(func(text string, index int) {
		controller.AddDownload(urlDownload, allQueues[index].ID, nameDownload)
		drawNewQueue(app)
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
		AddItem(tabHeader, 1, 0, false).
		AddItem(header, 1, 0, false).
		AddItem(inputFields[0], 1, 0, true).
		AddItem(inputFields[1], 1, 0, true).
		AddItem(queueDropDown, 1, 0, true).
		AddItem(tview.NewTextView().SetBackgroundColor(tcell.ColorBlack), 0, 1, false).
		AddItem(tview.NewTextView().SetText(errorView).SetTextColor(tcell.ColorRed), 1, 0, false).
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
