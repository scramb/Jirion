package main

import (
	"fyne.io/fyne/v2/app"
	"github.com/scramb/backlog-manager/ui"
	"os"
)

func main() {
	// Mit eindeutiger ID starten (fix für Preferences-Fehler)
	os.Setenv("FYNE_APP_STORAGE", "/Users/carstenmeininger/go/src/backlog-manager/.fyne")
	a := app.NewWithID("backlog-manager")
	w := a.NewWindow("Backlog Manager")

	// Persistierte Einstellungen aus Preferences lesen
	prefs := a.Preferences()
	domain := prefs.String("jira_domain")
	user := prefs.String("jira_user")
	token := prefs.String("jira_token")


	if domain == "" || user == "" || token == "" {
			ui.ShowSetupWizard(w, a)
	} else {
			ui.ShowMainApp(w, a, domain, user, token)
	}
	w.ShowAndRun()
}
