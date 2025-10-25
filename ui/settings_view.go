package ui

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/scramb/backlog-manager/internal/models"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// SettingsView builds the ⚙️ Settings tab with sub-tabs for Jira and AI configuration.
func SettingsView(app fyne.App, w fyne.Window) fyne.CanvasObject {
	prefs := app.Preferences()
	isRestoring := false
	isSettingLabels := false

	// === Jira Configuration ===
	domainEntry := widget.NewEntry()
	domainEntry.SetPlaceHolder("z. B. mycompany")
	domainEntry.SetText(prefs.String("jira_domain"))

	userEntry := widget.NewEntry()
	userEntry.SetPlaceHolder("E-Mail")
	userEntry.SetText(prefs.String("jira_user"))

	tokenEntry := widget.NewPasswordEntry()
	tokenEntry.SetPlaceHolder("API Token")
	tokenEntry.SetText(prefs.String("jira_token"))

	jiraSaveBtn := widget.NewButtonWithIcon("Speichern", theme.ConfirmIcon(), func() {
		prefs.SetString("jira_domain", domainEntry.Text)
		prefs.SetString("jira_user", userEntry.Text)
		encryptedToken, err := models.Encrypt(tokenEntry.Text)
		if err != nil {
			dialog.ShowError(fmt.Errorf("Fehler beim Verschlüsseln des Tokens: %w", err), w)
			return
		}
		prefs.SetString("jira_token", encryptedToken)
		dialog.ShowInformation("Gespeichert", "Jira-Einstellungen erfolgreich gespeichert!", w)
	})

	jiraForm := container.NewVBox(
		widget.NewLabelWithStyle("Jira Konfiguration", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabel("Jira-Domain"),
		domainEntry,
		widget.NewLabel("E-Mail"),
		userEntry,
		widget.NewLabel("API Token"),
		tokenEntry,
		jiraSaveBtn,
	)

	// === AI Configuration ===
	endpointEntry := widget.NewEntry()
	endpointEntry.SetPlaceHolder("z. B. https://api.openai.com/v1")
	endpointEntry.SetText(prefs.String("ai_endpoint"))
	if endpointEntry.Text == "" {
		endpointEntry.SetText("https://api.openai.com/v1") // default
	}

	systemPromptEntry := widget.NewMultiLineEntry()
	systemPromptEntry.SetText(prefs.String("system_prompt"))

	apiKeyEntry := widget.NewPasswordEntry()
	apiKeyEntry.SetText(prefs.String("openai_api_key"))

	aiSaveBtn := widget.NewButtonWithIcon("Speichern", theme.ConfirmIcon(), func() {
		encryptedKey, err := models.Encrypt(apiKeyEntry.Text)
		if err != nil {
			dialog.ShowError(fmt.Errorf("Fehler beim Verschlüsseln des API-Keys: %w", err), w)
			return
		}

		prefs.SetString("ai_endpoint", endpointEntry.Text)
		prefs.SetString("system_prompt", systemPromptEntry.Text)
		prefs.SetString("openai_api_key", encryptedKey)
		dialog.ShowInformation("Gespeichert", "AI-Einstellungen erfolgreich gespeichert!", w)
	})

	aiForm := container.NewVBox(
		widget.NewLabelWithStyle("AI Konfiguration", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabel("AI Endpoint"),
		endpointEntry,
		widget.NewLabel("System Prompt"),
		systemPromptEntry,
		widget.NewLabel("API Key"),
		apiKeyEntry,
		aiSaveBtn,
	)

	// === Label Configuration ===
	projectSelect := widget.NewSelect([]string{"Lade Projekte..."}, nil)
	labelsGroup := widget.NewCheckGroup([]string{}, nil) // selection handling will be added later

	labelsGroup.OnChanged = func(selected []string) {
		if isRestoring || isSettingLabels {
			return
		}
		selectedProj := projectSelect.Selected
		start := strings.LastIndex(selectedProj, "(")
		end := strings.LastIndex(selectedProj, ")")
		if start == -1 || end == -1 || start >= end {
			return
		}
		projectKey := selectedProj[start+1 : end]
		key := fmt.Sprintf("label_selection_%s", projectKey)
		data, _ := json.Marshal(selected)
		app.Preferences().SetString(key, string(data))
		fmt.Println("Saved labels for", projectKey, "=>", selected)
		fmt.Println("Confirm saved value:", app.Preferences().String(key))
	}

	saveLabelsBtn := widget.NewButtonWithIcon("Speichern", theme.ConfirmIcon(), func() {
		selectedProj := projectSelect.Selected
		start := strings.LastIndex(selectedProj, "(")
		end := strings.LastIndex(selectedProj, ")")
		if start == -1 || end == -1 || start >= end {
			dialog.ShowInformation("Hinweis", "Bitte zuerst ein Projekt auswählen.", w)
			return
		}
		projectKey := selectedProj[start+1 : end]
		selected := labelsGroup.Selected
		key := fmt.Sprintf("label_selection_%s", projectKey)
		data, _ := json.Marshal(selected)
		app.Preferences().SetString(key, string(data))
		fmt.Println("Manuell gespeichert für", projectKey, "=>", selected)
		fmt.Println("Confirm saved value:", app.Preferences().String(key))
		dialog.ShowInformation("Gespeichert", fmt.Sprintf("Labels für %s gespeichert!", projectKey), w)
	})

	// Load projects (all visible). Fallback to favourites on error.
	go func() {
		domain := prefs.String("jira_domain")
		user := prefs.String("jira_user")
		token := prefs.String("jira_token")
		if domain == "" || user == "" || token == "" {
			return
		}
		projs, err := models.FetchAllProjects(domain, user, token)
		if err != nil || len(projs) == 0 {
			projs, _ = models.FetchFavouriteProjects(domain, user, token)
		}
		names := make([]string, len(projs))
		for i, p := range projs {
			names[i] = fmt.Sprintf("%s (%s)", p.Name, p.Key)
		}
		fyne.Do(func() {
			projectSelect.Options = names
			if len(names) > 0 {
				// clear selection to ensure OnChanged triggers even if same value
				projectSelect.ClearSelected()
				projectSelect.Refresh()
				projectSelect.Selected = names[0]
				projectSelect.Refresh()
				// trigger initial label load manually
				projectSelect.OnChanged(names[0])
			} else {
				projectSelect.Refresh()
			}
		})
	}()

	// projectSelect.OnChanged keeps labels per project
	projectSelect.OnChanged = func(selected string) {
		start := strings.LastIndex(selected, "(")
		end := strings.LastIndex(selected, ")")
		if start == -1 || end == -1 || start >= end {
			return
		}
		projectKey := selected[start+1 : end]

		go func() {
			domain := prefs.String("jira_domain")
			user := prefs.String("jira_user")
			token := prefs.String("jira_token")
			if domain == "" || user == "" || token == "" {
				return
			}
			ls, err := models.FetchProjectLabels(domain, user, token, projectKey)
			if err != nil {
				fyne.Do(func() { dialog.ShowError(err, w) })
				return
			}

			fyne.Do(func() {
				// Guard programmatic changes to avoid empty saves
				isSettingLabels = true

				// Step 1: reset previous selection and options (this clears selection internally)
				labelsGroup.SetSelected([]string{})
				labelsGroup.Options = []string{}
				labelsGroup.Refresh()

				// Step 2: set new label options
				labelsGroup.Options = ls

				// Step 3: delayed restore of saved selection to ensure prefs sync
				time.AfterFunc(200*time.Millisecond, func() {
					fyne.Do(func() {
						key := fmt.Sprintf("label_selection_%s", projectKey)
						saved := prefs.String(key)
						fmt.Println("Loaded saved JSON for", projectKey, "=>", saved)
						if saved != "" {
							var selectedLabels []string
							if err := json.Unmarshal([]byte(saved), &selectedLabels); err == nil {
								isRestoring = true
								labelsGroup.SetSelected(selectedLabels)
								isRestoring = false
								fmt.Println("Restored labels for", projectKey, "=>", selectedLabels)
							}
						}
						labelsGroup.Refresh()
						// End of programmatic changes
						isSettingLabels = false
					})
				})
			})
		}()
	}

	labelConfig := container.NewBorder(
		container.NewVBox(
			widget.NewLabelWithStyle("Label Config", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			widget.NewLabel("Projekt"),
			projectSelect,
			saveLabelsBtn,
			widget.NewSeparator(),
			widget.NewLabel("Labels im Projekt"),
		),
		nil, nil, nil,
		container.NewVScroll(labelsGroup),
	)

	// === Sub-Tabs ===
	subTabs := container.NewAppTabs(
		container.NewTabItem("Jira Config", jiraForm),
		container.NewTabItem("AI Config", aiForm),
		container.NewTabItem("Label Config", labelConfig),
	)
	subTabs.SetTabLocation(container.TabLocationTop)

	return subTabs
}
