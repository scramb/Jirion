package servicedesk

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/scramb/backlog-manager/internal/i18n"
	"github.com/scramb/backlog-manager/internal/models"
)

func CreateView(app fyne.App, w fyne.Window) fyne.CanvasObject {
    // Title
    pageTitle := widget.NewLabelWithStyle(i18n.T("servicedesk.create_request"), fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

    prefs := app.Preferences()
    domain := prefs.String("jira_domain")
    user := prefs.String("jira_user")
    token := prefs.String("jira_token")

    // Load service desks
    var deskOptions []string
    desks, err := models.FetchServiceDesks(domain, user, token)
    if err == nil {
        for _, d := range desks {
            deskOptions = append(deskOptions, d.Name)
        }
    }

    // Dropdown for service desk selection
    deskSelect := widget.NewSelect(deskOptions, nil)
    deskSelect.PlaceHolder = i18n.T("servicedesk.select_desk_placeholder")

    // Priority select
    prioritySelect := widget.NewSelect([]string{"Low", "Medium", "High"}, nil)
    prioritySelect.PlaceHolder = i18n.T("servicedesk.select_priority_placeholder")

    // Request type placeholder
    requestTypeSelect := widget.NewSelect([]string{"Incident", "Service Request", "Access"}, nil)
    requestTypeSelect.PlaceHolder = i18n.T("servicedesk.select_type_placeholder")

    grid := container.New(layout.NewGridLayout(3),
        container.NewVBox(widget.NewLabel(i18n.T("servicedesk.select_desk")), deskSelect),
        container.NewVBox(widget.NewLabel(i18n.T("servicedesk.select_priority")), prioritySelect),
        container.NewVBox(widget.NewLabel(i18n.T("servicedesk.select_type")), requestTypeSelect),
    )

    // Summary
    summaryEntry := widget.NewEntry()
    summaryEntry.SetPlaceHolder(i18n.T("servicedesk.summary_placeholder"))

    // Description
    descriptionEntry := widget.NewMultiLineEntry()
    descriptionEntry.SetPlaceHolder(i18n.T("servicedesk.description_placeholder"))
    descriptionEntry.Wrapping = fyne.TextWrapWord

    // Submit button
    submitBtn := i18n.BindButton("servicedesk.submit_request", theme.ConfirmIcon(), func() {
        if deskSelect.Selected == "" || summaryEntry.Text == "" || descriptionEntry.Text == "" {
            err := fmt.Errorf(i18n.T("servicedesk.validation_error"))
            dialog.ShowError(err, w)
            return
        }

        // Beispielhafter Request-Aufruf mit Fehlerbehandlung
        _, err := models.CreateServiceRequest(
            domain,
            user,
            token,
            deskSelect.Selected,
            "1", // placeholder for requestTypeID
            summaryEntry.Text,
            descriptionEntry.Text,
            map[string]interface{}{"priority": prioritySelect.Selected},
        )
        if err != nil {
            dialog.ShowError(err, w)
            return
        }
        dialog.ShowInformation(
            i18n.T("servicedesk.submit_success_title"),
            i18n.T("servicedesk.submit_success_message"),
            w,
        )
    })

    // Layout
    content := container.NewVBox(
        pageTitle,
        widget.NewSeparator(),
        grid,
        widget.NewLabel(i18n.T("servicedesk.title_label")),
        summaryEntry,
        widget.NewLabel(i18n.T("servicedesk.description_label")),
        descriptionEntry,
    )

    scroll := container.NewVScroll(content)
    scroll.SetMinSize(fyne.NewSize(600, 400))

    // Button pinned at bottom
    pinned := container.NewBorder(nil, submitBtn, nil, nil, scroll)

    return pinned
}
