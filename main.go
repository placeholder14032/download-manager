package main

import (
	"github.com/placeholder14032/download-manager/cmd"
	//"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func main() {
	app := tview.NewApplication()

	cmd.DrawFirstStage(app)

	if err := app.Run(); err != nil {
		panic(err)
	}
}
