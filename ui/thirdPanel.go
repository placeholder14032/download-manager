package ui

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/placeholder14032/download-manager/internal/controller"
	"github.com/placeholder14032/download-manager/internal/util"
)

var selectedQueue util.QueueBody
var selectQueueFlex, selectOptionQueueFlex, addQueueFlex, editQueueFlex, deleteQueueFlex *tview.Flex
var nameQueue, directoryQueue, startTimeQueue, endTimeQueue string
var maxSimultaneousQueue, maxTryQueue, maxBandwidthQueue int64

func DrawMainQueuePage(app *tview.Application) {
	tabHeader := returnTabHeader()

	header := tview.NewTextView().
		SetText("[::b]SELECT ACTION[::-]").
		SetDynamicColors(true)

	footer := tview.NewTextView().SetText("Press arrow keys to navigate | Enter to confirm | f[1,2,3] to chnage tabs | Ctrl+q to quit")
	selectOptionQueueFlex = tview.NewFlex()

	var queueOptions = tview.NewList().
		AddItem("> NEW QUEUE", "adding new qeueu", 'a', func() { drawNewQueue(app) }).
		AddItem("> EDIT QUEUE", "edit existing queue", 'b', func() { drawSelectQueue(app) }).
		AddItem("> DELETE QUEUE", "delete existing queue", 'c', func() { drawDeleteQueue(app) })
	selectOptionQueueFlex = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(tabHeader, 1, 0, false).
		AddItem(header, 1, 0, false).
		AddItem(queueOptions, 6, 0, true).
		AddItem(tview.NewTextView().SetBackgroundColor(tcell.ColorBlack), 0, 1, false).
		AddItem(footer, 1, 0, false)

	app.SetRoot(selectOptionQueueFlex, true).SetFocus(queueOptions)
	StatePanel = "third"
}

