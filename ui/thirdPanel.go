// toDo: chnage the startTime after editing
// toDo: chnage the endTime after editing
// toDo: get queues in drawing selectQueueFlex

package cmd

import (
	"fmt"
	"regexp"
	"strconv"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/placeholder14032/download-manager/internal/queue"
)

var selectedQueue queue.Queue
var selectQueueFlex, selectOptionQueueFlex, addQueueFlex, editQueueFlex *tview.Flex
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
		AddItem("> EDIT QUEUE", "edit existing queue", 'b', func() { drawSelectQueue(app) })
	selectOptionQueueFlex = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(tabHeader, 1, 0, false).
		AddItem(header, 1, 0, false).
		AddItem(queueOptions, 4, 0, true).
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
			defer fmt.Println(nameQueue, directoryQueue, maxSimultaneousQueue, maxBandwidthQueue, startTimeQueue, endTimeQueue, maxTryQueue)
			app.Stop()
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
	tabHeader := returnTabHeader()

	header := tview.NewTextView().
		SetText("[::b]SELECT QUEUE[::-]").
		SetDynamicColors(true)

	// for test
	var listQueues []queue.Queue = []queue.Queue{queue.Queue{Name: "a", SaveDir: "A", MaxConcurrent: 1, MaxBandwidth: 2, TimeRange: queue.TimeRange{time.Now(), time.Now()}, MaxRetries: 3}}
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
		AddItem(footer, 1, 0, false)
	app.SetRoot(selectQueueFlex, true).SetFocus(selectOptions)
}

func getTimeString(t time.Time) string {
	hour := t.Hour()
	minute := t.Minute()
	return (strconv.FormatInt(int64(hour), 10) + ":" + strconv.FormatInt(int64(minute), 10))
}

func drawEditQueue(app *tview.Application) {
	tabHeader := returnTabHeader()

	header := tview.NewTextView().
		SetText("[::b]EDIT QUEUE[::-]").
		SetDynamicColors(true)

	footer := tview.NewTextView().SetText("Press arrow keys to navigate | Enter to confirm | f[1,2,3] to chnage tabs | Ctrl+q to quit")
	var currentStep, maxStep int = 0, 6
	directoryQueueInput := tview.NewInputField().SetLabel("Directory: ").SetText(selectedQueue.SaveDir).SetFieldBackgroundColor(tcell.ColorBlack)
	maxSimultaneousQueueInput := tview.NewInputField().SetLabel("Max Simultaneous: ").SetText(strconv.FormatInt(selectedQueue.MaxConcurrent, 10)).SetFieldBackgroundColor(tcell.ColorBlack)
	maxBandwidthQueueInput := tview.NewInputField().SetLabel("Max Bandwidth: ").SetText(strconv.FormatInt(selectedQueue.MaxBandwidth, 10)).SetFieldBackgroundColor(tcell.ColorBlack)
	startTimeQueueInput := tview.NewInputField().SetLabel("Start Time: ").SetText(getTimeString(selectedQueue.TimeRange.Start)).SetFieldBackgroundColor(tcell.ColorBlack)
	endTimeQueueInput := tview.NewInputField().SetLabel("End Time: ").SetText(getTimeString(selectedQueue.TimeRange.End)).SetFieldBackgroundColor(tcell.ColorBlack)
	maxTryQueueInput := tview.NewInputField().SetLabel("Max Retry: ").SetText(strconv.FormatInt(selectedQueue.MaxRetries, 10)).SetFieldBackgroundColor(tcell.ColorBlack)
	var inputFields []*tview.InputField = []*tview.InputField{directoryQueueInput, maxSimultaneousQueueInput, maxBandwidthQueueInput, startTimeQueueInput, endTimeQueueInput, maxTryQueueInput}

	directoryQueueInput.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter {
			directoryQueue = directoryQueueInput.GetText()
			currentStep++
			app.SetFocus(inputFields[currentStep])
		}
	})
	maxSimultaneousQueueInput.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter {
			selectedQueue.MaxConcurrent, _ = strconv.ParseInt(maxBandwidthQueueInput.GetText(), 10, 64)
			currentStep++
			app.SetFocus(inputFields[currentStep])
		}
	}).SetAcceptanceFunc(tview.InputFieldInteger)
	maxBandwidthQueueInput.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter {
			selectedQueue.MaxBandwidth, _ = strconv.ParseInt(maxBandwidthQueueInput.GetText(), 10, 64)
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
			DrawMainQueuePage(app)
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
