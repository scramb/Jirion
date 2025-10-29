package settings

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/scramb/backlog-manager/internal/i18n"
)

// BuildAppSettings builds the application settings tab.
func BuildAppSettings(app fyne.App, w fyne.Window) fyne.CanvasObject {
	prefs := app.Preferences()

	enableExperimental := i18n.BindCheckbox("settings.experimental")
	enableExperimental.SetChecked(prefs.Bool("experimental_enabled"))

	enableExperimental.OnChanged = func(checked bool) {
		prefs.SetBool("experimental_enabled", checked)
	}

	languages := map[string]string{
		"English": "en",
		"Deutsch": "de",
	}
	currentLang := prefs.String("language")
	if currentLang == "" {
		currentLang = "de"
	}

	langSelect := widget.NewSelect([]string{"English", "Deutsch"}, nil)
	for label, code := range languages {
		if code == currentLang {
			langSelect.SetSelected(label)
			break
		}
	}

	saveLangBtn := i18n.BindButton("settings.save", theme.ConfirmIcon(), func() {
		selectedLabel := langSelect.Selected
		selectedCode := languages[selectedLabel]
		prefs.SetString("language", selectedCode)
		i18n.LoadLanguage(selectedCode)
		dialog.ShowInformation(i18n.T("settings.saved_title"), fmt.Sprintf(i18n.T("settings.language_changed"), selectedCode), w)
	})

	resetBtn := i18n.BindButton("settings.reset_app", theme.DeleteIcon(), func() {
		dialog.ShowInformation(i18n.T("settings.reset_done_title"), i18n.T("settings.reset_done_message"), w)
	})

	formContent := container.NewVBox(
		enableExperimental,
		widget.NewLabelWithStyle(i18n.T("settings.app_config"), fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		i18n.BindLabel("settings.language"),
		langSelect,
		widget.NewSeparator(),
	)

	resetBtn.Importance = widget.DangerImportance

	scroll := container.NewVScroll(formContent)
	scroll.SetMinSize(fyne.NewSize(400, 250))

	pinnedButtons := container.NewBorder(nil, container.NewVBox(saveLangBtn, resetBtn), nil, nil, scroll)

	return pinnedButtons
}