package i18n

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

// BindLabel creates a label bound to a translation key that updates when the language changes.
func BindLabel(key string) *widget.Label {
	lbl := widget.NewLabel(T(key))

	RegisterOnLanguageChange(func() {
		fyne.Do(func() {
			lbl.SetText(T(key))
		})
	})

	return lbl
}

// BindButton creates a button with live-translated text that updates automatically.
func BindCheckbox(key string) *widget.Check {
	checkbox := widget.NewCheck(T(key), nil)
	RegisterOnLanguageChange(func() {
		fyne.Do(func() {
			checkbox.SetText(T(key))
		})
	})

	return checkbox
}

// BindButton creates a button with live-translated text that updates automatically.
func BindButton(key string, icon fyne.Resource, tapped func()) *widget.Button {
	btn := widget.NewButtonWithIcon(T(key), icon, tapped)

	RegisterOnLanguageChange(func() {
		fyne.Do(func() {
			btn.SetText(T(key))
		})
	})

	return btn
}

// BindEntry creates an Entry with live-updated placeholder text.
func BindEntry(key string, password bool) *widget.Entry {
	var entry *widget.Entry
	if password {
		entry = widget.NewPasswordEntry()
	} else {
		entry = widget.NewEntry()
	}
	entry.SetPlaceHolder(T(key))
	RegisterOnLanguageChange(func() {
		fyne.Do(func() {
			entry.SetPlaceHolder(T(key))
		})
	})
	return entry
}

func BindEntryWithPlaceholder(key string, password bool) *widget.Entry {
	entry := widget.NewEntry()
	entry.SetPlaceHolder(T(key))

	if password {
		entry.Password = true
	}

	RegisterOnLanguageChange(func() {
		entry.SetPlaceHolder(T(key))
	})

	return entry
}
