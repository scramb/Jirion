package components

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/scramb/backlog-manager/internal/models"
)

// CreateChatMessageCard zeigt eine Chat-Nachricht als stylische Bubble an.
func CreateChatMessageCard(comment models.JiraComment, currentUser string) fyne.CanvasObject {
	isCurrentUser := comment.Author.Email == currentUser

	// Nachrichtentext
	text := widget.NewLabel(models.ExtractDescriptionText(comment.Content))

	text.TextStyle = fyne.TextStyle{}

	// Bubble-Hintergrund
	bgColor := theme.Color(theme.ColorNameButton)
	if isCurrentUser {
		bgColor = theme.Color(theme.ColorNamePrimary)
	}
	bg := canvas.NewRectangle(bgColor)
	bg.CornerRadius = 16

	// Bubble mit Hintergrund und gepolstertem Text, füllt horizontal den verfügbaren Platz
	bubble := container.NewStack(bg, container.NewPadded(text))

	// Vertikale Anordnung mit Name
	nameLabel := canvas.NewText(comment.Author.DisplayName, theme.Color(theme.ColorNameForeground))

	if isCurrentUser {
		text.Alignment = fyne.TextAlignTrailing
	} else {
		text.Alignment = fyne.TextAlignLeading
	}
	nameLabel.TextSize = theme.TextSize() - 2
	nameLabel.Alignment = fyne.TextAlignLeading
	nameContainer := container.NewVBox(nameLabel, bubble)

	// Position links/rechts mit voller Breite
	if isCurrentUser {
		return container.NewHBox(
			layout.NewSpacer(),
			nameContainer,
		)
	} else {
		return container.NewHBox(
			nameContainer,
			layout.NewSpacer(),
		)
	}
}
