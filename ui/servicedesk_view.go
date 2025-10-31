package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"github.com/scramb/backlog-manager/internal/i18n"
	servicedesk "github.com/scramb/backlog-manager/ui/service_desk"
)

// ServiceDeskView – Hauptansicht für Jira ServiceDesk-Funktionen
func ServiceDeskView(app fyne.App, w fyne.Window) fyne.CanvasObject {
	createTab := container.NewTabItem(i18n.T("servicedesk.create_request"), servicedesk.CreateView(app, w))
	listTab := container.NewTabItem(i18n.T("servicedesk.my_requests"), servicedesk.ListView(app, w))

	tabs := container.NewAppTabs(
		createTab,
		listTab,
	)

	tabs.SetTabLocation(container.TabLocationTop)

	i18n.RegisterOnLanguageChange(func() {
		fyne.Do(func() {
			createTab.Text = i18n.T("servicedesk.create_request")
			listTab.Text = i18n.T("servicedesk.my_requests")
			tabs.Refresh()
		})
	})

	return tabs
}