func drawNewQueue(app *tview.Application) {
	tabHeader := returnTabHeader()

	header := tview.NewTextView().
		SetText("[::b]NEW QUEUE[::-]").
		SetDynamicColors(true)

	footer := tview.NewTextView().SetText("Press arrow keys to navigate | Enter to confirm | f[1,2,3] to chnage tabs | Ctrl+q to quit")
	var currentStep, maxStep int = 0, 7
	nameQueueInput := tview.NewInputField().SetLabel("Name: ").SetFieldBackgroundColor(tcell.ColorBlack)
	directoryQueueInput := tview.NewInputField().SetLabel("Directory: ").SetFieldBackgroundColor(tcell.ColorBlack)
	maxSimultaneousQueueInput := tview.NewInputField().SetLabel("Max Simultaneous: ").SetFieldBackgroundColor(tcell.ColorBlack)
	maxBandwidthQueueInput := tview.NewInputField().SetLabel("Max Bandwidth: ").SetFieldBackgroundColor(tcell.ColorBlack)
	startTimeQueueInput := tview.NewInputField().SetLabel("Start Time: ").SetFieldBackgroundColor(tcell.ColorBlack)
	endTimeQueueInput := tview.NewInputField().SetLabel("End Time: ").SetFieldBackgroundColor(tcell.ColorBlack)
	maxTryQueueInput := tview.NewInputField().SetLabel("Max Retry: ").SetFieldBackgroundColor(tcell.ColorBlack)
	var inputFields []*tview.InputField = []*tview.InputField{nameQueueInput, directoryQueueInput, maxSimultaneousQueueInput, maxBandwidthQueueInput, startTimeQueueInput, endTimeQueueInput, maxTryQueueInput}

	nameQueueInput.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter {
			nameQueue = nameQueueInput.GetText()
			currentStep++
			app.SetFocus(inputFields[currentStep])
		}
	})
	directoryQueueInput.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter {
			directoryQueue = directoryQueueInput.GetText()
			currentStep++
			app.SetFocus(inputFields[currentStep])
		}
	})
	maxSimultaneousQueueInput.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter {
			maxSimultaneousQueue, _ = strconv.ParseInt(maxSimultaneousQueueInput.GetText(), 10, 64)
			currentStep++
			app.SetFocus(inputFields[currentStep])
		}
	}).SetAcceptanceFunc(tview.InputFieldInteger)
	maxBandwidthQueueInput.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter {
			maxBandwidthQueue, _ = strconv.ParseInt(maxBandwidthQueueInput.GetText(), 10, 64)
			currentStep++
			app.SetFocus(inputFields[currentStep])
		}
	}).SetAcceptanceFunc(tview.InputFieldInteger)
	startTimeQueueInput.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter {
			startTimeQueue = startTimeQueueInput.GetText()
			currentStep++
			app.SetFocus(inputFields[currentStep])
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
			app.SetFocus(inputFields[currentStep])
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
			controller.AddQueue(directoryQueue, nameQueue, maxSimultaneousQueue, maxBandwidthQueue, maxTryQueue, true, strToTimeStart(startTimeQueue), strToTimeEnd(endTimeQueue))
			DrawMainQueuePage(app)
		}
	}).SetAcceptanceFunc(tview.InputFieldInteger)

	addQueueFlex = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(tabHeader, 1, 0, false).
		AddItem(header, 1, 0, false).
		AddItem(inputFields[0], 1, 0, true).
		AddItem(inputFields[1], 1, 0, true).
		AddItem(inputFields[2], 1, 0, true).
		AddItem(inputFields[3], 1, 0, true).
		AddItem(inputFields[4], 1, 0, true).
		AddItem(inputFields[5], 1, 0, true).
		AddItem(inputFields[6], 1, 0, true).
		AddItem(tview.NewTextView().SetBackgroundColor(tcell.ColorBlack), 0, 1, false).
		AddItem(footer, 1, 0, false)

	addQueueFlex.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyUp:
			if currentStep > 0 {
				currentStep--
				app.SetFocus(inputFields[currentStep])
				return nil
			}
		case tcell.KeyDown:
			if currentStep < maxStep-1 {
				currentStep++
				app.SetFocus(inputFields[currentStep])
				return nil
			}
		}
		return event
	})

	app.SetRoot(addQueueFlex, true).SetFocus(inputFields[0])
}
func drawSelectQueue(app *tview.Application) {
	errorView := ""
	tabHeader := returnTabHeader()

	header := tview.NewTextView().
		SetText("[::b]SELECT QUEUE[::-]").
		SetDynamicColors(true)

	listQueues := controller.GetQueues()
	if len(listQueues) == 0 {
		errorView = "No Queues Available! ** "
	}
	footer := tview.NewTextView().SetText("Press arrow keys to navigate | Enter to confirm | f[1,2,3] to chnage tabs | Ctrl+q to quit")

	var selectOptions = tview.NewList()
	for i, q := range listQueues {
		selectOptions.AddItem(fmt.Sprintf("> %s", q.Name), "", rune('a'+i), func() {
			selectedQueue = q
			drawEditQueue(app)
		})
	}

	listHeight := len(listQueues)

	selectQueueFlex = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(tabHeader, 1, 0, false).
		AddItem(header, 1, 0, false).
		AddItem(selectOptions, 2*listHeight, 0, true).
		AddItem(tview.NewTextView().SetBackgroundColor(tcell.ColorBlack), 0, 1, false).
		AddItem(tview.NewTextView().SetText(errorView).SetTextColor(tcell.ColorRed), 1, 0, false).
		AddItem(footer, 1, 0, false)
	app.SetRoot(selectQueueFlex, true).SetFocus(selectOptions)
}

func getTimeString(t time.Time) string {
	/*hour := t.Hour()
	minute := t.Minute()
	return (strconv.FormatInt(int64(hour), 10) + ":" + strconv.FormatInt(int64(minute), 10))*/
	return t.Format("15:04")
}

