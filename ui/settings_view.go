package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// NewSettingsView builds the ⚙️ Settings tab content.
// It lets the user configure Jira credentials and OpenAI settings.
func SettingsView(app fyne.App, w fyne.Window) fyne.CanvasObject {
	prefs := app.Preferences()

	// Jira credentials
	domainEntry := widget.NewEntry()
	domainEntry.SetPlaceHolder("z. B. mycompany")
	domainEntry.SetText(prefs.String("jira_domain"))

	userEntry := widget.NewEntry()
	userEntry.SetPlaceHolder("E-Mail")
	userEntry.SetText(prefs.String("jira_user"))

	tokenEntry := widget.NewPasswordEntry()
	tokenEntry.SetPlaceHolder("API Token")
	tokenEntry.SetText(prefs.String("jira_token"))

	// OpenAI system prompt + API key
	systemPromptEntry := widget.NewMultiLineEntry()
	systemPromptEntry.SetText(prefs.String("system_prompt"))

	apiKeyEntry := widget.NewPasswordEntry()
	apiKeyEntry.SetText(prefs.String("openai_api_key"))

	saveBtn := widget.NewButtonWithIcon("Speichern", theme.ConfirmIcon(), func() {
		prefs.SetString("jira_domain", domainEntry.Text)
		prefs.SetString("jira_user", userEntry.Text)
		prefs.SetString("jira_token", tokenEntry.Text)
		prefs.SetString("system_prompt", systemPromptEntry.Text)
		prefs.SetString("openai_api_key", apiKeyEntry.Text)
		dialog.ShowInformation("Gespeichert", "Einstellungen erfolgreich gespeichert!", w)
	})

	settingsForm := container.NewVBox(
		widget.NewLabelWithStyle("⚙️ Einstellungen", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),

		widget.NewLabel("Jira-Domain"),
		domainEntry,
		widget.NewLabel("Jira-User (E-Mail)"),
		userEntry,
		widget.NewLabel("Jira-Token"),
		tokenEntry,

		widget.NewSeparator(),
		widget.NewLabel("Systemprompt"),
		systemPromptEntry,
		widget.NewLabel("OpenAI API-Key"),
		apiKeyEntry,

		saveBtn,
	)

	return settingsForm
}