package components

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func CollapsibleSection(title string, content fyne.CanvasObject) fyne.CanvasObject {
	header := widget.NewButtonWithIcon(title, theme.MenuDropDownIcon(), nil)
	body := container.NewVBox(content)
	body.Hide()

	header.OnTapped = func() {
		if body.Visible() {
			body.Hide()
			header.SetIcon(theme.MenuDropDownIcon())
		} else {
			body.Show()
			header.SetIcon(theme.MenuDropUpIcon())
		}
	}

	return container.NewVBox(header, body)
}
