package ui

import (
	"github.com/scramb/backlog-manager/internal/i18n"
	"github.com/scramb/backlog-manager/ui/settings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
)

func SettingsView(app fyne.App, w fyne.Window) fyne.CanvasObject {
	subTabs := container.NewAppTabs(
		container.NewTabItem(i18n.T("settings.jira_config"), settings.BuildJiraSettings(app, w)),
		container.NewTabItem(i18n.T("settings.ai_config"), settings.BuildAISettings(app, w)),
		container.NewTabItem(i18n.T("settings.label_config"), settings.BuildLabelSettings(app, w)),
		container.NewTabItem(i18n.T("settings.app_config"), settings.BuildAppSettings(app, w)),
	)
	subTabs.SetTabLocation(container.TabLocationTop)

	i18n.RegisterOnLanguageChange(func() {
		fyne.Do(func() {
			subTabs.Items[0].Text = i18n.T("settings.jira_config")
			subTabs.Items[1].Text = i18n.T("settings.ai_config")
			subTabs.Items[2].Text = i18n.T("settings.label_config")
			subTabs.Items[3].Text = i18n.T("settings.app_config")
			subTabs.Refresh()
		})
	})

	return subTabs
}
