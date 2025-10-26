package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/scramb/backlog-manager/internal/i18n"
)

// ServiceDeskView – Platzhalter für Jira ServiceDesk-Funktionen
func ServiceDeskView(app fyne.App, w fyne.Window) fyne.CanvasObject {
	title := widget.NewLabelWithStyle(
		i18n.T("servicedesk.title"),
		fyne.TextAlignLeading,
		fyne.TextStyle{Bold: true},
	)

	info := widget.NewLabel(i18n.T("servicedesk.description"))

	refreshBtn := i18n.BindButton("servicedesk.refresh", theme.ViewRefreshIcon(), func() {
		// später: ServiceDesk API-Aufrufe (Tickets, Queues, etc.)
		dialog := widget.NewLabel(i18n.T("servicedesk.refresh_placeholder"))
		w.SetContent(container.NewVBox(title, info, dialog))
	})

	content := container.NewVBox(
		title,
		widget.NewSeparator(),
		info,
		refreshBtn,
	)

	// Reaktiver Sprachwechsel
	i18n.RegisterOnLanguageChange(func() {
		fyne.Do(func() {
			title.SetText(i18n.T("servicedesk.title"))
			info.SetText(i18n.T("servicedesk.description"))
			refreshBtn.SetText(i18n.T("servicedesk.refresh"))
		})
	})

	return container.NewBorder(nil, nil, nil, nil, content)
}