package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func ShowSetupWizard(w fyne.Window, a fyne.App) {
	prefs := a.Preferences()

	domainEntry := widget.NewEntry()
	domainEntry.SetPlaceHolder("z. B. my-jira-space")

	userEntry := widget.NewEntry()
	userEntry.SetPlaceHolder("E-Mail-Adresse")

	tokenEntry := widget.NewPasswordEntry()
	tokenEntry.SetPlaceHolder("Jira API Token")

	saveBtn := widget.NewButton("Speichern & Starten", func() {
		prefs.SetString("jira_domain", domainEntry.Text)
		prefs.SetString("jira_user", userEntry.Text)
		prefs.SetString("jira_token", tokenEntry.Text)
		ShowMainApp(w, a, domainEntry.Text, userEntry.Text, tokenEntry.Text)
	})

	form := container.NewVBox(
		widget.NewLabelWithStyle("Willkommen! Bitte Jira-Zugangsdaten eingeben:", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewLabel("Jira Domain:"),
		domainEntry,
		widget.NewLabel("E-Mail:"),
		userEntry,
		widget.NewLabel("API Token:"),
		tokenEntry,
		saveBtn,
	)

	w.SetContent(container.NewCenter(form))
	w.Resize(fyne.NewSize(400, 300))
}
