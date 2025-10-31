package servicedesk

import (
	"fmt"
	"os/exec"
	"runtime"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/scramb/backlog-manager/internal/i18n"
	"github.com/scramb/backlog-manager/internal/models"
)

func openBrowser(url string) error {
	switch runtime.GOOS {
	case "linux":
		return exec.Command("xdg-open", url).Start()
	case "windows":
		return exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		return exec.Command("open", url).Start()
	default:
		return fmt.Errorf("unsupported platform")
	}
}

func ListView(app fyne.App, w fyne.Window) fyne.CanvasObject {
	pageTitle := widget.NewLabelWithStyle(i18n.T("servicedesk.my_requests"), fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	prefs := app.Preferences()
	domain := prefs.String("jira_domain")
	user := prefs.String("jira_user")
	token := prefs.String("jira_token")

	requests, err := models.FetchMyServiceRequests(domain, user, token)
	if err != nil {
		dialog.ShowError(err, w)
	}

	var listData []models.JiraServiceRequest = requests

	list := widget.NewList(
		func() int {
			return len(listData)
		},
		func() fyne.CanvasObject {
			title := widget.NewLabel("")
			title.Wrapping = fyne.TextWrapWord
			status := widget.NewLabel("")
			status.TextStyle = fyne.TextStyle{Italic: true}
			status.Wrapping = fyne.TextWrapWord
			linkBtn := widget.NewButtonWithIcon("", theme.NavigateNextIcon(), nil)
			linkBtn.Importance = widget.LowImportance

			box := container.NewBorder(nil, nil, nil, linkBtn, container.NewVBox(title, status))
			return container.NewPadded(box)
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			item := listData[i]

			padded := o.(*fyne.Container)
			border := padded.Objects[0].(*fyne.Container)
			vbox := border.Objects[0].(*fyne.Container)
			linkBtn := border.Objects[1].(*widget.Button)
			title := vbox.Objects[0].(*widget.Label)
			status := vbox.Objects[1].(*widget.Label)

			title.SetText(item.Summary)
			status.SetText(fmt.Sprintf("[%s] %s", item.IssueKey, item.Status))

			linkBtn.OnTapped = func() {
				url := fmt.Sprintf("https://%s.atlassian.net/browse/%s", domain, item.IssueKey)
				if err := openBrowser(url); err != nil {
					dialog.ShowError(err, w)
				}
			}
		},
	)
	list.OnSelected = nil
	list.OnUnselected = nil

	searchEntry := widget.NewEntry()
	searchEntry.SetPlaceHolder(i18n.T("servicedesk.search_placeholder"))
	searchEntry.OnChanged = func(val string) {
		var filtered []models.JiraServiceRequest
		for _, r := range requests {
			if val == "" || containsInsensitive(r.Summary, val) || containsInsensitive(r.IssueKey, val) || containsInsensitive(r.Status, val) {
				filtered = append(filtered, r)
			}
		}
		listData = filtered
		list.Refresh()
	}

	content := container.NewBorder(
		container.NewVBox(
			pageTitle,
			widget.NewSeparator(),
			searchEntry,
		),
		nil,
		nil,
		nil,
		list,
	)

	return content
}

func containsInsensitive(s, substr string) bool {
	return len(substr) == 0 || (len(s) >= len(substr) && fyne.CurrentApp() != nil && stringContainsFold(s, substr))
}

func stringContainsFold(s, substr string) bool {
	sRunes := []rune(s)
	subRunes := []rune(substr)
	for i := 0; i+len(subRunes) <= len(sRunes); i++ {
		match := true
		for j := range subRunes {
			if toLower(sRunes[i+j]) != toLower(subRunes[j]) {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}

func toLower(r rune) rune {
	if r >= 'A' && r <= 'Z' {
		return r + ('a' - 'A')
	}
	return r
}
