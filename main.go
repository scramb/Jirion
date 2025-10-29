package main

import (
	"fmt"
	"os"

	"fyne.io/fyne/v2/app"
	"github.com/scramb/backlog-manager/internal/i18n"
	"github.com/scramb/backlog-manager/ui"
)

func main() {
	// Mit eindeutiger ID starten (fix für Preferences-Fehler)
	os.Setenv("FYNE_APP_STORAGE", "/Users/carstenmeininger/go/src/backlog-manager/.fyne")
	a := app.NewWithID("com.scramb.backlog-manager")
	w := a.NewWindow("Backlog Manager")

	// Persistierte Einstellungen aus Preferences lesen
	prefs := a.Preferences()
	domain := prefs.String("jira_domain")
	user := prefs.String("jira_user")
	token := prefs.String("jira_token")
	lang := prefs.StringWithFallback("language", "de")
	if err := i18n.LoadLanguage(lang); err != nil {
		fmt.Println("Failed to load language:", err)
	}
	if domain == "" || user == "" || token == "" {
		ui.ShowSetupWizard(w, a)
	} else {
		ui.ShowMainApp(w, a, domain, user, token)
	}

	w.ShowAndRun()
}