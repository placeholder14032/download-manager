package ui

import (
	"strconv"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/placeholder14032/download-manager/internal/controller"
	"github.com/placeholder14032/download-manager/internal/download"
	"github.com/placeholder14032/download-manager/internal/util"
)

var allDownloadFlex *tview.Flex

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
			SetSelectable(false).SetExpansion(1)
		allDownloadTable.SetCell(0, i, tempTableCell)
	}

	allDownloads := controller.GetAllDownloads()

	for i, download := range allDownloads {
		var downloadNameCell, URLCell, queueNameCell *tview.TableCell
		if len(download.FilePath) < 20 {
			downloadNameCell = tview.NewTableCell(download.FilePath).SetSelectable(true).SetExpansion(1)
		} else {
			downloadNameCell = tview.NewTableCell(download.FilePath[:20]).SetSelectable(false).SetExpansion(1)
		}
		allDownloadTable.SetCell(i+1, 0, downloadNameCell)
		if len(download.URL) < 20 {
			URLCell = tview.NewTableCell(download.URL).SetSelectable(false).SetExpansion(1)
		} else {
			URLCell = tview.NewTableCell(download.URL[:20]).SetSelectable(false).SetExpansion(1)
		}
		allDownloadTable.SetCell(i+1, 1, URLCell)
		if len(download.QueueName) < 20 {
			queueNameCell = tview.NewTableCell(download.QueueName).SetSelectable(false).SetExpansion(1)
		} else {
			queueNameCell = tview.NewTableCell(download.QueueName[:20]).SetSelectable(false).SetExpansion(1)
		}
		allDownloadTable.SetCell(i+1, 2, queueNameCell)
		statusCell := tview.NewTableCell(convertStateToString(download.Status)).SetSelectable(false).SetExpansion(1)
		allDownloadTable.SetCell(i+1, 3, statusCell)
		progressCell := tview.NewTableCell(strconv.FormatFloat(download.Progress, 'f', 2, 64)).SetSelectable(false).SetExpansion(1)
		allDownloadTable.SetCell(i+1, 4, progressCell)
		speedCell := tview.NewTableCell(download.Speed).SetSelectable(false).SetExpansion(1)
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
		if len(allDownloads) == 0 {
			return event
		}
		row, _ := allDownloadTable.GetSelection()
		tempDownload := allDownloads[row-1]
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
				if tempDownload.Status == download.Cancelled {
					controller.ModDownload(util.RetryDownload, tempDownload.ID)
				}
				return nil
			}
		case tcell.KeyCtrlS:
			if editMode {
				if tempDownload.Status == download.Paused {
					controller.ModDownload(util.ResumeDownload, selectedQueue.ID)
				} else if tempDownload.Status == download.Downloading {
					controller.ModDownload(util.PauseDownload, tempDownload.ID)
				}
				return nil
			}
		case tcell.KeyCtrlD:
			if editMode {
				controller.ModDownload(util.DeleteDownload, tempDownload.ID)
				DrawAllDownloads(app)
				return nil
			}
		case tcell.KeyCtrlC: // check for status ??
			if editMode {
				controller.ModDownload(util.CancelDownload, tempDownload.ID)
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
		"Pending",
		"Downloading",
		"Paused",
		"Cancelled",
		"Failed",
		"Retrying",
		"Done",
	}

	return states[int(state)]
}
