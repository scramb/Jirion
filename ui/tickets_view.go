package ui

import (
	"fmt"
	"runtime"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/scramb/backlog-manager/internal/models"
	"os/exec"
)

// NewTicketsView builds the “My Tickets” tab content.
// It lists all issues assigned to the logged-in user and allows opening them in the browser.
func TicketsView(app fyne.App, w fyne.Window, domain, user, token string) fyne.CanvasObject {
	var issues []models.JiraIssue

	ticketList := widget.NewList(
		func() int { return len(issues) },
		func() fyne.CanvasObject {
			icon := widget.NewIcon(nil)
			id := widget.NewLabel("ID")
			title := widget.NewLabel("Titel")
			return container.NewHBox(icon, id, title)
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			issue := issues[i]
			box := o.(*fyne.Container)

			iconWidget := box.Objects[0].(*widget.Icon)
			idLabel := box.Objects[1].(*widget.Label)
			titleLabel := box.Objects[2].(*widget.Label)

			var icon fyne.Resource
			switch issue.Fields.IssueType.Name {
			case "Story":
				icon = theme.FileIcon()
			case "Bug":
				icon = theme.ErrorIcon()
			case "Task":
				icon = theme.DocumentIcon()
			default:
				icon = theme.InfoIcon()
			}

			iconWidget.SetResource(icon)
			idLabel.SetText(issue.Key)
			titleLabel.SetText(issue.Fields.Summary)
		},
	)

	ticketList.OnSelected = func(id widget.ListItemID) {
		if id < 0 || id >= len(issues) {
			return
		}
		issue := issues[id]
		url := fmt.Sprintf("https://%s.atlassian.net/browse/%s", domain, issue.Key)
		openBrowser(url)
	}

	refreshBtn := widget.NewButtonWithIcon("Neu laden", theme.ViewRefreshIcon(), func() {
		go func() {
			fetched, err := models.FetchAssignedIssues(domain, user, token)
			if err != nil {
				fyne.Do(func() { dialog.ShowError(err, w) })
				return
			}
			fyne.Do(func() {
				issues = fetched
				ticketList.Refresh()
			})
		}()
	})

	ticketsView := container.NewBorder(
		refreshBtn, nil, nil, nil,
		ticketList,
	)

	// Initial auto-refresh when tab is opened (optional)
	refreshBtn.OnTapped()

	return ticketsView
}

// openBrowser opens a URL in the system's default browser.
func openBrowser(url string) {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "darwin":
		cmd = "open"
	case "windows":
		cmd = "rundll32"
		args = []string{"url.dll,FileProtocolHandler"}
	default:
		cmd = "xdg-open"
	}

	args = append(args, url)
	exec.Command(cmd, args...).Start()
}
