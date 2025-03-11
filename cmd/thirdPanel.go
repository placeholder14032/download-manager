package cmd

import (
	"fmt"
	"regexp"
	"strconv"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

var addQueueFlex, editQueueFlex *tview.Flex
var nameQueue, directoryQueue, startTimeQueue, endTimeQueue string
var maxSimultaneousQueue, maxTryQueue, maxBandwidthQueue int64

func DrawMainQueuePage(app *tview.Application) {
	var mainFlex = tview.NewFlex()

	var queueOptions = tview.NewList().
		AddItem("> NEW QUEUE", "adding new qeueu", 'a', func() { drawNewQueue(app) }).
		AddItem("> EDIT QUEUE", "edit existing queue", 'b', func() { drawEditQueue(app) })

	mainFlex = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(queueOptions, 4, 0, true)

	app.SetRoot(mainFlex, true).SetFocus(queueOptions)
}

func drawNewQueue(app *tview.Application) {
	footer := tview.NewTextView().SetText("Press arrow keys to navigate | Enter to confirm | Esc to go back | f[1,2,3] to chnage tabs | Ctrl+q to quit")
	var currentStep, maxStep int = 0, 7
	nameQueueInput := tview.NewInputField().SetLabel("Name: ")
	directoryQueueInput := tview.NewInputField().SetLabel("Directory: ")
	maxSimultaneousQueueInput := tview.NewInputField().SetLabel("Max Simultaneous: ")
	maxBandwidthQueueInput := tview.NewInputField().SetLabel("Max Bandwidth: ")
	startTimeQueueInput := tview.NewInputField().SetLabel("Start Time: ")
	endTimeQueueInput := tview.NewInputField().SetLabel("End Time: ")
	maxTryQueueInput := tview.NewInputField().SetLabel("Max Retry: ")
	var InputFields []*tview.InputField = []*tview.InputField{nameQueueInput, directoryQueueInput, maxSimultaneousQueueInput, maxBandwidthQueueInput, startTimeQueueInput, endTimeQueueInput, maxTryQueueInput}

	nameQueueInput.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter {
			nameQueue = nameQueueInput.GetText()
			currentStep++
			app.SetFocus(InputFields[currentStep])
		}
	})
	directoryQueueInput.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter {
			directoryQueue = directoryQueueInput.GetText()
			currentStep++
			app.SetFocus(InputFields[currentStep])
		}
	})
	maxSimultaneousQueueInput.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter {
			maxSimultaneousQueue, _ = strconv.ParseInt(maxSimultaneousQueueInput.GetText(), 10, 64)
			currentStep++
			app.SetFocus(InputFields[currentStep])
		}
	}).SetAcceptanceFunc(tview.InputFieldInteger)
	maxBandwidthQueueInput.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter {
			maxBandwidthQueue, _ = strconv.ParseInt(maxBandwidthQueueInput.GetText(), 10, 64)
			currentStep++
			app.SetFocus(InputFields[currentStep])
		}
	}).SetAcceptanceFunc(tview.InputFieldInteger)
	startTimeQueueInput.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter {
			startTimeQueue = startTimeQueueInput.GetText()
			currentStep++
			app.SetFocus(InputFields[currentStep])
		}
	}).SetAcceptanceFunc(func(textToCheck string, lastChar rune) bool {
		if len(textToCheck) == 0 {
			return true
		}
		pattern := `^([0-9]|[0-1][0-9]|2[0-3])(:[0-5]?[0-9]?)?$`
		matched, _ := regexp.MatchString(pattern, textToCheck)
		return matched
	})
	endTimeQueueInput.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter {
			endTimeQueue = endTimeQueueInput.GetText()
			currentStep++
			app.SetFocus(InputFields[currentStep])
		}
	}).SetAcceptanceFunc(func(textToCheck string, lastChar rune) bool {
		if len(textToCheck) == 0 {
			return true
		}
		pattern := `^([0-9]|[0-1][0-9]|2[0-3])(:[0-5]?[0-9]?)?$`
		matched, _ := regexp.MatchString(pattern, textToCheck)
		return matched
	})
	maxTryQueueInput.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter {
			maxTryQueue, _ = strconv.ParseInt(maxTryQueueInput.GetText(), 10, 64)
			defer fmt.Println(nameQueue, directoryQueue, maxSimultaneousQueue, maxBandwidthQueue, startTimeQueue, endTimeQueue, maxTryQueue)
			app.Stop()
		}
	}).SetAcceptanceFunc(tview.InputFieldInteger)

	addQueueFlex = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(InputFields[0], 1, 0, true).
		AddItem(InputFields[1], 1, 0, true).
		AddItem(InputFields[2], 1, 0, true).
		AddItem(InputFields[3], 1, 0, true).
		AddItem(InputFields[4], 1, 0, true).
		AddItem(InputFields[5], 1, 0, true).
		AddItem(InputFields[6], 1, 0, true).
		AddItem(nil, 0, 1, false).
		AddItem(footer, 1, 0, false)

	addQueueFlex.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyUp:
			if currentStep > 0 {
				currentStep--
				app.SetFocus(InputFields[currentStep])
				return nil
			}
		case tcell.KeyDown:
			if currentStep < maxStep-1 {
				currentStep++
				app.SetFocus(InputFields[currentStep])
				return nil
			}
		}
		return event
	})

	app.SetRoot(addQueueFlex, true).SetFocus(InputFields[0])
}
func drawEditQueue(app *tview.Application) {

}
