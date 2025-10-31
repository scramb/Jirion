package ui

import (
	"fmt"
	"runtime"
	"sort"
	"strings"

	"os/exec"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/scramb/backlog-manager/internal/i18n"
	"github.com/scramb/backlog-manager/internal/models"
)

// NewTicketsView builds the “My Tickets” tab content.
// It lists all issues assigned to the logged-in user and allows opening them in the browser.
func TicketsView(app fyne.App, w fyne.Window, domain, user, token string, reloadChan <-chan bool) fyne.CanvasObject {
	var issues []models.JiraIssue
	filteredIssues := []models.JiraIssue{}
	selectedProject := i18n.T("tickets.all_projects")

	ticketList := widget.NewList(
		func() int { return len(filteredIssues) },
		func() fyne.CanvasObject {
			icon := widget.NewIcon(nil)
			id := widget.NewLabel("ID")
			title := widget.NewLabel("Titel")
			openBtn := widget.NewButtonWithIcon("", theme.ViewFullScreenIcon(), nil)
			box := container.NewHBox(icon, id, title, layout.NewSpacer(), openBtn)
			return box
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			issue := filteredIssues[i]
			box := o.(*fyne.Container)

			iconWidget := box.Objects[0].(*widget.Icon)
			idLabel := box.Objects[1].(*widget.Label)
			titleLabel := box.Objects[2].(*widget.Label)
			openBtn := box.Objects[4].(*widget.Button)

			var icon fyne.Resource
			switch issue.Fields.IssueType.Name {
			case "Story":
				icon = theme.FileIcon()
			case "Bug":
				icon = theme.ErrorIcon()
			case "Task":
				icon = theme.DocumentIcon()
			default:
				icon = theme.InfoIcon()
			}

			iconWidget.SetResource(icon)
			idLabel.SetText(issue.Key)
			titleLabel.SetText(issue.Fields.Summary)

			openBtn.OnTapped = func() {
				url := fmt.Sprintf("https://%s.atlassian.net/browse/%s", domain, issue.Key)
				openBrowser(url)
			}
		},
	)

	searchQuery := ""

	applyFilter := func(project string) {
		selectedProject = project
		filteredIssues = nil
		for _, iss := range issues {
			if project != i18n.T("tickets.all_projects") && !strings.HasPrefix(iss.Key, project+"-") {
				continue
			}
			if searchQuery != "" {
				lowerQuery := strings.ToLower(searchQuery)
				if !strings.Contains(strings.ToLower(iss.Key), lowerQuery) && !strings.Contains(strings.ToLower(iss.Fields.Summary), lowerQuery) {
					continue
				}
			}
			filteredIssues = append(filteredIssues, iss)
		}
		ticketList.Refresh()
	}

	projectFilter := widget.NewSelect([]string{i18n.T("tickets.all_projects")}, func(selected string) {
		applyFilter(selected)
	})
	projectFilter.SetSelected(i18n.T("tickets.all_projects"))

	// Move these widget creations above showListView()
	reloadBtn := i18n.BindButton("tickets.reload", theme.ViewRefreshIcon(), nil)
	projectFilterLabel := i18n.BindLabel("tickets.project_filter")
	searchEntryWidget := i18n.BindEntryWithPlaceholder("tickets.search_placeholder", false)

	searchEntryWidget.OnChanged = func(text string) {
		searchQuery = text
		applyFilter(selectedProject)
	}

	reloadBtn.OnTapped = func() {
		go func() {
			fetched, err := models.FetchAssignedIssues(domain, user, token)
			if err != nil {
				fyne.Do(func() { dialog.ShowError(err, w) })
				return
			}
			fyne.Do(func() {
				issues = fetched
				projectSet := map[string]struct{}{}
				for _, iss := range issues {
					parts := strings.SplitN(iss.Key, "-", 2)
					if len(parts) > 1 {
						projectSet[parts[0]] = struct{}{}
					}
				}
				projects := []string{i18n.T("tickets.all_projects")}
				for p := range projectSet {
					projects = append(projects, p)
				}
				sort.Strings(projects)
				projectFilter.Options = projects
				projectFilter.Refresh()
				applyFilter(selectedProject)
			})
		}()
	}

	var contentContainer *fyne.Container

	showListView := func() {
		contentContainer.Objects = []fyne.CanvasObject{
			container.NewBorder(
				container.NewVBox(
					reloadBtn,
					projectFilterLabel,
					projectFilter,
					searchEntryWidget,
				),
				nil, nil, nil,
				ticketList,
			),
		}
		contentContainer.Refresh()
	}

	ticketList.OnSelected = func(id widget.ListItemID) {
		if id < 0 || id >= len(filteredIssues) {
			return
		}
		issue := filteredIssues[id]
		contentContainer.Objects = []fyne.CanvasObject{
			TicketDetailView(app, w, issue, domain, user, token, showListView),
		}
		contentContainer.Refresh()
	}

	contentContainer = container.NewMax()
	showListView()

	// Initial auto-refresh when tab is opened (optional)
	content := contentContainer.Objects[0].(*fyne.Container)
	if btn, ok := content.Objects[0].(*widget.Button); ok {
		btn.OnTapped()
	}

	// Add language change handler to update the "Alle Projekte" text dynamically
	i18n.RegisterOnLanguageChange(func() {
		fyne.Do(func() {
			projectFilterLabel.SetText(i18n.T("tickets.project_filter"))
			reloadBtn.SetText(i18n.T("tickets.reload"))
			searchEntryWidget.SetPlaceHolder(i18n.T("tickets.search_placeholder"))

			prevSelection := projectFilter.Selected
			translatedAll := i18n.T("tickets.all_projects")

			newOptions := []string{translatedAll}
			for _, opt := range projectFilter.Options {
				if opt != "All Projects" && opt != "Alle Projekte" {
					newOptions = append(newOptions, opt)
				}
			}

			newSelect := widget.NewSelect(newOptions, func(selected string) {
				applyFilter(selected)
			})

			if prevSelection == "All Projects" || prevSelection == "Alle Projekte" {
				newSelect.SetSelected(translatedAll)
			} else {
				newSelect.SetSelected(prevSelection)
			}

			// Versuche, den Container sicher zu ersetzen
			if len(contentContainer.Objects) == 0 {
				return
			}

			if content, ok := contentContainer.Objects[0].(*fyne.Container); ok {
				if len(content.Objects) > 0 {
					if border, ok := content.Objects[0].(*fyne.Container); ok {
						if len(border.Objects) > 0 {
							if header, ok := border.Objects[0].(*fyne.Container); ok {
								for i, obj := range header.Objects {
									if obj == projectFilter {
										header.Objects[i] = newSelect
										projectFilter = newSelect
										contentContainer.Refresh()
										return
									}
								}
							}
						}
					}
				}
			}

			// 🧹 Falls der View gerade nicht der erwartete Container ist
			// → einfach komplette Ansicht neu aufbauen
			showListView()
		})
	})

	go func() {
		for range reloadChan {
			fyne.Do(func() {
				fetched, err := models.FetchAssignedIssues(domain, user, token)
				if err != nil {
					dialog.ShowError(err, w)
					return
				}
				issues = fetched
				// build project list
				projectSet := map[string]struct{}{}
				for _, iss := range issues {
					parts := strings.SplitN(iss.Key, "-", 2)
					if len(parts) > 1 {
						projectSet[parts[0]] = struct{}{}
					}
				}
				projects := []string{i18n.T("tickets.all_projects")}
				for p := range projectSet {
					projects = append(projects, p)
				}
				sort.Strings(projects)
				projectFilter.Options = projects
				projectFilter.Refresh()

				applyFilter(selectedProject)
			})
		}
	}()

	return contentContainer
}

