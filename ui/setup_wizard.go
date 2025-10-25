package ui

import (
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/scramb/backlog-manager/internal/i18n"
	"github.com/scramb/backlog-manager/internal/models"
)

func ShowSetupWizard(w fyne.Window, a fyne.App) {
	prefs := a.Preferences()

	domainEntry := i18n.BindEntryWithPlaceholder("setup.jira_domain_placeholder", false)
	userEntry := i18n.BindEntryWithPlaceholder("setup.email_placeholder", false)
	tokenEntry := i18n.BindEntryWithPlaceholder("setup.api_token_placeholder", true)

	saveBtn := i18n.BindButton("setup.save_start", nil, func() {
		encryptedToken, err := models.Encrypt(tokenEntry.Text)
		if err != nil {
			log.Printf("%s %v", i18n.T("setup.encrypt_error"), err)
			return
		}

		prefs.SetString("jira_domain", domainEntry.Text)
		prefs.SetString("jira_user", userEntry.Text)
		prefs.SetString("jira_token", encryptedToken)

		ShowMainApp(w, a, domainEntry.Text, userEntry.Text, tokenEntry.Text)
	})

	form := container.NewVBox(
		widget.NewLabelWithStyle(i18n.BindLabel("setup.title").Text, fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		i18n.BindLabel("setup.jira_domain"),
		domainEntry,
		i18n.BindLabel("setup.email"),
		userEntry,
		i18n.BindLabel("setup.api_token"),
		tokenEntry,
		saveBtn,
	)

	w.SetContent(container.NewCenter(form))
	w.Resize(fyne.NewSize(400, 300))
}
