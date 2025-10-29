package ui

import (
	"fmt"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/scramb/backlog-manager/internal/i18n"
	"github.com/scramb/backlog-manager/internal/models"
)

// NewBacklogView builds the Create Backlog tab content
// It handles loading favourite projects, per-project issue types,
// OpenAI generation, and creating the Jira issue.
func BacklogView(app fyne.App, w fyne.Window, domain, user, token string) fyne.CanvasObject {
	// Inputs
	titleEntry := widget.NewEntry()
	contentEntry := widget.NewMultiLineEntry()
	contentEntry.Wrapping = fyne.TextWrapWord
	contentEntry.SetMinRowsVisible(10)

	projectSelect := widget.NewSelect([]string{i18n.T("backlog.load_projects")}, nil)
	issueType := widget.NewSelect([]string{i18n.T("backlog.load_types")}, nil)
	issueType.Disable()

	labelChecks := map[string]*widget.Check{}
	updateLabelsGrid := func(savedLabels []string) fyne.CanvasObject {
		numLabels := len(savedLabels)
		if numLabels == 0 {
			return i18n.BindLabel("backlog.no_labels")
		}

		labelsPerRow := 3
		labelChecks = make(map[string]*widget.Check)
		gridObjects := []fyne.CanvasObject{}

		for _, label := range savedLabels {
			check := widget.NewCheck(label, func(bool) {})
			labelChecks[label] = check
			gridObjects = append(gridObjects, check)
		}

		// container.NewGridWithColumns sorgt für gleichmäßige Spalten und saubere Ausrichtung
		grid := container.NewGridWithColumns(labelsPerRow, gridObjects...)
		return container.NewVBox(grid)
	}
	labelsGrid := container.NewVBox(i18n.BindLabel("backlog.load_types"))

	createBtn := i18n.BindButton("backlog.create", nil, nil)

	// zuerst deklarieren, aber noch ohne Handler
	generateBtn := i18n.BindButton("backlog.ai_generate", theme.ComputerIcon(), nil)
	// Disable generate if no AI endpoint configured
	if app.Preferences().String("ai_endpoint") == "" {
		generateBtn.Disable()
		fyne.Do(func() {
			dialog.ShowInformation(i18n.T("backlog.ai_disabled_title"), i18n.T("backlog.ai_disabled_message"), w)
		})
	}

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
					dialog.ShowInformation(i18n.T("backlog.error"), i18n.T("backlog.ai_missing_key"), w)
					generateBtn.Enable()
				})
				return
			}

			userPrompt := fmt.Sprintf("%s '%s'", i18n.T("backlog.ai_generate_prompt"), titleEntry.Text)
			endpoint := app.Preferences().String("ai_endpoint")
			result, err := models.GenerateBacklogContent(apiKey, endpoint, systemPrompt, userPrompt)
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

		issueType.Options = []string{i18n.T("backlog.load_types")}
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

		go func() {
			key := fmt.Sprintf("labels_%s", projectKey)
			saved := app.Preferences().String(key)
			var savedLabels []string
			if saved != "" {
				savedLabels = strings.Split(saved, ",")
			}
			fyne.Do(func() {
				if len(savedLabels) > 0 {
					grid := updateLabelsGrid(savedLabels)
					labelsGrid.Objects = []fyne.CanvasObject{grid}
				} else {
					labelsGrid.Objects = []fyne.CanvasObject{i18n.BindLabel("backlog.no_labels")}
				}
				labelsGrid.Refresh()
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
					dialog.ShowInformation(i18n.T("backlog.error"), i18n.T("backlog.error_fields"), w)
				})
				return
			}
			selected := projectSelect.Selected
			start := strings.LastIndex(selected, "(")
			end := strings.LastIndex(selected, ")")
			if start == -1 || end == -1 || start >= end {
				fyne.Do(func() {
					createBtn.Enable()
					dialog.ShowInformation(i18n.T("backlog.error"), i18n.T("backlog.error_project_format"), w)
				})
				return
			}
			projectKey := selected[start+1 : end]

			selectedType := issueType.Selected
			if selectedType == "" {
				fyne.Do(func() {
					createBtn.Enable()
					dialog.ShowInformation(i18n.T("backlog.error"), i18n.T("backlog.error_issue_type"), w)
				})
				return
			}

			var selectedLabels []string
			for label, check := range labelChecks {
				if check.Checked {
					selectedLabels = append(selectedLabels, label)
				}
			}

			err := models.CreateJiraIssue(domain, user, token, projectKey, selectedType, titleEntry.Text, contentEntry.Text, selectedLabels)
			fyne.Do(func() {
				createBtn.Enable()
				if err != nil {
					dialog.ShowError(err, w)
					return
				}
				dialog.ShowInformation(i18n.T("backlog.dialog_created"), i18n.T("backlog.created"), w)
				titleEntry.SetText("")
				contentEntry.SetText("")
			})
		}()
	}

	topControls := container.NewVBox(
		i18n.BindLabel("backlog.header"),
		i18n.BindLabel("backlog.project"),
		projectSelect,
		i18n.BindLabel("backlog.type"),
		issueType,
		i18n.BindLabel("backlog.title"),
		titleEntry,
		labelsGrid,
		i18n.BindLabel("backlog.description"),
		generateBtn,
	)
	createForm := container.NewBorder(topControls, createBtn, nil, nil, contentEntry)

	return createForm
}
