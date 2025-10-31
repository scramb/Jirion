package settings

import (
	"fmt"
	"strings"

	"github.com/scramb/backlog-manager/internal/i18n"
	"github.com/scramb/backlog-manager/internal/models"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// BuildLabelSettings builds the label configuration settings tab.
func BuildLabelSettings(app fyne.App, w fyne.Window) fyne.CanvasObject {
	prefs := app.Preferences()

	projectSelect := widget.NewSelect([]string{}, nil)
	projectSelect.PlaceHolder = i18n.T("settings.label_project")

	labelContainer := container.NewVBox()

	domain := prefs.String("jira_domain")
	user := prefs.String("jira_user")
	token := models.TryDecrypt(prefs.String("jira_token"))
	if domain != "" && user != "" && token != "" {
		projects, err := models.FetchFavouriteProjects(domain, user, token)
		if err == nil {
			var projectNames []string
			for _, p := range projects {
				projectNames = append(projectNames, p.Key)
			}
			projectSelect.Options = projectNames
			projectSelect.Refresh()
		}
	}

	saveBtn := i18n.BindButton("settings.save", theme.ConfirmIcon(), func() {
		currentProject := projectSelect.Selected
		if currentProject == "" {
			dialog.ShowError(fmt.Errorf(i18n.T("settings.no_project_selected")), w)
			return
		}
		var selected []string
		if len(labelContainer.Objects) > 0 {
			// labelContainer contains grid(s) of checkboxes
			for _, obj := range labelContainer.Objects {
				if grid, ok := obj.(*fyne.Container); ok {
					for _, cbObj := range grid.Objects {
						if cb, ok := cbObj.(*widget.Check); ok && cb.Checked {
							selected = append(selected, cb.Text)
						}
					}
				}
			}
		}
		prefs.SetString(fmt.Sprintf("labels_%s", currentProject), strings.Join(selected, ","))
		dialog.ShowInformation(i18n.T("settings.saved_title"), i18n.T("settings.labels_saved"), w)
	})

	projectSelect.OnChanged = func(project string) {
		labelContainer.Objects = nil
		labelContainer.Refresh()

		domain := prefs.String("jira_domain")
		user := prefs.String("jira_user")
		token := models.TryDecrypt(prefs.String("jira_token"))

		labels, err := models.FetchProjectLabels(domain, user, token, project)
		if err != nil {
			dialog.ShowError(fmt.Errorf(i18n.T("settings.error_load_labels")+": %w", err), w)
			return
		}

		// Restore saved selections
		saved := prefs.String(fmt.Sprintf("labels_%s", project))
		var selected []string
		if saved != "" {
			selected = strings.Split(saved, ",")
		}

		grid := container.NewGridWithColumns(5)
		for _, label := range labels {
			cb := widget.NewCheck(label, nil)
			for _, s := range selected {
				if s == label {
					cb.SetChecked(true)
					break
				}
			}
			grid.Add(cb)
		}
		labelContainer.Add(grid)
		labelContainer.Refresh()
	}

	formContent := container.NewVBox(
		widget.NewLabelWithStyle(i18n.T("settings.label_config"), fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		i18n.BindLabel("settings.label_project"),
		projectSelect,
		labelContainer,
	)

	scroll := container.NewVScroll(formContent)
	scroll.SetMinSize(fyne.NewSize(400, 300))

	pinnedSave := container.NewBorder(nil, saveBtn, nil, nil, scroll)

	return pinnedSave
}
