package ui

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/scramb/backlog-manager/internal/i18n"
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
	domainEntry := i18n.BindEntryWithPlaceholder("settings.jira_domain_placeholder", false)
	domainEntry.SetText(prefs.String("jira_domain"))

	userEntry := i18n.BindEntryWithPlaceholder("settings.jira_user_placeholder", false)
	userEntry.SetText(prefs.String("jira_user"))

	tokenEntry := i18n.BindEntryWithPlaceholder("settings.jira_token_placeholder", true)
	tokenEntry.SetText(prefs.String("jira_token"))

	jiraSaveBtn := i18n.BindButton("settings.save", theme.ConfirmIcon(), func() {
		prefs.SetString("jira_domain", domainEntry.Text)
		prefs.SetString("jira_user", userEntry.Text)
		encryptedToken, err := models.Encrypt(tokenEntry.Text)
		if err != nil {
			dialog.ShowError(fmt.Errorf(i18n.T("settings.error_encrypt_token")+": %w", err), w)
			return
		}
		prefs.SetString("jira_token", encryptedToken)
		dialog.ShowInformation(i18n.T("settings.saved_title"), i18n.T("settings.jira_saved"), w)
	})

	jiraForm := container.NewVBox(
		widget.NewLabelWithStyle(i18n.T("settings.jira_config"), fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		i18n.BindLabel("settings.jira_domain"),
		domainEntry,
		i18n.BindLabel("settings.jira_user"),
		userEntry,
		i18n.BindLabel("settings.jira_token"),
		tokenEntry,
		jiraSaveBtn,
	)

	// === AI Configuration ===
	endpointEntry := i18n.BindEntryWithPlaceholder("settings.ai_endpoint_placeholder", false)
	endpointEntry.SetText(prefs.String("ai_endpoint"))
	if endpointEntry.Text == "" {
		endpointEntry.SetText("https://api.openai.com/v1") // default
	}

	systemPromptEntry := widget.NewMultiLineEntry()
	systemPromptEntry.SetText(prefs.String("system_prompt"))

	apiKeyEntry := i18n.BindEntryWithPlaceholder("settings.api_key", true)

	// Model select for OpenAI-compatible models
	modelSelect := widget.NewSelect([]string{i18n.T("settings.loading_models")}, nil)
	modelSelect.PlaceHolder = i18n.T("settings.select_model")

	// Load available models in background
	go func() {
		domain := prefs.String("ai_endpoint")
		apiKey := prefs.String("openai_api_key")
		if apiKey == "" {
			return
		}
		modelsList, err := models.FetchAvailableModels(domain, apiKey)
		if err != nil {
			fyne.Do(func() { dialog.ShowError(err, w) })
			return
		}
		fyne.Do(func() {
			modelSelect.Options = modelsList
			modelSelect.Refresh()
			savedModel := prefs.String("openai_model")
			if savedModel != "" {
				modelSelect.SetSelected(savedModel)
			}
		})
	}()

	modelSelect.OnChanged = func(selected string) {
		prefs.SetString("openai_model", selected)
	}

	aiSaveBtn := i18n.BindButton("settings.save", theme.ConfirmIcon(), func() {
		encryptedKey, err := models.Encrypt(apiKeyEntry.Text)
		if err != nil {
			dialog.ShowError(fmt.Errorf(i18n.T("settings.error_encrypt_api_key")+": %w", err), w)
			return
		}
prefs.SetString("ai_endpoint", endpointEntry.Text)
prefs.SetString("system_prompt", systemPromptEntry.Text)

if strings.TrimSpace(apiKeyEntry.Text) != "" {
    encryptedKey, err := models.Encrypt(apiKeyEntry.Text)
    if err != nil {
        dialog.ShowError(fmt.Errorf(i18n.T("settings.error_encrypt_api_key")+": %w", err), w)
        return
    }
    prefs.SetString("openai_api_key", encryptedKey)
}
// Wenn leer gelassen, bleibt der bestehende (bereits verschlüsselte) Key erhalten
dialog.ShowInformation(i18n.T("settings.saved_title"), i18n.T("settings.ai_saved"), w)
		prefs.SetString("openai_api_key", encryptedKey)
		dialog.ShowInformation(i18n.T("settings.saved_title"), i18n.T("settings.ai_saved"), w)
	})

	aiForm := container.NewVBox(
		widget.NewLabelWithStyle(i18n.T("settings.ai_config"), fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		i18n.BindLabel("settings.ai_endpoint"),
		endpointEntry,
		i18n.BindLabel("settings.system_prompt"),
		systemPromptEntry,
		i18n.BindLabel("settings.api_key"),
		apiKeyEntry,
		i18n.BindLabel("settings.model"),
		modelSelect,
		aiSaveBtn,
	)

	// === Label Configuration ===
	projectSelect := widget.NewSelect([]string{i18n.T("settings.loading_projects")}, nil)
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
		fmt.Println(i18n.T("settings.saved_labels_for"), projectKey, "=>", selected)
		fmt.Println(i18n.T("settings.confirm_saved_value"), app.Preferences().String(key))
	}

	saveLabelsBtn := i18n.BindButton("settings.save", theme.ConfirmIcon(), func() {
		selectedProj := projectSelect.Selected
		start := strings.LastIndex(selectedProj, "(")
		end := strings.LastIndex(selectedProj, ")")
		if start == -1 || end == -1 || start >= end {
			dialog.ShowInformation(i18n.T("settings.notice"), i18n.T("settings.select_project_first"), w)
			return
		}
		projectKey := selectedProj[start+1 : end]
		selected := labelsGroup.Selected
		key := fmt.Sprintf("label_selection_%s", projectKey)
		data, _ := json.Marshal(selected)
		app.Preferences().SetString(key, string(data))
		fmt.Println(i18n.T("settings.manually_saved_for"), projectKey, "=>", selected)
		fmt.Println(i18n.T("settings.confirm_saved_value"), app.Preferences().String(key))
		dialog.ShowInformation(i18n.T("settings.saved_title"), fmt.Sprintf(i18n.T("settings.labels_saved_for"), projectKey), w)
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
						fmt.Println(i18n.T("settings.loaded_saved_json_for"), projectKey, "=>", saved)
						if saved != "" {
							var selectedLabels []string
							if err := json.Unmarshal([]byte(saved), &selectedLabels); err == nil {
								isRestoring = true
								labelsGroup.SetSelected(selectedLabels)
								isRestoring = false
								fmt.Println(i18n.T("settings.restored_labels_for"), projectKey, "=>", selectedLabels)
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
			widget.NewLabelWithStyle(i18n.T("settings.label_config"), fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			i18n.BindLabel("settings.project"),
			projectSelect,
			saveLabelsBtn,
			widget.NewSeparator(),
			i18n.BindLabel("settings.labels_in_project"),
		),
		nil, nil, nil,
		container.NewVScroll(labelsGroup),
	)

	// === App Configuration ===
	languageSelect := widget.NewSelect([]string{"en", "de"}, func(selected string) {
		prefs.SetString("lang", selected)
		if err := i18n.LoadLanguage(selected); err != nil {
			dialog.ShowError(fmt.Errorf(i18n.T("settings.error_load_language")+": %w", err), w)
			return
		}
		// Sprache wird stillschweigend geändert, kein Popup
	})
	languageSelect.SetSelected(prefs.StringWithFallback("lang", "en"))

	appConfigForm := container.NewVBox(
		widget.NewLabelWithStyle(i18n.T("settings.app_config"), fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		i18n.BindLabel("settings.language"),
		languageSelect,
	)

	// === Sub-Tabs ===
	subTabs := container.NewAppTabs(
		container.NewTabItem(i18n.T("settings.jira_config"), jiraForm),
		container.NewTabItem(i18n.T("settings.ai_config"), aiForm),
		container.NewTabItem(i18n.T("settings.label_config"), labelConfig),
		container.NewTabItem(i18n.T("settings.app_config"), appConfigForm),
	)
	subTabs.SetTabLocation(container.TabLocationTop)

i18n.RegisterOnLanguageChange(func() {
    fyne.Do(func() {
        subTabs.Items[0].Text = i18n.T("settings.jira_config")
        subTabs.Items[1].Text = i18n.T("settings.ai_config")
        subTabs.Items[2].Text = i18n.T("settings.label_config")
        subTabs.Items[3].Text = i18n.T("settings.app_config")

        // Jira section
        jiraSaveBtn.SetText(i18n.T("settings.save"))

        // AI section
        aiSaveBtn.SetText(i18n.T("settings.save"))

        // App Config
        languageSelect.PlaceHolder = i18n.T("settings.language")
        subTabs.Refresh()
    })
})

	return subTabs
}
