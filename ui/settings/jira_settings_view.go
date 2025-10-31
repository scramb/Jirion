package settings

import (
	"fmt"

	"github.com/scramb/backlog-manager/internal/i18n"
	"github.com/scramb/backlog-manager/internal/models"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// BuildJiraSettings builds the Jira settings tab UI.
func BuildJiraSettings(app fyne.App, w fyne.Window) fyne.CanvasObject {
	prefs := app.Preferences()

	domainEntry := i18n.BindEntryWithPlaceholder("settings.jira_domain_placeholder", false)
	domainEntry.SetText(prefs.String("jira_domain"))

	userEntry := i18n.BindEntryWithPlaceholder("settings.jira_user_placeholder", false)
	userEntry.SetText(prefs.String("jira_user"))

	tokenEntry := i18n.BindEntryWithPlaceholder("settings.jira_token_placeholder", true)

	// Synchronize Jira token behavior with AI token logic
	if enc := prefs.String("jira_token"); enc != "" {
		if dec := models.TryDecrypt(enc); dec != "" && tokenEntry.Text == "" {
			tokenEntry.SetText(dec)
		}
	}

	saveBtn := i18n.BindButton("settings.save", theme.ConfirmIcon(), func() {
		prefs.SetString("jira_domain", domainEntry.Text)
		prefs.SetString("jira_user", userEntry.Text)

		if tokenEntry.Text != "" {
			encryptedKey, err := models.Encrypt(tokenEntry.Text)
			if err != nil {
				dialog.ShowError(fmt.Errorf(i18n.T("settings.error_encrypt_api_key")+": %w", err), w)
				return
			}
			prefs.SetString("openai_api_key", encryptedKey)
		}

		dialog.ShowInformation(i18n.T("settings.saved_title"), i18n.T("settings.jira_saved"), w)
	})

	formContent := container.NewVBox(
		widget.NewLabelWithStyle(i18n.T("settings.jira_config"), fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		i18n.BindLabel("settings.jira_domain"),
		domainEntry,
		i18n.BindLabel("settings.jira_user"),
		userEntry,
		i18n.BindLabel("settings.jira_token"),
		tokenEntry,
	)

	scroll := container.NewVScroll(formContent)
	scroll.SetMinSize(fyne.NewSize(400, 300))

	pinnedSave := container.NewBorder(nil, saveBtn, nil, nil, scroll)

	return pinnedSave
}