func createChip(text string) fyne.CanvasObject {
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

// openBrowser opens a URL in the system's default browser.
func openBrowser(url string) {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "darwin":
		cmd = "open"
	case "windows":
		cmd = "rundll32"
		args = []string{"url.dll,FileProtocolHandler"}
	default:
		cmd = "xdg-open"
	}

	args = append(args, url)
	exec.Command(cmd, args...).Start()
}

// TicketDetailView shows detailed information about a Jira issue with a back button.
func TicketDetailView(app fyne.App, w fyne.Window, issue models.JiraIssue, domain, user, token string, back func()) fyne.CanvasObject {
	keyLabel := widget.NewLabelWithStyle(issue.Key, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	labels, transitions, comments := loadTicketContent(issue, domain, user, token)
	summaryHeader := i18n.BindLabel("tickets.summary")
	summaryLabel := widget.NewLabel(issue.Fields.Summary)

	descriptionHeader := i18n.BindLabel("tickets.description")
	description := models.ExtractDescriptionText(issue.Fields.Description)
	descriptionLabel := widget.NewLabel(description)
	descriptionLabel.Wrapping = fyne.TextWrapWord

	// Labels-Bereich vorbereiten
	labelTitle := widget.NewLabel("Labels:")
	labelsFlow := container.New(layout.NewGridWrapLayout(fyne.NewSize(180, 30)))

	go func() {
		fyne.Do(func() {
			for _, lbl := range labels.Fields.Labels {
				labelsFlow.Add(createChip(lbl))
			}
			labelsFlow.Refresh()
		})
	}()

	backBtn := i18n.BindButton("tickets.back", theme.NavigateBackIcon(), func() {
		back()
	})
	transitionOptions := []string{}
	transitionMap := map[string]string{}
	for _, t := range transitions {
		transitionOptions = append(transitionOptions, t.Name)
		transitionMap[t.Name] = t.ID
	}

	var selectedTransitionID string

	transitionSelect := widget.NewSelect(transitionOptions, func(selected string) {
		selectedTransitionID = transitionMap[selected]
		fmt.Println(selectedTransitionID)
	})

	transitionSelect.Disable()

	transitionContainer := container.NewVBox(i18n.BindLabel("tickets.transition_label"), transitionSelect)

	transitionSelect.OnChanged = func(selected string) {
		fmt.Printf("Selected: %s", transitionMap[selected])
	}

	commentsContainer := container.NewVBox()

	for _, c := range comments {
		commentsContainer.Add(CreateChatMessageCard(c, user))
	}

	detailsSection := NewCollapsibleSection("Comments", commentsContainer)

	content := container.NewVBox(
		backBtn,
		keyLabel,
		widget.NewSeparator(),
		transitionContainer,
		summaryHeader,
		summaryLabel,
		widget.NewSeparator(),
		descriptionHeader,
		descriptionLabel,
		labelTitle,
		labelsFlow, // statt labelsContainer
		detailsSection,
	)

	scroll := container.NewVScroll(content)
	scroll.SetMinSize(fyne.NewSize(400, 300))

	return scroll
}

func loadTicketContent(issue models.JiraIssue, domain, user, token string) (models.JiraIssueLabels, []models.JiraTransition, []models.JiraComment) {

	labels, errLabels := models.FetchIssueLabels(domain, user, token, issue.Id)
	comments, errComments := models.FetchIssueComments(domain, user, token, issue.Id)
	transitions, errTransitions := models.FetchIssueTransitions(domain, user, token, issue.Id)

	if errComments != nil || errLabels != nil || errTransitions != nil {
		fmt.Print(errLabels, errComments, errTransitions)
	}
	for _, c := range comments {
		if user == c.Author.Email {
			fmt.Println("Das ist meine Kommentar")
		} else {
			fmt.Println("Nicht mein Kommentar")
		}
	}

	return labels, transitions, comments
}

func NewCollapsibleSection(title string, content fyne.CanvasObject) fyne.CanvasObject {
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

// CreateChatMessageCard zeigt eine Chat-Nachricht als Card an, mit Avatar und Ausrichtung je nach Autor.
func CreateChatMessageCard(comment models.JiraComment, currentUser string) fyne.CanvasObject {
	// Avatar laden oder Platzhalter
	var avatar fyne.CanvasObject
	if comment.Author.AvatarUrls.Image != "" {
		img := canvas.NewImageFromURI(storage.NewURI(comment.Author.AvatarUrls.Image))
		img.FillMode = canvas.ImageFillContain
		img.SetMinSize(fyne.NewSize(32, 32))
		img.Resize(fyne.NewSize(32, 32))
		avatar = container.NewStack(img)
	} else {
		placeholder := canvas.NewCircle(theme.ForegroundColor())
		placeholder.Resize(fyne.NewSize(32, 32))
		avatar = container.NewStack(placeholder)
	}

	// Kommentartext
	msg := canvas.NewText(models.ExtractDescriptionText(comment.Content), theme.ForegroundColor())
	msg.TextSize = theme.TextSize()
	msg.Alignment = fyne.TextAlignLeading

	// Nachricht in Card
	card := widget.NewCard("", comment.Author.DisplayName, msg)
	card.Resize(fyne.NewSize(300, card.MinSize().Height))

	// Layout: links oder rechts
	if comment.Author.Email == currentUser {
		// Eigene Nachricht: links
		return container.NewHBox(
			avatar,
			layout.NewSpacer(),
			card,
		)
	} else {
		// Fremde Nachricht: rechts
		return container.NewHBox(
			card,
			layout.NewSpacer(),
			avatar,
		)
	}
}
