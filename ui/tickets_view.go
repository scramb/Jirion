package ui

import (
	"fmt"
	"runtime"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/scramb/backlog-manager/internal/models"
	"os/exec"
)

// NewTicketsView builds the “My Tickets” tab content.
// It lists all issues assigned to the logged-in user and allows opening them in the browser.
func TicketsView(app fyne.App, w fyne.Window, domain, user, token string, reloadChan <-chan bool) fyne.CanvasObject {
	var issues []models.JiraIssue

	ticketList := widget.NewList(
		func() int { return len(issues) },
		func() fyne.CanvasObject {
			icon := widget.NewIcon(nil)
			id := widget.NewLabel("ID")
			title := widget.NewLabel("Titel")
			openBtn := widget.NewButtonWithIcon("", theme.ViewFullScreenIcon(), nil)
			box := container.NewHBox(icon, id, title, layout.NewSpacer(), openBtn)
			return box
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			issue := issues[i]
			box := o.(*fyne.Container)

			iconWidget := box.Objects[0].(*widget.Icon)
			idLabel := box.Objects[1].(*widget.Label)
			titleLabel := box.Objects[2].(*widget.Label)
			openBtn := box.Objects[4].(*widget.Button)

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

			openBtn.OnTapped = func() {
				url := fmt.Sprintf("https://%s.atlassian.net/browse/%s", domain, issue.Key)
				openBrowser(url)
			}
		},
	)

	var contentContainer *fyne.Container

	showListView := func() {
		contentContainer.Objects = []fyne.CanvasObject{
			container.NewBorder(
				widget.NewButtonWithIcon("Neu laden", theme.ViewRefreshIcon(), func() {
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
				}),
				nil, nil, nil,
				ticketList,
			),
		}
		contentContainer.Refresh()
	}

	ticketList.OnSelected = func(id widget.ListItemID) {
		if id < 0 || id >= len(issues) {
			return
		}
		issue := issues[id]
		contentContainer.Objects = []fyne.CanvasObject{
			TicketDetailView(app, w, issue, domain, user, token, showListView),
		}
		contentContainer.Refresh()
	}

	contentContainer = container.NewMax()
	showListView()

	// Initial auto-refresh when tab is opened (optional)
	content := contentContainer.Objects[0].(*fyne.Container)
	if btn, ok := content.Objects[0].(*widget.Button); ok {
		btn.OnTapped()
	}

	go func() {
		for range reloadChan {
			fyne.Do(func() {
				fetched, err := models.FetchAssignedIssues(domain, user, token)
				if err != nil {
					dialog.ShowError(err, w)
					return
				}
				issues = fetched
				ticketList.Refresh()
			})
		}
	}()

	return contentContainer
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

// TicketDetailView shows detailed information about a Jira issue with a back button.
func TicketDetailView(app fyne.App, w fyne.Window, issue models.JiraIssue, domain, user, token string, back func()) fyne.CanvasObject {
	keyLabel := widget.NewLabelWithStyle(issue.Key, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	summaryLabel := widget.NewLabel(issue.Fields.Summary)

	description := models.ExtractDescriptionText(issue.Fields.Description)
	descriptionLabel := widget.NewLabel(description)
	descriptionLabel.Wrapping = fyne.TextWrapWord

	backBtn := widget.NewButtonWithIcon("Zurück", theme.NavigateBackIcon(), func() {
		back()
	})

	content := container.NewVBox(
		backBtn,
		keyLabel,
		summaryLabel,
		widget.NewSeparator(),
		descriptionLabel,
	)

	scroll := container.NewVScroll(content)
	scroll.SetMinSize(fyne.NewSize(400, 300))

	return scroll
}
