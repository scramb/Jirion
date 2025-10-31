package components

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
)

func CreateChip(text string) fyne.CanvasObject {
	txt := canvas.NewText(text, theme.ForegroundColor())
	txt.Alignment = fyne.TextAlignCenter
	txt.TextSize = theme.TextSize() - 4

	textSize := fyne.MeasureText(text, txt.TextSize, fyne.TextStyle{})
	paddingX := float32(16)
	paddingY := float32(6)

	bg := canvas.NewRectangle(theme.ButtonColor())
	bg.CornerRadius = 12

	chipWidth := textSize.Width + paddingX*2
	chipHeight := textSize.Height + paddingY*2

	box := container.NewStack(bg, container.NewCenter(txt))
	box.Resize(fyne.NewSize(chipWidth, chipHeight))
	return box
}