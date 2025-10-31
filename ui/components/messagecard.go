package components

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/scramb/backlog-manager/internal/models"
)

// CreateChatMessageCard zeigt eine Chat-Nachricht als stylische Bubble an.
func CreateChatMessageCard(comment models.JiraComment, currentUser string) fyne.CanvasObject {
	isCurrentUser := comment.Author.Email == currentUser

	// Avatar oder Platzhalter
	var avatar fyne.CanvasObject
	if comment.Author.AvatarUrls.Image != "" {
		img := canvas.NewImageFromURI(storage.NewURI(comment.Author.AvatarUrls.Image))
		img.FillMode = canvas.ImageFillContain
		img.SetMinSize(fyne.NewSize(32, 32))
		avatar = container.NewStack(img)
	} else {
		ph := canvas.NewCircle(theme.ForegroundColor())
		ph.Resize(fyne.NewSize(32, 32))
		avatar = container.NewStack(ph)
	}

	// Nachrichtentext
	text := widget.NewLabel(models.ExtractDescriptionText(comment.Content))
	text.Wrapping = fyne.TextWrapWord
	text.Alignment = fyne.TextAlignLeading
	text.TextStyle = fyne.TextStyle{}

	// Bubble-Hintergrund
	bgColor := theme.ButtonColor()
	if isCurrentUser {
		bgColor = theme.PrimaryColor()
	}
	bg := canvas.NewRectangle(bgColor)
	bg.CornerRadius = 16

	// Begrenzte Breite der Chat-Bubble
	// Bubble mit Hintergrund und gepolstertem Text, füllt horizontal den verfügbaren Platz
	bubble := container.NewStack(bg, container.NewPadded(text))

	// Vertikale Anordnung mit Name
	nameLabel := canvas.NewText(comment.Author.DisplayName, theme.ForegroundColor())
	nameLabel.TextSize = theme.TextSize() - 2
	nameLabel.Alignment = fyne.TextAlignLeading
	nameContainer := container.NewVBox(nameLabel, bubble)

	// Position links/rechts
	if isCurrentUser {
		return container.NewHBox(
			layout.NewSpacer(),
			nameContainer,
			avatar,
		)
	} else {
		return container.NewHBox(
			avatar,
			nameContainer,
			layout.NewSpacer(),
		)
	}
}