package cmd

import (
	"strconv"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

var mainFlex tview.Flex
var addQueueFlex, editQueueFlex *tview.Flex
var nameQueue, directoryQueue string
var maxSimultaneous, maxRetry, maxBandwidthQueue int64

// var timeRangeQueue queue.TimeRange
var startTime, endTime string

func DrawMainQueuePage(app *tview.Application) {
	var mainFlex = tview.NewFlex()

	var queueOptions = tview.NewList().
		AddItem("> NEW QUEUE", "adding new qeueu", 'a', func() { drawNewFirstStep(app) }).
		AddItem("> EDIT QUEUE", "edit existing queue", 'b', func() { drawEditQueue(app) })

	mainFlex = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(queueOptions, 4, 0, true)

	app.SetRoot(mainFlex, true).SetFocus(queueOptions)
}

func drawNewFirstStep(app *tview.Application) {
	var nameInputField = tview.NewInputField()

	nameInputField.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter {
			nameQueue = nameInputField.GetText()
			drawNewSecondStep(app)
			return
		}
	}).SetLabel("NAME: ")

	addQueueFlex = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(nameInputField, 1, 0, true).
		AddItem(tview.NewTextView().SetText("> Directory: \n> Max Simultaneous: \n> Max Bandwidth: \n> Time Range: \n> Max Retry: "), 5, 0, false)

	app.SetRoot(addQueueFlex, true).SetFocus(nameInputField)

}

func drawNewSecondStep(app *tview.Application) {
	var directoryInputField = tview.NewInputField()

	directoryInputField.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter {
			directoryQueue = directoryInputField.GetText()
			drawNewThirdStep(app)
			return
		}
	}).SetLabel("DIRECTORY: ")

	addQueueFlex = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(tview.NewTextView().SetText("> Name: "+nameQueue), 1, 0, false).
		AddItem(directoryInputField, 1, 0, true).
		AddItem(tview.NewTextView().SetText("> Max Simultaneous: \n> Max Bandwidth: \n> Time Range: \n> Max Retry: "), 4, 0, false)
	app.SetRoot(addQueueFlex, true).SetFocus(directoryInputField)

}
func drawNewThirdStep(app *tview.Application) {
	var maxSimultaneousInputField = tview.NewInputField().
		SetAcceptanceFunc(tview.InputFieldInteger)

	maxSimultaneousInputField.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter {
			maxSimultaneous, _ = strconv.ParseInt(maxSimultaneousInputField.GetText(), 10, 64)
			drawNewFourthStep(app)
			return
		}
	}).SetLabel("MAX SIMULTANEOUS: ")

	addQueueFlex = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(tview.NewTextView().SetText("> Name: "+nameQueue+"\n> Directory: "+directoryQueue), 2, 0, false).
		AddItem(maxSimultaneousInputField, 1, 0, true).
		AddItem(tview.NewTextView().SetText("> Max Bandwidth: \n> Time Range: \n> Max Retry: "), 3, 0, false)
	app.SetRoot(addQueueFlex, true).SetFocus(maxSimultaneousInputField)

}
func drawNewFourthStep(app *tview.Application) {
	var maxBandwidthInputField = tview.NewInputField().
		SetAcceptanceFunc(tview.InputFieldInteger)

	maxBandwidthInputField.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter {
			maxBandwidthQueue, _ = strconv.ParseInt(maxBandwidthInputField.GetText(), 10, 64)
			drawNewFifthStep(app)
			return
		}
	}).SetLabel("MAX BANDWIDTH: ")

	addQueueFlex = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(tview.NewTextView().SetText("> Name: "+nameQueue+"\n> Directory: "+directoryQueue+"\n> Max Simultaneous: "+strconv.FormatInt(maxSimultaneous, 10)), 3, 0, false).
		AddItem(maxBandwidthInputField, 1, 0, true).
		AddItem(tview.NewTextView().SetText("> Time Range: \n> Max Retry: "), 2, 0, false)
	app.SetRoot(addQueueFlex, true).SetFocus(maxBandwidthInputField)

}
func drawNewFifthStep(app *tview.Application) {
	var startTimeInputField = tview.NewInputField()

	var endTimeInputField = tview.NewInputField()

	startTimeInputField.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter {
			startTime = startTimeInputField.GetText()
			app.SetFocus(endTimeInputField)
			return
		}
	}).SetLabel("START: ").SetPlaceholder("hh:mm")

	endTimeInputField.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter {
			endTime = endTimeInputField.GetText()
			drawNewSixthStage(app)
			return
		}
	}).SetLabel("END: ").SetPlaceholder("hh:mm")

	addQueueFlex = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(tview.NewTextView().SetText("> Name: "+nameQueue+"\n> Directory: "+directoryQueue+"\n> Max Simultaneous: "+strconv.FormatInt(maxSimultaneous, 10)+
			"\n> Max Bandwidth: "+strconv.FormatInt(maxBandwidthQueue, 10)), 4, 0, false).
		AddItem(startTimeInputField, 1, 0, true).
		AddItem(endTimeInputField, 1, 0, true).
		AddItem(tview.NewTextView().SetText("> Max Retry: "), 1, 0, false)
	app.SetRoot(addQueueFlex, true).SetFocus(startTimeInputField)

}
func drawNewSixthStage(app *tview.Application) {
	var maxTryInputField = tview.NewInputField().
		SetAcceptanceFunc(tview.InputFieldInteger)

	maxTryInputField.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter {
			maxRetry, _ = strconv.ParseInt(maxTryInputField.GetText(), 10, 64)
			app.Stop()
			return
		}
	}).SetLabel("MAX TRY: ")

	addQueueFlex = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(tview.NewTextView().SetText("> Name: "+nameQueue+"\n> Directory: "+directoryQueue+"\n> Max Simultaneous: "+strconv.FormatInt(maxSimultaneous, 10)+
			"\n> Max Bandwidth: "+strconv.FormatInt(maxBandwidthQueue, 10)+"\n> Start: "+startTime+"\n> End: "+endTime), 6, 0, false).
		AddItem(maxTryInputField, 1, 0, true)
	app.SetRoot(addQueueFlex, true).SetFocus(maxTryInputField)
}
func drawEditQueue(app *tview.Application) {

}
