package ui

import (
	"backlog-manager/internal/models"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"os/exec"
	"runtime"
	"strings"
)

func CreateMainWindow(app fyne.App) fyne.Window {
	w := app.NewWindow("Backlog Manager – Login")

	prefs := app.Preferences()

	domain := widget.NewEntry()
	domain.SetPlaceHolder("Jira-Domain (z. B. mycompany)")
	domain.SetText(prefs.String("jira_domain"))

	username := widget.NewEntry()
	username.SetPlaceHolder("E-Mail")
	username.SetText(prefs.String("jira_user"))

	password := widget.NewPasswordEntry()
	password.SetPlaceHolder("API Token")
	password.SetText(prefs.String("jira_token"))

	loginButton := widget.NewButton("Login", func() {
		d := domain.Text
		u := username.Text
		p := password.Text

		if d == "" || u == "" || p == "" {
			dialog.ShowInformation("Fehler", "Bitte alle Felder ausfüllen.", w)
			return
		}

		// speichern der Daten
		prefs.SetString("jira_domain", d)
		prefs.SetString("jira_user", u)
		prefs.SetString("jira_token", p)

		showMainApp(w, app, d, u, p)
	})

	form := container.NewVBox(
		widget.NewLabelWithStyle("🗂️ Backlog Manager", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewLabel("Bitte Jira-Daten eingeben"),
		domain,
		username,
		password,
		loginButton,
	)

	w.SetContent(container.NewCenter(form))
	w.Resize(fyne.NewSize(400, 300))
	return w
}

func showMainApp(w fyne.Window, app fyne.App, domain, user, token string) {
	var issues []models.JiraIssue

	// Create backlog form widgets
	titleEntry := widget.NewEntry()
	contentEntry := widget.NewMultiLineEntry()

	projectSelect := widget.NewSelect([]string{"Lade Projekte..."}, nil)
	issueType := widget.NewSelect([]string{"Task", "Story", "Bug"}, nil)
	issueType.Selected = "Task"
	createBtn := widget.NewButton("Erstellen", nil)

	// Load projects asynchronously
	go func() {
		projects, err := models.FetchFavouriteProjects(domain, user, token)
		if err != nil {
			dialog.ShowError(err, w)
			return
		}
		projectNames := make([]string, len(projects))
		for i, p := range projects {
			projectNames[i] = fmt.Sprintf("%s (%s)", p.Name, p.Key)
		}
		// Update UI on main thread
		fyne.Do(func() {
			projectSelect.Options = projectNames
			if len(projectNames) > 0 {
				projectSelect.Selected = projectNames[0]
			}
			projectSelect.Refresh()
		})
	}()

	createBtn.OnTapped = func() {
		createBtn.Disable()
		go func() {
			if projectSelect.Selected == "" || issueType.Selected == "" || titleEntry.Text == "" {
				fyne.Do(func() {
					createBtn.Enable()
					dialog.ShowInformation("Fehler", "Bitte alle Felder ausfüllen.", w)
				})
				return
			}
			selected := projectSelect.Selected
			start := strings.LastIndex(selected, "(")
			end := strings.LastIndex(selected, ")")
			if start == -1 || end == -1 || start >= end {
				fyne.Do(func() {
					createBtn.Enable()
					dialog.ShowInformation("Fehler", "Ungültiges Projektformat.", w)
				})
				return
			}
			projectKey := selected[start+1 : end]

			err := models.CreateJiraIssue(domain, user, token, projectKey, issueType.Selected, titleEntry.Text, contentEntry.Text)
			fyne.Do(func() {
				createBtn.Enable()
				if err != nil {
					dialog.ShowError(err, w)
					return
				}
				dialog.ShowInformation("Erstellt", "Backlog Item erfolgreich erstellt!", w)
				titleEntry.SetText("")
				contentEntry.SetText("")
			})
		}()
	}

	createForm := container.NewVBox(
		widget.NewLabelWithStyle("📝 Create Backlog", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabel("Projekt"),
		projectSelect,
		widget.NewLabel("Typ"),
		issueType,
		widget.NewLabel("Titel"),
		titleEntry,
		widget.NewLabel("Beschreibung"),
		contentEntry,
		createBtn,
	)

	// --- Tickets ---
	ticketList := widget.NewList(
		func() int { return len(issues) },
		func() fyne.CanvasObject {
			icon := widget.NewIcon(nil)
			id := widget.NewLabel("ID")
			title := widget.NewLabel("Titel")
			box := container.NewHBox(icon, id, title)
			return box
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
				dialog.ShowError(err, w)
				return
			}

			// Update UI on main thread
			fyne.Do(func() {

				issues = fetched // 👈 hier speichern wir sie
				ticketList.Length = func() int { return len(issues) }
				ticketList.UpdateItem = func(i widget.ListItemID, o fyne.CanvasObject) {
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
				}
				ticketList.Refresh()
			})
		}()
	})

	ticketsView := container.NewBorder(
		refreshBtn, nil, nil, nil,
		ticketList,
	)

	tabs := container.NewAppTabs(
		container.NewTabItem("📝 Create Backlog", createForm),
		container.NewTabItem("🎫 My Tickets", ticketsView),
	)
tabs.OnChanged = func(tab *container.TabItem) {
    if tab.Text == "🎫 My Tickets" {
        // wenn Reiter My Tickets gewählt wurde → neu laden
        refreshBtn.OnTapped()
    }
}
	w.SetContent(tabs)
	w.SetTitle("Backlog Manager")
	w.Resize(fyne.NewSize(700, 500))
}

func openBrowser(url string) {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "darwin":
		cmd = "open"
	case "windows":
		cmd = "rundll32"
		args = []string{"url.dll,FileProtocolHandler"}
	default: // Linux
		cmd = "xdg-open"
	}

	args = append(args, url)
	exec.Command(cmd, args...).Start()
}
