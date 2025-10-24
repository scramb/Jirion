package main

import (
	"fyne.io/fyne/v2/app"
	"github.com/scramb/backlog-manager/ui"
)

func main() {
	// Mit eindeutiger ID starten (fix für Preferences-Fehler)
	a := app.NewWithID("com.scramb.backlogmanager")
	w := a.NewWindow("Backlog Manager")

	// Persistierte Einstellungen aus Preferences lesen
	prefs := a.Preferences()
	domain := prefs.String("jira_domain")
	user := prefs.String("jira_user")
	token := prefs.String("jira_token")

	ui.ShowMainApp(w, a, domain, user, token)
	w.ShowAndRun()
}