func drawEditQueue(app *tview.Application) {
	errorTextView := tview.NewTextView().SetText("").SetTextColor(tcell.ColorRed)
	errorView := ""
	tabHeader := returnTabHeader()

	header := tview.NewTextView().
		SetText("[::b]EDIT QUEUE[::-]").
		SetDynamicColors(true)

	footer := tview.NewTextView().SetText("Press arrow keys to navigate | Enter to confirm | f[1,2,3] to chnage tabs | Ctrl+q to quit")
	var currentStep, maxStep int = 0, 6
	directoryQueueInput := tview.NewInputField().SetLabel("Directory: ").SetText(selectedQueue.Directory).SetFieldBackgroundColor(tcell.ColorBlack)
	maxSimultaneousQueueInput := tview.NewInputField().SetLabel("Max Simultaneous: ").SetText(strconv.FormatInt(selectedQueue.MaxSimul, 10)).SetFieldBackgroundColor(tcell.ColorBlack)
	maxBandwidthQueueInput := tview.NewInputField().SetLabel("Max Bandwidth: ").SetText(strconv.FormatInt(selectedQueue.MaxBandWidth, 10)).SetFieldBackgroundColor(tcell.ColorBlack)
	startTimeQueueInput := tview.NewInputField().SetLabel("Start Time: ").SetText(getTimeString(selectedQueue.TimeRange.Start)).SetFieldBackgroundColor(tcell.ColorBlack)
	endTimeQueueInput := tview.NewInputField().SetLabel("End Time: ").SetText(getTimeString(selectedQueue.TimeRange.End)).SetFieldBackgroundColor(tcell.ColorBlack)
	maxTryQueueInput := tview.NewInputField().SetLabel("Max Retry: ").SetText(strconv.FormatInt(selectedQueue.MaxRetries, 10)).SetFieldBackgroundColor(tcell.ColorBlack)
	var inputFields []*tview.InputField = []*tview.InputField{directoryQueueInput, maxSimultaneousQueueInput, maxBandwidthQueueInput, startTimeQueueInput, endTimeQueueInput, maxTryQueueInput}

	startTimeQueue = getTimeString(selectedQueue.TimeRange.Start)
	endTimeQueue = getTimeString(selectedQueue.TimeRange.End)

	directoryQueueInput.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter {
			directoryQueue = directoryQueueInput.GetText()
			currentStep++
			app.SetFocus(inputFields[currentStep])
		}
	})
	maxSimultaneousQueueInput.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter {
			selectedQueue.MaxSimul, _ = strconv.ParseInt(maxBandwidthQueueInput.GetText(), 10, 64)
			currentStep++
			app.SetFocus(inputFields[currentStep])
		}
	}).SetAcceptanceFunc(tview.InputFieldInteger)
	// submit the edits !!!!!
	maxBandwidthQueueInput.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter {
			selectedQueue.MaxBandWidth, _ = strconv.ParseInt(maxBandwidthQueueInput.GetText(), 10, 64)
			currentStep++
			app.SetFocus(inputFields[currentStep])
		}
	}).SetAcceptanceFunc(tview.InputFieldInteger)
	startTimeQueueInput.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter {
			startTimeQueue = startTimeQueueInput.GetText()
			currentStep++
			app.SetFocus(inputFields[currentStep])
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
			app.SetFocus(inputFields[currentStep])
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
			selectedQueue.MaxRetries, _ = strconv.ParseInt(maxTryQueueInput.GetText(), 10, 64)
			err := controller.EditQueue(selectedQueue.ID, selectedQueue.Directory, selectedQueue.Name, selectedQueue.MaxSimul, selectedQueue.MaxBandWidth, selectedQueue.MaxRetries, true, strToTimeStart(startTimeQueue), strToTimeEnd(endTimeQueue))
			if err != nil {
				errorView = err.Error()
				errorTextView.SetText(errorView)
			} else {
				DrawMainQueuePage(app)
			}
		}
	}).SetAcceptanceFunc(tview.InputFieldInteger)

	editQueueFlex = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(tabHeader, 1, 0, false).
		AddItem(header, 1, 0, false).
		AddItem(inputFields[0], 1, 0, true).
		AddItem(inputFields[1], 1, 0, true).
		AddItem(inputFields[2], 1, 0, true).
		AddItem(inputFields[3], 1, 0, true).
		AddItem(inputFields[4], 1, 0, true).
		AddItem(inputFields[5], 1, 0, true).
		AddItem(tview.NewTextView().SetBackgroundColor(tcell.ColorBlack), 0, 1, false).
		AddItem(errorTextView, 1, 0, false).
		AddItem(footer, 1, 0, false)

	editQueueFlex.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyUp:
			if currentStep > 0 {
				currentStep--
				app.SetFocus(inputFields[currentStep])
				return nil
			}
		case tcell.KeyDown:
			if currentStep < maxStep-1 {
				currentStep++
				app.SetFocus(inputFields[currentStep])
				return nil
			}
		}
		return event
	})

	app.SetRoot(editQueueFlex, true).SetFocus(inputFields[0])
}

