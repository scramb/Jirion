package ui

import (
	"fmt"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/scramb/backlog-manager/internal/models"
)

// NewBacklogView builds the Create Backlog tab content
// It handles loading favourite projects, per-project issue types,
// OpenAI generation, and creating the Jira issue.
func BacklogView(app fyne.App, w fyne.Window, domain, user, token string) fyne.CanvasObject {
	// Inputs
	titleEntry := widget.NewEntry()
	contentEntry := widget.NewMultiLineEntry()

	projectSelect := widget.NewSelect([]string{"Lade Projekte..."}, nil)
	issueType := widget.NewSelect([]string{"Lade..."}, nil)
	issueType.Disable()

	createBtn := widget.NewButton("Erstellen", nil)

	// zuerst deklarieren, aber noch ohne Handler
	generateBtn := widget.NewButtonWithIcon("KI-Vorschlag generieren", theme.ComputerIcon(), nil)

	// dann Handler nachträglich setzen
	generateBtn.OnTapped = func() {
		generateBtn.Disable()
		go func() {
			systemPrompt := app.Preferences().String("system_prompt")
			if systemPrompt == "" {
				systemPrompt = "Du bist ein Jira-Assistent. Erstelle sinnvolle User Stories auf Basis von Titeln."
			}

			apiKey := app.Preferences().String("openai_api_key")
			if apiKey == "" {
				fyne.Do(func() {
					dialog.ShowInformation("Fehler", "Kein OpenAI API-Key hinterlegt. Bitte in den Einstellungen setzen.", w)
					generateBtn.Enable()
				})
				return
			}

			userPrompt := fmt.Sprintf("Erstelle eine Jira-Story basierend auf dem Titel: '%s'", titleEntry.Text)
			result, err := models.GenerateBacklogContent(apiKey, systemPrompt, userPrompt)

			fyne.Do(func() {
				generateBtn.Enable()
				if err != nil {
					dialog.ShowError(err, w)
					return
				}
				contentEntry.SetText(result)
			})
		}()
	}

	// Load favourite projects
	go func() {
		projects, err := models.FetchFavouriteProjects(domain, user, token)
		if err != nil {
			fyne.Do(func() { dialog.ShowError(err, w) })
			return
		}
		projectNames := make([]string, len(projects))
		for i, p := range projects {
			projectNames[i] = fmt.Sprintf("%s (%s)", p.Name, p.Key)
		}
		fyne.Do(func() {
			projectSelect.Options = projectNames
			if len(projectNames) > 0 {
				projectSelect.Selected = projectNames[0]
			}
			projectSelect.Refresh()
		})
	}()

	// Load issue types when project changes
	projectSelect.OnChanged = func(selected string) {
		start := strings.LastIndex(selected, "(")
		end := strings.LastIndex(selected, ")")
		if start == -1 || end == -1 {
			return
		}
		projectKey := selected[start+1 : end]

		issueType.Options = []string{"Lade..."}
		issueType.Disable()
		issueType.Refresh()

		go func() {
			types, err := models.FetchProjectIssueTypes(domain, user, token, projectKey)
			if err != nil {
				fyne.Do(func() { dialog.ShowError(err, w) })
				return
			}
			fyne.Do(func() {
				var names []string
				for _, t := range types {
					names = append(names, t.Name)
				}
				issueType.Options = names
				if len(names) > 0 {
					issueType.Selected = names[0]
					issueType.Enable()
				}
				issueType.Refresh()
			})
		}()
	}

	// Create issue
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

			selectedType := issueType.Selected
			if selectedType == "" {
				fyne.Do(func() {
					createBtn.Enable()
					dialog.ShowInformation("Fehler", "Bitte Issue Type auswählen.", w)
				})
				return
			}

			err := models.CreateJiraIssue(domain, user, token, projectKey, selectedType, titleEntry.Text, contentEntry.Text)
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

	// Layout
	createForm := container.NewVBox(
		widget.NewLabelWithStyle("📝 Create Backlog", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabel("Projekt"),
		projectSelect,
		widget.NewLabel("Typ"),
		issueType,
		widget.NewLabel("Titel"),
		titleEntry,
		widget.NewLabel("Beschreibung"),
		generateBtn,
		contentEntry,
		createBtn,
	)

	return createForm
}
