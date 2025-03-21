package ui

import (
	"strconv"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/placeholder14032/download-manager/internal/download"
)

var allDownloadFlex *tview.Flex

// temporary struct just so i can draw the table :)))

type DownloadBody struct {
	ID       int64
	URL      string
	FliePath string
	Status   download.State
	Progress int64
	Speed    int64
	// added now (should be added to the original DOwnalodBody)
	DownloadName string
	QueueName    string
}

func DrawAllDownloads(app *tview.Application) {
	// while true (call the function for the array of all Download bodies -> wait for the answer -> get the answer
	// draws the table -> wait something around 500ms and repeat the move)

	editMode := false

	tabHeader := tview.NewFlex().SetDirection(tview.FlexColumn)
	tab1 := tview.NewTextView().
		SetText("Tab 1").
		SetTextAlign(tview.AlignCenter)

	tab2 := tview.NewTextView().
		SetText("tab 2")
	// had to write these seperate for text to be shown
	tab2.SetTextAlign(tview.AlignCenter).
		SetBackgroundColor(tcell.ColorBlue)

	tab3 := tview.NewTextView().
		SetText("Tab 3").
		SetTextAlign(tview.AlignCenter)

	tabHeader.AddItem(tview.NewTextView().SetBackgroundColor(tcell.ColorBlack), 0, 1, false).
		AddItem(tab1, 10, 0, false).
		AddItem(tab2, 10, 0, false).
		AddItem(tab3, 10, 0, false).
		AddItem(tview.NewTextView().SetBackgroundColor(tcell.ColorBlack), 0, 1, false)

	header := tview.NewTextView().
		SetText("[::b]ALL DOWNLOADS[::-]").
		SetDynamicColors(true)

	footer := tview.NewTextView().SetText("Press arrow keys to navigate | Enter to confirm | f[1,2,3] to chnage tabs | Ctrl+q to quit")
	headers := []string{"Name", "URL", "Queue", "Status", "Progress", "Speed"}
	allDownloadFlex = tview.NewFlex()

	allDownloadTable := tview.NewTable()

	for i, header := range headers {
		tempTableCell := tview.NewTableCell(header).
			SetSelectable(false)
		allDownloadTable.SetCell(0, i, tempTableCell)
	}

	allDownloads := []DownloadBody{
		{URL: "https://example.com/file1.zip", Status: download.Paused, Progress: 0, Speed: 0, DownloadName: "name1", QueueName: "moz"},
		{URL: "https://example.com/file2.iso", Status: download.Done, Progress: 0, Speed: 0, DownloadName: "name2", QueueName: "khiar"},
		//{URL: "https://example.com/file3.pdf", Status: download.Starting, Progress: 0, Speed: 0, DownloadName: "name3", QueueName: "porteghal"},
		{URL: "https://example.com/file4.mp4", Status: download.Retrying, Progress: 0, Speed: 0, DownloadName: "name4", QueueName: "sib"},
	}

	for i, download := range allDownloads {
		downloadNameCell := tview.NewTableCell(download.DownloadName).SetSelectable(true)
		allDownloadTable.SetCell(i+1, 0, downloadNameCell)
		URLCell := tview.NewTableCell(download.URL).SetSelectable(false)
		allDownloadTable.SetCell(i+1, 1, URLCell)
		queueNameCell := tview.NewTableCell(download.QueueName).SetSelectable(false)
		allDownloadTable.SetCell(i+1, 2, queueNameCell)
		statusCell := tview.NewTableCell(convertStateToString(download.Status)).SetSelectable(false)
		allDownloadTable.SetCell(i+1, 3, statusCell)
		progressCell := tview.NewTableCell(strconv.FormatInt(download.Progress, 10)).SetSelectable(false)
		allDownloadTable.SetCell(i+1, 4, progressCell)
		speedCell := tview.NewTableCell(strconv.FormatInt(download.Speed, 10)).SetSelectable(false)
		allDownloadTable.SetCell(i+1, 5, speedCell)
	}

	allDownloadFlex = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(tabHeader, 1, 0, false).
		AddItem(header, 1, 0, false).
		AddItem(allDownloadTable, len(allDownloads)+3, 0, true).
		AddItem(tview.NewTextView().SetBackgroundColor(tcell.ColorBlack), 0, 1, false).
		AddItem(footer, 1, 0, false)

	allDownloadTable.SetBorder(true)
	//allDownloadTable.SetBorders(true) -> commented for better looks :)
	allDownloadTable.SetSelectable(true, false)
	allDownloadTable.SetSelectedStyle(tcell.StyleDefault.Background(tcell.ColorBlue))

	allDownloadTable.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		// row, _ := allDownloadTable.GetSelection() -> commented because not used
		// download = allDownloads[row-1] -> commented because not used
		switch event.Key() {
		case tcell.KeyCtrlE:
			editMode = !editMode
			if editMode {
				footer.SetText("Ctrl+S to Start/Stop | Ctrl+R to retry | Ctrl+C to cancel | Ctrl+D to delete")
			} else {
				footer.SetText("Press arrow keys to navigate | Ctrl+E to Edit | f[1,2,3] to chnage tabs | Ctrl+q to quit")
			}
			return nil
		case tcell.KeyCtrlR:
			if editMode {
				//retry
				app.Stop()
				return nil
			}
		case tcell.KeyCtrlS:
			if editMode {
				// start/stop
				return nil
			}
		case tcell.KeyCtrlD:
			if editMode {
				// delete
				return nil
			}
		case tcell.KeyCtrlC: //?
			if editMode {
				// cancle
				return nil
			}
		}
		return event
	})

	app.SetRoot(allDownloadFlex, true).SetFocus(allDownloadTable)
	StatePanel = "second"
}

func convertStateToString(state download.State) string {
	states := []string{
		"Pendoing",
		"Downloading",
		"Paused",
		"Cancelled",
		"Failed",
		"Retrying",
		"Done",
	}

	return states[int(state)]
}
