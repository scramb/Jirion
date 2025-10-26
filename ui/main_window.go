package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"

	"github.com/scramb/backlog-manager/internal/i18n"
)

// showMainApp initializes the main application window after login.
// It combines all UI views (Backlog, My Tickets, Settings) into tabs.
func ShowMainApp(w fyne.Window, app fyne.App, domain, user, token string) {
	// Initialize views
	createView := BacklogView(app, w, domain, user, token)
	reloadTickets := make(chan bool)
	ticketsView := TicketsView(app, w, domain, user, token, reloadTickets)
	settingsView := SettingsView(app, w)
	serviceDeskView := ServiceDeskView(app, w)

	// Build tab container
	tabs := container.NewAppTabs(
		container.NewTabItem(i18n.T("tab.create_backlog"), createView),
		container.NewTabItem(i18n.T("tab.my_tickets"), ticketsView),
		container.NewTabItem(i18n.T("tab.servicedesk"), serviceDeskView),
		container.NewTabItem(i18n.T("tab.settings"), settingsView),
	)

	i18n.RegisterOnLanguageChange(func() {
		fyne.Do(func() {
			tabs.Items[0].Text = i18n.T("tab.create_backlog")
			tabs.Items[1].Text = i18n.T("tab.my_tickets")
			tabs.Items[2].Text = i18n.T("tab.servicedesk")
			tabs.Items[3].Text = i18n.T("tab.settings")
			tabs.Refresh()
			w.SetTitle(i18n.T("app.title"))
		})
	})

	tabs.OnSelected = func(tab *container.TabItem) {
		if tab.Text == i18n.T("tab.my_tickets") {
			reloadTickets <- true
		}
	}

	// Set up window
	w.SetContent(tabs)

	w.SetTitle(i18n.T("app.title"))
	w.Resize(fyne.NewSize(800, 600))
}
