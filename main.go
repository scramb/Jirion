package main

import (
	"github.com/scramb/backlog-manager/ui"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/theme"
)

func main() {
	myApp := app.NewWithID("com.backlog.manager")
	myApp.Settings().SetTheme(theme.LightTheme())

	w := ui.CreateMainWindow(myApp) // Fenster-Objekt zurückgeben lassen
	w.ShowAndRun()                  // Fenster starten und App laufen lassen
}
