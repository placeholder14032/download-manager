package cmd

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

var url, name string

func DrawFirstStage(app *tview.Application) {
	var firstPanelFlex = tview.NewFlex()

	var urlInputField = tview.NewInputField()

	urlInputField.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter {
			url = urlInputField.GetText()
			drawSecondStage(app)
			return
		}
	}).SetLabel("URL: ")

	firstPanelFlex = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(urlInputField, 1, 0, true)

	app.SetRoot(firstPanelFlex, true).SetFocus(urlInputField)
}

func drawSecondStage(app *tview.Application) {
	var secondPanelFlex = tview.NewFlex()

	var nameInputField = tview.NewInputField()
	var urlTextField = tview.NewTextView().SetText("> URL: " + url)

	nameInputField.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter {
			name = nameInputField.GetText()
			drawThirdStage(app)
			return
		}
	}).SetLabel("NAME: ")

	secondPanelFlex = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(urlTextField, 1, 0, false).
		AddItem(nameInputField, 1, 0, true)

	app.SetRoot(secondPanelFlex, true).SetFocus(nameInputField)
}

func drawThirdStage(app *tview.Application) {
	var secondPanelFlex = tview.NewFlex()

	var queueDropDown = tview.NewDropDown().
		SetLabel("QUEUE: ").
		SetOptions([]string{"Queue1", "Queue2", "Queue3"}, func(text string, index int) {
			defer fmt.Println(url, name, text)
			app.Stop()
		})
	var urlTextField, nameTextField = tview.NewTextView().SetText("> URL: " + url), tview.NewTextView().SetText("> NAME: " + name)

	secondPanelFlex = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(urlTextField, 1, 0, false).
		AddItem(nameTextField, 1, 0, false).
		AddItem(queueDropDown, 1, 0, true)

	app.SetRoot(secondPanelFlex, true).SetFocus(queueDropDown)
}