func drawDeleteQueue(app *tview.Application) {
	errorView := ""
	tabHeader := returnTabHeader()

	header := tview.NewTextView().
		SetText("[::b]DELETE QUEUE[::-]").
		SetDynamicColors(true)

	listQueues := controller.GetQueues()
	if len(listQueues) == 0 {
		errorView = "No Queues Available! ** "
	}
	footer := tview.NewTextView().SetText("Press arrow keys to navigate | Enter to confirm | f[1,2,3] to chnage tabs | Ctrl+q to quit")

	var selectOptions = tview.NewList()
	for i, q := range listQueues {
		selectOptions.AddItem(fmt.Sprintf("> %s", q.Name), "", rune('a'+i), func() {
			controller.DeleteQueue(q.ID)
			DrawMainQueuePage(app)
		})
	}

	listHeight := len(listQueues)

	selectQueueFlex = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(tabHeader, 1, 0, false).
		AddItem(header, 1, 0, false).
		AddItem(selectOptions, 2*listHeight, 0, true).
		AddItem(tview.NewTextView().SetBackgroundColor(tcell.ColorBlack), 0, 1, false).
		AddItem(tview.NewTextView().SetText(errorView).SetTextColor(tcell.ColorRed), 1, 0, false).
		AddItem(footer, 1, 0, false)
	app.SetRoot(selectQueueFlex, true).SetFocus(selectOptions)
}

func returnTabHeader() *tview.Flex {
	tabHeader := tview.NewFlex().SetDirection(tview.FlexColumn)
	tab1 := tview.NewTextView().
		SetText("Tab 1").
		SetTextAlign(tview.AlignCenter)

	tab2 := tview.NewTextView().
		SetText("tab 2").
		SetTextAlign(tview.AlignCenter)

	tab3 := tview.NewTextView().
		SetText("Tab 3")
	// had to write these seperate for text to be shown
	tab3.SetTextAlign(tview.AlignCenter).
		SetBackgroundColor(tcell.ColorBlue)

	tabHeader.AddItem(tview.NewTextView().SetBackgroundColor(tcell.ColorBlack), 0, 1, false).
		AddItem(tab1, 10, 0, false).
		AddItem(tab2, 10, 0, false).
		AddItem(tab3, 10, 0, false).
		AddItem(tview.NewTextView().SetBackgroundColor(tcell.ColorBlack), 0, 1, false)
	return tabHeader
}

func strToTimeStart(tmpTime string) string {
	if tmpTime == "" {
		return "00:00:00"
	}
	parts := strings.Split(tmpTime, ":")
	hours, _ := strconv.Atoi(parts[0])

	minutes := 0
	if len(parts) > 1 && parts[1] != "" {
		minutes, _ = strconv.Atoi(parts[1])
	}
	return fmt.Sprintf("%02d:%02d:00", hours, minutes)
}

func strToTimeEnd(tmpTime string) string {
	if tmpTime == "" {
		return "23:59:00"
	}
	parts := strings.Split(tmpTime, ":")
	hours, _ := strconv.Atoi(parts[0])

	minutes := 59
	if len(parts) > 1 && parts[1] != "" {
		minutes, _ = strconv.Atoi(parts[1])
	}
	return fmt.Sprintf("%02d:%02d:00", hours, minutes)
}
