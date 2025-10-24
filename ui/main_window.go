package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
)

// showMainApp initializes the main application window after login.
// It combines all UI views (Backlog, My Tickets, Settings) into tabs.
func ShowMainApp(w fyne.Window, app fyne.App, domain, user, token string) {
	// Initialize views
	createView := BacklogView(app, w, domain, user, token)
	reloadTickets := make(chan bool)
	ticketsView := TicketsView(app, w, domain, user, token, reloadTickets)
	settingsView := SettingsView(app, w)

	// Build tab container
	tabs := container.NewAppTabs(
		container.NewTabItem("📝 Create Backlog", createView),
		container.NewTabItem("🎫 My Tickets", ticketsView),
		container.NewTabItem("⚙️ Settings", settingsView),
	)

	tabs.OnChanged = func(tab *container.TabItem) {
		if tab.Text == "🎫 My Tickets" {
			reloadTickets <- true
		}
	}

	// Set up window
	w.SetContent(tabs)
	w.SetTitle("Backlog Manager")
	w.Resize(fyne.NewSize(800, 600))
}
